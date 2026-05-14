// Package replay provides functionality to re-execute a recorded cron job run
// by replaying its command or function through the collector.
package replay

import (
	"errors"
	"fmt"
	"time"

	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/store"
)

// Replayer re-runs the most recent recorded execution of a named job.
type Replayer struct {
	store   *store.Store
	runners map[string]func() error
}

// New creates a new Replayer backed by the given store.
func New(s *store.Store) *Replayer {
	return &Replayer{
		store:   s,
		runners: make(map[string]func() error),
	}
}

// Register associates a runnable function with a job name so it can be replayed.
func (r *Replayer) Register(name string, fn func() error) {
	r.runners[name] = fn
}

// Replay executes the registered function for the given job name, records the
// result in the store, and returns the resulting Run.
func (r *Replayer) Replay(name string) (*job.Run, error) {
	fn, ok := r.runners[name]
	if !ok {
		return nil, fmt.Errorf("replay: no runner registered for job %q", name)
	}

	run := job.NewRun(name)
	err := fn()
	if err != nil {
		run.Finish(err)
	} else {
		run.Finish(nil)
	}

	if storeErr := r.store.Record(run); storeErr != nil {
		return run, fmt.Errorf("replay: store record: %w", storeErr)
	}
	return run, nil
}

// LastRun returns the most recent run for the given job name, or an error if
// none exists.
func (r *Replayer) LastRun(name string) (*job.Run, error) {
	runs := r.store.Latest(name, 1)
	if len(runs) == 0 {
		return nil, errors.New("replay: no previous run found for job " + name)
	}
	return runs[0], nil
}

// TimeSinceLastRun returns the duration since the last recorded run of the job.
func (r *Replayer) TimeSinceLastRun(name string) (time.Duration, error) {
	run, err := r.LastRun(name)
	if err != nil {
		return 0, err
	}
	return time.Since(run.StartedAt), nil
}
