// Package export provides functionality to export job run history
// in common formats (JSON, CSV) for external consumption.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/yourorg/cronlens/internal/job"
	"github.com/yourorg/cronlens/internal/store"
)

// Exporter writes job run data to an io.Writer in a specified format.
type Exporter struct {
	store *store.Store
}

// New creates a new Exporter backed by the given store.
func New(s *store.Store) *Exporter {
	return &Exporter{store: s}
}

// WriteJSON encodes all runs for the given job names since t as JSON.
// If names is empty, all jobs are included.
func (e *Exporter) WriteJSON(w io.Writer, names []string, since time.Time) error {
	runs, err := e.collect(names, since)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(runs)
}

// WriteCSV encodes all runs for the given job names since t as CSV.
// If names is empty, all jobs are included.
func (e *Exporter) WriteCSV(w io.Writer, names []string, since time.Time) error {
	runs, err := e.collect(names, since)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"job", "started_at", "finished_at", "success", "error", "duration_ms"}); err != nil {
		return err
	}
	for _, r := range runs {
		durMs := r.FinishedAt.Sub(r.StartedAt).Milliseconds()
		errStr := ""
		if r.Err != nil {
			errStr = r.Err.Error()
		}
		row := []string{
			r.JobName,
			r.StartedAt.UTC().Format(time.RFC3339),
			r.FinishedAt.UTC().Format(time.RFC3339),
			fmt.Sprintf("%v", r.Success),
			errStr,
			fmt.Sprintf("%d", durMs),
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func (e *Exporter) collect(names []string, since time.Time) ([]job.Run, error) {
	if len(names) == 0 {
		names = e.store.JobNames()
	}
	var all []job.Run
	for _, name := range names {
		runs := e.store.Since(name, since)
		all = append(all, runs...)
	}
	return all, nil
}
