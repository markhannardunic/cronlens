package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronlens/internal/api"
	"github.com/yourorg/cronlens/internal/job"
	"github.com/yourorg/cronlens/internal/store"
)

func recordExportRun(s *store.Store, name string, success bool) {
	r := job.NewRun(name)
	r.StartedAt = time.Now().Add(-time.Minute)
	r.FinishedAt = time.Now()
	r.Success = success
	s.Record(r)
}

func TestExportHandler_UnsupportedFormat(t *testing.T) {
	s := store.New()
	h := api.NewExportHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/api/export?format=xml", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestExportHandler_MethodNotAllowed(t *testing.T) {
	s := store.New()
	h := api.NewExportHandler(s)
	req := httptest.NewRequest(http.MethodPost, "/api/export", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestExportHandler_JSON(t *testing.T) {
	s := store.New()
	recordExportRun(s, "backup", true)
	recordExportRun(s, "cleanup", false)
	h := api.NewExportHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/api/export?format=json", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected content-type: %s", ct)
	}
	var runs []job.Run
	if err := json.NewDecoder(rec.Body).Decode(&runs); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestExportHandler_CSV(t *testing.T) {
	s := store.New()
	recordExportRun(s, "sync", true)
	h := api.NewExportHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/api/export?format=csv", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("unexpected content-type: %s", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "sync") {
		t.Errorf("expected job name in CSV output")
	}
}

func TestExportHandler_InvalidSince(t *testing.T) {
	s := store.New()
	h := api.NewExportHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/api/export?since=not-a-date", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
