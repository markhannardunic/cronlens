package api

import (
	"encoding/json"
	"net/http"

	"cronlens/internal/scheduler"
)

// ScheduleHandler serves information about registered scheduled jobs.
type ScheduleHandler struct {
	sched *scheduler.Scheduler
}

// NewScheduleHandler creates a handler backed by the given scheduler.
func NewScheduleHandler(s *scheduler.Scheduler) *ScheduleHandler {
	return &ScheduleHandler{sched: s}
}

// scheduleEntry is the JSON representation of a scheduled job.
type scheduleEntry struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
}

// ServeHTTP returns the list of registered scheduled jobs as JSON.
func (h *ScheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	entries := h.sched.Entries()
	out := make([]scheduleEntry, len(entries))
	for i, e := range entries {
		out[i] = scheduleEntry{Name: e.Name, Schedule: e.Schedule}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
