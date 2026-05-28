package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/family/crab-server/internal/models"
	"github.com/family/crab-server/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// DefaultPairingCodeExpiry is the default duration until a pairing code expires
const DefaultPairingCodeExpiry = 24 * time.Hour

// FamilyService handles family-related business logic
type FamilyService struct {
	familyRepo         repository.FamilyRepository
	childRepo          repository.ChildRepository
	jwtSecret          []byte
	pairingCodeExpiry  time.Duration
}

// NewFamilyService creates a new FamilyService
func NewFamilyService(familyRepo repository.FamilyRepository, childRepo repository.ChildRepository, jwtSecret string) *FamilyService {
	return &FamilyService{
		familyRepo:        familyRepo,
		childRepo:         childRepo,
		jwtSecret:         []byte(jwtSecret),
		pairingCodeExpiry: DefaultPairingCodeExpiry,
	}
}

// CreateFamily creates a new family and returns the family ID, pairing code, and parent token
func (s *FamilyService) CreateFamily(ctx context.Context, req *models.CreateFamilyRequest) (*models.CreateFamilyResponse, error) {
	// Generate random 8-character pairing code
	pairingCode, err := generatePairingCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate pairing code: %w", err)
	}

	// Create family with expiration time
	family := &models.Family{
		ID:                   uuid.New(),
		PairingCode:          pairingCode,
		PairingCodeExpiresAt: time.Now().Add(s.pairingCodeExpiry),
		CreatedAt:            time.Now(),
	}

	if err := s.familyRepo.CreateFamily(ctx, family); err != nil {
		return nil, fmt.Errorf("failed to create family: %w", err)
	}

	// Generate parent JWT token
	token, err := s.generateToken(family.ID, "parent")
	if err != nil {
		return nil, fmt.Errorf("failed to generate parent token: %w", err)
	}

	return &models.CreateFamilyResponse{
		FamilyID:    family.ID,
		PairingCode: pairingCode,
		ParentToken: token,
	}, nil
}

// PairChild pairs a child device to a family using the pairing code
func (s *FamilyService) PairChild(ctx context.Context, req *models.PairChildRequest) (*models.PairChildResponse, error) {
	// Find family by pairing code
	family, err := s.familyRepo.GetFamilyByPairingCode(ctx, req.PairingCode)
	if err != nil {
		return nil, fmt.Errorf("failed to find family by pairing code: %w", err)
	}

	// Check if pairing code has expired
	if family.IsPairingCodeExpired() {
		return nil, models.ErrPairingCodeExpired
	}

	// Create child
	child := &models.Child{
		ID:        uuid.New(),
		FamilyID:  family.ID,
		Name:      req.ChildName,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	if err := s.childRepo.CreateChild(ctx, child); err != nil {
		return nil, fmt.Errorf("failed to create child: %w", err)
	}

	// Generate child JWT token
	token, err := s.generateToken(child.ID, "child")
	if err != nil {
		return nil, fmt.Errorf("failed to generate child token: %w", err)
	}

	return &models.PairChildResponse{
		ChildID:    child.ID,
		FamilyID:   family.ID,
		ChildName:  child.Name,
		ChildToken: token,
	}, nil
}

// GetChildrenByFamilyID retrieves all children for a family
func (s *FamilyService) GetChildrenByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.ChildListResponse, error) {
	children, err := s.childRepo.GetChildrenByFamilyID(ctx, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children for family %s: %w", familyID, err)
	}

	result := make([]*models.ChildListResponse, len(children))
	for i, c := range children {
		result[i] = &models.ChildListResponse{
			ID:        c.ID,
			Name:      c.Name,
			Status:    c.Status,
			CreatedAt: c.CreatedAt,
		}
	}

	return result, nil
}

// GetChildByID retrieves a child by ID
func (s *FamilyService) GetChildByID(ctx context.Context, childID uuid.UUID) (*models.Child, error) {
	return s.childRepo.GetChildByID(ctx, childID)
}

// generatePairingCode generates a random 8-character pairing code
func generatePairingCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// generateToken generates a JWT token for a user
func (s *FamilyService) generateToken(id uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  id.String(),
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken validates a JWT token and returns the claims
func (s *FamilyService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
