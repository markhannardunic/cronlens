package tag_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronlens/internal/job"
	"github.com/yourorg/cronlens/internal/tag"
)

func makeRun(jobName string) job.Run {
	r := job.NewRun(jobName)
	r.Finish(nil)
	return r
}

func TestRegistry_SetAndGet(t *testing.T) {
	r := tag.New()
	r.Set("backup", []string{"infra", "nightly"})

	got := r.Get("backup")
	if len(got) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(got))
	}
	if got[0] != "infra" || got[1] != "nightly" {
		t.Errorf("unexpected tags: %v", got)
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := tag.New()
	if got := r.Get("unknown"); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestRegistry_SetDeduplicates(t *testing.T) {
	r := tag.New()
	r.Set("job", []string{"a", "b", "a", "b"})
	got := r.Get("job")
	if len(got) != 2 {
		t.Errorf("expected 2 unique tags, got %d: %v", len(got), got)
	}
}

func TestRegistry_Filter_Empty(t *testing.T) {
	r := tag.New()
	runs := []job.Run{makeRun("job-a"), makeRun("job-b")}
	got := r.Filter(runs, nil)
	if len(got) != 2 {
		t.Errorf("expected all runs when no filter tags, got %d", len(got))
	}
}

func TestRegistry_Filter_Match(t *testing.T) {
	r := tag.New()
	r.Set("job-a", []string{"nightly"})
	r.Set("job-b", []string{"hourly"})

	runs := []job.Run{makeRun("job-a"), makeRun("job-b"), makeRun("job-a")}
	got := r.Filter(runs, []string{"nightly"})
	if len(got) != 2 {
		t.Errorf("expected 2 runs for 'nightly', got %d", len(got))
	}
	for _, run := range got {
		if run.JobName != "job-a" {
			t.Errorf("unexpected job in result: %s", run.JobName)
		}
	}
}

func TestRegistry_Filter_NoMatch(t *testing.T) {
	r := tag.New()
	r.Set("job-a", []string{"infra"})

	runs := []job.Run{makeRun("job-a")}
	got := r.Filter(runs, []string{"nonexistent"})
	if len(got) != 0 {
		t.Errorf("expected 0 runs, got %d", len(got))
	}
}

func TestRegistry_JobNames(t *testing.T) {
	r := tag.New()
	r.Set("zebra", []string{"z"})
	r.Set("alpha", []string{"a"})

	names := r.JobNames()
	if len(names) != 2 || names[0] != "alpha" || names[1] != "zebra" {
		t.Errorf("expected sorted names, got %v", names)
	}
}

// Ensure zero-value time fields don't cause panics.
func TestRegistry_Filter_ZeroTime(t *testing.T) {
	r := tag.New()
	r.Set("j", []string{"t"})
	run := job.Run{JobName: "j", StartedAt: time.Time{}}
	got := r.Filter([]job.Run{run}, []string{"t"})
	if len(got) != 1 {
		t.Errorf("expected 1 run, got %d", len(got))
	}
}
