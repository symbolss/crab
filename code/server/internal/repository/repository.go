package repository

import (
	"context"

	"github.com/family/crab-server/internal/models"
	"github.com/google/uuid"
)

// FamilyRepository defines the interface for family data operations
type FamilyRepository interface {
	// CreateFamily creates a new family and returns the created family
	CreateFamily(ctx context.Context, family *models.Family) error

	// GetFamilyByID retrieves a family by ID
	GetFamilyByID(ctx context.Context, id uuid.UUID) (*models.Family, error)

	// GetFamilyByPairingCode retrieves a family by pairing code
	GetFamilyByPairingCode(ctx context.Context, pairingCode string) (*models.Family, error)

	// UpdatePairingCode updates the pairing code for a family
	UpdatePairingCode(ctx context.Context, id uuid.UUID, pairingCode string) error
}

// ChildRepository defines the interface for child data operations
type ChildRepository interface {
	// CreateChild creates a new child and returns the created child
	CreateChild(ctx context.Context, child *models.Child) error

	// GetChildByID retrieves a child by ID
	GetChildByID(ctx context.Context, id uuid.UUID) (*models.Child, error)

	// GetChildrenByFamilyID retrieves all children for a family
	GetChildrenByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Child, error)

	// UpdateChildStatus updates the status of a child
	UpdateChildStatus(ctx context.Context, id uuid.UUID, status string) error
}
