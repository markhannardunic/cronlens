package ratelimit_test

import (
	"testing"
	"time"

	"github.com/cronlens/internal/ratelimit"
)

func TestAllow_NoRule(t *testing.T) {
	l := ratelimit.New()
	ok, err := l.Allow("backup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow=true when no rule is configured")
	}
}

func TestAllow_FirstRun_WithRule(t *testing.T) {
	l := ratelimit.New()
	l.SetMinInterval("backup", 5*time.Minute)

	ok, err := l.Allow("backup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow=true on first run")
	}
}

func TestAllow_SecondRun_TooSoon(t *testing.T) {
	l := ratelimit.New()
	l.SetMinInterval("backup", 10*time.Minute)

	// First run succeeds.
	l.Allow("backup") //nolint:errcheck

	// Immediate second run should be blocked.
	ok, err := l.Allow("backup")
	if ok {
		t.Fatal("expected Allow=false when interval has not elapsed")
	}
	if err == nil {
		t.Fatal("expected a descriptive error when rate-limited")
	}
}

func TestAllow_SecondRun_AfterInterval(t *testing.T) {
	l := ratelimit.New()
	l.SetMinInterval("sync", time.Millisecond)

	l.Allow("sync") //nolint:errcheck
	time.Sleep(5 * time.Millisecond)

	ok, err := l.Allow("sync")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow=true after interval has elapsed")
	}
}

func TestReset_AllowsImmediateRun(t *testing.T) {
	l := ratelimit.New()
	l.SetMinInterval("report", 1*time.Hour)

	l.Allow("report") //nolint:errcheck
	l.Reset("report")

	ok, err := l.Allow("report")
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow=true after Reset")
	}
}

func TestRules_ReturnsSnapshot(t *testing.T) {
	l := ratelimit.New()
	l.SetMinInterval("a", time.Minute)
	l.SetMinInterval("b", time.Hour)

	rules := l.Rules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules["a"] != time.Minute {
		t.Errorf("expected 1m for job a, got %v", rules["a"])
	}
	if rules["b"] != time.Hour {
		t.Errorf("expected 1h for job b, got %v", rules["b"])
	}
}
