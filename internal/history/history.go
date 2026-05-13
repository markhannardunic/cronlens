// Package history provides aggregated run history statistics per job.
package history

import (
	"time"

	"github.com/user/cronlens/internal/job"
	"github.com/user/cronlens/internal/store"
)

// Stats holds aggregated statistics for a single job.
type Stats struct {
	JobName        string        `json:"job_name"`
	TotalRuns      int           `json:"total_runs"`
	SuccessCount   int           `json:"success_count"`
	FailureCount   int           `json:"failure_count"`
	SuccessRate    float64       `json:"success_rate"`
	AvgDuration    time.Duration `json:"avg_duration_ns"`
	MaxDuration    time.Duration `json:"max_duration_ns"`
	LastRun        *job.Run      `json:"last_run,omitempty"`
}

// Aggregator computes history stats from a store.
type Aggregator struct {
	store *store.Store
}

// New returns a new Aggregator backed by the given store.
func New(s *store.Store) *Aggregator {
	return &Aggregator{store: s}
}

// StatsFor returns aggregated Stats for the named job using all recorded runs.
func (a *Aggregator) StatsFor(jobName string) Stats {
	runs := a.store.Since(jobName, time.Time{})
	return compute(jobName, runs)
}

// AllStats returns aggregated Stats for every known job.
func (a *Aggregator) AllStats() []Stats {
	names := a.store.JobNames()
	out := make([]Stats, 0, len(names))
	for _, name := range names {
		out = append(out, a.StatsFor(name))
	}
	return out
}

func compute(jobName string, runs []job.Run) Stats {
	s := Stats{JobName: jobName, TotalRuns: len(runs)}
	if len(runs) == 0 {
		return s
	}
	var total time.Duration
	for i := range runs {
		r := &runs[i]
		if r.Success {
			s.SuccessCount++
		} else {
			s.FailureCount++
		}
		if r.Duration > s.MaxDuration {
			s.MaxDuration = r.Duration
		}
		total += r.Duration
		if s.LastRun == nil || r.StartedAt.After(s.LastRun.StartedAt) {
			copy := *r
			s.LastRun = &copy
		}
	}
	s.AvgDuration = total / time.Duration(len(runs))
	if s.TotalRuns > 0 {
		s.SuccessRate = float64(s.SuccessCount) / float64(s.TotalRuns) * 100.0
	}
	return s
}
