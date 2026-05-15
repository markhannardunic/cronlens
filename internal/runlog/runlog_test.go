package runlog_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/runlog"
)

func makeRun(name string, failed bool, finished bool) job.Run {
	r := job.NewRun(name)
	if finished {
		var err error
		if failed {
			err = errors.New("something went wrong")
		}
		r.FinishedAt = r.StartedAt.Add(2 * time.Second)
		r.Err = err
	}
	return r
}

func TestRecord_StartedEvent(t *testing.T) {
	l := runlog.New(100)
	r := makeRun("backup", false, false)
	l.Record(r)

	entries := l.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Event != "started" {
		t.Errorf("expected event=started, got %q", entries[0].Event)
	}
	if entries[0].JobName != "backup" {
		t.Errorf("expected job_name=backup, got %q", entries[0].JobName)
	}
}

func TestRecord_SuccessEvent(t *testing.T) {
	l := runlog.New(100)
	l.Record(makeRun("sync", false, true))

	entries := l.All()
	if entries[0].Event != "success" {
		t.Errorf("expected success, got %q", entries[0].Event)
	}
	if entries[0].Duration <= 0 {
		t.Errorf("expected positive duration, got %f", entries[0].Duration)
	}
}

func TestRecord_FailureEvent(t *testing.T) {
	l := runlog.New(100)
	l.Record(makeRun("cleanup", true, true))

	entries := l.All()
	if entries[0].Event != "failure" {
		t.Errorf("expected failure, got %q", entries[0].Event)
	}
	if entries[0].Message == "" {
		t.Error("expected non-empty message for failure")
	}
}

func TestForJob_FiltersCorrectly(t *testing.T) {
	l := runlog.New(100)
	l.Record(makeRun("job-a", false, true))
	l.Record(makeRun("job-b", false, true))
	l.Record(makeRun("job-a", true, true))

	results := l.ForJob("job-a")
	if len(results) != 2 {
		t.Fatalf("expected 2 entries for job-a, got %d", len(results))
	}
	for _, e := range results {
		if e.JobName != "job-a" {
			t.Errorf("unexpected job name %q in filtered results", e.JobName)
		}
	}
}

func TestMaxSize_Eviction(t *testing.T) {
	l := runlog.New(3)
	for i := 0; i < 5; i++ {
		l.Record(makeRun("job", false, true))
	}
	if len(l.All()) != 3 {
		t.Errorf("expected log to be capped at 3, got %d", len(l.All()))
	}
}

func TestClear_RemovesAllEntries(t *testing.T) {
	l := runlog.New(100)
	l.Record(makeRun("job", false, true))
	l.Clear()
	if len(l.All()) != 0 {
		t.Errorf("expected empty log after Clear, got %d entries", len(l.All()))
	}
}
