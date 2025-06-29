package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/etoneja/go-metrics/internal/agent"
	"github.com/etoneja/go-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := agent.PrepareConfig()

	logger.Init(false)
	defer logger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Get().Info("Agent started",
		zap.String("ServerEndpoint", cfg.ServerEndpoint),
		zap.Uint("PollInterval", cfg.PollInterval),
		zap.Uint("ReportInterval", cfg.ReportInterval),
	)

	service := agent.NewService(cfg)
	err := service.Run(ctx)
	if err != nil {
		logger.Get().Info("Service stopped", zap.Error(err))
	}
}
