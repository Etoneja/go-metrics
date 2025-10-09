package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/etoneja/go-metrics/internal/agent"
	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/version"
	"go.uber.org/zap"
)

func main() {
	version.Print()

	logger.Init(false)

	cfg, err := agent.PrepareConfig()
	if err != nil {
		logger.Get().Fatal("Failed prepare config", zap.Error(err))
	}

	publicKey, err := common.LoadPublicKey(cfg.CryptoKey)
	if err != nil {
		logger.Get().Fatal("Failed to load public key:", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	logger.Get().Info("Agent started",
		zap.String("ServerEndpoint", cfg.ServerEndpoint),
		zap.Uint("PollInterval", cfg.PollInterval),
		zap.Uint("ReportInterval", cfg.ReportInterval),
		zap.Uint("RateLimit", cfg.RateLimit),
		zap.String("CryptoKey", cfg.CryptoKey),
		zap.String("ConfigFile", cfg.ConfigFile),
	)

	service := agent.NewService(cfg, publicKey)
	err = service.Run(ctx)
	if err != nil {
		logger.Get().Info("Service stopped", zap.Error(err))
	}
}
