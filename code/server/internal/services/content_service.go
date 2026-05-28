package services

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/family/crab-server/internal/models"
	"github.com/family/crab-server/internal/repository"
	"github.com/google/uuid"
)

// URLNormalizer handles URL normalization and source detection
type URLNormalizer struct {
	douyinPattern         *regexp.Regexp
	xPattern              *regexp.Regexp
	wechatChannelsPattern *regexp.Regexp
}

// NewURLNormalizer creates a new URL normalizer
func NewURLNormalizer() *URLNormalizer {
	return &URLNormalizer{
		douyinPattern:         regexp.MustCompile(`(?i)(www\.)?douyin\.com|v\.douyin\.com`),
		xPattern:              regexp.MustCompile(`(?i)(www\.)?(twitter|x)\.com`),
		wechatChannelsPattern: regexp.MustCompile(`(?i)channels\.weixin\.qq\.com|finder\.video\.qq\.com`),
	}
}

// NormalizeURL normalizes a URL by removing tracking parameters
func (n *URLNormalizer) NormalizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Remove tracking parameters
	trackingParams := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "fbclid", "gclid"}
	q := parsed.Query()
	for _, param := range trackingParams {
		q.Del(param)
	}
	parsed.RawQuery = q.Encode()

	// Default scheme
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}

	// Remove trailing slash
	parsed.Path = strings.TrimRight(parsed.Path, "/")

	return parsed.String(), nil
}

// DetectSourceType detects the source platform from a URL
func (n *URLNormalizer) DetectSourceType(rawURL string) models.SourceType {
	lower := strings.ToLower(rawURL)

	if n.douyinPattern.MatchString(lower) {
		return models.SourceTypeDouyin
	}
	if n.xPattern.MatchString(lower) {
		return models.SourceTypeX
	}
	if n.wechatChannelsPattern.MatchString(lower) {
		return models.SourceTypeWechatChannels
	}
	return models.SourceTypeWebArticle
}

// ItemRepository defines the interface for item data operations
type ItemRepository interface {
	CreateItem(ctx context.Context, item *models.Item) error
	GetItemByID(ctx context.Context, id uuid.UUID) (*models.Item, error)
	GetItemByNormalizedURL(ctx context.Context, familyID uuid.UUID, normalizedURL string) (*models.Item, error)
	GetItemsByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Item, error)
	UpdateItem(ctx context.Context, item *models.Item) error
}

// ContentService handles content business logic
type ContentService struct {
	familyRepo repository.FamilyRepository
	itemRepo   ItemRepository
	normalizer *URLNormalizer
}

// NewContentService creates a new content service
func NewContentService(familyRepo repository.FamilyRepository, itemRepo ItemRepository) *ContentService {
	return &ContentService{
		familyRepo: familyRepo,
		itemRepo:   itemRepo,
		normalizer: NewURLNormalizer(),
	}
}

// ImportItem imports a new content item to a family
// Items are shared at family level - all children in the family can see them
func (s *ContentService) ImportItem(ctx context.Context, familyID uuid.UUID, req *models.ImportItemRequest) (*models.ImportItemResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify family exists
	family, err := s.familyRepo.GetFamilyByID(ctx, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify family: %w", err)
	}
	if family == nil {
		return nil, models.ErrFamilyNotFound
	}

	// Normalize URL
	normalizedURL, err := s.normalizer.NormalizeURL(req.SourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize URL: %w", err)
	}

	// Check for duplicates
	existing, err := s.itemRepo.GetItemByNormalizedURL(ctx, familyID, normalizedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicates: %w", err)
	}
	if existing != nil {
		return &models.ImportItemResponse{
			ID:               existing.ID,
			ProcessingStatus: existing.ProcessingStatus,
			IsDuplicate:      true,
		}, nil
	}

	// Detect source type
	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = s.normalizer.DetectSourceType(req.SourceURL)
	}

	// Determine playback mode
	playbackMode := models.PlaybackModeWebView
	if sourceType == models.SourceTypeWebArticle {
		playbackMode = models.PlaybackModeArticle
	}

	// Create item
	now := time.Now()
	item := &models.Item{
		ID:               uuid.New(),
		FamilyID:         familyID,
		SourceType:       sourceType,
		SourceURL:        req.SourceURL,
		NormalizedURL:    normalizedURL,
		PlaybackMode:     playbackMode,
		ProcessingStatus: models.ProcessingStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.itemRepo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return &models.ImportItemResponse{
		ID:               item.ID,
		ProcessingStatus: item.ProcessingStatus,
		IsDuplicate:      false,
	}, nil
}

// GetItemsByFamily returns all items for a family (used by child app)
func (s *ContentService) GetItemsByFamily(ctx context.Context, familyID uuid.UUID) ([]*models.Item, error) {
	return s.itemRepo.GetItemsByFamilyID(ctx, familyID)
}
