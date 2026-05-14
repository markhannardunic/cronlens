// Package dependency tracks inter-job dependencies and surfaces
// blocking relationships in the cronlens dashboard.
package dependency

import (
	"errors"
	"sync"
)

// ErrCycle is returned when registering a dependency would create a cycle.
var ErrCycle = errors.New("dependency: cycle detected")

// ErrUnknownJob is returned when referencing a job that has not been registered.
var ErrUnknownJob = errors.New("dependency: unknown job")

// Registry holds directed dependency edges between named cron jobs.
// An edge A → B means job A must complete before job B starts.
type Registry struct {
	mu    sync.RWMutex
	edges map[string][]string // job → jobs it depends on
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{edges: make(map[string][]string)}
}

// Register declares that jobName depends on each of deps.
// All names are registered implicitly. Returns ErrCycle if the
// new edges would introduce a cycle.
func (r *Registry) Register(jobName string, deps []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure all nodes exist.
	if _, ok := r.edges[jobName]; !ok {
		r.edges[jobName] = nil
	}
	for _, d := range deps {
		if _, ok := r.edges[d]; !ok {
			r.edges[d] = nil
		}
	}

	// Tentatively add edges, then check for cycles.
	original := make([]string, len(r.edges[jobName]))
	copy(original, r.edges[jobName])
	r.edges[jobName] = append(r.edges[jobName], deps...)

	if r.hasCycle() {
		r.edges[jobName] = original
		return ErrCycle
	}
	return nil
}

// DepsOf returns the direct dependencies of jobName.
func (r *Registry) DepsOf(jobName string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	deps, ok := r.edges[jobName]
	if !ok {
		return nil, ErrUnknownJob
	}
	out := make([]string, len(deps))
	copy(out, deps)
	return out, nil
}

// Jobs returns all registered job names.
func (r *Registry) Jobs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.edges))
	for k := range r.edges {
		names = append(names, k)
	}
	return names
}

// hasCycle performs a DFS over the current edge map.
// Must be called with r.mu held.
func (r *Registry) hasCycle() bool {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var dfs func(n string) bool
	dfs = func(n string) bool {
		if inStack[n] {
			return true
		}
		if visited[n] {
			return false
		}
		visited[n] = true
		inStack[n] = true
		for _, dep := range r.edges[n] {
			if dfs(dep) {
				return true
			}
		}
		inStack[n] = false
		return false
	}
	for node := range r.edges {
		if dfs(node) {
			return true
		}
	}
	return false
}
