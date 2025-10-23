package server

import (
	"github.com/etoneja/go-metrics/internal/logger"
)

func NewStorageFromConfig(cfg *config) Storager {
	var store Storager
	if cfg.DatabaseDSN == "" {
		logger.Get().Info("Init memstorage")
		storageConfig := &StorageConfig{
			StoreInterval:   cfg.StoreInterval,
			FileStoragePath: cfg.FileStoragePath,
			Restore:         cfg.Restore,
		}
		store = NewMemStorageFromStorageConfig(storageConfig)
	} else {
		logger.Get().Info("Init memstorage")
		store = NewDBStorage(cfg.DatabaseDSN)
	}
	return store
}
