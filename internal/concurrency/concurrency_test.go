package concurrency_test

import (
	"testing"

	"github.com/cronlens/internal/concurrency"
)

func TestActive_ZeroByDefault(t *testing.T) {
	tr := concurrency.New()
	if got := tr.Active("backup"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestAcquire_IncrementsCount(t *testing.T) {
	tr := concurrency.New()
	if err := tr.Acquire("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := tr.Active("backup"); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestRelease_DecrementsCount(t *testing.T) {
	tr := concurrency.New()
	_ = tr.Acquire("backup")
	tr.Release("backup")
	if got := tr.Active("backup"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestRelease_BelowZeroIsNoop(t *testing.T) {
	tr := concurrency.New()
	tr.Release("backup") // never acquired
	if got := tr.Active("backup"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestAcquire_NoLimit_AllowsMany(t *testing.T) {
	tr := concurrency.New()
	for i := 0; i < 10; i++ {
		if err := tr.Acquire("backup"); err != nil {
			t.Fatalf("unexpected error on acquire %d: %v", i, err)
		}
	}
	if got := tr.Active("backup"); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
}

func TestAcquire_LimitEnforced(t *testing.T) {
	tr := concurrency.New()
	tr.SetLimit("report", 2)

	if err := tr.Acquire("report"); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	if err := tr.Acquire("report"); err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}
	if err := tr.Acquire("report"); err == nil {
		t.Fatal("expected error on third acquire, got nil")
	}
}

func TestAcquire_AfterRelease_AllowsAgain(t *testing.T) {
	tr := concurrency.New()
	tr.SetLimit("report", 1)

	_ = tr.Acquire("report")
	tr.Release("report")

	if err := tr.Acquire("report"); err != nil {
		t.Fatalf("expected acquire to succeed after release: %v", err)
	}
}

func TestSnapshot_ReflectsState(t *testing.T) {
	tr := concurrency.New()
	_ = tr.Acquire("alpha")
	_ = tr.Acquire("alpha")
	_ = tr.Acquire("beta")

	snap := tr.Snapshot()
	if snap["alpha"] != 2 {
		t.Fatalf("expected alpha=2, got %d", snap["alpha"])
	}
	if snap["beta"] != 1 {
		t.Fatalf("expected beta=1, got %d", snap["beta"])
	}
}
