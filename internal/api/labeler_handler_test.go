package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronlens/internal/api"
	"github.com/cronlens/internal/labeler"
)

func newLabelerHandler() http.Handler {
	return api.NewLabelerHandler(labeler.New())
}

func TestLabelerHandler_SetAndGet(t *testing.T) {
	h := newLabelerHandler()

	body := `{"key":"env","value":"prod"}`
	req := httptest.NewRequest(http.MethodPost, "/labels/backup", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/labels/backup", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	var lbls map[string]string
	json.NewDecoder(w2.Body).Decode(&lbls)
	if lbls["env"] != "prod" {
		t.Errorf("expected prod, got %s", lbls["env"])
	}
}

func TestLabelerHandler_Delete(t *testing.T) {
	reg := labeler.New()
	_ = reg.Set("job", "team", "ops")
	h := api.NewLabelerHandler(reg)

	req := httptest.NewRequest(http.MethodDelete, "/labels/job/team", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if lbls := reg.Get("job"); lbls["team"] != "" {
		t.Error("expected label to be deleted")
	}
}

func TestLabelerHandler_JobsWithLabel(t *testing.T) {
	reg := labeler.New()
	_ = reg.Set("job-a", "env", "prod")
	_ = reg.Set("job-b", "env", "prod")
	h := api.NewLabelerHandler(reg)

	req := httptest.NewRequest(http.MethodGet, "/labels?key=env&value=prod", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	var resp map[string][]string
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp["jobs"]) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(resp["jobs"]))
	}
}

func TestLabelerHandler_InvalidJSON(t *testing.T) {
	h := newLabelerHandler()
	req := httptest.NewRequest(http.MethodPost, "/labels/job", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLabelerHandler_MethodNotAllowed(t *testing.T) {
	h := newLabelerHandler()
	req := httptest.NewRequest(http.MethodPut, "/labels/job", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
