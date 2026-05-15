package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronlens/internal/snapshot"
)

// SnapshotHandler serves the current and last captured snapshot of all jobs.
type SnapshotHandler struct {
	capturer *snapshot.Capturer
}

// NewSnapshotHandler returns an http.Handler backed by the given Capturer.
func NewSnapshotHandler(c *snapshot.Capturer) *SnapshotHandler {
	return &SnapshotHandler{capturer: c}
}

func (h *SnapshotHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var snap snapshot.Snapshot

	switch r.URL.Query().Get("mode") {
	case "last":
		var ok bool
		snap, ok = h.capturer.Last()
		if !ok {
			http.Error(w, "no snapshot available", http.StatusNotFound)
			return
		}
	default:
		snap = h.capturer.Capture()
	}

	type response struct {
		CapturedAt time.Time                        `json:"captured_at"`
		Jobs       map[string]snapshot.JobSummary   `json:"jobs"`
	}

	resp := response{
		CapturedAt: snap.CapturedAt,
		Jobs:       snap.Jobs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
