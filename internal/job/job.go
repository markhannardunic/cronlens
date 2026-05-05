package job

import "time"

// Status represents the outcome of a cron job execution.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusRunning Status = "running"
)

// Run holds the recorded data for a single cron job execution.
type Run struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   *time.Time    `json:"ended_at,omitempty"`
	Duration  time.Duration `json:"duration_ms"`
	Status    Status        `json:"status"`
	ExitCode  int           `json:"exit_code"`
	Output    string        `json:"output,omitempty"`
}

// Finish marks the run as complete, recording end time, duration, and status.
func (r *Run) Finish(exitCode int, output string) {
	now := time.Now()
	r.EndedAt = &now
	r.Duration = now.Sub(r.StartedAt)
	r.ExitCode = exitCode
	r.Output = output
	if exitCode == 0 {
		r.Status = StatusSuccess
	} else {
		r.Status = StatusFailure
	}
}

// NewRun creates a new Run in the running state.
func NewRun(id, name string) *Run {
	return &Run{
		ID:        id,
		Name:      name,
		StartedAt: time.Now(),
		Status:    StatusRunning,
	}
}
