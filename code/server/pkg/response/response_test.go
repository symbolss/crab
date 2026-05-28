package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/family/crab-server/pkg/response"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	response.JSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !result.Success {
		t.Error("Success should be true for 2xx status")
	}
}

func TestJSON_Created(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123"}

	response.JSON(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var result response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !result.Success {
		t.Error("Success should be true for 201 status")
	}
}

func TestJSON_ErrorStatus(t *testing.T) {
	w := httptest.NewRecorder()

	response.JSON(w, http.StatusBadRequest, nil)

	var result response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Success {
		t.Error("Success should be false for 4xx status")
	}
}

func TestOK(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}

	response.OK(w, data)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123"}

	response.Created(w, data)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()

	response.BadRequest(w, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var result response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Success {
		t.Error("Success should be false")
	}
	if result.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if result.Error.Code != "BAD_REQUEST" {
		t.Errorf("Error.Code = %s, want BAD_REQUEST", result.Error.Code)
	}
	if result.Error.Message != "invalid input" {
		t.Errorf("Error.Message = %s, want invalid input", result.Error.Message)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()

	response.NotFound(w, "resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var result response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Error.Code != "NOT_FOUND" {
		t.Errorf("Error.Code = %s, want NOT_FOUND", result.Error.Code)
	}
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()

	response.InternalError(w, "internal server error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var result response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("Error.Code = %s, want INTERNAL_ERROR", result.Error.Code)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()

	response.Error(w, http.StatusForbidden, "FORBIDDEN", "access denied")

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}

	var result response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Error.Code != "FORBIDDEN" {
		t.Errorf("Error.Code = %s, want FORBIDDEN", result.Error.Code)
	}
	if result.Error.Message != "access denied" {
		t.Errorf("Error.Message = %s, want access denied", result.Error.Message)
	}
}

func TestPaginated(t *testing.T) {
	w := httptest.NewRecorder()
	data := []map[string]string{
		{"id": "1"},
		{"id": "2"},
	}

	response.Paginated(w, data, 100, 1, 10)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !result.Success {
		t.Error("Success should be true")
	}
	if result.Meta == nil {
		t.Fatal("Meta should not be nil")
	}
	if result.Meta.Total != 100 {
		t.Errorf("Meta.Total = %d, want 100", result.Meta.Total)
	}
	if result.Meta.Page != 1 {
		t.Errorf("Meta.Page = %d, want 1", result.Meta.Page)
	}
	if result.Meta.Limit != 10 {
		t.Errorf("Meta.Limit = %d, want 10", result.Meta.Limit)
	}
}

func TestContentType(t *testing.T) {
	w := httptest.NewRecorder()

	response.OK(w, map[string]string{})

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", contentType)
	}
}
