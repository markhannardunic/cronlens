// Package ratelimit provides per-job execution rate limiting to prevent
// runaway cron jobs from overwhelming downstream systems.
package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// Limiter enforces a minimum interval between successive runs of the same job.
type Limiter struct {
	mu       sync.Mutex
	lastRun  map[string]time.Time
	interval map[string]time.Duration
}

// New returns a new Limiter with no rules configured.
func New() *Limiter {
	return &Limiter{
		lastRun:  make(map[string]time.Time),
		interval: make(map[string]time.Duration),
	}
}

// SetMinInterval registers a minimum interval between runs for the given job.
// Subsequent calls for the same job overwrite the previous interval.
func (l *Limiter) SetMinInterval(jobName string, d time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.interval[jobName] = d
}

// Allow reports whether the job is allowed to run now.
// If no interval is configured for the job, Allow always returns true.
// When allowed, the internal last-run timestamp is updated atomically.
func (l *Limiter) Allow(jobName string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	interval, ok := l.interval[jobName]
	if !ok {
		l.lastRun[jobName] = time.Now()
		return true, nil
	}

	last, seen := l.lastRun[jobName]
	if seen {
		elapsed := time.Since(last)
		if elapsed < interval {
			return false, fmt.Errorf(
				"job %q rate-limited: %v remaining before next allowed run",
				jobName, interval-elapsed,
			)
		}
	}

	l.lastRun[jobName] = time.Now()
	return true, nil
}

// Reset clears the last-run timestamp for a job, allowing it to run immediately
// regardless of the configured interval. Useful for manual retries.
func (l *Limiter) Reset(jobName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.lastRun, jobName)
}

// Rules returns a snapshot of all configured job intervals.
func (l *Limiter) Rules() map[string]time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[string]time.Duration, len(l.interval))
	for k, v := range l.interval {
		out[k] = v
	}
	return out
}
