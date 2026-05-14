package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourorg/cronlens/internal/dependency"
)

// DependencyHandler exposes the dependency registry over HTTP.
type DependencyHandler struct {
	reg *dependency.Registry
}

// NewDependencyHandler returns a handler backed by reg.
func NewDependencyHandler(reg *dependency.Registry) *DependencyHandler {
	return &DependencyHandler{reg: reg}
}

// ServeHTTP routes GET /api/dependencies and POST /api/dependencies.
func (h *DependencyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleList(w, r)
	case http.MethodPost:
		h.handleRegister(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type depEdge struct {
	Job  string   `json:"job"`
	Deps []string `json:"deps"`
}

func (h *DependencyHandler) handleList(w http.ResponseWriter, _ *http.Request) {
	jobs := h.reg.Jobs()
	out := make([]depEdge, 0, len(jobs))
	for _, j := range jobs {
		deps, _ := h.reg.DepsOf(j)
		out = append(out, depEdge{Job: j, Deps: deps})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *DependencyHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Job  string   `json:"job"`
		Deps []string `json:"deps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Job) == "" {
		http.Error(w, "job name required", http.StatusBadRequest)
		return
	}
	if err := h.reg.Register(req.Job, req.Deps); err != nil {
		switch err {
		case dependency.ErrCycle:
			http.Error(w, "cycle detected", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}
