// Package scheduler manages the lifecycle of registered cron jobs.
//
// It wraps a cron runner (github.com/robfig/cron/v3) and exposes a simple
// API to register named jobs with standard cron expressions. Each job is
// backed by a collector.Collector which handles execution and result
// recording.
//
// Usage:
//
//	s := scheduler.New()
//	s.Register("health-check", "@every 1m", myCollector)
//	s.Start()
//	defer s.Stop()
package scheduler
