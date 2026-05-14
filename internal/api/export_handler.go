package api

import (
	"net/http"
	"time"

	"github.com/yourorg/cronlens/internal/export"
	"github.com/yourorg/cronlens/internal/store"
)

// ExportHandler serves job run data in JSON or CSV format.
type ExportHandler struct {
	exporter *export.Exporter
}

// NewExportHandler creates an ExportHandler backed by the given store.
func NewExportHandler(s *store.Store) *ExportHandler {
	return &ExportHandler{exporter: export.New(s)}
}

// ServeHTTP handles GET /api/export?format=json|csv&since=<RFC3339>&job=<name>
func (h *ExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	format := q.Get("format")
	if format == "" {
		format = "json"
	}

	var since time.Time
	if s := q.Get("since"); s != "" {
		parsed, err := time.Parse(time.RFC3339, s)
		if err != nil {
			http.Error(w, "invalid since parameter: use RFC3339", http.StatusBadRequest)
			return
		}
		since = parsed
	}

	names := q["job"]

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="cronlens-export.json"`)
		if err := h.exporter.WriteJSON(w, names, since); err != nil {
			http.Error(w, "export failed", http.StatusInternalServerError)
		}
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="cronlens-export.csv"`)
		if err := h.exporter.WriteCSV(w, names, since); err != nil {
			http.Error(w, "export failed", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "unsupported format; use json or csv", http.StatusBadRequest)
	}
}
