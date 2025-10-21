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

func NewService(cfg *config, publicKey *rsa.PublicKey) (*service, error) {
	stats := newStats()

	pollDuration := time.Second * time.Duration(cfg.PollInterval)
	poller := newPoller(stats, pollDuration)

	reporter, err := newReporter(stats, cfg, publicKey)
	if err != nil {
		return nil, err
	}

	return &service{
		stats:    stats,
		poller:   poller,
		reporter: reporter,
	}, nil

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
