package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/family/crab-server/internal/models"
	"github.com/family/crab-server/pkg/response"
	"github.com/google/uuid"
)

// PlayEventHandler handles play event HTTP requests
type PlayEventHandler struct {
	playEventService *services.PlayEventService
}

// NewPlayEventHandler creates a new play event handler
func NewPlayEventHandler(playEventService *services.PlayEventService) *PlayEventHandler {
	return &PlayEventHandler{
		playEventService: playEventService,
	}
}

// ReportEvent handles POST /api/play-events
func (h *PlayEventHandler) ReportEvent(w http.ResponseWriter, r *http.Request) {
	// Get child ID from context (set by auth middleware)
	childID, ok := r.Context().Value("childId").(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "child id not found")
		return
	}

	childUUID, err := uuid.Parse(childID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid child id")
		return
	}

	// Parse request
	var req PlayEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Report event
	err = h.playEventService.ReportEvent(r.Context(), childUUID, req.ToModel())
	if err != nil {
		switch err {
		case models.ErrUsageExceeded:
			response.Error(w, http.StatusForbidden, "USAGE_EXCEEDED", "daily usage limit exceeded")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to report event")
		}
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}

// PlayEventRequest represents the play event request body
type PlayEventRequest struct {
	ItemID      string `json:"itemId"`
	EventType   string `json:"eventType"`
	PositionSec int    `json:"positionSec"`
	OccurredAt  string `json:"occurredAt"`
}

// Validate validates the request
func (r *PlayEventRequest) Validate() error {
	if r.ItemID == "" {
		return models.ErrInvalidChildID
	}
	if r.EventType == "" {
		return models.ErrSourceURLRequired
	}
	validTypes := map[string]bool{"open": true, "start": true, "pause": true, "complete": true, "favorite": true}
	if !validTypes[r.EventType] {
		return models.ErrInvalidPairingCode
	}
	return nil
}

// ToModel converts request to model
func (r *PlayEventRequest) ToModel() *models.PlayEvent {
	itemUUID, _ := uuid.Parse(r.ItemID)
	occurredAt, _ := time.Parse(time.RFC3339, r.OccurredAt)
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	return &models.PlayEvent{
		ItemID:      itemUUID,
		EventType:   r.EventType,
		PositionSec: r.PositionSec,
		OccurredAt:  occurredAt,
	}
}
