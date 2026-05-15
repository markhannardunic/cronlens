// Package jobgroup provides grouping and aggregation of related cron jobs
// under a named group, allowing bulk queries and status summaries.
package jobgroup

import (
	"errors"
	"fmt"
	"sync"
)

// Registry maps group names to sets of job names.
type Registry struct {
	mu     sync.RWMutex
	groups map[string]map[string]struct{}
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{
		groups: make(map[string]map[string]struct{}),
	}
}

// AddJob adds jobName to the named group, creating the group if necessary.
func (r *Registry) AddJob(group, jobName string) error {
	if group == "" {
		return errors.New("group name must not be empty")
	}
	if jobName == "" {
		return errors.New("job name must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.groups[group]; !ok {
		r.groups[group] = make(map[string]struct{})
	}
	r.groups[group][jobName] = struct{}{}
	return nil
}

// RemoveJob removes jobName from the named group.
// Returns an error if the group or job is not found.
func (r *Registry) RemoveJob(group, jobName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	jobs, ok := r.groups[group]
	if !ok {
		return fmt.Errorf("group %q not found", group)
	}
	if _, ok := jobs[jobName]; !ok {
		return fmt.Errorf("job %q not in group %q", jobName, group)
	}
	delete(jobs, jobName)
	if len(jobs) == 0 {
		delete(r.groups, group)
	}
	return nil
}

// JobsIn returns a sorted slice of job names belonging to group.
func (r *Registry) JobsIn(group string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	jobs, ok := r.groups[group]
	if !ok {
		return nil, fmt.Errorf("group %q not found", group)
	}
	out := make([]string, 0, len(jobs))
	for j := range jobs {
		out = append(out, j)
	}
	return out, nil
}

// Groups returns all group names currently registered.
func (r *Registry) Groups() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.groups))
	for g := range r.groups {
		names = append(names, g)
	}
	return names
}

// GroupOf returns all groups that contain jobName.
func (r *Registry) GroupOf(jobName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []string
	for g, jobs := range r.groups {
		if _, ok := jobs[jobName]; ok {
			out = append(out, g)
		}
	}
	return out
}
