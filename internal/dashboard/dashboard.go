// Package dashboard serves a minimal HTML dashboard for cronlens.
package dashboard

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/cronlens/internal/store"
)

//go:embed templates/index.html
var templateFS embed.FS

// Handler serves the HTML dashboard.
type Handler struct {
	store *store.Store
	tmpl  *template.Template
}

// NewHandler creates a new dashboard Handler.
func NewHandler(s *store.Store) (*Handler, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		return nil, err
	}
	return &Handler{store: s, tmpl: tmpl}, nil
}

type pageData struct {
	Jobs      []jobSummary
	Rendered  string
}

type jobSummary struct {
	Name      string
	LastRun   string
	Duration  string
	Status    string
	Failures  int
}

// ServeHTTP renders the dashboard page.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	names := h.store.JobNames()
	summaries := make([]jobSummary, 0, len(names))

	for _, name := range names {
		runs := h.store.Latest(name, 10)
		if len(runs) == 0 {
			continue
		}
		last := runs[0]
		status := "ok"
		if last.Error != "" {
			status = "error"
		}
		failures := 0
		for _, run := range runs {
			if run.Error != "" {
				failures++
			}
		}
		summaries = append(summaries, jobSummary{
			Name:     name,
			LastRun:  last.StartedAt.Format(time.RFC3339),
			Duration: last.Duration.Round(time.Millisecond).String(),
			Status:   status,
			Failures: failures,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.Execute(w, pageData{
		Jobs:     summaries,
		Rendered: time.Now().Format(time.RFC3339),
	})
}
