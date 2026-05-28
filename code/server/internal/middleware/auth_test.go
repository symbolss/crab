package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/family/crab-server/internal/middleware"
	"github.com/family/crab-server/internal/models"
	memrepo "github.com/family/crab-server/internal/repository/memory"
	"github.com/family/crab-server/internal/services"
	"github.com/google/uuid"
)

const testJWTSecret = "test-secret-key-minimum-32-characters"

func setupAuthMiddleware() (*middleware.AuthMiddleware, *services.FamilyService) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)
	return middleware.NewAuthMiddleware(svc), svc
}

func TestAuthMiddleware_RequireAuth_Success(t *testing.T) {
	m, svc := setupAuthMiddleware()

	// Create a family to get a token
	resp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Create a handler that checks context
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		userId := r.Context().Value("userId").(uuid.UUID)
		role := r.Context().Value("role").(string)
		familyId := r.Context().Value("familyId").(uuid.UUID)

		if userId != resp.FamilyID {
			t.Errorf("userId = %v, want %v", userId, resp.FamilyID)
		}
		if role != "parent" {
			t.Errorf("role = %v, want parent", role)
		}
		if familyId != resp.FamilyID {
			t.Errorf("familyId = %v, want %v", familyId, resp.FamilyID)
		}
	})

	// Create request with Bearer token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+resp.ParentToken)
	w := httptest.NewRecorder()

	m.RequireAuth(testHandler).ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("handler should have been called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_RequireAuth_MissingHeader(t *testing.T) {
	m, _ := setupAuthMiddleware()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	m.RequireAuth(testHandler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("handler should not have been called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_RequireAuth_InvalidFormat(t *testing.T) {
	m, _ := setupAuthMiddleware()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	m.RequireAuth(testHandler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("handler should not have been called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	m, _ := setupAuthMiddleware()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	m.RequireAuth(testHandler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("handler should not have been called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_RequireAuth_ChildToken(t *testing.T) {
	m, svc := setupAuthMiddleware()

	// Create a family and child
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	childResp, err := svc.PairChild(context.Background(), &models.PairChildRequest{
		PairingCode: familyResp.PairingCode,
		ChildName:   "Little John",
	})
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	// Create a handler that checks context
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		userId := r.Context().Value("userId").(uuid.UUID)
		role := r.Context().Value("role").(string)
		familyId := r.Context().Value("familyId").(uuid.UUID)

		if userId != childResp.ChildID {
			t.Errorf("userId = %v, want %v", userId, childResp.ChildID)
		}
		if role != "child" {
			t.Errorf("role = %v, want child", role)
		}
		// Child tokens should now have familyId set from database lookup
		if familyId != familyResp.FamilyID {
			t.Errorf("familyId = %v, want %v", familyId, familyResp.FamilyID)
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+childResp.ChildToken)
	w := httptest.NewRecorder()

	m.RequireAuth(testHandler).ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("handler should have been called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_RequireParent_Success(t *testing.T) {
	m, svc := setupAuthMiddleware()

	// Create a family to get a parent token
	resp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+resp.ParentToken)
	w := httptest.NewRecorder()

	m.RequireParent(testHandler).ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("handler should have been called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_RequireParent_ChildBlocked(t *testing.T) {
	m, svc := setupAuthMiddleware()

	// Create a family and child
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	childResp, err := svc.PairChild(context.Background(), &models.PairChildRequest{
		PairingCode: familyResp.PairingCode,
		ChildName:   "Little John",
	})
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+childResp.ChildToken)
	w := httptest.NewRecorder()

	m.RequireParent(testHandler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("handler should not have been called for child token")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestAuthMiddleware_RequireParent_MissingHeader(t *testing.T) {
	m, _ := setupAuthMiddleware()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	m.RequireParent(testHandler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("handler should not have been called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// Test edge case: Bearer with no token
func TestAuthMiddleware_RequireAuth_BearerNoToken(t *testing.T) {
	m, _ := setupAuthMiddleware()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	m.RequireAuth(testHandler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("handler should not have been called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
