package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := server.PrepareConfig()

	logger.Init(false)
	defer logger.Sync()

	storageConfig := &server.StorageConfig{
		StoreInterval:   cfg.StoreInterval,
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	serverErrChan := make(chan error, 1)

	store := server.NewMemStorageFromStorageConfig(storageConfig)
	router := server.NewRouter(store)

	logger.Get().Info("Server started",
		zap.String("ServerAddress", cfg.ServerAddress),
		zap.Uint("StoreInterval", cfg.StoreInterval),
		zap.String("FileStoragePath", cfg.FileStoragePath),
		zap.Bool("Restore", cfg.Restore),
	)

	go func() {
		err := http.ListenAndServe(cfg.ServerAddress, router)
		if err != nil {
			logger.Get().Error("Server failed",
				zap.Error(err),
			)
			serverErrChan <- err
		}
	}()

	select {
	case sig := <-signalChan:
		logger.Get().Info("Received signal",
			zap.String("signal", sig.String()),
		)
		store.ShutDown()
	case err := <-serverErrChan:
		logger.Get().Info("Server error",
			zap.Error(err),
		)
		store.ShutDown()
	}

}
