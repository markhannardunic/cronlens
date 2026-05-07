package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronlens/internal/store"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	store *store.Store
}

// NewHandler creates a new Handler with the given store.
func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

// JobSummary is the JSON response shape for a single job.
type JobSummary struct {
	Name      string     `json:"name"`
	LastRanAt *time.Time `json:"last_ran_at,omitempty"`
	LastStatus string    `json:"last_status,omitempty"`
	LastError  string    `json:"last_error,omitempty"`
}

// HandleJobs returns a summary of all known jobs.
func (h *Handler) HandleJobs(w http.ResponseWriter, r *http.Request) {
	names := h.store.JobNames()
	summaries := make([]JobSummary, 0, len(names))

	for _, name := range names {
		run := h.store.Latest(name)
		s := JobSummary{Name: name}
		if run != nil {
			s.LastRanAt = &run.StartedAt
			if run.Success {
				s.LastStatus = "success"
			} else {
				s.LastStatus = "failure"
				s.LastError = run.Error
			}
		}
		summaries = append(summaries, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// HandleRuns returns recent runs for all jobs since a given duration.
func (h *Handler) HandleRuns(w http.ResponseWriter, r *http.Request) {
	sinceStr := r.URL.Query().Get("since")
	dur := 24 * time.Hour
	if sinceStr != "" {
		if d, err := time.ParseDuration(sinceStr); err == nil {
			dur = d
		}
	}

	runs := h.store.Since(time.Now().Add(-dur))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}
