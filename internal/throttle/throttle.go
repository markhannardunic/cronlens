// Package throttle limits how frequently a job can be triggered
// by enforcing a minimum gap between consecutive runs.
package throttle

import (
	"errors"
	"sync"
	"time"
)

// Rule defines the minimum interval between runs for a job.
type Rule struct {
	JobName  string
	Interval time.Duration
}

// Limiter tracks the last run time per job and enforces throttle rules.
type Limiter struct {
	mu      sync.Mutex
	rules   map[string]time.Duration
	lastRun map[string]time.Time
}

// New returns a new Limiter.
func New() *Limiter {
	return &Limiter{
		rules:   make(map[string]time.Duration),
		lastRun: make(map[string]time.Time),
	}
}

// Register adds or replaces a throttle rule for the given job.
func (l *Limiter) Register(r Rule) error {
	if r.JobName == "" {
		return errors.New("throttle: job name must not be empty")
	}
	if r.Interval <= 0 {
		return errors.New("throttle: interval must be positive")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rules[r.JobName] = r.Interval
	return nil
}

// Allow reports whether the job is allowed to run now.
// Jobs with no registered rule are always allowed.
func (l *Limiter) Allow(jobName string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	interval, ok := l.rules[jobName]
	if !ok {
		return true
	}
	last, seen := l.lastRun[jobName]
	if !seen {
		return true
	}
	return time.Since(last) >= interval
}

// Record marks the job as having run at the current time.
func (l *Limiter) Record(jobName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lastRun[jobName] = time.Now()
}

// Rules returns a copy of all registered rules.
func (l *Limiter) Rules() []Rule {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Rule, 0, len(l.rules))
	for name, interval := range l.rules {
		out = append(out, Rule{JobName: name, Interval: interval})
	}
	return out
}
