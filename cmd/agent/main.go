package main

import (
	"github.com/etoneja/go-metrics/internal/agent"
	"github.com/etoneja/go-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := agent.PrepareConfig()

	logger.Init(false)
	defer logger.Sync()

	logger.Get().Info("Agent started",
		zap.String("ServerEndpoint", cfg.ServerEndpoint),
		zap.Uint("PollInterval", cfg.PollInterval),
		zap.Uint("ReportInterval", cfg.ReportInterval),
	)

	service := agent.NewService(cfg)
	service.Run()
}
