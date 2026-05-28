package memory

import (
	"context"
	"sync"

	"github.com/family/crab-server/internal/models"
	"github.com/google/uuid"
)

// InMemoryFamilyRepository is an in-memory implementation of FamilyRepository for testing
type InMemoryFamilyRepository struct {
	mu       sync.RWMutex
	families map[uuid.UUID]*models.Family
	byCode   map[string]*models.Family
}

// NewInMemoryFamilyRepository creates a new in-memory family repository
func NewInMemoryFamilyRepository() *InMemoryFamilyRepository {
	return &InMemoryFamilyRepository{
		families: make(map[uuid.UUID]*models.Family),
		byCode:   make(map[string]*models.Family),
	}
}

// CreateFamily creates a new family
func (r *InMemoryFamilyRepository) CreateFamily(ctx context.Context, family *models.Family) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store a copy to avoid external mutations
	f := &models.Family{
		ID:                   family.ID,
		PairingCode:          family.PairingCode,
		PairingCodeExpiresAt: family.PairingCodeExpiresAt,
		CreatedAt:            family.CreatedAt,
	}
	r.families[f.ID] = f
	r.byCode[f.PairingCode] = f
	return nil
}

// GetFamilyByID retrieves a family by ID
func (r *InMemoryFamilyRepository) GetFamilyByID(ctx context.Context, id uuid.UUID) (*models.Family, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	family, ok := r.families[id]
	if !ok {
		return nil, models.ErrFamilyNotFound
	}

	// Return a copy to avoid external mutations
	return &models.Family{
		ID:                   family.ID,
		PairingCode:          family.PairingCode,
		PairingCodeExpiresAt: family.PairingCodeExpiresAt,
		CreatedAt:            family.CreatedAt,
	}, nil
}

// GetFamilyByPairingCode retrieves a family by pairing code
func (r *InMemoryFamilyRepository) GetFamilyByPairingCode(ctx context.Context, pairingCode string) (*models.Family, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	family, ok := r.byCode[pairingCode]
	if !ok {
		return nil, models.ErrPairingCodeInvalid
	}

	// Return a copy to avoid external mutations
	return &models.Family{
		ID:                   family.ID,
		PairingCode:          family.PairingCode,
		PairingCodeExpiresAt: family.PairingCodeExpiresAt,
		CreatedAt:            family.CreatedAt,
	}, nil
}

// UpdatePairingCode updates the pairing code for a family
func (r *InMemoryFamilyRepository) UpdatePairingCode(ctx context.Context, id uuid.UUID, pairingCode string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	family, ok := r.families[id]
	if !ok {
		return models.ErrFamilyNotFound
	}

	// Remove old code mapping
	delete(r.byCode, family.PairingCode)

	// Update code
	family.PairingCode = pairingCode
	r.byCode[pairingCode] = family

	return nil
}

// InMemoryChildRepository is an in-memory implementation of ChildRepository for testing
type InMemoryChildRepository struct {
	mu       sync.RWMutex
	children map[uuid.UUID]*models.Child
	byFamily map[uuid.UUID][]*models.Child
}

// NewInMemoryChildRepository creates a new in-memory child repository
func NewInMemoryChildRepository() *InMemoryChildRepository {
	return &InMemoryChildRepository{
		children: make(map[uuid.UUID]*models.Child),
		byFamily: make(map[uuid.UUID][]*models.Child),
	}
}

// CreateChild creates a new child
func (r *InMemoryChildRepository) CreateChild(ctx context.Context, child *models.Child) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store a copy to avoid external mutations
	c := &models.Child{
		ID:        child.ID,
		FamilyID:  child.FamilyID,
		Name:      child.Name,
		Status:    child.Status,
		CreatedAt: child.CreatedAt,
	}
	r.children[c.ID] = c
	r.byFamily[c.FamilyID] = append(r.byFamily[c.FamilyID], c)
	return nil
}

// GetChildByID retrieves a child by ID
func (r *InMemoryChildRepository) GetChildByID(ctx context.Context, id uuid.UUID) (*models.Child, error) {
	return r.GetByID(ctx, id)
}

// GetByID retrieves a child by ID
func (r *InMemoryChildRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Child, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	child, ok := r.children[id]
	if !ok {
		return nil, models.ErrChildNotFound
	}

	// Return a copy to avoid external mutations
	return &models.Child{
		ID:        child.ID,
		FamilyID:  child.FamilyID,
		Name:      child.Name,
		Status:    child.Status,
		CreatedAt: child.CreatedAt,
	}, nil
}

