package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/etoneja/go-metrics/internal/agent"
	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/version"
	"go.uber.org/zap"
)

func main() {
	version.Print()

	cfg := agent.PrepareConfig()

	logger.Init(false)
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Get().Warn("failed to sync logger", zap.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Get().Info("Agent started",
		zap.String("ServerEndpoint", cfg.ServerEndpoint),
		zap.Uint("PollInterval", cfg.PollInterval),
		zap.Uint("ReportInterval", cfg.ReportInterval),
		zap.Uint("RateLimit", cfg.RateLimit),
	)

	service := agent.NewService(cfg)
	err := service.Run(ctx)
	if err != nil {
		logger.Get().Info("Service stopped", zap.Error(err))
	}
}
