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
	// Patterns for source detection
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

// NormalizeURL normalizes a URL by removing tracking parameters and expanding short links
func (n *URLNormalizer) NormalizeURL(rawURL string) (string, error) {
	// Parse the URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Remove common tracking parameters
	trackingParams := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "fbclid", "gclid"}
	q := parsed.Query()
	for _, param := range trackingParams {
		q.Del(param)
	}
	parsed.RawQuery = q.Encode()

	// Normalize scheme
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}

	// Remove trailing slash from path
	parsed.Path = strings.TrimRight(parsed.Path, "/")

	return parsed.String(), nil
}

// DetectSourceType detects the source platform from a URL
func (n *URLNormalizer) DetectSourceType(rawURL string) models.SourceType {
	normalized := strings.ToLower(rawURL)

	if n.douyinPattern.MatchString(normalized) {
		return models.SourceTypeDouyin
	}
	if n.xPattern.MatchString(normalized) {
		return models.SourceTypeX
	}
	if n.wechatChannelsPattern.MatchString(normalized) {
		return models.SourceTypeWechatChannels
	}

	return models.SourceTypeWebArticle
}

// ContentService handles content business logic
type ContentService struct {
	familyRepo repository.FamilyRepository
	itemRepo   ItemRepository
	normalizer *URLNormalizer
}

// ItemRepository defines the interface for item data operations
type ItemRepository interface {
	CreateItem(ctx context.Context, item *models.Item) error
	GetItemByID(ctx context.Context, id uuid.UUID) (*models.Item, error)
	GetItemByNormalizedURL(ctx context.Context, familyID uuid.UUID, normalizedURL string) (*models.Item, error)
	GetItemsByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*models.Item, error)
	UpdateItem(ctx context.Context, item *models.Item) error

	CreateAssignment(ctx context.Context, assignment *models.Assignment) error
	GetAssignmentsByItemID(ctx context.Context, itemID uuid.UUID) ([]*models.Assignment, error)
	GetAssignmentsByChildID(ctx context.Context, childID uuid.UUID) ([]*models.Assignment, error)
	UpdateAssignmentState(ctx context.Context, id uuid.UUID, state string) error
}

// NewContentService creates a new content service
func NewContentService(familyRepo repository.FamilyRepository, itemRepo ItemRepository) *ContentService {
	return &ContentService{
		familyRepo: familyRepo,
		itemRepo:   itemRepo,
		normalizer: NewURLNormalizer(),
	}
}

// ImportItem imports a new content item
func (s *ContentService) ImportItem(ctx context.Context, familyID uuid.UUID, req *models.ImportItemRequest) (*models.ImportItemResponse, error) {
	// Validate request
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
	existingItem, err := s.itemRepo.GetItemByNormalizedURL(ctx, familyID, normalizedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}
	if existingItem != nil {
		return &models.ImportItemResponse{
			ID:               existingItem.ID,
			ProcessingStatus: existingItem.ProcessingStatus,
			IsDuplicate:      true,
		}, nil
	}

	// Detect source type if not provided
	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = s.normalizer.DetectSourceType(req.SourceURL)
	}

	// Determine initial playback mode
	playbackMode := models.PlaybackModeWebView
	if sourceType == models.SourceTypeWebArticle {
		playbackMode = models.PlaybackModeArticle
	}

	// Create new item
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

	// Save to repository
	if err := s.itemRepo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return &models.ImportItemResponse{
		ID:               item.ID,
		ProcessingStatus: item.ProcessingStatus,
		IsDuplicate:      false,
	}, nil
}

// AssignItem assigns an item to children
func (s *ContentService) AssignItem(ctx context.Context, familyID, itemID, parentID uuid.UUID, req *models.AssignItemRequest) (*models.AssignItemResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify item exists and belongs to family
	item, err := s.itemRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return nil, models.ErrItemNotFound
	}
	if item.FamilyID != familyID {
		return nil, models.ErrUnauthorized
	}

	// Create assignments for each child
	now := time.Now()
	assignedTo := make([]uuid.UUID, 0, len(req.ChildIDs))

	for _, childID := range req.ChildIDs {
		assignment := &models.Assignment{
			ID:         uuid.New(),
			ItemID:     itemID,
			ChildID:    childID,
			AssignedBy: parentID,
			AssignedAt: now,
			State:      "active",
		}

		if err := s.itemRepo.CreateAssignment(ctx, assignment); err != nil {
			return nil, fmt.Errorf("failed to create assignment for child %s: %w", childID, err)
		}
		assignedTo = append(assignedTo, childID)
	}

	return &models.AssignItemResponse{
		ItemID:     itemID,
		AssignedTo: assignedTo,
	}, nil
}
