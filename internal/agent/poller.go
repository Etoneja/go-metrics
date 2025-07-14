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

func (p *Poller) poll(ctx context.Context) {
	p.iteration++
	log.Println("Poll - start iteration", p.iteration)
	err := p.stats.collect(ctx)
	if err != nil {
		log.Printf("Error while polling: %v", err)
	}
}

func (p *Poller) runRoutine(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(p.pollInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func newPoller(stats *Stats, pollInterval time.Duration) *Poller {
	return &Poller{
		stats:        stats,
		pollInterval: pollInterval,
	}
}
