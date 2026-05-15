package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronlens/internal/api"
	"github.com/user/cronlens/internal/healthcheck"
	"github.com/user/cronlens/internal/job"
	"github.com/user/cronlens/internal/store"
)

func recordHealthRun(s *store.Store, name string, success bool, ago time.Duration) {
	r := job.NewRun(name)
	r.Finish(success, "")
	r.StartedAt = time.Now().Add(-ago)
	r.FinishedAt = r.StartedAt.Add(time.Second)
	s.Record(r)
}

func newHCHandler() (*api.HealthCheckHandler, *store.Store) {
	s := store.New()
	c := healthcheck.New()
	return api.NewHealthCheckHandler(c, s), s
}

func TestHealthCheckHandler_MethodNotAllowed(t *testing.T) {
	h, _ := newHCHandler()
	req := httptest.NewRequest(http.MethodDelete, "/healthcheck", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHealthCheckHandler_RegisterInvalidJSON(t *testing.T) {
	h, _ := newHCHandler()
	req := httptest.NewRequest(http.MethodPost, "/healthcheck", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHealthCheckHandler_RegisterAndQueryHealthy(t *testing.T) {
	h, s := newHCHandler()

	body := `{"job_name":"backup","interval_seconds":3600}`
	req := httptest.NewRequest(http.MethodPost, "/healthcheck", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", w.Code)
	}

	recordHealthRun(s, "backup", true, 10*time.Minute)

	req2 := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	var results []healthcheck.Result
	if err := json.NewDecoder(w2.Body).Decode(&results); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Healthy {
		t.Errorf("expected healthy, got: %s", results[0].Message)
	}
}

func TestHealthCheckHandler_UnhealthyWhenStale(t *testing.T) {
	h, s := newHCHandler()

	body := `{"job_name":"sync","interval_seconds":60}`
	req := httptest.NewRequest(http.MethodPost, "/healthcheck", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	recordHealthRun(s, "sync", true, 10*time.Minute)

	req2 := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	var results []healthcheck.Result
	_ = json.NewDecoder(w2.Body).Decode(&results)
	if results[0].Healthy {
		t.Errorf("expected unhealthy for stale job")
	}
}
