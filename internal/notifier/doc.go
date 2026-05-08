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
//
// Consecutive failure counting
//
// The Notifier inspects runs in reverse-chronological order (most recent
// first). It increments an internal counter for each failed run and resets
// the counter as soon as a successful run is encountered. An alert is
// emitted only once per threshold crossing — subsequent calls to Check will
// not repeat the alert unless the counter is reset by an intervening
// success.
package notifier
