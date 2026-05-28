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

func TestCreateFamilyHandler_ResponseFields(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)
	handler := handlers.NewFamilyHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body := models.CreateFamilyRequest{ParentName: "John Doe"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/family/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
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

	// Verify all fields are present
	if data.FamilyID == uuid.Nil {
		t.Error("FamilyID should not be nil")
	}
	if data.PairingCode == "" {
		t.Error("PairingCode should not be empty")
	}
	if data.ParentToken == "" {
		t.Error("ParentToken should not be empty")
	}
}

func TestPairChildHandler_ResponseFields(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)
	handler := handlers.NewFamilyHandler(svc)

	// Create family first
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

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
	}

	var resp testResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	var data models.PairChildResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	// Verify all fields
	if data.ChildID == uuid.Nil {
		t.Error("ChildID should not be nil")
	}
	if data.FamilyID == uuid.Nil {
		t.Error("FamilyID should not be nil")
	}
	if data.ChildName == "" {
		t.Error("ChildName should not be empty")
	}
	if data.ChildToken == "" {
		t.Error("ChildToken should not be empty")
	}
}

func TestGetChildrenHandler_MultipleChildren(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)
	handler := handlers.NewFamilyHandler(svc)

	// Create family and multiple children
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	childNames := []string{"Alice", "Bob", "Charlie"}
	for _, name := range childNames {
		_, err := svc.PairChild(context.Background(), &models.PairChildRequest{
			PairingCode: familyResp.PairingCode,
			ChildName:   name,
		})
		if err != nil {
			t.Fatalf("PairChild() error = %v", err)
		}
	}

	r := chi.NewRouter()
	r.Get("/api/children", handler.GetChildren)

	req := httptest.NewRequest("GET", "/api/children", nil)
	req = req.WithContext(contextWithFamilyID(req, familyResp.FamilyID))

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

	if len(children) != 3 {
		t.Errorf("len(children) = %d, want 3", len(children))
	}

	// Verify each child has required fields
	for _, child := range children {
		if child.ID == uuid.Nil {
			t.Error("Child ID should not be nil")
		}
		if child.Name == "" {
			t.Error("Child Name should not be empty")
		}
		if child.Status != "active" {
			t.Errorf("Child Status = %v, want active", child.Status)
		}
	}
}

func TestCreateFamilyHandler_ContentType(t *testing.T) {
	r, _ := setupTestRouter()

	body := models.CreateFamilyRequest{ParentName: "John Doe"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/family/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", contentType)
	}
}
