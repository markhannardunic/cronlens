package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronlens/internal/alert"
	"github.com/yourorg/cronlens/internal/api"
	"github.com/yourorg/cronlens/internal/job"
	"github.com/yourorg/cronlens/internal/store"
)

func recordAlertRun(s *store.Store, jobName string, success bool, dur time.Duration) {
	r := job.NewRun(jobName)
	r.Finish(success, "")
	r.Duration = dur
	s.Record(r)
}

func TestAlertHandler_NoAlerts(t *testing.T) {
	s := store.New(100)
	recordAlertRun(s, "backup", true, time.Second)

	e := alert.NewEvaluator([]alert.Rule{
		{JobName: "backup", ConsecutiveFailures: 2},
	})
	h := api.NewAlertHandler(s, e)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/alerts", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var alerts []map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&alerts)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestAlertHandler_FiresOnConsecutiveFailures(t *testing.T) {
	s := store.New(100)
	recordAlertRun(s, "backup", false, time.Second)
	recordAlertRun(s, "backup", false, time.Second)
	recordAlertRun(s, "backup", false, time.Second)

	e := alert.NewEvaluator([]alert.Rule{
		{JobName: "backup", ConsecutiveFailures: 3},
	})
	h := api.NewAlertHandler(s, e)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/alerts", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var alerts []map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&alerts)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0]["job_name"] != "backup" {
		t.Errorf("unexpected job_name: %v", alerts[0]["job_name"])
	}
}

func TestAlertHandler_EmptyStore(t *testing.T) {
	s := store.New(100)
	e := alert.NewEvaluator([]alert.Rule{{ConsecutiveFailures: 1}})
	h := api.NewAlertHandler(s, e)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/alerts", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var alerts []map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&alerts)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts for empty store, got %d", len(alerts))
	}
}
