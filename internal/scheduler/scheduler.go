package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"cronlens/internal/collector"
)

// Entry represents a registered cron job with its schedule.
type Entry struct {
	Name     string
	Schedule string
	Collect  collector.Collector
}

// Scheduler wraps a cron runner and manages registered jobs.
type Scheduler struct {
	mu      sync.Mutex
	cron    *cron.Cron
	entries []Entry
}

// New creates a new Scheduler.
func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

// Register adds a job to the scheduler with the given cron expression.
func (s *Scheduler) Register(name, schedule string, c collector.Collector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.cron.AddFunc(schedule, func() {
		log.Printf("[scheduler] running job: %s", name)
		if err := c.Run(); err != nil {
			log.Printf("[scheduler] job %s failed: %v", name, err)
		}
	})
	if err != nil {
		return err
	}

	s.entries = append(s.entries, Entry{Name: name, Schedule: schedule, Collect: c})
	return nil
}

// Start begins the scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop gracefully stops the scheduler and waits for running jobs.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(30 * time.Second):
		log.Println("[scheduler] timed out waiting for jobs to finish")
	}
}

// Entries returns a copy of all registered entries.
func (s *Scheduler) Entries() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}
