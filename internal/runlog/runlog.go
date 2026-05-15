// Package runlog provides a structured, append-only log of job run events
// that can be streamed or queried for audit and debugging purposes.
package runlog

import (
	"sync"
	"time"

	"github.com/cronlens/internal/job"
)

// Entry represents a single logged event for a job run.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	JobName   string    `json:"job_name"`
	RunID     string    `json:"run_id"`
	Event     string    `json:"event"` // "started", "success", "failure"
	Message   string    `json:"message,omitempty"`
	Duration  float64   `json:"duration_seconds,omitempty"`
}

// Log is a thread-safe append-only run event log.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
	maxSize int
}

// New creates a new Log with the given maximum number of entries.
// When maxSize is exceeded, the oldest entries are evicted.
func New(maxSize int) *Log {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &Log{maxSize: maxSize}
}

// Record appends an entry derived from the given run.
func (l *Log) Record(r job.Run) {
	entry := Entry{
		Timestamp: time.Now().UTC(),
		JobName:   r.JobName,
		RunID:     r.ID,
	}

	switch {
	case !r.FinishedAt.IsZero() && r.Err == nil:
		entry.Event = "success"
		entry.Duration = r.FinishedAt.Sub(r.StartedAt).Seconds()
	case !r.FinishedAt.IsZero() && r.Err != nil:
		entry.Event = "failure"
		entry.Message = r.Err.Error()
		entry.Duration = r.FinishedAt.Sub(r.StartedAt).Seconds()
	default:
		entry.Event = "started"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
}

// All returns a copy of all log entries in chronological order.
func (l *Log) All() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// ForJob returns all entries for the given job name.
func (l *Log) ForJob(name string) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Entry
	for _, e := range l.entries {
		if e.JobName == name {
			out = append(out, e)
		}
	}
	return out
}

// Clear removes all entries from the log.
func (l *Log) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
}
