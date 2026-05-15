// Package circuitbreaker provides a per-job circuit breaker that temporarily
// disables job execution after a configurable number of consecutive failures,
// allowing the system to recover before retrying.
package circuitbreaker

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// State represents the current state of a circuit breaker.
type State string

const (
	StateClosed   State = "closed"   // normal operation
	StateOpen     State = "open"     // tripped, rejecting runs
	StateHalfOpen State = "half_open" // testing recovery
)

// ErrOpen is returned when a job is blocked by an open circuit.
var ErrOpen = errors.New("circuit breaker is open")

// Rule configures a circuit breaker for a specific job.
type Rule struct {
	JobName      string
	Threshold    int           // consecutive failures before opening
	ResetTimeout time.Duration // time before transitioning to half-open
}

// entry tracks per-job circuit state.
type entry struct {
	state       State
	failures    int
	lastTripped time.Time
	rule        Rule
}

// Breaker manages circuit breaker state for multiple jobs.
type Breaker struct {
	mu      sync.Mutex
	circuits map[string]*entry
}

// New returns an initialised Breaker with no rules registered.
func New() *Breaker {
	return &Breaker{circuits: make(map[string]*entry)}
}

// Register adds or replaces the circuit breaker rule for a job.
func (b *Breaker) Register(r Rule) error {
	if r.Threshold <= 0 {
		return fmt.Errorf("threshold must be > 0")
	}
	if r.ResetTimeout <= 0 {
		return fmt.Errorf("reset timeout must be > 0")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.circuits[r.JobName] = &entry{state: StateClosed, rule: r}
	return nil
}

// Allow returns nil if the job may run, or ErrOpen if the circuit is open.
// It also handles the half-open probe window.
func (b *Breaker) Allow(jobName string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	e, ok := b.circuits[jobName]
	if !ok {
		return nil // no rule → always allow
	}
	if e.state == StateOpen {
		if time.Since(e.lastTripped) >= e.rule.ResetTimeout {
			e.state = StateHalfOpen
		} else {
			return ErrOpen
		}
	}
	return nil
}

// RecordSuccess resets the failure counter and closes the circuit.
func (b *Breaker) RecordSuccess(jobName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if e, ok := b.circuits[jobName]; ok {
		e.failures = 0
		e.state = StateClosed
	}
}

// RecordFailure increments the failure counter and may trip the circuit.
func (b *Breaker) RecordFailure(jobName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e, ok := b.circuits[jobName]
	if !ok {
		return
	}
	e.failures++
	if e.failures >= e.rule.Threshold {
		e.state = StateOpen
		e.lastTripped = time.Now()
	}
}

// StateOf returns the current State for a job (StateClosed if no rule exists).
func (b *Breaker) StateOf(jobName string) State {
	b.mu.Lock()
	defer b.mu.Unlock()
	if e, ok := b.circuits[jobName]; ok {
		return e.state
	}
	return StateClosed
}
