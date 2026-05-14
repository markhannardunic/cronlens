// Package webhook provides outbound HTTP webhook notifications
// for job run events such as failures or completions.
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cronlens/internal/job"
)

// Config holds the configuration for a single webhook endpoint.
type Config struct {
	URL     string            `json:"url"`
	Events  []string          `json:"events"` // "failure", "success", "all"
	Headers map[string]string `json:"headers,omitempty"`
}

// Payload is the JSON body sent to the webhook endpoint.
type Payload struct {
	JobName   string        `json:"job_name"`
	Status    string        `json:"status"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration_ms"`
	Error     string        `json:"error,omitempty"`
}

// Dispatcher sends webhook notifications based on job run outcomes.
type Dispatcher struct {
	client  *http.Client
	configs []Config
}

// New creates a new Dispatcher with the given webhook configs.
func New(configs []Config) *Dispatcher {
	return &Dispatcher{
		client:  &http.Client{Timeout: 5 * time.Second},
		configs: configs,
	}
}

// NewWithClient creates a Dispatcher with a custom HTTP client (useful for testing).
func NewWithClient(configs []Config, client *http.Client) *Dispatcher {
	return &Dispatcher{client: client, configs: configs}
}

// Notify sends webhook notifications for the given run to all matching configs.
func (d *Dispatcher) Notify(r job.Run) error {
	var errs []error
	for _, cfg := range d.configs {
		if !d.shouldSend(cfg, r) {
			continue
		}
		if err := d.send(cfg, r); err != nil {
			errs = append(errs, fmt.Errorf("webhook %s: %w", cfg.URL, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("webhook dispatch errors: %v", errs)
	}
	return nil
}

func (d *Dispatcher) shouldSend(cfg Config, r job.Run) bool {
	for _, ev := range cfg.Events {
		if ev == "all" {
			return true
		}
		if ev == "failure" && !r.Success {
			return true
		}
		if ev == "success" && r.Success {
			return true
		}
	}
	return false
}

func (d *Dispatcher) send(cfg Config, r job.Run) error {
	status := "success"
	if !r.Success {
		status = "failure"
	}
	p := Payload{
		JobName:   r.JobName,
		Status:    status,
		StartedAt: r.StartedAt,
		Duration:  r.Duration,
		Error:     r.Err,
	}
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, cfg.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
