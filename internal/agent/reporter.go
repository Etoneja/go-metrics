package agent

import (
	"context"
	"log"
	"time"
)

type Reporter struct {
	cfg          *config
	stats        *Stats
	iteration    uint
	metricClient MetricClienter
}

func (r *Reporter) report(ctx context.Context) {
	r.iteration++
	log.Println("Report - start iteration", r.iteration)
	metrics := r.stats.GetMetrics()

	if len(metrics) == 0 {
		log.Println("No metrics. Skip report")
		return
	}

	err := r.metricClient.SendBatch(ctx, metrics)

	if err != nil {
		log.Printf("Error occurred sending metrcs %v", err)
	}

	log.Println("Report - finish iteration", r.iteration)
}

func (r *Reporter) stop() {
	err := r.metricClient.Close()
	if err != nil {
		log.Printf("Failed to close gRPC client: %v", err)
	}
}

func (r *Reporter) runRoutine(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * time.Duration(r.cfg.ReportInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.stop()
			return ctx.Err()
		case <-ticker.C:
			r.report(ctx)
		}
	}
}

func newReporter(stats *Stats, cfg *config, metricClient MetricClienter) *Reporter {
	return &Reporter{
		stats:        stats,
		cfg:          cfg,
		metricClient: metricClient,
	}
}
