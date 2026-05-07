// Package notifier implements failure alerting for cronlens.
//
// A Notifier watches recent job runs and emits an alert when the number of
// consecutive failures reaches a configurable threshold. Alerts are written
// to stderr by default, making them easy to capture in system logs or pipe
// to external tooling.
//
// Basic usage:
//
//	n := notifier.New(3) // alert after 3 consecutive failures
//	n.Check("backup-job", runs)
//
// For testing or custom sinks, use NewWithWriter to redirect output:
//
//	n := notifier.NewWithWriter(3, myWriter)
package notifier
