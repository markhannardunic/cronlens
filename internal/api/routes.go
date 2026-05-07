package api

import (
	"net/http"

	"github.com/cronlens/internal/store"
)

// NewRouter wires up all API routes and returns an http.Handler.
func NewRouter(s *store.Store) http.Handler {
	h := NewHandler(s)
	mux := http.NewServeMux()

	mux.HandleFunc("/api/jobs", h.HandleJobs)
	mux.HandleFunc("/api/runs", h.HandleRuns)

	return mux
}
