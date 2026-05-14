package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronlens/internal/retention"
)

// RetentionHandler exposes a manual prune endpoint and reports the current
// retention policy configuration.
type RetentionHandler struct {
	pruner *retention.Pruner
	maxAge time.Duration
}

// NewRetentionHandler creates a RetentionHandler backed by the given Pruner.
func NewRetentionHandler(p *retention.Pruner, maxAge time.Duration) *RetentionHandler {
	return &RetentionHandler{pruner: p, maxAge: maxAge}
}

func (h *RetentionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handlePrune(w, r)
	case http.MethodGet:
		h.handlePolicy(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type pruneResponse struct {
	Pruned int    `json:"pruned"`
	Message string `json:"message"`
}

type policyResponse struct {
	MaxAgeSec int64 `json:"max_age_seconds"`
}

func (h *RetentionHandler) handlePrune(w http.ResponseWriter, _ *http.Request) {
	n := h.pruner.Prune()
	resp := pruneResponse{
		Pruned:  n,
		Message: "pruned successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *RetentionHandler) handlePolicy(w http.ResponseWriter, _ *http.Request) {
	resp := policyResponse{
		MaxAgeSec: int64(h.maxAge.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
