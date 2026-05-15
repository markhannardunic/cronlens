// Package pinned provides a thread-safe registry for marking cron jobs as
// "pinned" — a way to flag jobs as critical so that monitoring surfaces and
// dashboards can give them elevated visibility.
//
// Typical usage:
//
//	reg := pinned.New()
//	_ = reg.Pin("nightly-backup")
//
//	if reg.IsPinned("nightly-backup") {
//		// render with priority badge
//	}
//
// Pinning a job that is already pinned, or unpinning a job that is not pinned,
// returns a descriptive error so callers can distinguish state conflicts from
// unexpected failures.
package pinned
