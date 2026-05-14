// Package webhook implements outbound HTTP webhook dispatch for cronlens.
//
// A Dispatcher is configured with one or more Config entries, each specifying
// a target URL, the events that should trigger a notification ("success",
// "failure", or "all"), and optional HTTP headers (e.g. for authentication).
//
// Example usage:
//
//	cfgs := []webhook.Config{
//		{
//			URL:    "https://hooks.example.com/cronlens",
//			Events: []string{"failure"},
//			Headers: map[string]string{"Authorization": "Bearer token"},
//		},
//	}
//	d := webhook.New(cfgs)
//	// later, after a job run completes:
//	d.Notify(run)
//
// The Dispatcher sends a JSON payload containing the job name, status,
// start time, duration, and any error message to every matching endpoint.
package webhook
