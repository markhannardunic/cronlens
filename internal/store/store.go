package store

import (
	"sync"
	"time"

	"github.com/cronlens/internal/job"
)

// Store holds an in-memory record of job runs, keyed by job name.
type Store struct {
	mu   sync.RWMutex
	runs map[string][]*job.Run
}

// New creates and returns an initialised Store.
func New() *Store {
	return &Store{
		runs: make(map[string][]*job.Run),
	}
}

// Record appends a completed Run to the store for the given job name.
func (s *Store) Record(name string, r *job.Run) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs[name] = append(s.runs[name], r)
}

// Latest returns the most recent Run for the given job name, or nil if none
// exist.
func (s *Store) Latest(name string) *job.Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	runs := s.runs[name]
	if len(runs) == 0 {
		return nil
	}
	return runs[len(runs)-1]
}

// Since returns all runs for the given job name that started at or after t.
func (s *Store) Since(name string, t time.Time) []*job.Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*job.Run
	for _, r := range s.runs[name] {
		if !r.StartedAt.Before(t) {
			result = append(result, r)
		}
	}
	return result
}

// JobNames returns a sorted list of all job names that have at least one
// recorded run.
func (s *Store) JobNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.runs))
	for name := range s.runs {
		names = append(names, name)
	}
	return names
}
