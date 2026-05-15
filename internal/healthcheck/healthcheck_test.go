package healthcheck_test

import (
	"testing"
	"time"

	"github.com/user/cronlens/internal/healthcheck"
	"github.com/user/cronlens/internal/job"
)

var (
	now   = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	hour  = time.Hour
)

func makeRun(name string, success bool, finishedAt time.Time) job.Run {
	return job.Run{
		JobName:    name,
		Success:    success,
		FinishedAt: finishedAt,
		StartedAt:  finishedAt.Add(-time.Second),
	}
}

func TestEvaluate_NoRules(t *testing.T) {
	c := healthcheck.New()
	results := c.Evaluate(nil, now)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestEvaluate_NoRuns_Unhealthy(t *testing.T) {
	c := healthcheck.New()
	_ = c.Register(healthcheck.Rule{JobName: "backup", Interval: hour})

	results := c.Evaluate(nil, now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Healthy {
		t.Errorf("expected unhealthy when no runs recorded")
	}
}

func TestEvaluate_RecentSuccess_Healthy(t *testing.T) {
	c := healthcheck.New()
	_ = c.Register(healthcheck.Rule{JobName: "backup", Interval: hour})

	runs := []job.Run{makeRun("backup", true, now.Add(-30*time.Minute))}
	results := c.Evaluate(runs, now)

	if !results[0].Healthy {
		t.Errorf("expected healthy, got: %s", results[0].Message)
	}
}

func TestEvaluate_StaleSuccess_Unhealthy(t *testing.T) {
	c := healthcheck.New()
	_ = c.Register(healthcheck.Rule{JobName: "backup", Interval: hour})

	runs := []job.Run{makeRun("backup", true, now.Add(-2*hour))}
	results := c.Evaluate(runs, now)

	if results[0].Healthy {
		t.Errorf("expected unhealthy for stale run")
	}
}

func TestEvaluate_FailureRunIgnored(t *testing.T) {
	c := healthcheck.New()
	_ = c.Register(healthcheck.Rule{JobName: "sync", Interval: hour})

	runs := []job.Run{makeRun("sync", false, now.Add(-10*time.Minute))}
	results := c.Evaluate(runs, now)

	if results[0].Healthy {
		t.Errorf("failed run should not count as healthy")
	}
}

func TestRegister_InvalidRule(t *testing.T) {
	c := healthcheck.New()
	if err := c.Register(healthcheck.Rule{JobName: "", Interval: hour}); err == nil {
		t.Error("expected error for empty job name")
	}
	if err := c.Register(healthcheck.Rule{JobName: "x", Interval: 0}); err == nil {
		t.Error("expected error for zero interval")
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	c := healthcheck.New()
	_ = c.Register(healthcheck.Rule{JobName: "a", Interval: hour})
	_ = c.Register(healthcheck.Rule{JobName: "b", Interval: 2 * hour})
	if len(c.Rules()) != 2 {
		t.Errorf("expected 2 rules")
	}
}
