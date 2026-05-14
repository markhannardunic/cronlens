package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronlens/internal/webhook"
)

// webhookStore is the minimal store interface needed by WebhookHandler.
type webhookStore interface {
	JobNames() []string
}

// WebhookHandler exposes webhook configuration management over HTTP.
type WebhookHandler struct {
	dispatcher *webhook.Dispatcher
	configs    []webhook.Config
	store      webhookStore
}

// NewWebhookHandler creates a WebhookHandler.
func NewWebhookHandler(store webhookStore, configs []webhook.Config) *WebhookHandler {
	return &WebhookHandler{
		dispatcher: webhook.New(configs),
		configs:    configs,
		store:      store,
	}
}

// ServeHTTP routes GET /api/webhooks to listConfigs.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listConfigs(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type webhookConfigResponse struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

func (h *WebhookHandler) listConfigs(w http.ResponseWriter, r *http.Request) {
	out := make([]webhookConfigResponse, 0, len(h.configs))
	for _, c := range h.configs {
		out = append(out, webhookConfigResponse{URL: c.URL, Events: c.Events})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
