package server

import (
	"testing"

	"go.uber.org/zap"
)

func TestStartGRPCServer(t *testing.T) {
	store := NewMemStorage()
	logger := zap.NewNop()
	cfg := &config{ServerGRPCAddress: ":0"}
	errChan := make(chan error, 1)

	server, err := StartGRPCServer(store, logger, cfg, errChan)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if server == nil {
		t.Error("Expected server instance")
	}
	server.Stop()
}
