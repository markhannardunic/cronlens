package pause_test

import (
	"testing"

	"github.com/cronlens/internal/pause"
)

func TestPause_And_IsPaused(t *testing.T) {
	r := pause.New()

	if r.IsPaused("backup") {
		t.Fatal("expected job not to be paused initially")
	}

	if err := r.Pause("backup", "maintenance window"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !r.IsPaused("backup") {
		t.Fatal("expected job to be paused")
	}
}

func TestPause_Duplicate_ReturnsError(t *testing.T) {
	r := pause.New()

	_ = r.Pause("backup", "")
	if err := r.Pause("backup", ""); err == nil {
		t.Fatal("expected error when pausing an already-paused job")
	}
}

func TestResume_ClearsPausedState(t *testing.T) {
	r := pause.New()
	_ = r.Pause("sync", "testing")

	if err := r.Resume("sync"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.IsPaused("sync") {
		t.Fatal("expected job to no longer be paused after resume")
	}
}

func TestResume_NotPaused_ReturnsError(t *testing.T) {
	r := pause.New()

	if err := r.Resume("nonexistent"); err == nil {
		t.Fatal("expected error when resuming a job that is not paused")
	}
}

func TestList_ReturnsAllPausedEntries(t *testing.T) {
	r := pause.New()

	_ = r.Pause("jobA", "reason A")
	_ = r.Pause("jobB", "reason B")

	entries := r.List()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	names := map[string]bool{}
	for _, e := range entries {
		names[e.JobName] = true
		if e.PausedAt.IsZero() {
			t.Errorf("expected PausedAt to be set for %q", e.JobName)
		}
	}

	if !names["jobA"] || !names["jobB"] {
		t.Errorf("expected both jobA and jobB in list, got %v", names)
	}
}

func TestList_Empty(t *testing.T) {
	r := pause.New()
	if entries := r.List(); len(entries) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(entries))
	}
}
