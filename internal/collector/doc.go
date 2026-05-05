// Package collector provides a Collector type that bridges job execution
// with the store layer. It records the start time, outcome, and duration
// of each job run — whether triggered by a shell command or a Go function —
// and persists the result via the embedded Store.
//
// Typical usage:
//
//	s := store.New()
//	c := collector.New(s)
//
//	// Wrap a Go function:
//	c.RunFunc("nightly-backup", func() error {
//		return backup.Run()
//	})
//
//	// Wrap a shell command:
//	c.Run("ping-check", "ping", "-c", "1", "example.com")
package collector
