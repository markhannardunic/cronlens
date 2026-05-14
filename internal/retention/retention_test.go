package retention_test

import (
	"testing"
	"time"

	"github.com/user/cronlens/internal/job"
	"github.com/user/cronlens/internal/retention"
	"github.com/user/cronlens/internal/store"
)

func makeRun(name string, start time.Time, success bool) job.Run {
	r := job.NewRun(name)
	r.StartedAt = start
	err := ""
	if !success {
		err = "boom"
	}
	r.Finish(err)
	return r
}

func TestPrune_RemovesOldRuns(t *testing.T) {
	s := store.New()
	now := time.Now()

	s.Record(makeRun("job1", now.Add(-3*time.Hour), true))
	s.Record(makeRun("job1", now.Add(-2*time.Hour), true))
	s.Record(makeRun("job1", now.Add(-30*time.Minute), true))

	p := retention.New(s, 1*time.Hour, 24*time.Hour)
	n := p.Prune()

	if n != 2 {
		t.Fatalf("expected 2 pruned, got %d", n)
	}

	runs := s.Latest("job1", 10)
	if len(runs) != 1 {
		t.Fatalf("expected 1 remaining run, got %d", len(runs))
	}
}

func TestPrune_NothingToRemove(t *testing.T) {
	s := store.New()
	now := time.Now()

	s.Record(makeRun("jobA", now.Add(-10*time.Minute), true))

	p := retention.New(s, 1*time.Hour, 24*time.Hour)
	n := p.Prune()

	if n != 0 {
		t.Fatalf("expected 0 pruned, got %d", n)
	}
}

func TestPrune_EmptyStore(t *testing.T) {
	s := store.New()
	p := retention.New(s, 1*time.Hour, 24*time.Hour)
	n := p.Prune()
	if n != 0 {
		t.Fatalf("expected 0 pruned on empty store, got %d", n)
	}
}

func TestPrune_MultipleJobs(t *testing.T) {
	s := store.New()
	now := time.Now()

	s.Record(makeRun("alpha", now.Add(-5*time.Hour), true))
	s.Record(makeRun("beta", now.Add(-4*time.Hour), false))
	s.Record(makeRun("alpha", now.Add(-20*time.Minute), true))
	s.Record(makeRun("beta", now.Add(-10*time.Minute), true))

	p := retention.New(s, 1*time.Hour, 24*time.Hour)
	n := p.Prune()

	if n != 2 {
		t.Fatalf("expected 2 pruned, got %d", n)
	}

	for _, name := range []string{"alpha", "beta"} {
		runs := s.Latest(name, 10)
		if len(runs) != 1 {
			t.Errorf("job %s: expected 1 run remaining, got %d", name, len(runs))
		}
	}
}
