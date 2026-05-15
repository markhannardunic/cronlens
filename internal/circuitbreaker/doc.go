// Package circuitbreaker implements a per-job circuit breaker pattern for
// cronlens. When a job exceeds a configured number of consecutive failures the
// circuit trips to the open state, causing subsequent Allow calls to return
// ErrOpen until the reset timeout elapses.
//
// States:
//
//	closed   – normal operation; all runs are permitted.
//	open     – circuit tripped; runs are rejected with ErrOpen.
//	half_open – probe window; one run is permitted to test recovery.
//
// Usage:
//
//	br := circuitbreaker.New()
//	br.Register(circuitbreaker.Rule{
//	    JobName:      "backup",
//	    Threshold:    3,
//	    ResetTimeout: 5 * time.Minute,
//	})
//
//	if err := br.Allow("backup"); err != nil {
//	    // skip execution
//	}
//	// after run completes:
//	br.RecordSuccess("backup") // or br.RecordFailure("backup")
package circuitbreaker
