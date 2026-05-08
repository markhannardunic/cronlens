package scheduler_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"cronlens/internal/scheduler"
)

// fakeCollector is a minimal collector stub for testing.
type fakeCollector struct {
	calls int32
	fail  bool
}

func (f *fakeCollector) Run() error {
	atomic.AddInt32(&f.calls, 1)
	if f.fail {
		return errors.New("simulated failure")
	}
	return nil
}

func TestScheduler_Register_InvalidCron(t *testing.T) {
	s := scheduler.New()
	err := s.Register("bad", "not-a-cron", &fakeCollector{})
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestScheduler_Register_ValidCron(t *testing.T) {
	s := scheduler.New()
	err := s.Register("ping", "* * * * * *", &fakeCollector{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := s.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "ping" {
		t.Errorf("expected name 'ping', got %q", entries[0].Name)
	}
}

func TestScheduler_JobRuns(t *testing.T) {
	s := scheduler.New()
	fc := &fakeCollector{}

	if err := s.Register("tick", "* * * * * *", fc); err != nil {
		t.Fatalf("register: %v", err)
	}

	s.Start()
	time.Sleep(2200 * time.Millisecond)
	s.Stop()

	calls := atomic.LoadInt32(&fc.calls)
	if calls < 2 {
		t.Errorf("expected at least 2 calls, got %d", calls)
	}
}

func TestScheduler_MultipleJobs(t *testing.T) {
	s := scheduler.New()

	for _, name := range []string{"a", "b", "c"} {
		if err := s.Register(name, "@every 1m", &fakeCollector{}); err != nil {
			t.Fatalf("register %s: %v", name, err)
		}
	}

	if got := len(s.Entries()); got != 3 {
		t.Errorf("expected 3 entries, got %d", got)
	}
}
