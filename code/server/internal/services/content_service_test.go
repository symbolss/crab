package services

import (
	"context"
	"testing"

	"github.com/family/crab-server/internal/models"
	memrepo "github.com/family/crab-server/internal/repository/memory"
	"github.com/google/uuid"
)

func TestURLNormalizer_NormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid URL with tracking params",
			input:   "https://www.douyin.com/video/123?utm_source=test&id=456",
			wantErr: false,
		},
		{
			name:    "URL without scheme",
			input:   "www.example.com/page",
			wantErr: false,
		},
		{
			name:    "URL with trailing slash",
			input:   "https://example.com/page/",
			wantErr: false,
		},
	}

	normalizer := NewURLNormalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizer.NormalizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("NormalizeURL() returned empty string")
			}
		})
	}
}

func TestURLNormalizer_DetectSourceType(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected models.SourceType
	}{
		{
			name:     "douyin URL",
			url:      "https://www.douyin.com/video/123",
			expected: models.SourceTypeDouyin,
		},
		{
			name:     "douyin short URL",
			url:      "https://v.douyin.com/abc123",
			expected: models.SourceTypeDouyin,
		},
		{
			name:     "x/twitter URL",
			url:      "https://twitter.com/user/status/123",
			expected: models.SourceTypeX,
		},
		{
			name:     "x.com URL",
			url:      "https://x.com/user/status/123",
			expected: models.SourceTypeX,
		},
		{
			name:     "wechat channels URL",
			url:      "https://channels.weixin.qq.com/video/123",
			expected: models.SourceTypeWechatChannels,
		},
		{
			name:     "generic web URL",
			url:      "https://example.com/article/123",
			expected: models.SourceTypeWebArticle,
		},
	}

	normalizer := NewURLNormalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.DetectSourceType(tt.url)
			if got != tt.expected {
				t.Errorf("DetectSourceType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContentService_ImportItem(t *testing.T) {
	ctx := context.Background()
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	itemRepo := memrepo.NewInMemoryItemRepository()
	service := NewContentService(familyRepo, itemRepo)

	// Create a test family
	family := &models.Family{
		ID:          uuid.New(),
		PairingCode: "TEST1234",
	}
	_ = familyRepo.CreateFamily(ctx, family)

	tests := []struct {
		name     string
		familyID uuid.UUID
		req      *models.ImportItemRequest
		wantErr  bool
	}{
		{
			name:     "valid import",
			familyID: family.ID,
			req: &models.ImportItemRequest{
				SourceURL: "https://www.douyin.com/video/123",
			},
			wantErr: false,
		},
		{
			name:     "empty source URL",
			familyID: family.ID,
			req: &models.ImportItemRequest{
				SourceURL: "",
			},
			wantErr: true,
		},
		{
			name:     "non-existent family",
			familyID: uuid.New(),
			req: &models.ImportItemRequest{
				SourceURL: "https://example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.ImportItem(ctx, tt.familyID, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImportItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID == uuid.Nil {
					t.Error("ImportItem() returned nil ID")
				}
				if got.ProcessingStatus != models.ProcessingStatusPending {
					t.Errorf("ImportItem() status = %v, want pending", got.ProcessingStatus)
				}
			}
		})
	}
}

func TestContentService_ImportItem_Duplicate(t *testing.T) {
	ctx := context.Background()
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	itemRepo := memrepo.NewInMemoryItemRepository()
	service := NewContentService(familyRepo, itemRepo)

	// Create a test family
	family := &models.Family{
		ID:          uuid.New(),
		PairingCode: "TEST1234",
	}
	_ = familyRepo.CreateFamily(ctx, family)

	// Import first item
	req := &models.ImportItemRequest{
		SourceURL: "https://www.douyin.com/video/123",
	}
	first, err := service.ImportItem(ctx, family.ID, req)
	if err != nil {
		t.Fatalf("First import failed: %v", err)
	}

	// Import duplicate
	second, err := service.ImportItem(ctx, family.ID, req)
	if err != nil {
		t.Fatalf("Second import failed: %v", err)
	}

	if !second.IsDuplicate {
		t.Error("Second import should be marked as duplicate")
	}
	if second.ID != first.ID {
		t.Error("Duplicate import should return same item ID")
	}
}

func TestContentService_AssignItem(t *testing.T) {
	ctx := context.Background()
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	itemRepo := memrepo.NewInMemoryItemRepository()
	service := NewContentService(familyRepo, itemRepo)

	// Create test family
	family := &models.Family{
		ID:          uuid.New(),
		PairingCode: "TEST1234",
	}
	_ = familyRepo.CreateFamily(ctx, family)

	// Create test item
	item := &models.Item{
		ID:               uuid.New(),
		FamilyID:         family.ID,
		SourceType:       models.SourceTypeDouyin,
		SourceURL:        "https://example.com",
		NormalizedURL:    "https://example.com",
		ProcessingStatus: models.ProcessingStatusReady,
	}
	_ = itemRepo.CreateItem(ctx, item)

	childID := uuid.New()
	parentID := uuid.New()

	tests := []struct {
		name    string
		itemID  uuid.UUID
		req     *models.AssignItemRequest
		wantErr bool
	}{
		{
			name:    "valid assignment",
			itemID:  item.ID,
			req:     &models.AssignItemRequest{ChildIDs: []uuid.UUID{childID}},
			wantErr: false,
		},
		{
			name:    "empty child IDs",
			itemID:  item.ID,
			req:     &models.AssignItemRequest{ChildIDs: []uuid.UUID{}},
			wantErr: true,
		},
		{
			name:    "non-existent item",
			itemID:  uuid.New(),
			req:     &models.AssignItemRequest{ChildIDs: []uuid.UUID{childID}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.AssignItem(ctx, family.ID, tt.itemID, parentID, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssignItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got.AssignedTo) != len(tt.req.ChildIDs) {
					t.Errorf("AssignItem() assigned %d children, want %d", len(got.AssignedTo), len(tt.req.ChildIDs))
				}
			}
		})
	}
}
