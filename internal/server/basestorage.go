package server

import "log"

func NewStorageFromConfig(cfg *config) Storager {
	var store Storager
	if cfg.DatabaseDSN == "" {
		log.Println("Init memstorage")
		storageConfig := &StorageConfig{
			StoreInterval:   cfg.StoreInterval,
			FileStoragePath: cfg.FileStoragePath,
			Restore:         cfg.Restore,
		}
		store = NewMemStorageFromStorageConfig(storageConfig)
	} else {
		log.Println("Init dbstorage")
		store = NewDBStorage(cfg.DatabaseDSN)
	}
	return store
}
