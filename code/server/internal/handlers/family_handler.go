package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/family/crab-server/internal/models"
	"github.com/family/crab-server/internal/services"
	"github.com/family/crab-server/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// FamilyHandler handles family-related HTTP requests
type FamilyHandler struct {
	svc *services.FamilyService
}

// NewFamilyHandler creates a new FamilyHandler
func NewFamilyHandler(svc *services.FamilyService) *FamilyHandler {
	return &FamilyHandler{svc: svc}
}

// RegisterRoutes registers family routes on the given router
func (h *FamilyHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/family/create", h.CreateFamily)
	r.Post("/api/family/pair", h.PairChild)
	r.Get("/api/children", h.GetChildren)
}

// CreateFamily handles POST /api/family/create
func (h *FamilyHandler) CreateFamily(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Create family
	resp, err := h.svc.CreateFamily(r.Context(), &req)
	if err != nil {
		response.InternalError(w, "Failed to create family")
		return
	}

	response.Created(w, resp)
}

// PairChild handles POST /api/family/pair
func (h *FamilyHandler) PairChild(w http.ResponseWriter, r *http.Request) {
	var req models.PairChildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Pair child
	resp, err := h.svc.PairChild(r.Context(), &req)
	if err != nil {
		if errors.Is(err, models.ErrPairingCodeInvalid) || errors.Is(err, models.ErrPairingCodeExpired) {
			response.Error(w, http.StatusNotFound, "PAIRING_CODE_INVALID", "Invalid or expired pairing code")
			return
		}
		response.InternalError(w, "Failed to pair child")
		return
	}

	response.Created(w, resp)
}

// GetChildren handles GET /api/children
func (h *FamilyHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	// Get family ID from context (set by auth middleware)
	familyID, ok := r.Context().Value("familyId").(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Get children
	children, err := h.svc.GetChildrenByFamilyID(r.Context(), familyID)
	if err != nil {
		response.InternalError(w, "Failed to get children")
		return
	}

	response.OK(w, children)
}
