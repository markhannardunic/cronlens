// Package alert implements rule-based alerting for cronlens.
//
// An Evaluator holds a set of Rules and checks them against a job's run
// history each time it is invoked. Two rule types are supported:
//
//   - ConsecutiveFailures: fires when the tail of the history contains N or
//     more consecutive failed runs.
//   - MaxDuration: fires when the most recent run's duration exceeds the
//     configured threshold.
//
// Rules may be scoped to a specific job by setting JobName, or applied
// globally by leaving JobName empty.
//
// Example:
//
//	e := alert.NewEvaluator([]alert.Rule{
//	    {JobName: "backup", ConsecutiveFailures: 3},
//	    {MaxDuration: 10 * time.Minute},
//	})
//	alerts := e.Evaluate(history)
package alert
