package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronlens/internal/api"
	"github.com/user/cronlens/internal/history"
	"github.com/user/cronlens/internal/job"
	"github.com/user/cronlens/internal/store"
)

func recordHistoryRun(s *store.Store, name string, success bool, dur time.Duration) {
	r := job.NewRun(name)
	r.Finish(success, "", dur)
	s.Record(r)
}

func TestHistoryHandler_Empty(t *testing.T) {
	s := store.New(100)
	agg := history.New(s)
	h := api.NewHistoryHandler(agg)

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []history.Stats
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestHistoryHandler_AllStats(t *testing.T) {
	s := store.New(100)
	recordHistoryRun(s, "backup", true, 5*time.Second)
	recordHistoryRun(s, "backup", false, 2*time.Second)
	recordHistoryRun(s, "sync", true, 1*time.Second)

	agg := history.New(s)
	h := api.NewHistoryHandler(agg)

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var result []history.Stats
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 job stats, got %d", len(result))
	}
}

func TestHistoryHandler_SingleJob(t *testing.T) {
	s := store.New(100)
	recordHistoryRun(s, "backup", true, 3*time.Second)
	recordHistoryRun(s, "backup", true, 7*time.Second)

	agg := history.New(s)
	h := api.NewHistoryHandler(agg)

	req := httptest.NewRequest(http.MethodGet, "/api/history/backup", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var stats history.Stats
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if stats.TotalRuns != 2 {
		t.Errorf("expected 2 runs, got %d", stats.TotalRuns)
	}
	if stats.SuccessRate != 100.0 {
		t.Errorf("expected 100%% success rate, got %f", stats.SuccessRate)
	}
	if stats.MaxDuration != 7*time.Second {
		t.Errorf("expected max 7s, got %v", stats.MaxDuration)
	}
}
