package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cronlens/internal/labeler"
)

// NewLabelerHandler returns an http.Handler that exposes label CRUD over HTTP.
//
// Routes (all under the prefix this handler is mounted at):
//
//	GET  /labels/{job}              — retrieve all labels for a job
//	POST /labels/{job}              — set a label  {"key":"...","value":"..."}
//	DELETE /labels/{job}/{key}      — remove a single label
//	GET  /labels?key=k&value=v      — list jobs matching a label
func NewLabelerHandler(reg *labeler.Registry) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/labels", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if key == "" {
			http.Error(w, "key query param required", http.StatusBadRequest)
			return
		}
		jobs := reg.JobsWithLabel(key, value)
		if jobs == nil {
			jobs = []string{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"jobs": jobs})
	})

	mux.HandleFunc("/labels/", func(w http.ResponseWriter, r *http.Request) {
		// path: /labels/{job} or /labels/{job}/{key}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/labels/"), "/")
		job := parts[0]
		if job == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			lbls := reg.Get(job)
			if lbls == nil {
				lbls = map[string]string{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(lbls)

		case http.MethodPost:
			var body struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := reg.Set(job, body.Key, body.Value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if len(parts) < 2 || parts[1] == "" {
				http.Error(w, "key required in path", http.StatusBadRequest)
				return
			}
			reg.Delete(job, parts[1])
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
