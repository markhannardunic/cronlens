package alert_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronlens/internal/alert"
	"github.com/yourorg/cronlens/internal/job"
)

func makeRun(jobName string, success bool, dur time.Duration) job.Run {
	r := job.NewRun(jobName)
	r.Finish(success, "")
	r.Duration = dur
	return r
}

func TestEvaluate_NoRules(t *testing.T) {
	e := alert.NewEvaluator(nil)
	got := e.Evaluate([]job.Run{makeRun("backup", false, time.Second)})
	if len(got) != 0 {
		t.Fatalf("expected no alerts, got %d", len(got))
	}
}

func TestEvaluate_ConsecutiveFailures_NotReached(t *testing.T) {
	e := alert.NewEvaluator([]alert.Rule{
		{JobName: "backup", ConsecutiveFailures: 3},
	})
	history := []job.Run{
		makeRun("backup", false, time.Second),
		makeRun("backup", false, time.Second),
	}
	got := e.Evaluate(history)
	if len(got) != 0 {
		t.Fatalf("expected no alerts, got %d", len(got))
	}
}

func TestEvaluate_ConsecutiveFailures_Reached(t *testing.T) {
	e := alert.NewEvaluator([]alert.Rule{
		{JobName: "backup", ConsecutiveFailures: 3},
	})
	history := []job.Run{
		makeRun("backup", true, time.Second),
		makeRun("backup", false, time.Second),
		makeRun("backup", false, time.Second),
		makeRun("backup", false, time.Second),
	}
	got := e.Evaluate(history)
	if len(got) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(got))
	}
	if got[0].JobName != "backup" {
		t.Errorf("unexpected job name: %s", got[0].JobName)
	}
}

func TestEvaluate_MaxDuration_Exceeded(t *testing.T) {
	e := alert.NewEvaluator([]alert.Rule{
		{MaxDuration: 5 * time.Second},
	})
	history := []job.Run{makeRun("sync", true, 10*time.Second)}
	got := e.Evaluate(history)
	if len(got) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(got))
	}
}

func TestEvaluate_MaxDuration_NotExceeded(t *testing.T) {
	e := alert.NewEvaluator([]alert.Rule{
		{MaxDuration: 5 * time.Second},
	})
	history := []job.Run{makeRun("sync", true, 2*time.Second)}
	got := e.Evaluate(history)
	if len(got) != 0 {
		t.Fatalf("expected no alerts, got %d", len(got))
	}
}

func TestEvaluate_JobNameFilter(t *testing.T) {
	e := alert.NewEvaluator([]alert.Rule{
		{JobName: "backup", ConsecutiveFailures: 1},
	})
	history := []job.Run{makeRun("sync", false, time.Second)}
	got := e.Evaluate(history)
	if len(got) != 0 {
		t.Fatalf("rule for 'backup' should not fire for 'sync', got %d alerts", len(got))
	}
}
