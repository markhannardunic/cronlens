package jobpriority_test

import (
	"testing"

	"github.com/cronlens/internal/jobpriority"
)

func TestSet_And_Get(t *testing.T) {
	r := jobpriority.New()
	if err := r.Set("backup", jobpriority.High); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := r.Get("backup"); got != jobpriority.High {
		t.Errorf("expected High, got %v", got)
	}
}

func TestGet_UnknownJob_ReturnsNormal(t *testing.T) {
	r := jobpriority.New()
	if got := r.Get("nonexistent"); got != jobpriority.Normal {
		t.Errorf("expected Normal default, got %v", got)
	}
}

func TestSet_EmptyJobName_ReturnsError(t *testing.T) {
	r := jobpriority.New()
	if err := r.Set("", jobpriority.Low); err == nil {
		t.Error("expected error for empty job name")
	}
}

func TestSet_InvalidLevel_ReturnsError(t *testing.T) {
	r := jobpriority.New()
	if err := r.Set("myjob", jobpriority.Level(99)); err == nil {
		t.Error("expected error for unknown level")
	}
}

func TestDelete_RemovesPriority(t *testing.T) {
	r := jobpriority.New()
	_ = r.Set("cleanup", jobpriority.Critical)
	r.Delete("cleanup")
	if got := r.Get("cleanup"); got != jobpriority.Normal {
		t.Errorf("expected Normal after delete, got %v", got)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	r := jobpriority.New()
	_ = r.Set("jobA", jobpriority.Low)
	_ = r.Set("jobB", jobpriority.Critical)
	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the copy should not affect the registry.
	delete(all, "jobA")
	if r.Get("jobA") != jobpriority.Low {
		t.Error("registry was mutated via All() return value")
	}
}

func TestLevel_String(t *testing.T) {
	cases := []struct {
		level jobpriority.Level
		want  string
	}{
		{jobpriority.Low, "low"},
		{jobpriority.Normal, "normal"},
		{jobpriority.High, "high"},
		{jobpriority.Critical, "critical"},
	}
	for _, tc := range cases {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("Level(%d).String() = %q, want %q", int(tc.level), got, tc.want)
		}
	}
}
