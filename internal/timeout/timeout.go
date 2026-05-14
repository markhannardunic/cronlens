// Package timeout provides per-job execution timeout enforcement.
// A Watcher tracks registered timeout rules and reports any jobs whose
// last run exceeded the allowed duration.
package timeout

import (
	"fmt"
	"sync"
	"time"

	"github.com/cronlens/internal/job"
)

// Rule defines the maximum allowed duration for a named job.
type Rule struct {
	JobName string
	Max     time.Duration
}

// Violation is returned when a run exceeded its configured timeout.
type Violation struct {
	JobName  string
	Ran      time.Duration
	Max      time.Duration
	RunID    string
}

func (v Violation) Error() string {
	return fmt.Sprintf("job %q exceeded timeout: ran %s, max %s (run %s)",
		v.JobName, v.Ran.Round(time.Millisecond), v.Max.Round(time.Millisecond), v.RunID)
}

// Watcher holds timeout rules and evaluates runs against them.
type Watcher struct {
	mu    sync.RWMutex
	rules map[string]time.Duration
}

// New returns an initialised Watcher with no rules.
func New() *Watcher {
	return &Watcher{rules: make(map[string]time.Duration)}
}

// Register adds or replaces the timeout rule for jobName.
func (w *Watcher) Register(jobName string, max time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.rules[jobName] = max
}

// Rules returns a snapshot of all registered rules.
func (w *Watcher) Rules() []Rule {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]Rule, 0, len(w.rules))
	for name, max := range w.rules {
		out = append(out, Rule{JobName: name, Max: max})
	}
	return out
}

// Evaluate checks the provided runs against registered rules and returns
// any violations. Runs for jobs without a rule are ignored.
func (w *Watcher) Evaluate(runs []job.Run) []Violation {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var violations []Violation
	for _, r := range runs {
		max, ok := w.rules[r.JobName]
		if !ok || !r.Finished {
			continue
		}
		if r.Duration > max {
			violations = append(violations, Violation{
				JobName: r.JobName,
				Ran:     r.Duration,
				Max:     max,
				RunID:   r.ID,
			})
		}
	}
	return violations
}
