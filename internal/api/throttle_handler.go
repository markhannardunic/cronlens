package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronlens/internal/throttle"
)

// ThrottleHandler exposes throttle rule management over HTTP.
type ThrottleHandler struct {
	limiter *throttle.Limiter
}

// NewThrottleHandler returns a ThrottleHandler backed by the given Limiter.
func NewThrottleHandler(l *throttle.Limiter) *ThrottleHandler {
	return &ThrottleHandler{limiter: l}
}

func (h *ThrottleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.register(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ThrottleHandler) list(w http.ResponseWriter, _ *http.Request) {
	type ruleJSON struct {
		JobName  string `json:"job_name"`
		Interval string `json:"interval"`
	}
	rules := h.limiter.Rules()
	out := make([]ruleJSON, 0, len(rules))
	for _, r := range rules {
		out = append(out, ruleJSON{JobName: r.JobName, Interval: r.Interval.String()})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *ThrottleHandler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		JobName  string `json:"job_name"`
		Interval string `json:"interval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	d, err := time.ParseDuration(body.Interval)
	if err != nil {
		http.Error(w, "invalid interval: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.limiter.Register(throttle.Rule{JobName: body.JobName, Interval: d}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
