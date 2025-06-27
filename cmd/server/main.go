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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	serverErrChan := make(chan error, 1)

	var store server.Storager
	if cfg.DatabaseDSN == "" {
		logger.Get().Info("Init memstorage")
		storageConfig := &server.StorageConfig{
			StoreInterval:   cfg.StoreInterval,
			FileStoragePath: cfg.FileStoragePath,
			Restore:         cfg.Restore,
		}
		store = server.NewMemStorageFromStorageConfig(storageConfig)
	} else {
		logger.Get().Info("Init dbstorage")
		store = server.NewDBStorage(cfg.DatabaseDSN)
	}

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
	case err := <-serverErrChan:
		logger.Get().Info("Server error",
			zap.Error(err),
		)
	}

	store.ShutDown()
	logger.Get().Info("Server stopped")

}
