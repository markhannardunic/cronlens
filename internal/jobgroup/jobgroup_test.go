package jobgroup_test

import (
	"testing"

	"cronlens/internal/jobgroup"
)

func TestAddJob_And_JobsIn(t *testing.T) {
	r := jobgroup.New()
	if err := r.AddJob("billing", "invoice-send"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.AddJob("billing", "payment-check"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	jobs, err := r.JobsIn("billing")
	if err != nil {
		t.Fatalf("JobsIn error: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestAddJob_EmptyGroupReturnsError(t *testing.T) {
	r := jobgroup.New()
	if err := r.AddJob("", "some-job"); err == nil {
		t.Fatal("expected error for empty group")
	}
}

func TestAddJob_EmptyJobReturnsError(t *testing.T) {
	r := jobgroup.New()
	if err := r.AddJob("g", ""); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestJobsIn_UnknownGroupReturnsError(t *testing.T) {
	r := jobgroup.New()
	if _, err := r.JobsIn("nope"); err == nil {
		t.Fatal("expected error for unknown group")
	}
}

func TestRemoveJob_RemovesAndCleansGroup(t *testing.T) {
	r := jobgroup.New()
	_ = r.AddJob("ops", "cleanup")
	if err := r.RemoveJob("ops", "cleanup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	groups := r.Groups()
	if len(groups) != 0 {
		t.Fatalf("expected group to be removed, got %v", groups)
	}
}

func TestRemoveJob_UnknownGroupReturnsError(t *testing.T) {
	r := jobgroup.New()
	if err := r.RemoveJob("ghost", "job"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveJob_UnknownJobReturnsError(t *testing.T) {
	r := jobgroup.New()
	_ = r.AddJob("g", "a")
	if err := r.RemoveJob("g", "b"); err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestGroups_ReturnAllGroups(t *testing.T) {
	r := jobgroup.New()
	_ = r.AddJob("alpha", "j1")
	_ = r.AddJob("beta", "j2")
	groups := r.Groups()
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestGroupOf_ReturnsGroupsContainingJob(t *testing.T) {
	r := jobgroup.New()
	_ = r.AddJob("alpha", "shared")
	_ = r.AddJob("beta", "shared")
	_ = r.AddJob("gamma", "other")
	result := r.GroupOf("shared")
	if len(result) != 2 {
		t.Fatalf("expected 2 groups for 'shared', got %d: %v", len(result), result)
	}
}

func TestAddJob_Idempotent(t *testing.T) {
	r := jobgroup.New()
	_ = r.AddJob("g", "j")
	_ = r.AddJob("g", "j") // duplicate
	jobs, _ := r.JobsIn("g")
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job after duplicate add, got %d", len(jobs))
	}
}
