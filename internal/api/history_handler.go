package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/cronlens/internal/history"
)

// HistoryHandler serves aggregated job history statistics.
type HistoryHandler struct {
	agg *history.Aggregator
}

// NewHistoryHandler creates a HistoryHandler backed by the given Aggregator.
func NewHistoryHandler(agg *history.Aggregator) *HistoryHandler {
	return &HistoryHandler{agg: agg}
}

// ServeHTTP routes GET /api/history and GET /api/history/{job}.
func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract optional job name from path: /api/history/{job}
	path := strings.TrimPrefix(r.URL.Path, "/api/history")
	path = strings.Trim(path, "/")

	if path != "" {
		stats := h.agg.StatsFor(path)
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
		return
	}

	all := h.agg.AllStats()
	if all == nil {
		all = []history.Stats{}
	}
	if err := json.NewEncoder(w).Encode(all); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}
