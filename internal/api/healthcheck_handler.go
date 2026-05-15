package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronlens/internal/healthcheck"
	"github.com/user/cronlens/internal/store"
)

// HealthCheckHandler serves health status for registered job rules.
type HealthCheckHandler struct {
	checker *healthcheck.Checker
	store   *store.Store
}

// NewHealthCheckHandler returns a handler wired to the given checker and store.
func NewHealthCheckHandler(c *healthcheck.Checker, s *store.Store) *HealthCheckHandler {
	return &HealthCheckHandler{checker: c, store: s}
}

func (h *HealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleStatus(w, r)
	case http.MethodPost:
		h.handleRegister(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HealthCheckHandler) handleStatus(w http.ResponseWriter, _ *http.Request) {
	runs := h.store.Latest(1000)
	results := h.checker.Evaluate(runs, time.Now())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

func (h *HealthCheckHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		JobName  string  `json:"job_name"`
		Interval float64 `json:"interval_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	rule := healthcheck.Rule{
		JobName:  body.JobName,
		Interval: time.Duration(body.Interval) * time.Second,
	}
	if err := h.checker.Register(rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
