// Package audit records a lightweight audit trail of configuration
// changes made to cronlens at runtime (e.g. rule registrations, pauses,
// tag assignments). Each entry captures who changed what and when.
package audit

import (
	"sync"
	"time"
)

// ActionKind describes the type of change that was recorded.
type ActionKind string

const (
	ActionRegister ActionKind = "register"
	ActionUpdate   ActionKind = "update"
	ActionDelete   ActionKind = "delete"
	ActionPause    ActionKind = "pause"
	ActionResume   ActionKind = "resume"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time  `json:"timestamp"`
	Actor     string     `json:"actor"`
	Action    ActionKind `json:"action"`
	Target    string     `json:"target"`
	Detail    string     `json:"detail,omitempty"`
}

// Log stores audit entries in memory.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
}

// New returns an empty audit Log.
func New() *Log {
	return &Log{}
}

// Record appends a new entry to the log.
func (l *Log) Record(actor string, action ActionKind, target, detail string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, Entry{
		Timestamp: time.Now().UTC(),
		Actor:     actor,
		Action:    action,
		Target:    target,
		Detail:    detail,
	})
}

// All returns a copy of all recorded entries, oldest first.
func (l *Log) All() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// ForTarget returns all entries whose Target matches the given name.
func (l *Log) ForTarget(target string) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Entry
	for _, e := range l.entries {
		if e.Target == target {
			out = append(out, e)
		}
	}
	return out
}
