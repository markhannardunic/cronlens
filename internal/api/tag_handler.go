package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourorg/cronlens/internal/tag"
)

// TagHandler exposes tag management over HTTP.
type TagHandler struct {
	reg *tag.Registry
}

// NewTagHandler creates a TagHandler backed by the given Registry.
func NewTagHandler(reg *tag.Registry) *TagHandler {
	return &TagHandler{reg: reg}
}

// ServeHTTP routes GET /api/tags and PUT /api/tags/{job}.
func (h *TagHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleList(w, r)
	case http.MethodPut:
		h.handleSet(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type tagEntry struct {
	Job  string   `json:"job"`
	Tags []string `json:"tags"`
}

func (h *TagHandler) handleList(w http.ResponseWriter, _ *http.Request) {
	names := h.reg.JobNames()
	out := make([]tagEntry, 0, len(names))
	for _, name := range names {
		out = append(out, tagEntry{Job: name, Tags: h.reg.Get(name)})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *TagHandler) handleSet(w http.ResponseWriter, r *http.Request) {
	// Expect path: /api/tags/{jobName}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "job name required", http.StatusBadRequest)
		return
	}
	jobName := parts[2]

	var payload struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	h.reg.Set(jobName, payload.Tags)
	w.WriteHeader(http.StatusNoContent)
}
