package agent

import (
	"context"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	cfg := &config{
		PollInterval:   2,
		ReportInterval: 10,
		RateLimit:      5,
		ServerEndpoint: "http://localhost:8080",
		HashKey:        "test-key",
	}

	service, err := NewService(cfg, nil)
	if err != nil {
		t.Fatalf("Expected service instance, got error: %v", err)
	}

	if service == nil {
		t.Fatal("Expected service instance, got nil")
	}

	if service.stats == nil {
		t.Error("Stats should be initialized")
	}

	if service.poller == nil {
		t.Error("Poller should be initialized")
	}

	if service.reporter == nil {
		t.Error("Reporter should be initialized")
	}

	if service.poller.pollInterval != 2*time.Second {
		t.Errorf("Expected poll interval 2s, got %v", service.poller.pollInterval)
	}
}

func TestService_Run_ContextCancel(t *testing.T) {
	cfg := &config{
		PollInterval:   100,
		ReportInterval: 100,
		ServerEndpoint: "http://test",
	}
	service, err := NewService(cfg, nil)
	if err != nil {
		t.Fatalf("Expected service instance, got error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = service.Run(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context canceled, got %v", err)
	}
}
