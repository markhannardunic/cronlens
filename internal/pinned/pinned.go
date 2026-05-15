// Package pinned allows jobs to be pinned, marking them as critical so that
// dashboards and alerts can surface them with higher priority.
package pinned

import (
	"errors"
	"sync"
)

// Registry tracks which jobs are pinned.
type Registry struct {
	mu     sync.RWMutex
	pinned map[string]struct{}
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{pinned: make(map[string]struct{})}
}

// Pin marks a job as pinned. Returns an error if already pinned.
func (r *Registry) Pin(job string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.pinned[job]; ok {
		return errors.New("job already pinned: " + job)
	}
	r.pinned[job] = struct{}{}
	return nil
}

// Unpin removes a job from the pinned set. Returns an error if not pinned.
func (r *Registry) Unpin(job string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.pinned[job]; !ok {
		return errors.New("job not pinned: " + job)
	}
	delete(r.pinned, job)
	return nil
}

// IsPinned reports whether a job is currently pinned.
func (r *Registry) IsPinned(job string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.pinned[job]
	return ok
}

// List returns all currently pinned job names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.pinned))
	for job := range r.pinned {
		out = append(out, job)
	}
	return out
}
