package snapshot_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/snapshot"
	"github.com/cronlens/internal/store"
)

func makeRun(name string, success bool, dur time.Duration) job.Run {
	r := job.NewRun(name)
	time.Sleep(0) // ensure StartedAt is set
	if success {
		r.Finish(nil)
	} else {
		r.Finish(errors.New("failed"))
	}
	_ = dur
	return r
}

func TestCapture_Empty(t *testing.T) {
	s := store.New()
	c := snapshot.New(s)
	snap := c.Capture()

	if len(snap.Jobs) != 0 {
		t.Fatalf("expected 0 jobs, got %d", len(snap.Jobs))
	}
	if snap.CapturedAt.IsZero() {
		t.Fatal("expected CapturedAt to be set")
	}
}

func TestCapture_WithRuns(t *testing.T) {
	s := store.New()
	s.Record(makeRun("backup", true, 0))
	s.Record(makeRun("backup", false, 0))
	s.Record(makeRun("sync", true, 0))

	c := snapshot.New(s)
	snap := c.Capture()

	if len(snap.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(snap.Jobs))
	}

	backup, ok := snap.Jobs["backup"]
	if !ok {
		t.Fatal("expected backup job in snapshot")
	}
	if backup.LastStatus != "failure" {
		t.Errorf("expected last status failure, got %s", backup.LastStatus)
	}
	if backup.RunCount != 2 {
		t.Errorf("expected run count 2, got %d", backup.RunCount)
	}

	sync, ok := snap.Jobs["sync"]
	if !ok {
		t.Fatal("expected sync job in snapshot")
	}
	if sync.LastStatus != "success" {
		t.Errorf("expected last status success, got %s", sync.LastStatus)
	}
}

func TestLast_BeforeCapture(t *testing.T) {
	s := store.New()
	c := snapshot.New(s)
	_, ok := c.Last()
	if ok {
		t.Fatal("expected no last snapshot before first capture")
	}
}

func TestLast_AfterCapture(t *testing.T) {
	s := store.New()
	s.Record(makeRun("job1", true, 0))
	c := snapshot.New(s)
	c.Capture()
	snap, ok := c.Last()
	if !ok {
		t.Fatal("expected last snapshot after capture")
	}
	if _, exists := snap.Jobs["job1"]; !exists {
		t.Error("expected job1 in last snapshot")
	}
}
