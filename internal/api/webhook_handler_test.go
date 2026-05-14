package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronlens/internal/api"
	"github.com/cronlens/internal/webhook"
)

type stubWebhookStore struct{}

func (s *stubWebhookStore) JobNames() []string { return []string{} }

func TestWebhookHandler_MethodNotAllowed(t *testing.T) {
	h := api.NewWebhookHandler(&stubWebhookStore{}, nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestWebhookHandler_EmptyConfigs(t *testing.T) {
	h := api.NewWebhookHandler(&stubWebhookStore{}, nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/webhooks", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestWebhookHandler_ListConfigs(t *testing.T) {
	cfgs := []webhook.Config{
		{URL: "https://example.com/hook1", Events: []string{"failure"}},
		{URL: "https://example.com/hook2", Events: []string{"all"}},
	}
	h := api.NewWebhookHandler(&stubWebhookStore{}, cfgs)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/webhooks", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var result []struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(result))
	}
	if result[0].URL != "https://example.com/hook1" {
		t.Errorf("unexpected URL: %q", result[0].URL)
	}
	if len(result[1].Events) != 1 || result[1].Events[0] != "all" {
		t.Errorf("unexpected events for second config: %v", result[1].Events)
	}
}

func TestWebhookHandler_ContentTypeJSON(t *testing.T) {
	h := api.NewWebhookHandler(&stubWebhookStore{}, []webhook.Config{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/webhooks", nil)
	h.ServeHTTP(rr, req)
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
