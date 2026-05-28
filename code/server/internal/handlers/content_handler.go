package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/family/crab-server/internal/models"
	"github.com/family/crab-server/internal/services"
	"github.com/family/crab-server/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ContentHandler handles content-related HTTP requests
type ContentHandler struct {
	contentService *services.ContentService
	familyService  *services.FamilyService
}

// NewContentHandler creates a new content handler
func NewContentHandler(contentService *services.ContentService, familyService *services.FamilyService) *ContentHandler {
	return &ContentHandler{
		contentService: contentService,
		familyService:  familyService,
	}
}

// ImportItem handles POST /api/items/import
func (h *ContentHandler) ImportItem(w http.ResponseWriter, r *http.Request) {
	// Get family ID from context (set by auth middleware)
	familyID, ok := r.Context().Value("familyId").(uuid.UUID)
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

	// Import item
	result, err := h.contentService.ImportItem(r.Context(), familyID, &req)
	if err != nil {
		switch err {
		case models.ErrSourceURLRequired:
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case models.ErrSourceURLTooLong:
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case models.ErrFamilyNotFound:
			response.Error(w, http.StatusNotFound, "FAMILY_NOT_FOUND", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to import item")
		}
		return
	}

	// Return response
	status := http.StatusCreated
	if result.IsDuplicate {
		status = http.StatusOK
	}
	response.JSON(w, status, result)
}

// AssignItem handles POST /api/items/{id}/assign
func (h *ContentHandler) AssignItem(w http.ResponseWriter, r *http.Request) {
	// Get family ID and parent ID from context
	familyID, ok := r.Context().Value("familyId").(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "family id not found in context")
		return
	}

	parentID, ok := r.Context().Value("userId").(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "user id not found in context")
		return
	}

	// Get item ID from URL
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ITEM_ID", "invalid item id format")
		return
	}

	// Parse request
	var req models.AssignItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to parse request body")
		return
	}

	// Assign item
	result, err := h.contentService.AssignItem(r.Context(), familyID, itemID, parentID, &req)
	if err != nil {
		switch err {
		case models.ErrChildIDsRequired:
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case models.ErrInvalidChildID:
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case models.ErrItemNotFound:
			response.Error(w, http.StatusNotFound, "ITEM_NOT_FOUND", err.Error())
		case models.ErrUnauthorized:
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "not authorized to access this item")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to assign item")
		}
		return
	}

	// Return response
	response.JSON(w, http.StatusOK, result)
}
