// Package retention provides configurable pruning of old job run records
// from the store, keeping the dataset bounded over time.
package retention

import (
	"log"
	"time"

	"github.com/user/cronlens/internal/store"
)

// Pruner periodically removes runs older than a configured age.
type Pruner struct {
	store    *store.Store
	maxAge   time.Duration
	interval time.Duration
	stop     chan struct{}
}

// New creates a Pruner that deletes runs older than maxAge, checked every interval.
func New(s *store.Store, maxAge, interval time.Duration) *Pruner {
	return &Pruner{
		store:    s,
		maxAge:   maxAge,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start begins the background pruning loop. Call Stop to halt it.
func (p *Pruner) Start() {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				n := p.Prune()
				if n > 0 {
					log.Printf("retention: pruned %d run(s) older than %s", n, p.maxAge)
				}
			case <-p.stop:
				return
			}
		}
	}()
}

// Stop halts the background pruning loop.
func (p *Pruner) Stop() {
	close(p.stop)
}

// Prune removes all runs whose start time is older than maxAge and returns
// the number of runs deleted.
func (p *Pruner) Prune() int {
	cutoff := time.Now().Add(-p.maxAge)
	return p.store.DeleteBefore(cutoff)
}
