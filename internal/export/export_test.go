package export_test

import {
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronlens/internal/export"
	"github.com/yourorg/cronlens/internal/job"
	"github.com/yourorg/cronlens/internal/store"
)

func makeRun(name string, success bool, start time.Time, dur time.Duration) job.Run {
	r := job.NewRun(name)
	r.StartedAt = start
	r.FinishedAt = start.Add(dur)
	r.Success = success
	if !success {
		r.Err = fmt.Errorf("job failed")
	}
	return r
}

func newExporter(t *testing.T, runs ...job.Run) *export.Exporter {
	t.Helper()
	s := store.New()
	for _, r := range runs {
		s.Record(r)
	}
	return export.New(s)
}

func TestWriteJSON_Empty(t *testing.T) {
	ex := newExporter(t)
	var buf bytes.Buffer
	if err := ex.WriteJSON(&buf, nil, time.Time{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var runs []job.Run
	if err := json.Unmarshal(buf.Bytes(), &runs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestWriteJSON_WithRuns(t *testing.T) {
	now := time.Now()
	r1 := makeRun("backup", true, now.Add(-2*time.Minute), time.Second)
	r2 := makeRun("cleanup", false, now.Add(-time.Minute), 2*time.Second)
	ex := newExporter(t, r1, r2)
	var buf bytes.Buffer
	if err := ex.WriteJSON(&buf, nil, time.Time{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var runs []job.Run
	if err := json.Unmarshal(buf.Bytes(), &runs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestWriteCSV_Headers(t *testing.T) {
	ex := newExporter(t)
	var buf bytes.Buffer
	if err := ex.WriteCSV(&buf, nil, time.Time{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 1 {
		t.Fatal("expected at least a header line")
	}
	if !strings.Contains(lines[0], "job") || !strings.Contains(lines[0], "duration_ms") {
		t.Errorf("unexpected header: %s", lines[0])
	}
}

func TestWriteCSV_WithRuns(t *testing.T) {
	now := time.Now()
	r := makeRun("sync", true, now.Add(-time.Minute), 500*time.Millisecond)
	ex := newExporter(t, r)
	var buf bytes.Buffer
	if err := ex.WriteCSV(&buf, nil, time.Time{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// header + 1 data row
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "sync") {
		t.Errorf("expected job name in CSV row: %s", lines[1])
	}
}
