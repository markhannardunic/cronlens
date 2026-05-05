package store_test

import (
	"testing"
	"time"

	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/store"
)

func makeRun(t *testing.T, success bool) *job.Run {
	t.Helper()
	r := job.NewRun()
	if success {
		r.FinishSuccess()
	} else {
		r.FinishFailure("something went wrong")
	}
	return r
}

func TestStore_RecordAndLatest(t *testing.T) {
	s := store.New()

	if got := s.Latest("backup"); got != nil {
		t.Fatalf("expected nil for unknown job, got %v", got)
	}

	r1 := makeRun(t, true)
	s.Record("backup", r1)

	if got := s.Latest("backup"); got != r1 {
		t.Fatalf("expected r1, got %v", got)
	}

	r2 := makeRun(t, false)
	s.Record("backup", r2)

	if got := s.Latest("backup"); got != r2 {
		t.Fatalf("expected r2 as latest, got %v", got)
	}
}

func TestStore_Since(t *testing.T) {
	s := store.New()

	r1 := makeRun(t, true)
	s.Record("sync", r1)

	time.Sleep(2 * time.Millisecond)
	cutoff := time.Now()
	time.Sleep(2 * time.Millisecond)

	r2 := makeRun(t, true)
	s.Record("sync", r2)

	results := s.Since("sync", cutoff)
	if len(results) != 1 {
		t.Fatalf("expected 1 run since cutoff, got %d", len(results))
	}
	if results[0] != r2 {
		t.Fatalf("expected r2 in results")
	}
}

func TestStore_JobNames(t *testing.T) {
	s := store.New()

	if names := s.JobNames(); len(names) != 0 {
		t.Fatalf("expected empty names, got %v", names)
	}

	s.Record("jobA", makeRun(t, true))
	s.Record("jobB", makeRun(t, false))

	names := s.JobNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 job names, got %d", len(names))
	}
}