// GetChildrenByFamilyID retrieves all children for a family
func (r *InMemoryChildRepository) GetChildrenByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Child, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	children := r.byFamily[familyID]
	result := make([]*models.Child, len(children))
	for i, c := range children {
		result[i] = &models.Child{
			ID:        c.ID,
			FamilyID:  c.FamilyID,
			Name:      c.Name,
			Status:    c.Status,
			CreatedAt: c.CreatedAt,
		}
	}
	return result, nil
}

// UpdateChildStatus updates the status of a child
func (r *InMemoryChildRepository) UpdateChildStatus(ctx context.Context, id uuid.UUID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	child, ok := r.children[id]
	if !ok {
		return models.ErrChildNotFound
	}

	child.Status = status
	return nil
}

// InMemoryItemRepository is an in-memory implementation of ItemRepository for testing
type InMemoryItemRepository struct {
	mu              sync.RWMutex
	items           map[uuid.UUID]*models.Item
	byFamily        map[uuid.UUID][]*models.Item
	byNormalizedURL map[string]*models.Item // key: familyID:normalizedURL
}

// NewInMemoryItemRepository creates a new in-memory item repository
func NewInMemoryItemRepository() *InMemoryItemRepository {
	return &InMemoryItemRepository{
		items:           make(map[uuid.UUID]*models.Item),
		byFamily:        make(map[uuid.UUID][]*models.Item),
		byNormalizedURL: make(map[string]*models.Item),
	}
}

// CreateItem creates a new item
func (r *InMemoryItemRepository) CreateItem(ctx context.Context, item *models.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store a copy
	i := &models.Item{
		ID:               item.ID,
		FamilyID:         item.FamilyID,
		SourceType:       item.SourceType,
		SourceURL:        item.SourceURL,
		NormalizedURL:    item.NormalizedURL,
		Title:            item.Title,
		CoverURL:         item.CoverURL,
		Summary:          item.Summary,
		PlaybackMode:     item.PlaybackMode,
		ProcessingStatus: item.ProcessingStatus,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
	r.items[i.ID] = i
	r.byFamily[i.FamilyID] = append(r.byFamily[i.FamilyID], i)
	r.byNormalizedURL[i.FamilyID.String()+":"+i.NormalizedURL] = i
	return nil
}

// GetItemByID retrieves an item by ID
func (r *InMemoryItemRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	if !ok {
		return nil, nil
	}

	return &models.Item{
		ID:               item.ID,
		FamilyID:         item.FamilyID,
		SourceType:       item.SourceType,
		SourceURL:        item.SourceURL,
		NormalizedURL:    item.NormalizedURL,
		Title:            item.Title,
		CoverURL:         item.CoverURL,
		Summary:          item.Summary,
		PlaybackMode:     item.PlaybackMode,
		ProcessingStatus: item.ProcessingStatus,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}, nil
}

// GetItemByNormalizedURL retrieves an item by normalized URL within a family
func (r *InMemoryItemRepository) GetItemByNormalizedURL(ctx context.Context, familyID uuid.UUID, normalizedURL string) (*models.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.byNormalizedURL[familyID.String()+":"+normalizedURL]
	if !ok {
		return nil, nil
	}

	return &models.Item{
		ID:               item.ID,
		FamilyID:         item.FamilyID,
		SourceType:       item.SourceType,
		SourceURL:        item.SourceURL,
		NormalizedURL:    item.NormalizedURL,
		Title:            item.Title,
		CoverURL:         item.CoverURL,
		Summary:          item.Summary,
		PlaybackMode:     item.PlaybackMode,
		ProcessingStatus: item.ProcessingStatus,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}, nil
}

// GetItemsByFamilyID retrieves all items for a family
func (r *InMemoryItemRepository) GetItemsByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := r.byFamily[familyID]
	result := make([]*models.Item, len(items))
	for i, item := range items {
		result[i] = &models.Item{
			ID:               item.ID,
			FamilyID:         item.FamilyID,
			SourceType:       item.SourceType,
			SourceURL:        item.SourceURL,
			NormalizedURL:    item.NormalizedURL,
			Title:            item.Title,
			CoverURL:         item.CoverURL,
			Summary:          item.Summary,
			PlaybackMode:     item.PlaybackMode,
			ProcessingStatus: item.ProcessingStatus,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
		}
	}
	return result, nil
}

// UpdateItem updates an item
func (r *InMemoryItemRepository) UpdateItem(ctx context.Context, item *models.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.items[item.ID]
	if !ok {
		return models.ErrItemNotFound
	}

	existing.Title = item.Title
	existing.CoverURL = item.CoverURL
	existing.Summary = item.Summary
	existing.ProcessingStatus = item.ProcessingStatus
	existing.UpdatedAt = item.UpdatedAt
	return nil
}
