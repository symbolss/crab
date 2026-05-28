package models

import (
	"time"

	"github.com/google/uuid"
)

// SourceType represents the source platform of content
type SourceType string

const (
	SourceTypeDouyin         SourceType = "douyin"
	SourceTypeX              SourceType = "x"
	SourceTypeWechatChannels SourceType = "wechat_channels"
	SourceTypeWebArticle     SourceType = "web_article"
)

// ProcessingStatus represents the processing state of an item
type ProcessingStatus string

const (
	ProcessingStatusPending ProcessingStatus = "pending"
	ProcessingStatusReady   ProcessingStatus = "ready"
	ProcessingStatusFailed  ProcessingStatus = "failed"
)

// PlaybackMode represents how content should be played
type PlaybackMode string

const (
	PlaybackModeWebView      PlaybackMode = "webview"
	PlaybackModePrivateVideo PlaybackMode = "private_video"
	PlaybackModeArticle      PlaybackMode = "article"
	PlaybackModeAudio        PlaybackMode = "audio"
)

// Item represents a content item in the system
// Items are shared at family level - all children in the family can see them
type Item struct {
	ID               uuid.UUID        `json:"id"`
	FamilyID         uuid.UUID        `json:"familyId"`
	SourceType       SourceType       `json:"sourceType"`
	SourceURL        string           `json:"sourceUrl"`
	NormalizedURL    string           `json:"normalizedUrl"`
	Title            string           `json:"title,omitempty"`
	CoverURL         string           `json:"coverUrl,omitempty"`
	Summary          string           `json:"summary,omitempty"`
	PlaybackMode     PlaybackMode     `json:"playbackMode"`
	ProcessingStatus ProcessingStatus `json:"processingStatus"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

// ImportItemRequest represents the request body for importing content
type ImportItemRequest struct {
	SourceURL  string     `json:"sourceUrl"`
	SourceType SourceType `json:"sourceType,omitempty"` // Optional, auto-detected if not provided
}

// ImportItemResponse represents the response for importing content
type ImportItemResponse struct {
	ID               uuid.UUID        `json:"id"`
	ProcessingStatus ProcessingStatus `json:"processingStatus"`
	IsDuplicate      bool             `json:"isDuplicate"`
}

// Validate validates the ImportItemRequest
func (r *ImportItemRequest) Validate() error {
	if r.SourceURL == "" {
		return ErrSourceURLRequired
	}
	if len(r.SourceURL) > 2048 {
		return ErrSourceURLTooLong
	}
	return nil
}
