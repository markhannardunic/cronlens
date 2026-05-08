package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cronlens/internal/api"
	"cronlens/internal/scheduler"
)

type nopCollector struct{}

func (n *nopCollector) Run() error { return nil }

func newTestScheduler(t *testing.T, jobs map[string]string) *scheduler.Scheduler {
	t.Helper()
	s := scheduler.New()
	for name, expr := range jobs {
		if err := s.Register(name, expr, &nopCollector{}); err != nil {
			t.Fatalf("register %s: %v", name, err)
		}
	}
	return s
}

func TestScheduleHandler_Empty(t *testing.T) {
	s := scheduler.New()
	h := api.NewScheduleHandler(s)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/schedule", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var out []map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty list, got %d items", len(out))
	}
}

func TestScheduleHandler_WithJobs(t *testing.T) {
	s := newTestScheduler(t, map[string]string{
		"backup": "@every 1h",
		"report": "0 9 * * *",
	})
	h := api.NewScheduleHandler(s)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/schedule", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var out []map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 entries, got %d", len(out))
	}
	for _, entry := range out {
		if entry["name"] == "" || entry["schedule"] == "" {
			t.Errorf("entry missing fields: %v", entry)
		}
	}
}
