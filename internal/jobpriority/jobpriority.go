// Package jobpriority assigns and manages priority levels for cron jobs,
// allowing the scheduler and dashboard to surface high-priority job failures
// more prominently.
package jobpriority

import (
	"errors"
	"fmt"
	"sync"
)

// Level represents the priority of a job.
type Level int

const (
	Low    Level = 1
	Normal Level = 5
	High   Level = 10
	Critical Level = 20
)

func (l Level) String() string {
	switch l {
	case Low:
		return "low"
	case Normal:
		return "normal"
	case High:
		return "high"
	case Critical:
		return "critical"
	default:
		return fmt.Sprintf("unknown(%d)", int(l))
	}
}

// Registry stores priority levels keyed by job name.
type Registry struct {
	mu       sync.RWMutex
	priority map[string]Level
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{priority: make(map[string]Level)}
}

// Set assigns a priority level to a job. Returns an error if jobName is empty
// or the level is not one of the defined constants.
func (r *Registry) Set(jobName string, level Level) error {
	if jobName == "" {
		return errors.New("jobpriority: job name must not be empty")
	}
	if level != Low && level != Normal && level != High && level != Critical {
		return fmt.Errorf("jobpriority: unknown priority level %d", int(level))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.priority[jobName] = level
	return nil
}

// Get returns the priority level for a job. If no priority has been set,
// Normal is returned as the default.
func (r *Registry) Get(jobName string) Level {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if l, ok := r.priority[jobName]; ok {
		return l
	}
	return Normal
}

// Delete removes the priority entry for a job.
func (r *Registry) Delete(jobName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.priority, jobName)
}

// All returns a snapshot of all explicitly set job priorities.
func (r *Registry) All() map[string]Level {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]Level, len(r.priority))
	for k, v := range r.priority {
		out[k] = v
	}
	return out
}
