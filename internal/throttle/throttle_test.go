package throttle_test

import (
	"testing"
	"time"

	"github.com/cronlens/internal/throttle"
)

func TestAllow_NoRule(t *testing.T) {
	l := throttle.New()
	if !l.Allow("job-a") {
		t.Fatal("expected job with no rule to be allowed")
	}
}

func TestRegister_InvalidRule(t *testing.T) {
	l := throttle.New()
	if err := l.Register(throttle.Rule{JobName: "", Interval: time.Second}); err == nil {
		t.Fatal("expected error for empty job name")
	}
	if err := l.Register(throttle.Rule{JobName: "job-a", Interval: 0}); err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestAllow_FirstRun(t *testing.T) {
	l := throttle.New()
	_ = l.Register(throttle.Rule{JobName: "job-a", Interval: time.Minute})
	if !l.Allow("job-a") {
		t.Fatal("expected first run to be allowed")
	}
}

func TestAllow_TooSoon(t *testing.T) {
	l := throttle.New()
	_ = l.Register(throttle.Rule{JobName: "job-a", Interval: time.Hour})
	l.Record("job-a")
	if l.Allow("job-a") {
		t.Fatal("expected run to be throttled immediately after record")
	}
}

func TestAllow_AfterInterval(t *testing.T) {
	l := throttle.New()
	_ = l.Register(throttle.Rule{JobName: "job-a", Interval: time.Millisecond})
	l.Record("job-a")
	time.Sleep(5 * time.Millisecond)
	if !l.Allow("job-a") {
		t.Fatal("expected run to be allowed after interval elapsed")
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	l := throttle.New()
	_ = l.Register(throttle.Rule{JobName: "job-a", Interval: time.Second})
	_ = l.Register(throttle.Rule{JobName: "job-b", Interval: 2 * time.Second})
	rules := l.Rules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestRecord_NoRule_DoesNotPanic(t *testing.T) {
	l := throttle.New()
	l.Record("unregistered-job") // should not panic
	if !l.Allow("unregistered-job") {
		t.Fatal("unregistered job should always be allowed")
	}
}
