package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := server.PrepareConfig()

	logger.Init(false)
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Get().Warn("failed to sync logger", zap.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store := server.NewStorageFromConfig(cfg)

	logger.Get().Info("Server started",
		zap.String("ServerAddress", cfg.ServerAddress),
		zap.Uint("StoreInterval", cfg.StoreInterval),
		zap.String("FileStoragePath", cfg.FileStoragePath),
		zap.Bool("Restore", cfg.Restore),
	)

	router := server.NewRouter(store, cfg.HashKey)
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
