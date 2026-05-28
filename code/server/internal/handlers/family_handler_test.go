package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/family/crab-server/internal/handlers"
	"github.com/family/crab-server/internal/models"
	memrepo "github.com/family/crab-server/internal/repository/memory"
	"github.com/family/crab-server/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const testJWTSecret = "test-secret-key-minimum-32-characters"

// Response envelope matching pkg/response.Response
type testResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *testErrorInfo  `json:"error,omitempty"`
}

type testErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func setupTestRouter() (*chi.Mux, *services.FamilyService) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)
	handler := handlers.NewFamilyHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, svc
}

func TestCreateFamilyHandler_Success(t *testing.T) {
	r, _ := setupTestRouter()

	body := models.CreateFamilyRequest{ParentName: "John Doe"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/family/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		t.Logf("body: %s", w.Body.String())
	}

	var resp testResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("response.Success should be true")
	}

	var data models.CreateFamilyResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if data.FamilyID == uuid.Nil {
		t.Error("FamilyID should not be nil")
	}
	if len(data.PairingCode) != 8 {
		t.Errorf("PairingCode length = %d, want 8", len(data.PairingCode))
	}
	if data.ParentToken == "" {
		t.Error("ParentToken should not be empty")
	}
}

func TestCreateFamilyHandler_InvalidJSON(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest("POST", "/api/family/create", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateFamilyHandler_EmptyParentName(t *testing.T) {
	r, _ := setupTestRouter()

	body := models.CreateFamilyRequest{ParentName: ""}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/family/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPairChildHandler_Success(t *testing.T) {
	r, svc := setupTestRouter()

	// First create a family
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Pair a child
	body := models.PairChildRequest{
		PairingCode: familyResp.PairingCode,
		ChildName:   "Little John",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/family/pair", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		t.Logf("body: %s", w.Body.String())
	}

	var resp testResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	var data models.PairChildResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if data.ChildID == uuid.Nil {
		t.Error("ChildID should not be nil")
	}
	if data.FamilyID != familyResp.FamilyID {
		t.Errorf("FamilyID = %v, want %v", data.FamilyID, familyResp.FamilyID)
	}
	if data.ChildName != "Little John" {
		t.Errorf("ChildName = %v, want Little John", data.ChildName)
	}
	if data.ChildToken == "" {
		t.Error("ChildToken should not be empty")
	}
}

func TestPairChildHandler_InvalidPairingCode(t *testing.T) {
	r, _ := setupTestRouter()

	body := models.PairChildRequest{
		PairingCode: "INVALID1",
		ChildName:   "Little John",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/family/pair", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestPairChildHandler_EmptyChildName(t *testing.T) {
	r, _ := setupTestRouter()

	body := models.PairChildRequest{
		PairingCode: "ABCD1234",
		ChildName:   "",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/family/pair", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPairChildHandler_InvalidPairingCodeLength(t *testing.T) {
	r, _ := setupTestRouter()

	body := models.PairChildRequest{
		PairingCode: "TOOLONG12",
		ChildName:   "Little John",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/family/pair", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetChildrenHandler_Success(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)
	handler := handlers.NewFamilyHandler(svc)

	// Create family and children
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	_, err = svc.PairChild(context.Background(), &models.PairChildRequest{
		PairingCode: familyResp.PairingCode,
		ChildName:   "Child A",
	})
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	// Create router and handler
	r := chi.NewRouter()
	r.Get("/api/children", handler.GetChildren)

	req := httptest.NewRequest("GET", "/api/children", nil)
	// Set familyId in context
	req = req.WithContext(contextWithFamilyID(req, familyResp.FamilyID))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		t.Logf("body: %s", w.Body.String())
	}

	var resp testResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	var children []*models.ChildListResponse
	if err := json.Unmarshal(resp.Data, &children); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(children) != 1 {
		t.Errorf("len(children) = %d, want 1", len(children))
	}
}

func TestGetChildrenHandler_Empty(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)
	handler := handlers.NewFamilyHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/children", handler.GetChildren)

	req := httptest.NewRequest("GET", "/api/children", nil)
	// Set familyId in context
	req = req.WithContext(contextWithFamilyID(req, uuid.New()))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp testResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	var children []*models.ChildListResponse
	if err := json.Unmarshal(resp.Data, &children); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(children) != 0 {
		t.Errorf("len(children) = %d, want 0", len(children))
	}
}

func TestGetChildrenHandler_Unauthorized(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest("GET", "/api/children", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// Helper function to create context with family ID
func contextWithFamilyID(r *http.Request, familyID uuid.UUID) context.Context {
	return context.WithValue(r.Context(), "familyId", familyID)
}
