package agent

import (
	"context"
	"time"

	"github.com/etoneja/go-metrics/internal/logger"
	"go.uber.org/zap"
)

type Reporter struct {
	cfg          *config
	stats        *Stats
	iteration    uint
	metricClient MetricClienter
}

func (r *Reporter) report(ctx context.Context) {
	r.iteration++
	logger.Get().Info("Report iteration started",
		zap.Uint("iteration", r.iteration),
	)
	metrics := r.stats.GetMetrics()

	if len(metrics) == 0 {
		logger.Get().Info("No metrics, skipping report")
		return
	}

	err := r.metricClient.SendBatch(ctx, metrics)
	if err != nil {
		logger.Get().Error("Error sending metrics", zap.Error(err))
	}

	logger.Get().Info("Report iteration finished",
		zap.Uint("iteration", r.iteration),
	)
}

func (r *Reporter) stop() {
	err := r.metricClient.Close()
	if err != nil {
		logger.Get().Error("Failed to close metric client", zap.Error(err))
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
