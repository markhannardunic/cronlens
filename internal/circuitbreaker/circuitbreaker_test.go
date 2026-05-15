package circuitbreaker

import (
	"testing"
	"time"
)

func defaultRule(job string) Rule {
	return Rule{JobName: job, Threshold: 3, ResetTimeout: 50 * time.Millisecond}
}

func TestAllow_NoRule(t *testing.T) {
	b := New()
	if err := b.Allow("unknown"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_ClosedCircuit(t *testing.T) {
	b := New()
	_ = b.Register(defaultRule("job1"))
	if err := b.Allow("job1"); err != nil {
		t.Fatalf("closed circuit should allow: %v", err)
	}
}

func TestRegister_InvalidThreshold(t *testing.T) {
	b := New()
	if err := b.Register(Rule{JobName: "j", Threshold: 0, ResetTimeout: time.Second}); err == nil {
		t.Fatal("expected error for zero threshold")
	}
}

func TestRegister_InvalidTimeout(t *testing.T) {
	b := New()
	if err := b.Register(Rule{JobName: "j", Threshold: 1, ResetTimeout: 0}); err == nil {
		t.Fatal("expected error for zero reset timeout")
	}
}

func TestCircuit_TripsAfterThreshold(t *testing.T) {
	b := New()
	_ = b.Register(defaultRule("job1"))
	for i := 0; i < 3; i++ {
		b.RecordFailure("job1")
	}
	if b.StateOf("job1") != StateOpen {
		t.Fatalf("expected open, got %s", b.StateOf("job1"))
	}
	if err := b.Allow("job1"); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestCircuit_ResetsOnSuccess(t *testing.T) {
	b := New()
	_ = b.Register(defaultRule("job1"))
	b.RecordFailure("job1")
	b.RecordFailure("job1")
	b.RecordSuccess("job1")
	if b.StateOf("job1") != StateClosed {
		t.Fatalf("expected closed after success, got %s", b.StateOf("job1"))
	}
	if err := b.Allow("job1"); err != nil {
		t.Fatalf("should allow after reset: %v", err)
	}
}

func TestCircuit_HalfOpenAfterTimeout(t *testing.T) {
	b := New()
	_ = b.Register(Rule{JobName: "job2", Threshold: 1, ResetTimeout: 20 * time.Millisecond})
	b.RecordFailure("job2")
	if b.StateOf("job2") != StateOpen {
		t.Fatal("expected open")
	}
	time.Sleep(30 * time.Millisecond)
	if err := b.Allow("job2"); err != nil {
		t.Fatalf("expected half-open probe to be allowed: %v", err)
	}
	if b.StateOf("job2") != StateHalfOpen {
		t.Fatalf("expected half_open, got %s", b.StateOf("job2"))
	}
}

func TestStateOf_UnknownJob(t *testing.T) {
	b := New()
	if s := b.StateOf("nope"); s != StateClosed {
		t.Fatalf("expected closed for unknown job, got %s", s)
	}
}
