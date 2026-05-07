// Package notifier provides alerting for failed cron jobs.
package notifier

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/user/cronlens/internal/job"
)

// Notifier sends alerts when jobs fail.
type Notifier struct {
	threshold int
	writer    io.Writer
	logger    *log.Logger
}

// New creates a Notifier that alerts after a job fails threshold times consecutively.
func New(threshold int) *Notifier {
	w := os.Stderr
	return &Notifier{
		threshold: threshold,
		writer:    w,
		logger:    log.New(w, "[cronlens] ", 0),
	}
}

// NewWithWriter creates a Notifier writing alerts to the given writer (useful for testing).
func NewWithWriter(threshold int, w io.Writer) *Notifier {
	return &Notifier{
		threshold: threshold,
		writer:    w,
		logger:    log.New(w, "[cronlens] ", 0),
	}
}

// Check inspects recent runs for a job and emits an alert if consecutive
// failures meet or exceed the configured threshold.
func (n *Notifier) Check(jobName string, runs []job.Run) {
	if len(runs) == 0 {
		return
	}

	consecutiveFails := 0
	for i := len(runs) - 1; i >= 0; i-- {
		if !runs[i].Success {
			consecutiveFails++
		} else {
			break
		}
	}

	if consecutiveFails >= n.threshold {
		last := runs[len(runs)-1]
		n.logger.Printf(
			ALERT: job %q has failed %d consecutive time(s). Last failure at %s (exit: %q)\n",
			jobName,
			consecutiveFails,
			last.FinishedAt.Format(time.RFC3339),
			last.ExitError,
		)
	}
}

// Alert formats and writes a plain alert message directly.
func (n *Notifier) Alert(msg string) {
	fmt.Fprintln(n.writer, "[cronlens] ALERT: "+msg)
}
