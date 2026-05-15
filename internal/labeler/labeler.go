// Package labeler provides key-value label management for cron jobs,
// allowing arbitrary metadata to be attached and queried.
package labeler

import (
	"fmt"
	"sync"
)

// Registry stores labels (key-value pairs) per job name.
type Registry struct {
	mu     sync.RWMutex
	labels map[string]map[string]string
}

// New creates an empty label Registry.
func New() *Registry {
	return &Registry{
		labels: make(map[string]map[string]string),
	}
}

// Set attaches a label key=value to the named job.
// Returns an error if key is empty.
func (r *Registry) Set(job, key, value string) error {
	if key == "" {
		return fmt.Errorf("labeler: key must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.labels[job] == nil {
		r.labels[job] = make(map[string]string)
	}
	r.labels[job][key] = value
	return nil
}

// Get returns the labels attached to a job. Returns nil if the job has no labels.
func (r *Registry) Get(job string) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.labels[job]
	if src == nil {
		return nil
	}
	copy := make(map[string]string, len(src))
	for k, v := range src {
		copy[k] = v
	}
	return copy
}

// Delete removes a single label key from a job.
func (r *Registry) Delete(job, key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.labels[job], key)
}

// JobsWithLabel returns all job names that have the given key=value label.
func (r *Registry) JobsWithLabel(key, value string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []string
	for job, lbls := range r.labels {
		if lbls[key] == value {
			result = append(result, job)
		}
	}
	return result
}
