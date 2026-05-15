package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cronlens/internal/jobpriority"
)

// NewJobPriorityHandler returns an http.Handler that exposes job priority
// management over HTTP.
//
//   GET  /api/priority          — list all explicitly set priorities
//   GET  /api/priority/{job}    — get priority for a single job
//   POST /api/priority/{job}    — set priority for a job
//   DELETE /api/priority/{job}  — remove priority for a job
func NewJobPriorityHandler(reg *jobpriority.Registry) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/priority", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		all := reg.All()
		out := make(map[string]string, len(all))
		for job, lvl := range all {
			out[job] = lvl.String()
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("/api/priority/", func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimPrefix(r.URL.Path, "/api/priority/")
		if job == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			lvl := reg.Get(job)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"job": job, "priority": lvl.String()})

		case http.MethodPost:
			var body struct {
				Level int `json:"level"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := reg.Set(job, jobpriority.Level(body.Level)); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			reg.Delete(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
