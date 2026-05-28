package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/family/crab-server/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// FamilyRepository is a PostgreSQL implementation of repository.FamilyRepository
type FamilyRepository struct {
	db *sqlx.DB
}

// NewFamilyRepository creates a new PostgreSQL family repository
func NewFamilyRepository(db *sqlx.DB) *FamilyRepository {
	return &FamilyRepository{db: db}
}

// CreateFamily creates a new family
func (r *FamilyRepository) CreateFamily(ctx context.Context, family *models.Family) error {
	query := `
		INSERT INTO families (id, pairing_code, pairing_code_expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		family.ID,
		family.PairingCode,
		family.PairingCodeExpiresAt,
		family.CreatedAt,
	)
	return err
}

// GetFamilyByID retrieves a family by ID
func (r *FamilyRepository) GetFamilyByID(ctx context.Context, id uuid.UUID) (*models.Family, error) {
	query := `
		SELECT id, pairing_code, pairing_code_expires_at, created_at
		FROM families WHERE id = $1
	`
	var family models.Family
	err := r.db.GetContext(ctx, &family, query, id)
	if err == sql.ErrNoRows {
		return nil, models.ErrFamilyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &family, nil
}

// GetFamilyByPairingCode retrieves a family by pairing code
func (r *FamilyRepository) GetFamilyByPairingCode(ctx context.Context, pairingCode string) (*models.Family, error) {
	query := `
		SELECT id, pairing_code, pairing_code_expires_at, created_at
		FROM families WHERE pairing_code = $1
	`
	var family models.Family
	err := r.db.GetContext(ctx, &family, query, pairingCode)
	if err == sql.ErrNoRows {
		return nil, models.ErrPairingCodeInvalid
	}
	if err != nil {
		return nil, err
	}
	return &family, nil
}

// UpdatePairingCode updates the pairing code for a family
func (r *FamilyRepository) UpdatePairingCode(ctx context.Context, id uuid.UUID, pairingCode string) error {
	query := `UPDATE families SET pairing_code = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, pairingCode, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErrFamilyNotFound
	}
	return nil
}

// ChildRepository is a PostgreSQL implementation of repository.ChildRepository
type ChildRepository struct {
	db *sqlx.DB
}

// NewChildRepository creates a new PostgreSQL child repository
func NewChildRepository(db *sqlx.DB) *ChildRepository {
	return &ChildRepository{db: db}
}

// CreateChild creates a new child
func (r *ChildRepository) CreateChild(ctx context.Context, child *models.Child) error {
	query := `
		INSERT INTO children (id, family_id, name, avatar_url, device_id, daily_time_limit_sec, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		child.ID,
		child.FamilyID,
		child.Name,
		child.AvatarURL,
		child.DeviceID,
		child.DailyTimeLimitSec,
		child.Status,
		child.CreatedAt,
	)
	return err
}

// GetByID retrieves a child by ID
func (r *ChildRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Child, error) {
	query := `
		SELECT id, family_id, name, avatar_url, device_id, daily_time_limit_sec, status, created_at
		FROM children WHERE id = $1
	`
	var child models.Child
	err := r.db.GetContext(ctx, &child, query, id)
	if err == sql.ErrNoRows {
		return nil, models.ErrChildNotFound
	}
	if err != nil {
		return nil, err
	}
	return &child, nil
}

// GetChildByID retrieves a child by ID (alias for GetByID)
func (r *ChildRepository) GetChildByID(ctx context.Context, id uuid.UUID) (*models.Child, error) {
	return r.GetByID(ctx, id)
}

// GetChildrenByFamilyID retrieves all children for a family
func (r *ChildRepository) GetChildrenByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Child, error) {
	query := `
		SELECT id, family_id, name, avatar_url, device_id, daily_time_limit_sec, status, created_at
		FROM children WHERE family_id = $1 ORDER BY created_at
	`
	var children []*models.Child
	err := r.db.SelectContext(ctx, &children, query, familyID)
	if err != nil {
		return nil, err
	}
	return children, nil
}

// UpdateChildStatus updates the status of a child
func (r *ChildRepository) UpdateChildStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE children SET status = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErrChildNotFound
	}
	return nil
}

// ItemRepository is a PostgreSQL implementation of repository.ItemRepository
type ItemRepository struct {
	db *sqlx.DB
}

// NewItemRepository creates a new PostgreSQL item repository
func NewItemRepository(db *sqlx.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// CreateItem creates a new item
func (r *ItemRepository) CreateItem(ctx context.Context, item *models.Item) error {
	query := `
		INSERT INTO items (id, family_id, title, cover_url, summary, source_type, source_url,
			normalized_url, playback_mode, processing_status, playback_status, age_band, duration_sec, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.FamilyID, item.Title, item.CoverURL, item.Summary,
		item.SourceType, item.SourceURL, item.NormalizedURL, item.PlaybackMode,
		item.ProcessingStatus, item.PlaybackStatus, item.AgeBand, item.DurationSec,
		item.CreatedAt, item.UpdatedAt,
	)
	return err
}

