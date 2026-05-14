package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronlens/internal/store"
	"github.com/cronlens/internal/timeout"
)

// TimeoutHandler exposes timeout rules and current violations over HTTP.
type TimeoutHandler struct {
	watcher *timeout.Watcher
	store   *store.Store
}

// NewTimeoutHandler creates a TimeoutHandler backed by the given watcher and store.
func NewTimeoutHandler(w *timeout.Watcher, s *store.Store) *TimeoutHandler {
	return &TimeoutHandler{watcher: w, store: s}
}

func (h *TimeoutHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listViolations(rw, r)
	case http.MethodPost:
		h.registerRule(rw, r)
	default:
		http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type timeoutRuleRequest struct {
	JobName string        `json:"job_name"`
	Max     time.Duration `json:"max_ns"` // nanoseconds for JSON simplicity
}

type timeoutViolationResponse struct {
	JobName string        `json:"job_name"`
	RunID   string        `json:"run_id"`
	RanNs   time.Duration `json:"ran_ns"`
	MaxNs   time.Duration `json:"max_ns"`
	Message string        `json:"message"`
}

func (h *TimeoutHandler) listViolations(rw http.ResponseWriter, _ *http.Request) {
	runs := h.store.Latest(1000)
	violations := h.watcher.Evaluate(runs)

	out := make([]timeoutViolationResponse, 0, len(violations))
	for _, v := range violations {
		out = append(out, timeoutViolationResponse{
			JobName: v.JobName,
			RunID:   v.RunID,
			RanNs:   v.Ran,
			MaxNs:   v.Max,
			Message: v.Error(),
		})
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(out)
}

func (h *TimeoutHandler) registerRule(rw http.ResponseWriter, r *http.Request) {
	var req timeoutRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.JobName == "" || req.Max <= 0 {
		http.Error(rw, "job_name and positive max_ns required", http.StatusBadRequest)
		return
	}
	h.watcher.Register(req.JobName, req.Max)
	rw.WriteHeader(http.StatusNoContent)
}
