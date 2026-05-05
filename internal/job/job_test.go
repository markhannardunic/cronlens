package job

import (
	"testing"
	"time"
)

func TestNewRun_InitialState(t *testing.T) {
	before := time.Now()
	r := NewRun("abc123", "backup-db")
	after := time.Now()

	if r.ID != "abc123" {
		t.Errorf("expected ID abc123, got %s", r.ID)
	}
	if r.Name != "backup-db" {
		t.Errorf("expected name backup-db, got %s", r.Name)
	}
	if r.Status != StatusRunning {
		t.Errorf("expected status running, got %s", r.Status)
	}
	if r.EndedAt != nil {
		t.Error("expected EndedAt to be nil for a new run")
	}
	if r.StartedAt.Before(before) || r.StartedAt.After(after) {
		t.Error("StartedAt is outside expected range")
	}
}

func TestRun_FinishSuccess(t *testing.T) {
	r := NewRun("id1", "sync-files")
	time.Sleep(2 * time.Millisecond)
	r.Finish(0, "done")

	if r.Status != StatusSuccess {
		t.Errorf("expected success, got %s", r.Status)
	}
	if r.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", r.ExitCode)
	}
	if r.EndedAt == nil {
		t.Fatal("expected EndedAt to be set")
	}
	if r.Duration <= 0 {
		t.Error("expected positive duration")
	}
	if r.Output != "done" {
		t.Errorf("expected output 'done', got %s", r.Output)
	}
}

func TestRun_FinishFailure(t *testing.T) {
	r := NewRun("id2", "clean-logs")
	r.Finish(1, "error: permission denied")

	if r.Status != StatusFailure {
		t.Errorf("expected failure, got %s", r.Status)
	}
	if r.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", r.ExitCode)
	}
}
