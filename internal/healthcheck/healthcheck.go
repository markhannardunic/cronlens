// Package healthcheck evaluates whether cron jobs are running on schedule
// by comparing the last successful run time against an expected interval.
package healthcheck

import (
	"fmt"
	"sync"
	"time"

	"github.com/user/cronlens/internal/job"
)

// Rule defines the expected maximum interval between successful runs for a job.
type Rule struct {
	JobName  string
	Interval time.Duration
}

// Result holds the outcome of a health evaluation for a single job.
type Result struct {
	JobName   string
	Healthy   bool
	LastOK    time.Time
	Message   string
}

// Checker holds registered rules and evaluates job health against a run source.
type Checker struct {
	mu    sync.RWMutex
	rules map[string]Rule
}

// New returns an initialised Checker.
func New() *Checker {
	return &Checker{rules: make(map[string]Rule)}
}

// Register adds or replaces a health rule for the named job.
func (c *Checker) Register(r Rule) error {
	if r.JobName == "" {
		return fmt.Errorf("healthcheck: job name must not be empty")
	}
	if r.Interval <= 0 {
		return fmt.Errorf("healthcheck: interval must be positive")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules[r.JobName] = r
	return nil
}

// Rules returns a copy of all registered rules.
func (c *Checker) Rules() []Rule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Rule, 0, len(c.rules))
	for _, r := range c.rules {
		out = append(out, r)
	}
	return out
}

// Evaluate checks each registered rule against the provided runs and returns
// a Result per rule. now is injectable for deterministic testing.
func (c *Checker) Evaluate(runs []job.Run, now time.Time) []Result {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]Result, 0, len(c.rules))
	for _, rule := range c.rules {
		result := c.evaluate(rule, runs, now)
		results = append(results, result)
	}
	return results
}

func (c *Checker) evaluate(rule Rule, runs []job.Run, now time.Time) Result {
	var lastOK time.Time
	for _, r := range runs {
		if r.JobName == rule.JobName && r.Success && r.FinishedAt.After(lastOK) {
			lastOK = r.FinishedAt
		}
	}
	if lastOK.IsZero() {
		return Result{JobName: rule.JobName, Healthy: false, Message: "no successful run recorded"}
	}
	if now.Sub(lastOK) > rule.Interval {
		return Result{
			JobName: rule.JobName,
			Healthy: false,
			LastOK:  lastOK,
			Message: fmt.Sprintf("last success was %s ago (limit %s)", now.Sub(lastOK).Round(time.Second), rule.Interval),
		}
	}
	return Result{JobName: rule.JobName, Healthy: true, LastOK: lastOK, Message: "ok"}
}