// GetItemByID retrieves an item by ID
func (r *ItemRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	query := `
		SELECT id, family_id, title, cover_url, summary, source_type, source_url,
			normalized_url, playback_mode, processing_status, playback_status, age_band, duration_sec, created_at, updated_at
		FROM items WHERE id = $1
	`
	var item models.Item
	err := r.db.GetContext(ctx, &item, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetItemByNormalizedURL retrieves an item by normalized URL within a family
func (r *ItemRepository) GetItemByNormalizedURL(ctx context.Context, familyID uuid.UUID, normalizedURL string) (*models.Item, error) {
	query := `
		SELECT id, family_id, title, cover_url, summary, source_type, source_url,
			normalized_url, playback_mode, processing_status, playback_status, age_band, duration_sec, created_at, updated_at
		FROM items WHERE family_id = $1 AND normalized_url = $2
	`
	var item models.Item
	err := r.db.GetContext(ctx, &item, query, familyID, normalizedURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetItemsByFamilyID retrieves all items for a family
func (r *ItemRepository) GetItemsByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Item, error) {
	query := `
		SELECT id, family_id, title, cover_url, summary, source_type, source_url,
			normalized_url, playback_mode, processing_status, playback_status, age_band, duration_sec, created_at, updated_at
		FROM items WHERE family_id = $1 ORDER BY created_at DESC
	`
	var items []*models.Item
	err := r.db.SelectContext(ctx, &items, query, familyID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateItem updates an item
func (r *ItemRepository) UpdateItem(ctx context.Context, item *models.Item) error {
	query := `
		UPDATE items SET
			title = $1, cover_url = $2, summary = $3, processing_status = $4,
			playback_status = $5, age_band = $6, duration_sec = $7, updated_at = $8
		WHERE id = $9
	`
	result, err := r.db.ExecContext(ctx, query,
		item.Title, item.CoverURL, item.Summary, item.ProcessingStatus,
		item.PlaybackStatus, item.AgeBand, item.DurationSec, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return models.ErrItemNotFound
	}
	return nil
}

// PlayEventRepository is a PostgreSQL implementation of repository.PlayEventRepository
type PlayEventRepository struct {
	db *sqlx.DB
}

// NewPlayEventRepository creates a new PostgreSQL play event repository
func NewPlayEventRepository(db *sqlx.DB) *PlayEventRepository {
	return &PlayEventRepository{db: db}
}

// Create records a new play event
func (r *PlayEventRepository) Create(ctx context.Context, childID uuid.UUID, event *models.PlayEvent) error {
	query := `
		INSERT INTO play_events (id, child_id, item_id, event_type, position_sec, occurred_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		event.ID, childID, event.ItemID, event.EventType, event.PositionSec, event.OccurredAt, event.CreatedAt,
	)
	return err
}

// GetByChildID retrieves all events for a child
func (r *PlayEventRepository) GetByChildID(ctx context.Context, childID uuid.UUID) ([]*models.PlayEvent, error) {
	query := `
		SELECT id, child_id, item_id, event_type, position_sec, occurred_at, created_at
		FROM play_events WHERE child_id = $1 ORDER BY occurred_at DESC
	`
	var events []*models.PlayEvent
	err := r.db.SelectContext(ctx, &events, query, childID)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GetByChildAndItem retrieves all events for a child and item
func (r *PlayEventRepository) GetByChildAndItem(ctx context.Context, childID, itemID uuid.UUID) ([]*models.PlayEvent, error) {
	query := `
		SELECT id, child_id, item_id, event_type, position_sec, occurred_at, created_at
		FROM play_events WHERE child_id = $1 AND item_id = $2 ORDER BY occurred_at DESC
	`
	var events []*models.PlayEvent
	err := r.db.SelectContext(ctx, &events, query, childID, itemID)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// DailyUsageRepository is a PostgreSQL implementation of repository.DailyUsageRepository
type DailyUsageRepository struct {
	db *sqlx.DB
}

// NewDailyUsageRepository creates a new PostgreSQL daily usage repository
func NewDailyUsageRepository(db *sqlx.DB) *DailyUsageRepository {
	return &DailyUsageRepository{db: db}
}

// GetByChildAndDate retrieves daily usage for a child on a specific date
func (r *DailyUsageRepository) GetByChildAndDate(ctx context.Context, childID uuid.UUID, date time.Time) (*models.DailyUsage, error) {
	query := `SELECT id, child_id, date, total_play_sec, updated_at FROM daily_usage WHERE child_id = $1 AND date = $2`
	var usage models.DailyUsage
	err := r.db.GetContext(ctx, &usage, query, childID, date.Format("2006-01-02"))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

// Upsert creates or updates daily usage
func (r *DailyUsageRepository) Upsert(ctx context.Context, usage *models.DailyUsage) error {
	query := `
		INSERT INTO daily_usage (id, child_id, date, total_play_sec, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (child_id, date) DO UPDATE SET
			total_play_sec = EXCLUDED.total_play_sec, updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, usage.ID, usage.ChildID, usage.Date.Format("2006-01-02"), usage.TotalPlaySec, usage.UpdatedAt)
	return err
}
