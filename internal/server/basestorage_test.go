package server

import (
	"testing"
)

func TestNewStorageFromConfig_MemStorage(t *testing.T) {
	cfg := &config{
		DatabaseDSN:     "",
		StoreInterval:   10,
		FileStoragePath: "/tmp/test",
		Restore:         false,
	}

	store := NewStorageFromConfig(cfg)

	_, ok := store.(*MemStorage)
	if !ok {
		t.Error("Expected MemStorage for empty DatabaseDSN")
	}
}
