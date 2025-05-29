package agent

import "time"

type service struct {
	stats    *Stats
	poller   *Poller
	reporter *Reporter
}

func NewService(cfg *config) *service {
	stats := newStats()

	pollDuration := time.Second * time.Duration(cfg.PollInterval)
	poller := newPoller(stats, pollDuration)

	reportDuration := time.Second * time.Duration(cfg.ReportInterval)
	reporter := newReporter(stats, cfg.ServerEndpoint, reportDuration)

	return &service{
		stats:    stats,
		poller:   poller,
		reporter: reporter,
	}

}

func (s *service) Run() {
	go s.poller.runRoutine()
	go s.reporter.runRoutine()

	select {}
}
