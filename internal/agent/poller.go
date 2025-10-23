package agent

import (
	"context"
	"time"

	"github.com/etoneja/go-metrics/internal/logger"
	"go.uber.org/zap"
)

type Poller struct {
	stats        *Stats
	iteration    uint
	pollInterval time.Duration
}

func (p *Poller) poll(ctx context.Context) {
	p.iteration++
	logger.Get().Info("Poll iteration started",
		zap.Uint("iteration", p.iteration),
	)
	err := p.stats.collect(ctx)
	if err != nil {
		logger.Get().Error("Error while polling", zap.Error(err))
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
