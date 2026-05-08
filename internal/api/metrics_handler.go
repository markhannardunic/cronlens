package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronlens/internal/store"
)

// MetricsHandler serves aggregated metrics for all tracked cron jobs.
type MetricsHandler struct {
	store *store.Store
}

// NewMetricsHandler creates a new MetricsHandler backed by the given store.
func NewMetricsHandler(s *store.Store) *MetricsHandler {
	return &MetricsHandler{store: s}
}

// JobMetrics holds aggregated statistics for a single job.
type JobMetrics struct {
	JobName     string  `json:"job_name"`
	TotalRuns   int     `json:"total_runs"`
	Failures    int     `json:"failures"`
	SuccessRate float64 `json:"success_rate"`
	AvgDuration float64 `json:"avg_duration_ms"`
	LastRun     *time.Time `json:"last_run,omitempty"`
}

// ServeHTTP handles GET /api/metrics and returns per-job aggregated stats.
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	names := h.store.JobNames()
	result := make([]JobMetrics, 0, len(names))

	for _, name := range names {
		runs := h.store.Latest(name, 1000)
		if len(runs) == 0 {
			continue
		}

		var failures int
		var totalDuration float64
		for _, run := range runs {
			if run.Error != nil {
				failures++
			}
			totalDuration += float64(run.Duration.Milliseconds())
		}

		total := len(runs)
		successRate := 0.0
		if total > 0 {
			successRate = float64(total-failures) / float64(total) * 100
		}
		avgDuration := 0.0
		if total > 0 {
			avgDuration = totalDuration / float64(total)
		}

		lastStart := runs[0].StartedAt
		m := JobMetrics{
			JobName:     name,
			TotalRuns:   total,
			Failures:    failures,
			SuccessRate: successRate,
			AvgDuration: avgDuration,
			LastRun:     &lastStart,
		}
		result = append(result, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
