package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/server"
	"github.com/etoneja/go-metrics/internal/version"
	"go.uber.org/zap"
)

func main() {
	version.Print()

	logger.Init(false)

	cfg, err := server.PrepareConfig()
	if err != nil {
		logger.Get().Fatal("Failed prepare config", zap.Error(err))
	}

	privateKey, err := common.LoadPrivateKey(cfg.CryptoKey)
	if err != nil {
		logger.Get().Fatal("Failed to load private key:", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	store := server.NewStorageFromConfig(cfg)

	logger.Get().Info("Server started",
		zap.String("ServerAddress", cfg.ServerAddress),
		zap.Uint("StoreInterval", cfg.StoreInterval),
		zap.String("FileStoragePath", cfg.FileStoragePath),
		zap.Bool("Restore", cfg.Restore),
		zap.String("CryptoKey", cfg.CryptoKey),
		zap.String("ConfigFile", cfg.ConfigFile),
		zap.String("TrustedSubnet", cfg.TrustedSubnet),
	)

	router := server.NewRouter(store, cfg.HashKey, privateKey, cfg.TrustedSubnet)
	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	serverErrChan := make(chan error, 1)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			logger.Get().Error("Server failed",
				zap.Error(err),
			)
			serverErrChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Get().Info("Received shutdown signal")
	case err := <-serverErrChan:
		logger.Get().Info("Server error",
			zap.Error(err),
		)
	}

	logger.Get().Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Get().Error("Server shutdown error",
			zap.Error(err),
		)
	}

	store.ShutDown()
	logger.Get().Info("Server stopped")

}
