package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronlens/internal/api"
	"github.com/cronlens/internal/pinned"
)

func newPinnedHandler() *api.PinnedHandler {
	return api.NewPinnedHandler(pinned.New())
}

func TestPinnedHandler_MethodNotAllowed(t *testing.T) {
	h := newPinnedHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/pinned", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestPinnedHandler_EmptyList(t *testing.T) {
	h := newPinnedHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/pinned", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string][]string
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if len(body["pinned"]) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestPinnedHandler_PinAndList(t *testing.T) {
	h := newPinnedHandler()

	// Pin a job
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/pinned/backup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// List should include it
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/pinned", nil))
	var body map[string][]string
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if len(body["pinned"]) != 1 || body["pinned"][0] != "backup" {
		t.Errorf("expected [backup], got %v", body["pinned"])
	}
}

func TestPinnedHandler_DuplicatePin_Conflict(t *testing.T) {
	h := newPinnedHandler()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPut, "/api/pinned/sync", nil))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/pinned/sync", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestPinnedHandler_Unpin(t *testing.T) {
	h := newPinnedHandler()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPut, "/api/pinned/cleanup", nil))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/pinned/cleanup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestPinnedHandler_UnpinUnknown_NotFound(t *testing.T) {
	h := newPinnedHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/pinned/ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
