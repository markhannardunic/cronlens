// Package snapshot captures a point-in-time summary of all job states,
// suitable for persistence, debugging, or diffing across time.
package snapshot

import (
	"sync"
	"time"

	"github.com/cronlens/internal/job"
)

// JobSummary holds a condensed view of a single job's recent state.
type JobSummary struct {
	Name       string        `json:"name"`
	LastStatus string        `json:"last_status"` // "success", "failure", "unknown"
	LastRun    time.Time     `json:"last_run"`
	Duration   time.Duration `json:"duration_ns"`
	RunCount   int           `json:"run_count"`
}

// Snapshot is an immutable point-in-time capture of all job summaries.
type Snapshot struct {
	CapturedAt time.Time              `json:"captured_at"`
	Jobs       map[string]JobSummary  `json:"jobs"`
}

// Store is the interface snapshot depends on to read run history.
type Store interface {
	JobNames() []string
	Latest(name string) (job.Run, bool)
	Since(name string, t time.Time) []job.Run
}

// Capturer takes snapshots of the current store state.
type Capturer struct {
	mu    sync.Mutex
	store Store
	last  *Snapshot
}

// New creates a new Capturer backed by the given store.
func New(s Store) *Capturer {
	return &Capturer{store: s}
}

// Capture reads all jobs from the store and returns a new Snapshot.
func (c *Capturer) Capture() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	snap := Snapshot{
		CapturedAt: time.Now(),
		Jobs:       make(map[string]JobSummary),
	}

	for _, name := range c.store.JobNames() {
		latest, ok := c.store.Latest(name)
		summary := JobSummary{Name: name, LastStatus: "unknown"}
		if ok {
			summary.LastRun = latest.StartedAt
			summary.Duration = latest.Duration()
			if latest.Err == nil {
				summary.LastStatus = "success"
			} else {
				summary.LastStatus = "failure"
			}
		}
		allRuns := c.store.Since(name, time.Time{})
		summary.RunCount = len(allRuns)
		snap.Jobs[name] = summary
	}

	c.last = &snap
	return snap
}

// Last returns the most recently captured snapshot, or false if none exists.
func (c *Capturer) Last() (Snapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.last == nil {
		return Snapshot{}, false
	}
	return *c.last, true
}
