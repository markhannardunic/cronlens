// Package jobstatus tracks the current operational status of named jobs,
// allowing callers to mark a job as enabled or disabled and query that state.
package jobstatus

import (
	"errors"
	"sync"
)

// Status represents the operational state of a job.
type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
)

// Entry holds the status and an optional reason for the current state.
type Entry struct {
	Job    string `json:"job"`
	Status Status `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// Registry manages per-job operational statuses.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{entries: make(map[string]Entry)}
}

// Set records the status for a job. An empty job name returns an error.
func (r *Registry) Set(job string, s Status, reason string) error {
	if job == "" {
		return errors.New("jobstatus: job name must not be empty")
	}
	if s != StatusEnabled && s != StatusDisabled {
		return errors.New("jobstatus: status must be 'enabled' or 'disabled'")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[job] = Entry{Job: job, Status: s, Reason: reason}
	return nil
}

// Get returns the Entry for a job. If no entry exists, the job is treated as enabled.
func (r *Registry) Get(job string) Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if e, ok := r.entries[job]; ok {
		return e
	}
	return Entry{Job: job, Status: StatusEnabled}
}

// IsEnabled reports whether the job is currently enabled.
func (r *Registry) IsEnabled(job string) bool {
	return r.Get(job).Status == StatusEnabled
}

// All returns a snapshot of all explicitly recorded entries.
func (r *Registry) All() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Entry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, e)
	}
	return out
}

// Delete removes a job's status record, reverting it to the default enabled state.
func (r *Registry) Delete(job string) error {
	if job == "" {
		return errors.New("jobstatus: job name must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, job)
	return nil
}
