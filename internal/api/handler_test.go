package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronlens/internal/api"
	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/store"
)

func makeStore(t *testing.T) *store.Store {
	t.Helper()
	return store.New(100)
}

func recordRun(s *store.Store, name string, success bool, errMsg string) {
	r := job.NewRun(name)
	r.Finish(success, errMsg)
	s.Record(r)
}

func TestHandleJobs_Empty(t *testing.T) {
	s := makeStore(t)
	router := api.NewRouter(s)

	req := httptest.NewRequest(http.MethodGet, "/api/jobs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var summaries []api.JobSummary
	if err := json.NewDecoder(rec.Body).Decode(&summaries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestHandleJobs_WithRuns(t *testing.T) {
	s := makeStore(t)
	recordRun(s, "backup", true, "")
	recordRun(s, "cleanup", false, "disk full")

	router := api.NewRouter(s)
	req := httptest.NewRequest(http.MethodGet, "/api/jobs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var summaries []api.JobSummary
	if err := json.NewDecoder(rec.Body).Decode(&summaries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
}

func TestHandleRuns_Since(t *testing.T) {
	s := makeStore(t)
	recordRun(s, "nightly", true, "")

	router := api.NewRouter(s)
	req := httptest.NewRequest(http.MethodGet, "/api/runs?since=1h", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var runs []job.Run
	if err := json.NewDecoder(rec.Body).Decode(&runs); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(runs) == 0 {
		t.Error("expected at least one run")
	}
	_ = time.Now() // ensure time import used
}
