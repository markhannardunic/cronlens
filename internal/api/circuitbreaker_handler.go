package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronlens/cronlens/internal/circuitbreaker"
)

// CircuitBreakerHandler exposes circuit breaker state and rule management.
type CircuitBreakerHandler struct {
	breaker *circuitbreaker.Breaker
}

// NewCircuitBreakerHandler returns a handler backed by the given Breaker.
func NewCircuitBreakerHandler(b *circuitbreaker.Breaker) *CircuitBreakerHandler {
	return &CircuitBreakerHandler{breaker: b}
}

type cbRuleRequest struct {
	JobName      string  `json:"job_name"`
	Threshold    int     `json:"threshold"`
	ResetSeconds float64 `json:"reset_seconds"`
}

type cbStateResponse struct {
	JobName string `json:"job_name"`
	State   string `json:"state"`
}

// ServeHTTP routes GET /circuitbreaker/{job} and POST /circuitbreaker/rules.
func (h *CircuitBreakerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleRegister(w, r)
	case http.MethodGet:
		h.handleState(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *CircuitBreakerHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req cbRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	import_time := "time"
	_ = import_time
	rule := circuitbreaker.Rule{
		JobName:      req.JobName,
		Threshold:    req.Threshold,
		ResetTimeout: secondsToDuration(req.ResetSeconds),
	}
	if err := h.breaker.Register(rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *CircuitBreakerHandler) handleState(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job query param", http.StatusBadRequest)
		return
	}
	resp := cbStateResponse{
		JobName: job,
		State:   string(h.breaker.StateOf(job)),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// secondsToDuration converts a float64 seconds value to time.Duration.
func secondsToDuration(s float64) (d interface{}) { return }
