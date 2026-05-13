package history_test

import (
	"testing"
	"time"

	"github.com/user/cronlens/internal/history"
	"github.com/user/cronlens/internal/job"
	"github.com/user/cronlens/internal/store"
)

func makeRun(name string, success bool, dur time.Duration, start time.Time) job.Run {
	r := job.NewRun(name)
	r.StartedAt = start
	r.Finish(success, "", dur)
	return r
}

func TestStatsFor_Empty(t *testing.T) {
	s := store.New(100)
	a := history.New(s)
	stats := a.StatsFor("nojob")
	if stats.TotalRuns != 0 {
		t.Fatalf("expected 0 runs, got %d", stats.TotalRuns)
	}
	if stats.SuccessRate != 0 {
		t.Fatalf("expected 0 success rate, got %f", stats.SuccessRate)
	}
}

func TestStatsFor_MixedRuns(t *testing.T) {
	s := store.New(100)
	now := time.Now()
	s.Record(makeRun("job1", true, 2*time.Second, now.Add(-3*time.Minute)))
	s.Record(makeRun("job1", true, 4*time.Second, now.Add(-2*time.Minute)))
	s.Record(makeRun("job1", false, 1*time.Second, now.Add(-1*time.Minute)))

	a := history.New(s)
	stats := a.StatsFor("job1")

	if stats.TotalRuns != 3 {
		t.Errorf("expected 3 total runs, got %d", stats.TotalRuns)
	}
	if stats.SuccessCount != 2 {
		t.Errorf("expected 2 successes, got %d", stats.SuccessCount)
	}
	if stats.FailureCount != 1 {
		t.Errorf("expected 1 failure, got %d", stats.FailureCount)
	}
	if stats.MaxDuration != 4*time.Second {
		t.Errorf("expected max 4s, got %v", stats.MaxDuration)
	}
	expectedAvg := (2 + 4 + 1) * time.Second / 3
	if stats.AvgDuration != expectedAvg {
		t.Errorf("expected avg %v, got %v", expectedAvg, stats.AvgDuration)
	}
	if stats.SuccessRate < 66.6 || stats.SuccessRate > 66.8 {
		t.Errorf("expected ~66.7%% success rate, got %f", stats.SuccessRate)
	}
	if stats.LastRun == nil || stats.LastRun.StartedAt != now.Add(-1*time.Minute) {
		t.Error("LastRun should be the most recent run")
	}
}

func TestAllStats(t *testing.T) {
	s := store.New(100)
	now := time.Now()
	s.Record(makeRun("alpha", true, time.Second, now))
	s.Record(makeRun("beta", false, time.Second, now))

	a := history.New(s)
	all := a.AllStats()

	if len(all) != 2 {
		t.Fatalf("expected 2 job stats, got %d", len(all))
	}
}
