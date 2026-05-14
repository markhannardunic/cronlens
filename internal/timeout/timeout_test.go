package timeout_test

import (
	"testing"
	"time"

	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/timeout"
)

func makeRun(name, id string, dur time.Duration, success bool) job.Run {
	r := job.NewRun(name)
	r.ID = id
	r.Finish(success, dur)
	return r
}

func TestEvaluate_NoRules(t *testing.T) {
	w := timeout.New()
	runs := []job.Run{makeRun("backup", "r1", 5*time.Second, true)}
	if v := w.Evaluate(runs); len(v) != 0 {
		t.Fatalf("expected no violations, got %d", len(v))
	}
}

func TestEvaluate_WithinLimit(t *testing.T) {
	w := timeout.New()
	w.Register("backup", 10*time.Second)
	runs := []job.Run{makeRun("backup", "r1", 5*time.Second, true)}
	if v := w.Evaluate(runs); len(v) != 0 {
		t.Fatalf("expected no violations, got %d", len(v))
	}
}

func TestEvaluate_ExceedsLimit(t *testing.T) {
	w := timeout.New()
	w.Register("backup", 3*time.Second)
	runs := []job.Run{makeRun("backup", "r1", 5*time.Second, true)}
	v := w.Evaluate(runs)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].JobName != "backup" {
		t.Errorf("unexpected job name %q", v[0].JobName)
	}
	if v[0].Ran != 5*time.Second {
		t.Errorf("unexpected ran duration %v", v[0].Ran)
	}
}

func TestEvaluate_UnfinishedRunIgnored(t *testing.T) {
	w := timeout.New()
	w.Register("sync", time.Second)
	// unfinished run — Finished == false
	r := job.NewRun("sync")
	if v := w.Evaluate([]job.Run{r}); len(v) != 0 {
		t.Fatalf("expected no violations for unfinished run, got %d", len(v))
	}
}

func TestEvaluate_MultipleJobs(t *testing.T) {
	w := timeout.New()
	w.Register("alpha", 2*time.Second)
	w.Register("beta", 10*time.Second)
	runs := []job.Run{
		makeRun("alpha", "a1", 5*time.Second, false), // violation
		makeRun("beta", "b1", 3*time.Second, true),  // ok
		makeRun("gamma", "g1", 99*time.Second, true), // no rule
	}
	v := w.Evaluate(runs)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].JobName != "alpha" {
		t.Errorf("expected alpha violation, got %q", v[0].JobName)
	}
}

func TestRegister_OverwritesRule(t *testing.T) {
	w := timeout.New()
	w.Register("job", 5*time.Second)
	w.Register("job", 1*time.Second)
	runs := []job.Run{makeRun("job", "j1", 3*time.Second, true)}
	v := w.Evaluate(runs)
	if len(v) != 1 {
		t.Fatalf("expected violation after rule tightened, got %d", len(v))
	}
}

func TestRules_Snapshot(t *testing.T) {
	w := timeout.New()
	w.Register("a", time.Minute)
	w.Register("b", time.Hour)
	rules := w.Rules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}
