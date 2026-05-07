package notifier_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/cronlens/internal/job"
	"github.com/user/cronlens/internal/notifier"
)

func makeRun(success bool, exitErr string) job.Run {
	now := time.Now()
	r := job.Run{
		JobName:   "test-job",
		StartedAt: now.Add(-5 * time.Second),
	}
	r.Success = success
	r.ExitError = exitErr
	r.FinishedAt = now
	return r
}

func TestNotifier_NoAlertOnSuccess(t *testing.T) {
	var buf bytes.Buffer
	n := notifier.NewWithWriter(2, &buf)
	runs := []job.Run{makeRun(true, ""), makeRun(true, "")}
	n.Check("test-job", runs)
	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %s", buf.String())
	}
}

func TestNotifier_NoAlertBelowThreshold(t *testing.T) {
	var buf bytes.Buffer
	n := notifier.NewWithWriter(3, &buf)
	runs := []job.Run{makeRun(true, ""), makeRun(false, "exit 1"), makeRun(false, "exit 1")}
	n.Check("test-job", runs)
	if buf.Len() != 0 {
		t.Errorf("expected no output below threshold, got: %s", buf.String())
	}
}

func TestNotifier_AlertAtThreshold(t *testing.T) {
	var buf bytes.Buffer
	n := notifier.NewWithWriter(2, &buf)
	runs := []job.Run{makeRun(true, ""), makeRun(false, "exit 2"), makeRun(false, "exit 2")}
	n.Check("test-job", runs)
	out := buf.String()
	if !strings.Contains(out, "ALERT") {
		t.Errorf("expected ALERT in output, got: %s", out)
	}
	if !strings.Contains(out, "test-job") {
		t.Errorf("expected job name in output, got: %s", out)
	}
}

func TestNotifier_EmptyRuns(t *testing.T) {
	var buf bytes.Buffer
	n := notifier.NewWithWriter(1, &buf)
	n.Check("test-job", []job.Run{})
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty runs, got: %s", buf.String())
	}
}

func TestNotifier_Alert(t *testing.T) {
	var buf bytes.Buffer
	n := notifier.NewWithWriter(1, &buf)
	n.Alert("something went wrong")
	if !strings.Contains(buf.String(), "something went wrong") {
		t.Errorf("expected message in output, got: %s", buf.String())
	}
}
