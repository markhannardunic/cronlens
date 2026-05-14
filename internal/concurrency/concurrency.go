// Package concurrency tracks how many instances of a job are currently running
// and enforces an optional per-job concurrency limit.
package concurrency

import (
	"errors"
	"fmt"
	"sync"
)

// ErrLimitExceeded is returned when a job has reached its concurrency limit.
var ErrLimitExceeded = errors.New("concurrency limit exceeded")

// Tracker records in-flight job runs and optionally enforces per-job limits.
type Tracker struct {
	mu     sync.Mutex
	active map[string]int
	limits map[string]int
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{
		active: make(map[string]int),
		limits: make(map[string]int),
	}
}

// SetLimit configures the maximum number of concurrent runs allowed for a job.
// A limit of 0 means unlimited.
func (t *Tracker) SetLimit(jobName string, limit int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.limits[jobName] = limit
}

// Acquire marks a new in-flight run for jobName.
// It returns ErrLimitExceeded if the job's limit would be breached.
func (t *Tracker) Acquire(jobName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if limit, ok := t.limits[jobName]; ok && limit > 0 {
		if t.active[jobName] >= limit {
			return fmt.Errorf("%w: job %q limit %d", ErrLimitExceeded, jobName, limit)
		}
	}
	t.active[jobName]++
	return nil
}

// Release decrements the in-flight count for jobName.
func (t *Tracker) Release(jobName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active[jobName] > 0 {
		t.active[jobName]--
	}
}

// Active returns the current number of in-flight runs for jobName.
func (t *Tracker) Active(jobName string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.active[jobName]
}

// Snapshot returns a copy of the current active counts keyed by job name.
func (t *Tracker) Snapshot() map[string]int {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make(map[string]int, len(t.active))
	for k, v := range t.active {
		out[k] = v
	}
	return out
}
