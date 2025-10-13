package agent

import (
	"context"
	"crypto/rsa"
	"time"

	"golang.org/x/sync/errgroup"
)

type service struct {
	stats    *Stats
	poller   *Poller
	reporter *Reporter
}

func NewService(cfg *config, publicKey *rsa.PublicKey) *service {
	stats := newStats()

	pollDuration := time.Second * time.Duration(cfg.PollInterval)
	poller := newPoller(stats, pollDuration)

	reporter := newReporter(stats, cfg, publicKey)

	return &service{
		stats:    stats,
		poller:   poller,
		reporter: reporter,
	}

}

func (s *service) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.poller.runRoutine(ctx)
	})

	g.Go(func() error {
		return s.reporter.runRoutine(ctx)
	})

	return g.Wait()
}
