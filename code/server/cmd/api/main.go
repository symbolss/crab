package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/family/crab-server/internal/config"
	"github.com/family/crab-server/internal/database"
	"github.com/family/crab-server/internal/handlers"
	"github.com/family/crab-server/internal/middleware"
	memrepo "github.com/family/crab-server/internal/repository/memory"
	postgresrepo "github.com/family/crab-server/internal/repository/postgres"
	"github.com/family/crab-server/internal/services"
	"github.com/family/crab-server/pkg/response"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if cfg.IsProduction() {
			log.Fatal("JWT_SECRET must be set in production")
		}
		jwtSecret = generateRandomSecret()
		log.Println("WARNING: Using auto-generated JWT secret in development mode. Set JWT_SECRET for production.")
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	var familySvc *services.FamilyService
	var contentSvc *services.ContentService

	if dbURL != "" {
		// Use PostgreSQL
		db, err := database.Connect(database.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     5432,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Database: getEnv("DB_NAME", "crab"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		log.Println("Connected to PostgreSQL database")

		familyRepo := postgresrepo.NewFamilyRepository(db)
		childRepo := postgresrepo.NewChildRepository(db)
		itemRepo := postgresrepo.NewItemRepository(db)

		familySvc = services.NewFamilyService(familyRepo, childRepo, jwtSecret)
		contentSvc = services.NewContentService(familyRepo, itemRepo)
	} else {
		// Use in-memory storage (for development/testing)
		log.Println("WARNING: Using in-memory storage. Set DATABASE_URL for production.")

		familyRepo := memrepo.NewInMemoryFamilyRepository()
		childRepo := memrepo.NewInMemoryChildRepository()
		itemRepo := memrepo.NewInMemoryItemRepository()

		familySvc = services.NewFamilyService(familyRepo, childRepo, jwtSecret)
		contentSvc = services.NewContentService(familyRepo, itemRepo)
	}

	// Initialize handlers
	familyHandler := handlers.NewFamilyHandler(familySvc)
	contentHandler := handlers.NewContentHandler(contentSvc)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(familySvc)

	// Initialize rate limiter: 10 requests per minute per IP
	rateLimiter := middleware.NewRateLimiter(10, time.Minute)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)

	// Routes
	r.Get("/healthz", healthHandler)
	
	// Family routes with rate limiting
	r.Group(func(r chi.Router) {
		r.Use(rateLimiter.RateLimit)
		r.Post("/api/family/create", familyHandler.CreateFamily)
		r.Post("/api/family/pair", familyHandler.PairChild)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)
		r.Get("/api/children", familyHandler.GetChildren)
		r.Post("/api/items/import", contentHandler.ImportItem)
	})

	// Start server
	addr := cfg.ServerHost + ":" + cfg.ServerPort
	log.Printf("Starting server on %s (environment: %s)", addr, cfg.Environment)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// generateRandomSecret generates a 32-byte random hex string for development use
func generateRandomSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate random secret: %v", err)
	}
	return hex.EncodeToString(bytes)
}

// healthHandler returns server health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]string{
		"status": "ok",
	})
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
