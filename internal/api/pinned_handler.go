package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cronlens/internal/pinned"
)

// PinnedHandler exposes the pinned job registry over HTTP.
type PinnedHandler struct {
	reg *pinned.Registry
}

// NewPinnedHandler returns a new PinnedHandler backed by reg.
func NewPinnedHandler(reg *pinned.Registry) *PinnedHandler {
	return &PinnedHandler{reg: reg}
}

func (h *PinnedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route: /api/pinned[/<job>]
	job := strings.TrimPrefix(r.URL.Path, "/api/pinned")
	job = strings.TrimPrefix(job, "/")

	switch {
	case job == "" && r.Method == http.MethodGet:
		h.handleList(w, r)
	case job != "" && r.Method == http.MethodPut:
		h.handlePin(w, r, job)
	case job != "" && r.Method == http.MethodDelete:
		h.handleUnpin(w, r, job)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PinnedHandler) handleList(w http.ResponseWriter, _ *http.Request) {
	list := h.reg.List()
	if list == nil {
		list = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{"pinned": list})
}

func (h *PinnedHandler) handlePin(w http.ResponseWriter, _ *http.Request, job string) {
	if err := h.reg.Pin(job); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PinnedHandler) handleUnpin(w http.ResponseWriter, _ *http.Request, job string) {
	if err := h.reg.Unpin(job); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
