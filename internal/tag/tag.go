// Package tag provides tagging support for cron jobs, allowing runs to be
// grouped and filtered by arbitrary string labels.
package tag

import (
	"sort"
	"sync"

	"github.com/yourorg/cronlens/internal/job"
)

// Registry maps job names to their associated tags.
type Registry struct {
	mu   sync.RWMutex
	tags map[string][]string // jobName -> sorted unique tags
}

// New returns an empty tag Registry.
func New() *Registry {
	return &Registry{
		tags: make(map[string][]string),
	}
}

// Set replaces all tags for the given job name.
func (r *Registry) Set(jobName string, tags []string) {
	normalized := uniqueSorted(tags)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tags[jobName] = normalized
}

// Get returns the tags associated with the given job name.
// Returns nil if no tags are registered.
func (r *Registry) Get(jobName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tags[jobName]
}

// Filter returns the subset of runs whose job name carries at least one of
// the requested tags. If tags is empty, all runs are returned.
func (r *Registry) Filter(runs []job.Run, tags []string) []job.Run {
	if len(tags) == 0 {
		return runs
	}
	want := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		want[t] = struct{}{}
	}

	var out []job.Run
	for _, run := range runs {
		if r.hasAny(run.JobName, want) {
			out = append(out, run)
		}
	}
	return out
}

// JobNames returns all job names that have at least one registered tag.
func (r *Registry) JobNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tags))
	for k := range r.tags {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) hasAny(jobName string, want map[string]struct{}) bool {
	for _, t := range r.tags[jobName] {
		if _, ok := want[t]; ok {
			return true
		}
	}
	return false
}

func uniqueSorted(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := in[:0:0]
	for _, s := range in {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
