package dependency_test

import (
	"testing"

	"github.com/yourorg/cronlens/internal/dependency"
)

func TestRegister_NoDeps(t *testing.T) {
	r := dependency.New()
	if err := r.Register("jobA", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	jobs := r.Jobs()
	if len(jobs) != 1 || jobs[0] != "jobA" {
		t.Fatalf("expected [jobA], got %v", jobs)
	}
}

func TestRegister_WithDeps(t *testing.T) {
	r := dependency.New()
	if err := r.Register("jobB", []string{"jobA"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deps, err := r.DepsOf("jobB")
	if err != nil {
		t.Fatalf("DepsOf error: %v", err)
	}
	if len(deps) != 1 || deps[0] != "jobA" {
		t.Fatalf("expected [jobA], got %v", deps)
	}
}

func TestRegister_CycleDetected(t *testing.T) {
	r := dependency.New()
	if err := r.Register("jobA", []string{"jobB"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.Register("jobB", []string{"jobA"}); err == nil {
		t.Fatal("expected ErrCycle, got nil")
	} else if err != dependency.ErrCycle {
		t.Fatalf("expected ErrCycle, got %v", err)
	}
}

func TestRegister_SelfCycle(t *testing.T) {
	r := dependency.New()
	err := r.Register("jobA", []string{"jobA"})
	if err != dependency.ErrCycle {
		t.Fatalf("expected ErrCycle for self-loop, got %v", err)
	}
}

func TestDepsOf_UnknownJob(t *testing.T) {
	r := dependency.New()
	_, err := r.DepsOf("ghost")
	if err != dependency.ErrUnknownJob {
		t.Fatalf("expected ErrUnknownJob, got %v", err)
	}
}

func TestRegister_ChainNoCycle(t *testing.T) {
	r := dependency.New()
	// A → B → C is acyclic
	if err := r.Register("jobB", []string{"jobA"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.Register("jobC", []string{"jobB"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// C → A would complete the cycle
	if err := r.Register("jobA", []string{"jobC"}); err != dependency.ErrCycle {
		t.Fatalf("expected ErrCycle for closing chain, got %v", err)
	}
}

func TestJobs_ReturnsAll(t *testing.T) {
	r := dependency.New()
	_ = r.Register("jobA", nil)
	_ = r.Register("jobB", []string{"jobA"})
	_ = r.Register("jobC", []string{"jobA"})
	jobs := r.Jobs()
	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d: %v", len(jobs), jobs)
	}
}
