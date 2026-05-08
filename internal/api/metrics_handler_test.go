package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/store"
)

func recordMetricRun(s *store.Store, name string, dur time.Duration, fail bool) {
	r := job.NewRun(name)
	time.Sleep(0)
	var err error
	if fail {
		err = fmt.Errorf("simulated failure")
	}
	r.Finish(err)
	// Override duration for determinism
	_ = dur
	s.Record(r)
}

func TestMetricsHandler_Empty(t *testing.T) {
	s := store.New(100)
	h := NewMetricsHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var metrics []JobMetrics
	if err := json.NewDecoder(rec.Body).Decode(&metrics); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(metrics) != 0 {
		t.Fatalf("expected empty metrics, got %d entries", len(metrics))
	}
}

func TestMetricsHandler_SuccessRate(t *testing.T) {
	s := store.New(100)

	for i := 0; i < 3; i++ {
		r := job.NewRun("backup")
		r.Finish(nil)
		s.Record(r)
	}
	r := job.NewRun("backup")
	r.Finish(fmt.Errorf("disk full"))
	s.Record(r)

	h := NewMetricsHandler(s)
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var metrics []JobMetrics
	if err := json.NewDecoder(rec.Body).Decode(&metrics); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("expected 1 job metric, got %d", len(metrics))
	}
	m := metrics[0]
	if m.JobName != "backup" {
		t.Errorf("expected job name 'backup', got %q", m.JobName)
	}
	if m.TotalRuns != 4 {
		t.Errorf("expected 4 total runs, got %d", m.TotalRuns)
	}
	if m.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", m.Failures)
	}
	if m.SuccessRate != 75.0 {
		t.Errorf("expected success rate 75.0, got %f", m.SuccessRate)
	}
	if m.LastRun == nil {
		t.Error("expected LastRun to be set")
	}
}
