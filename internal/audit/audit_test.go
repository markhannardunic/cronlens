package audit

import (
	"testing"
)

func TestRecord_And_All(t *testing.T) {
	l := New()

	if got := l.All(); len(got) != 0 {
		t.Fatalf("expected empty log, got %d entries", len(got))
	}

	l.Record("admin", ActionRegister, "backup-job", "cron: 0 * * * *")
	l.Record("admin", ActionPause, "backup-job", "")

	all := l.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Action != ActionRegister {
		t.Errorf("expected first action %q, got %q", ActionRegister, all[0].Action)
	}
	if all[1].Action != ActionPause {
		t.Errorf("expected second action %q, got %q", ActionPause, all[1].Action)
	}
}

func TestRecord_TimestampSet(t *testing.T) {
	l := New()
	l.Record("ci", ActionUpdate, "report-job", "changed schedule")
	all := l.All()
	if all[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestForTarget_FiltersCorrectly(t *testing.T) {
	l := New()
	l.Record("admin", ActionRegister, "job-a", "")
	l.Record("admin", ActionRegister, "job-b", "")
	l.Record("admin", ActionDelete, "job-a", "")

	results := l.ForTarget("job-a")
	if len(results) != 2 {
		t.Fatalf("expected 2 entries for job-a, got %d", len(results))
	}
	for _, e := range results {
		if e.Target != "job-a" {
			t.Errorf("unexpected target %q in filtered results", e.Target)
		}
	}
}

func TestForTarget_NoMatch(t *testing.T) {
	l := New()
	l.Record("admin", ActionRegister, "job-x", "")

	results := l.ForTarget("job-z")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	l := New()
	l.Record("admin", ActionResume, "job-a", "")

	a := l.All()
	a[0].Actor = "tampered"

	b := l.All()
	if b[0].Actor == "tampered" {
		t.Error("All() should return a copy, not a reference")
	}
}
