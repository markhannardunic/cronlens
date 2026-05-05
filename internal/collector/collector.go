package collector

import (
	"os/exec"
	"time"

	"github.com/user/cronlens/internal/job"
	"github.com/user/cronlens/internal/store"
)

// Collector wraps a Store and provides a way to execute a named job,
// recording the outcome automatically.
type Collector struct {
	store *store.Store
}

// New creates a Collector backed by the given Store.
func New(s *store.Store) *Collector {
	return &Collector{store: s}
}

// Run executes the provided shell command under the given job name,
// records start/finish times and any error, then persists the Run.
func (c *Collector) Run(jobName string, command string, args ...string) error {
	run := job.NewRun(jobName)

	cmd := exec.Command(command, args...)
	err := cmd.Run()

	if err != nil {
		run.Finish(err)
	} else {
		run.Finish(nil)
	}

	c.store.Record(run)
	return err
}

// RunFunc executes an arbitrary Go function under the given job name,
// records the outcome, and persists the Run.
func (c *Collector) RunFunc(jobName string, fn func() error) error {
	run := job.NewRun(jobName)

	err := fn()
	run.Finish(err)

	c.store.Record(run)
	return err
}

// Since returns all runs recorded after the given time.
func (c *Collector) Since(t time.Time) []job.Run {
	return c.store.Since(t)
}
