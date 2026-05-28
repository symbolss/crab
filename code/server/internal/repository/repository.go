package repository

import (
	"context"
	"time"

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

	// GetByID retrieves a child by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Child, error)

	// GetChildByID retrieves a child by ID (deprecated, use GetByID)
	GetChildByID(ctx context.Context, id uuid.UUID) (*models.Child, error)

	// GetChildrenByFamilyID retrieves all children for a family
	GetChildrenByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Child, error)

	// UpdateChildStatus updates the status of a child
	UpdateChildStatus(ctx context.Context, id uuid.UUID, status string) error
}

// ItemRepository defines the interface for item data operations
type ItemRepository interface {
	// CreateItem creates a new item
	CreateItem(ctx context.Context, item *models.Item) error

	// GetItemByID retrieves an item by ID
	GetItemByID(ctx context.Context, id uuid.UUID) (*models.Item, error)

	// GetItemByNormalizedURL retrieves an item by normalized URL within a family
	GetItemByNormalizedURL(ctx context.Context, familyID uuid.UUID, normalizedURL string) (*models.Item, error)

	// GetItemsByFamilyID retrieves all items for a family
	GetItemsByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Item, error)

	// UpdateItem updates an item
	UpdateItem(ctx context.Context, item *models.Item) error
}

// PlayEventRepository defines the interface for play event data operations
type PlayEventRepository interface {
	// Create records a new play event
	Create(ctx context.Context, childID uuid.UUID, event *models.PlayEvent) error

	// GetByChildID retrieves all events for a child
	GetByChildID(ctx context.Context, childID uuid.UUID) ([]*models.PlayEvent, error)

	// GetByChildAndItem retrieves all events for a child and item
	GetByChildAndItem(ctx context.Context, childID, itemID uuid.UUID) ([]*models.PlayEvent, error)
}

// DailyUsageRepository defines the interface for daily usage data operations
type DailyUsageRepository interface {
	// GetByChildAndDate retrieves daily usage for a child on a specific date
	GetByChildAndDate(ctx context.Context, childID uuid.UUID, date time.Time) (*models.DailyUsage, error)

	// Upsert creates or updates daily usage
	Upsert(ctx context.Context, usage *models.DailyUsage) error
}
