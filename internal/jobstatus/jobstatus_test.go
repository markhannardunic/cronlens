package jobstatus_test

import (
	"testing"

	"github.com/example/cronlens/internal/jobstatus"
)

func TestGet_UnknownJob_DefaultsToEnabled(t *testing.T) {
	r := jobstatus.New()
	e := r.Get("unknown")
	if e.Status != jobstatus.StatusEnabled {
		t.Fatalf("expected enabled, got %s", e.Status)
	}
}

func TestSet_And_Get(t *testing.T) {
	r := jobstatus.New()
	if err := r.Set("backup", jobstatus.StatusDisabled, "maintenance"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e := r.Get("backup")
	if e.Status != jobstatus.StatusDisabled {
		t.Fatalf("expected disabled, got %s", e.Status)
	}
	if e.Reason != "maintenance" {
		t.Fatalf("expected reason 'maintenance', got %q", e.Reason)
	}
}

func TestSet_EmptyJobName_ReturnsError(t *testing.T) {
	r := jobstatus.New()
	if err := r.Set("", jobstatus.StatusEnabled, ""); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestSet_InvalidStatus_ReturnsError(t *testing.T) {
	r := jobstatus.New()
	if err := r.Set("myjob", jobstatus.Status("unknown"), ""); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestIsEnabled(t *testing.T) {
	r := jobstatus.New()
	_ = r.Set("job-a", jobstatus.StatusDisabled, "")
	if r.IsEnabled("job-a") {
		t.Fatal("expected job-a to be disabled")
	}
	if !r.IsEnabled("job-b") {
		t.Fatal("expected unknown job-b to be enabled by default")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	r := jobstatus.New()
	_ = r.Set("alpha", jobstatus.StatusEnabled, "")
	_ = r.Set("beta", jobstatus.StatusDisabled, "paused")
	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	r := jobstatus.New()
	_ = r.Set("gamma", jobstatus.StatusDisabled, "")
	if err := r.Delete("gamma"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.IsEnabled("gamma") {
		t.Fatal("expected gamma to revert to enabled after deletion")
	}
}

func TestDelete_EmptyJobName_ReturnsError(t *testing.T) {
	r := jobstatus.New()
	if err := r.Delete(""); err == nil {
		t.Fatal("expected error for empty job name")
	}
}
