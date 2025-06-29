package agent

import (
	"context"
	"log"
	"time"
)

type Poller struct {
	stats        *Stats
	iteration    uint
	pollInterval time.Duration
}

func (p *Poller) poll() {
	p.iteration++
	log.Println("Poll - start iteration", p.iteration)
	p.stats.collect()
}

func (p *Poller) runRoutine(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(p.pollInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			p.poll()
		}
	}
}

func newPoller(stats *Stats, pollInterval time.Duration) *Poller {
	return &Poller{
		stats:        stats,
		pollInterval: pollInterval,
	}
}
