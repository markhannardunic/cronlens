// Package alert provides configurable alerting rules for cron job failures.
package alert

import (
	"time"

	"github.com/yourorg/cronlens/internal/job"
)

// Rule defines conditions under which an alert should fire.
type Rule struct {
	// JobName is the cron job this rule applies to. Empty string matches all jobs.
	JobName string
	// ConsecutiveFailures triggers an alert after N consecutive failures.
	ConsecutiveFailures int
	// MaxDuration triggers an alert if a run exceeds this duration.
	MaxDuration time.Duration
}

// Alert represents a fired alert.
type Alert struct {
	JobName string
	Reason  string
	FiredAt time.Time
	Run     job.Run
}

// Evaluator checks runs against a set of rules and returns fired alerts.
type Evaluator struct {
	rules []Rule
}

// NewEvaluator creates an Evaluator with the given rules.
func NewEvaluator(rules []Rule) *Evaluator {
	return &Evaluator{rules: rules}
}

// Evaluate checks the most-recent run against all matching rules.
// history should be ordered oldest→newest for the same job.
func (e *Evaluator) Evaluate(history []job.Run) []Alert {
	if len(history) == 0 {
		return nil
	}
	latest := history[len(history)-1]
	var alerts []Alert

	for _, r := range e.rules {
		if r.JobName != "" && r.JobName != latest.JobName {
			continue
		}
		if r.ConsecutiveFailures > 0 {
			count := consecutiveFailures(history)
			if count >= r.ConsecutiveFailures {
				alerts = append(alerts, Alert{
					JobName: latest.JobName,
					Reason:  "consecutive failures threshold reached",
					FiredAt: time.Now(),
					Run:     latest,
				})
			}
		}
		if r.MaxDuration > 0 && latest.Duration >= r.MaxDuration {
			alerts = append(alerts, Alert{
				JobName: latest.JobName,
				Reason:  "run exceeded max duration",
				FiredAt: time.Now(),
				Run:     latest,
			})
		}
	}
	return alerts
}

func consecutiveFailures(history []job.Run) int {
	count := 0
	for i := len(history) - 1; i >= 0; i-- {
		if !history[i].Success {
			count++
		} else {
			break
		}
	}
	return count
}
