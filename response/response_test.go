package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/response"
)

func decode(t *testing.T, rr *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.NewDecoder(rr.Body).Decode(&env); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return env
}

func TestJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	response.JSON(rr, http.StatusOK, map[string]string{"hello": "world"})
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	env := decode(t, rr)
	if !env.Success {
		t.Error("expected success=true")
	}
}

func TestOK(t *testing.T) {
	rr := httptest.NewRecorder()
	response.OK(rr, "payload")
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestCreated(t *testing.T) {
	rr := httptest.NewRecorder()
	response.Created(rr, map[string]int{"id": 1})
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestNoContent(t *testing.T) {
	rr := httptest.NewRecorder()
	response.NoContent(rr)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestError(t *testing.T) {
	rr := httptest.NewRecorder()
	response.Error(rr, http.StatusBadRequest, "INVALID_INPUT", "bad input")
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	env := decode(t, rr)
	if env.Success {
		t.Error("expected success=false")
	}
	if env.Error == nil {
		t.Fatal("expected error object")
	}
	if env.Error.Code != "INVALID_INPUT" {
		t.Errorf("expected code INVALID_INPUT, got %s", env.Error.Code)
	}
}

func TestBadRequest(t *testing.T) {
	rr := httptest.NewRecorder()
	response.BadRequest(rr, "ERR", "bad")
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestUnauthorized(t *testing.T) {
	rr := httptest.NewRecorder()
	response.Unauthorized(rr, "UNAUTH", "unauthorized")
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestForbidden(t *testing.T) {
	rr := httptest.NewRecorder()
	response.Forbidden(rr, "FORBIDDEN", "no access")
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestNotFound(t *testing.T) {
	rr := httptest.NewRecorder()
	response.NotFound(rr, "NOT_FOUND", "not found")
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestInternalServerError(t *testing.T) {
	rr := httptest.NewRecorder()
	response.InternalServerError(rr, "ISE", "internal")
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestWithMeta(t *testing.T) {
	rr := httptest.NewRecorder()
	meta := &response.Meta{Page: 1, Limit: 10, Total: 100, HasNext: true}
	response.WithMeta(rr, http.StatusOK, []string{"a", "b"}, meta)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	env := decode(t, rr)
	if env.Meta == nil {
		t.Fatal("expected meta")
	}
	if env.Meta.Total != 100 {
		t.Errorf("expected total=100, got %d", env.Meta.Total)
	}
}
