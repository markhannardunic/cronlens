// Package scheduler manages the lifecycle of registered cron jobs.
//
// It wraps a cron runner (github.com/robfig/cron/v3) and exposes a simple
// API to register named jobs with standard cron expressions. Each job is
// backed by a collector.Collector which handles execution and result
// recording.
//
// The scheduler supports graceful shutdown via Stop, which blocks until all
// currently running jobs have completed before returning.
//
// Job names must be unique within a scheduler instance. Attempting to
// register a duplicate name will return an error.
//
// Usage:
//
//	s := scheduler.New()
//	if err := s.Register("health-check", "@every 1m", myCollector); err != nil {
//		log.Fatal(err)
//	}
//	s.Start()
//	defer s.Stop()
package scheduler
