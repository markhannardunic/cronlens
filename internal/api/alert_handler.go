package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/cronlens/internal/alert"
	"github.com/yourorg/cronlens/internal/store"
)

// AlertHandler serves evaluated alerts for all tracked jobs.
type AlertHandler struct {
	store     *store.Store
	evaluator *alert.Evaluator
}

// NewAlertHandler creates an AlertHandler.
func NewAlertHandler(s *store.Store, e *alert.Evaluator) *AlertHandler {
	return &AlertHandler{store: s, evaluator: e}
}

type alertResponse struct {
	JobName string    `json:"job_name"`
	Reason  string    `json:"reason"`
	FiredAt time.Time `json:"fired_at"`
}

// ServeHTTP evaluates rules against stored history and returns active alerts.
func (h *AlertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	names := h.store.JobNames()
	var responses []alertResponse

	for _, name := range names {
		history := h.store.Latest(name, 50)
		// Latest returns newest-first; reverse for evaluator (oldest→newest).
		for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
			history[i], history[j] = history[j], history[i]
		}
		fired := h.evaluator.Evaluate(history)
		for _, a := range fired {
			responses = append(responses, alertResponse{
				JobName: a.JobName,
				Reason:  a.Reason,
				FiredAt: a.FiredAt,
			})
		}
	}

	if responses == nil {
		responses = []alertResponse{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
