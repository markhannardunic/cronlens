package dashboard_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cronlens/internal/dashboard"
	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/store"
)

func makeStore(t *testing.T) *store.Store {
	t.Helper()
	return store.New(100)
}

func recordRun(s *store.Store, name string, dur time.Duration, errMsg string) {
	r := job.NewRun(name)
	time.Sleep(dur)
	if errMsg != "" {
		r.Finish(fmt.Errorf("%s", errMsg))
	} else {
		r.Finish(nil)
	}
	s.Record(r)
}

func TestDashboard_Empty(t *testing.T) {
	s := makeStore(t)
	h, err := dashboard.NewHandler(s)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "No jobs recorded yet") {
		t.Errorf("expected empty message in body")
	}
}

func TestDashboard_WithRuns(t *testing.T) {
	s := makeStore(t)

	r := job.NewRun("backup")
	r.Finish(nil)
	s.Record(r)

	h, err := dashboard.NewHandler(s)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "backup") {
		t.Errorf("expected job name in body")
	}
	if !strings.Contains(body, "class=\"ok\"") && !strings.Contains(body, `class="ok"`) {
		t.Errorf("expected ok status in body")
	}
}
