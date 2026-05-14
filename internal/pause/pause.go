// Package pause provides the ability to temporarily pause and resume
// scheduled cron jobs without removing them from the scheduler.
package pause

import (
	"fmt"
	"sync"
	"time"
)

// Entry records when a job was paused and an optional reason.
type Entry struct {
	JobName  string    `json:"job_name"`
	PausedAt time.Time `json:"paused_at"`
	Reason   string    `json:"reason,omitempty"`
}

// Registry tracks which jobs are currently paused.
type Registry struct {
	mu      sync.RWMutex
	paused  map[string]Entry
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{paused: make(map[string]Entry)}
}

// Pause marks a job as paused. Calling Pause on an already-paused job
// is a no-op and returns an error so callers can surface the duplicate.
func (r *Registry) Pause(jobName, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.paused[jobName]; exists {
		return fmt.Errorf("job %q is already paused", jobName)
	}
	r.paused[jobName] = Entry{
		JobName:  jobName,
		PausedAt: time.Now().UTC(),
		Reason:   reason,
	}
	return nil
}

// Resume removes the paused state for a job. Returns an error if the
// job was not paused.
func (r *Registry) Resume(jobName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.paused[jobName]; !exists {
		return fmt.Errorf("job %q is not paused", jobName)
	}
	delete(r.paused, jobName)
	return nil
}

// IsPaused reports whether the named job is currently paused.
func (r *Registry) IsPaused(jobName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.paused[jobName]
	return ok
}

// List returns all currently paused entries.
func (r *Registry) List() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Entry, 0, len(r.paused))
	for _, e := range r.paused {
		out = append(out, e)
	}
	return out
}
