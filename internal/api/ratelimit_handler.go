package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronlens/internal/ratelimit"
)

// RateLimitHandler exposes HTTP endpoints for managing per-job rate-limit rules.
type RateLimitHandler struct {
	limiter *ratelimit.Limiter
}

// NewRateLimitHandler returns a handler backed by the given Limiter.
func NewRateLimitHandler(l *ratelimit.Limiter) *RateLimitHandler {
	return &RateLimitHandler{limiter: l}
}

func (h *RateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listRules(w, r)
	case http.MethodPost:
		h.setRule(w, r)
	case http.MethodDelete:
		h.resetRule(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type ruleResponse struct {
	Job      string `json:"job"`
	Interval string `json:"min_interval"`
}

func (h *RateLimitHandler) listRules(w http.ResponseWriter, _ *http.Request) {
	rules := h.limiter.Rules()
	out := make([]ruleResponse, 0, len(rules))
	for job, d := range rules {
		out = append(out, ruleResponse{Job: job, Interval: d.String()})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out) //nolint:errcheck
}

type setRuleRequest struct {
	Job      string `json:"job"`
	Interval string `json:"min_interval"`
}

func (h *RateLimitHandler) setRule(w http.ResponseWriter, r *http.Request) {
	var req setRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Job == "" || req.Interval == "" {
		http.Error(w, "job and min_interval are required", http.StatusBadRequest)
		return
	}
	d, err := time.ParseDuration(req.Interval)
	if err != nil {
		http.Error(w, "invalid duration: "+err.Error(), http.StatusBadRequest)
		return
	}
	h.limiter.SetMinInterval(req.Job, d)
	w.WriteHeader(http.StatusNoContent)
}

func (h *RateLimitHandler) resetRule(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job query parameter is required", http.StatusBadRequest)
		return
	}
	h.limiter.Reset(job)
	w.WriteHeader(http.StatusNoContent)
}
