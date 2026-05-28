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
	"github.com/family/crab-server/internal/handlers"
	"github.com/family/crab-server/internal/middleware"
	memrepo "github.com/family/crab-server/internal/repository/memory"
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
		// Generate random secret for development mode
		jwtSecret = generateRandomSecret()
		log.Println("WARNING: Using auto-generated JWT secret in development mode. Set JWT_SECRET for production.")
	}

	// Initialize repositories (using in-memory for now)
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()

	// Initialize services
	familySvc := services.NewFamilyService(familyRepo, childRepo, jwtSecret)

	// Initialize handlers
	familyHandler := handlers.NewFamilyHandler(familySvc)

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
