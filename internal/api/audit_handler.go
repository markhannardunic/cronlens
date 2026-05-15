package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronlens/cronlens/internal/audit"
)

// AuditHandler exposes the audit log over HTTP.
type AuditHandler struct {
	log *audit.Log
}

// NewAuditHandler returns a new AuditHandler backed by the given Log.
func NewAuditHandler(l *audit.Log) *AuditHandler {
	return &AuditHandler{log: l}
}

// ServeHTTP routes GET /audit and GET /audit?target=<name>.
func (h *AuditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var entries []audit.Entry
	if target := r.URL.Query().Get("target"); target != "" {
		entries = h.log.ForTarget(target)
	} else {
		entries = h.log.All()
	}

	if entries == nil {
		entries = []audit.Entry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries) //nolint:errcheck
}
