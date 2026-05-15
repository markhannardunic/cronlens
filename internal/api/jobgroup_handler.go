package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"cronlens/internal/jobgroup"
)

// JobGroupHandler exposes CRUD operations for job groups over HTTP.
type JobGroupHandler struct {
	reg *jobgroup.Registry
}

// NewJobGroupHandler returns a JobGroupHandler backed by reg.
func NewJobGroupHandler(reg *jobgroup.Registry) *JobGroupHandler {
	return &JobGroupHandler{reg: reg}
}

// ServeHTTP routes requests:
//
//	GET    /jobgroups                  → list all groups
//	GET    /jobgroups/{group}           → list jobs in group
//	POST   /jobgroups/{group}/{job}     → add job to group
//	DELETE /jobgroups/{group}/{job}     → remove job from group
func (h *JobGroupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// parts[0] == "jobgroups"
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		h.listGroups(w)
	case len(parts) == 2 && r.Method == http.MethodGet:
		h.listJobs(w, parts[1])
	case len(parts) == 3 && r.Method == http.MethodPost:
		h.addJob(w, parts[1], parts[2])
	case len(parts) == 3 && r.Method == http.MethodDelete:
		h.removeJob(w, parts[1], parts[2])
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *JobGroupHandler) listGroups(w http.ResponseWriter) {
	groups := h.reg.Groups()
	if groups == nil {
		groups = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"groups": groups})
}

func (h *JobGroupHandler) listJobs(w http.ResponseWriter, group string) {
	jobs, err := h.reg.JobsIn(group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"jobs": jobs})
}

func (h *JobGroupHandler) addJob(w http.ResponseWriter, group, job string) {
	if err := h.reg.AddJob(group, job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *JobGroupHandler) removeJob(w http.ResponseWriter, group, job string) {
	if err := h.reg.RemoveJob(group, job); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
