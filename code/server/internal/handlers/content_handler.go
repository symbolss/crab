package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/family/crab-server/internal/models"
	"github.com/family/crab-server/internal/services"
	"github.com/family/crab-server/pkg/response"
)

// ContentHandler handles content-related HTTP requests
type ContentHandler struct {
	contentService *services.ContentService
}

// NewContentHandler creates a new content handler
func NewContentHandler(contentService *services.ContentService) *ContentHandler {
	return &ContentHandler{
		contentService: contentService,
	}
}

// ImportItem handles POST /api/items/import
// Items are shared at family level - all children in the family can see them
func (h *ContentHandler) ImportItem(w http.ResponseWriter, r *http.Request) {
	// Get family ID from context (set by auth middleware)
	familyID, ok := r.Context().Value("familyId").(interface{ String() string })
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "family id not found in context")
		return
	}

	// Parse request
	var req models.ImportItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	// Convert familyID to uuid.UUID
	familyUUID := parseUUID(familyID.String())
	if familyUUID == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid family id")
		return
	}

	// Import item
	result, err := h.contentService.ImportItem(r.Context(), *familyUUID, &req)
	if err != nil {
		switch err {
		case models.ErrSourceURLRequired, models.ErrSourceURLTooLong:
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case models.ErrFamilyNotFound:
			response.Error(w, http.StatusNotFound, "FAMILY_NOT_FOUND", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to import item")
		}
		return
	}

	status := http.StatusCreated
	if result.IsDuplicate {
		status = http.StatusOK
	}
	response.JSON(w, status, result)
}

// parseUUID parses a UUID string, returns nil if invalid
func parseUUID(s string) *[16]byte {
	if len(s) != 36 {
		return nil
	}
	var uuid [16]byte
	for i := 0; i < 16; i++ {
		var hi, lo byte
		switch {
		case s[i*2] >= '0' && s[i*2] <= '9':
			hi = s[i*2] - '0'
		case s[i*2] >= 'a' && s[i*2] <= 'f':
			hi = s[i*2] - 'a' + 10
		case s[i*2] >= 'A' && s[i*2] <= 'F':
			hi = s[i*2] - 'A' + 10
		default:
			return nil
		}
		switch {
		case s[i*2+1] >= '0' && s[i*2+1] <= '9':
			lo = s[i*2+1] - '0'
		case s[i*2+1] >= 'a' && s[i*2+1] <= 'f':
			lo = s[i*2+1] - 'a' + 10
		case s[i*2+1] >= 'A' && s[i*2+1] <= 'F':
			lo = s[i*2+1] - 'A' + 10
		default:
			return nil
		}
		uuid[i] = hi<<4 | lo
		if i == 3 || i == 5 || i == 7 || i == 9 {
			if s[i*2+2] != '-' {
				return nil
			}
		}
	}
	return &uuid
}

// GetItems handles GET /api/items
// Returns all items for the authenticated family
func (h *ContentHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	// Get family ID from context
	familyID, ok := r.Context().Value("familyId").(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "family id not found in context")
		return
	}

	familyUUID, err := uuid.Parse(familyID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid family id")
		return
	}

	// Get items
	items, err := h.contentService.GetItemsByFamily(r.Context(), familyUUID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get items")
		return
	}

	response.OK(w, items)
}

// GetItem handles GET /api/items/{id}
// Returns a single item by ID
func (h *ContentHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid item id")
		return
	}

	// Get family ID from context for authorization
	familyID, ok := r.Context().Value("familyId").(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "family id not found")
		return
	}

	familyUUID, err := uuid.Parse(familyID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid family id")
		return
	}

	// Get item
	item, err := h.contentService.GetItemByID(r.Context(), itemID, familyUUID)
	if err != nil {
		switch err {
		case models.ErrItemNotFound:
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "item not found")
		case models.ErrUnauthorized:
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get item")
		}
		return
	}

	response.OK(w, item)
}

// GetPlayback handles GET /api/items/{id}/playback
// Returns playback configuration for an item
func (h *ContentHandler) GetPlayback(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid item id")
		return
	}

	// Get family ID from context
	familyID, ok := r.Context().Value("familyId").(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "family id not found")
		return
	}

	familyUUID, err := uuid.Parse(familyID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid family id")
		return
	}

	// Get playback config
	config, err := h.contentService.GetPlaybackConfig(r.Context(), itemID, familyUUID)
	if err != nil {
		switch err {
		case models.ErrItemNotFound:
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "item not found")
		case models.ErrUnauthorized:
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get playback config")
		}
		return
	}

	response.OK(w, config)
}

// PlaybackConfigResponse represents the playback configuration response
type PlaybackConfigResponse struct {
	ItemID        uuid.UUID       `json:"itemId"`
	PlaybackMode  string          `json:"playbackMode"`
	Asset         *PlaybackAsset  `json:"asset,omitempty"`
	WebViewPolicy *WebViewPolicy  `json:"webviewPolicy,omitempty"`
}

// PlaybackAsset represents a playback asset
type PlaybackAsset struct {
	Kind string `json:"kind"`
	URL  string `json:"url"`
}

// WebViewPolicy represents WebView security policy
type WebViewPolicy struct {
	AllowDomains       []string `json:"allowDomains"`
	BlockExternalScheme bool    `json:"blockExternalScheme"`
	BlockNewWindow      bool    `json:"blockNewWindow"`
}
