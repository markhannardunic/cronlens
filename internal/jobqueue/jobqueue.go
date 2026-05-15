// Package jobqueue provides a priority-aware pending job queue for cronlens.
// Jobs can be enqueued and dequeued in order of their assigned priority level.
package jobqueue

import (
	"errors"
	"sync"
	"time"
)

// Entry represents a queued job waiting to be dispatched.
type Entry struct {
	JobName   string
	Priority  int
	EnqueuedAt time.Time
}

// Queue holds pending job entries ordered by priority (higher value = higher priority).
type Queue struct {
	mu      sync.Mutex
	entries []Entry
}

// New returns an empty Queue.
func New() *Queue {
	return &Queue{}
}

// Enqueue adds a job to the queue with the given priority.
func (q *Queue) Enqueue(jobName string, priority int) error {
	if jobName == "" {
		return errors.New("jobqueue: job name must not be empty")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.entries = append(q.entries, Entry{
		JobName:    jobName,
		Priority:   priority,
		EnqueuedAt: time.Now(),
	})
	return nil
}

// Dequeue removes and returns the highest-priority entry.
// If multiple entries share the same priority, the earliest enqueued is returned.
// Returns an error if the queue is empty.
func (q *Queue) Dequeue() (Entry, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.entries) == 0 {
		return Entry{}, errors.New("jobqueue: queue is empty")
	}
	best := 0
	for i := 1; i < len(q.entries); i++ {
		e := q.entries[i]
		b := q.entries[best]
		if e.Priority > b.Priority || (e.Priority == b.Priority && e.EnqueuedAt.Before(b.EnqueuedAt)) {
			best = i
		}
	}
	entry := q.entries[best]
	q.entries = append(q.entries[:best], q.entries[best+1:]...)
	return entry, nil
}

// Len returns the current number of queued entries.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.entries)
}

// Peek returns all entries in their current (unsorted) order without removing them.
func (q *Queue) Peek() []Entry {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Entry, len(q.entries))
	copy(out, q.entries)
	return out
}
