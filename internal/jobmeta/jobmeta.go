// Package jobmeta stores arbitrary key-value metadata associated with
// registered cron job names. Unlike labeler (which stores operational
// labels), jobmeta is intended for descriptive, human-facing attributes
// such as owner, description, or team.
package jobmeta

import (
	"errors"
	"sync"
)

// Registry holds metadata entries per job.
type Registry struct {
	mu   sync.RWMutex
	data map[string]map[string]string
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{data: make(map[string]map[string]string)}
}

// Set stores key=value metadata for the given job name.
// Returns an error if jobName or key is empty.
func (r *Registry) Set(jobName, key, value string) error {
	if jobName == "" {
		return errors.New("jobmeta: job name must not be empty")
	}
	if key == "" {
		return errors.New("jobmeta: key must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.data[jobName] == nil {
		r.data[jobName] = make(map[string]string)
	}
	r.data[jobName][key] = value
	return nil
}

// Get returns the metadata map for the given job. The returned map is a
// shallow copy; the caller may read but should not mutate it.
func (r *Registry) Get(jobName string) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.data[jobName]
	if src == nil {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// Delete removes a single metadata key for the given job.
func (r *Registry) Delete(jobName, key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.data[jobName] != nil {
		delete(r.data[jobName], key)
	}
}

// JobNames returns the sorted list of job names that have at least one
// metadata entry.
func (r *Registry) JobNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.data))
	for name, m := range r.data {
		if len(m) > 0 {
			names = append(names, name)
		}
	}
	return names
}
