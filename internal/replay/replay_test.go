package replay_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cronlens/internal/replay"
	"github.com/cronlens/internal/store"
)

func newReplayer() *replay.Replayer {
	s := store.New()
	return replay.New(s)
}

func TestReplay_UnregisteredJob(t *testing.T) {
	r := newReplayer()
	_, err := r.Replay("missing-job")
	if err == nil {
		t.Fatal("expected error for unregistered job, got nil")
	}
}

func TestReplay_SuccessfulRun(t *testing.T) {
	r := newReplayer()
	r.Register("backup", func() error { return nil })

	run, err := r.Replay("backup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run == nil {
		t.Fatal("expected non-nil run")
	}
	if run.JobName != "backup" {
		t.Errorf("job name: got %q, want %q", run.JobName, "backup")
	}
	if !run.Success {
		t.Error("expected run to be successful")
	}
}

func TestReplay_FailedRun(t *testing.T) {
	r := newReplayer()
	r.Register("flaky", func() error { return errors.New("disk full") })

	run, err := r.Replay("flaky")
	if err != nil {
		t.Fatalf("unexpected error recording run: %v", err)
	}
	if run.Success {
		t.Error("expected run to be marked failed")
	}
	if run.Err == nil {
		t.Error("expected non-nil Err on failed run")
	}
}

func TestReplay_LastRun_NotFound(t *testing.T) {
	r := newReplayer()
	_, err := r.LastRun("nojob")
	if err == nil {
		t.Fatal("expected error when no runs exist")
	}
}

func TestReplay_TimeSinceLastRun(t *testing.T) {
	r := newReplayer()
	r.Register("cleanup", func() error { return nil })

	if _, err := r.Replay("cleanup"); err != nil {
		t.Fatalf("replay failed: %v", err)
	}

	d, err := r.TimeSinceLastRun("cleanup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d < 0 || d > 5*time.Second {
		t.Errorf("unexpected duration since last run: %v", d)
	}
}
