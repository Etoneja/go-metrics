package agent

import (
	"flag"
	"os"
	"testing"
)

// TestNormalizeConfig tests the endpoint URL normalization
func TestNormalizeConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no_protocol", "localhost:8080", "http://localhost:8080"},
		{"has_http", "http://server:8080", "http://server:8080"},
		{"has_https", "https://server:8080", "https://server:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{ServerEndpoint: tt.input}
			normalizeConfig(cfg)
			if cfg.ServerEndpoint != tt.expected {
				t.Errorf("normalizeConfig(%s) = %s, want %s", tt.input, cfg.ServerEndpoint, tt.expected)
			}
		})
	}
}

// TestPrepareConfigIntegration tests the complete config preparation with flags and env
func TestPrepareConfigIntegration(t *testing.T) {
	// Set environment variables
	t.Setenv("ADDRESS", "https://env-server:8080")
	t.Setenv("POLL_INTERVAL", "3")
	t.Setenv("KEY", "env-key")

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// env should override flags variables
	os.Args = []string{"test", "-a", "https://flag-server:9090", "-k", "flag-key"}

	// Reset flags for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg, err := PrepareConfig()
	if err != nil {
		t.Fatalf("PrepareConfig failed: %v", err)
	}

	// env should take precedence over flags
	if cfg.ServerEndpoint != "https://env-server:8080" {
		t.Errorf("ServerEndpoint = %s, want %s", cfg.ServerEndpoint, "https://env-server:8080")
	}
	if cfg.HashKey != "env-key" {
		t.Errorf("HashKey = %s, want %s", cfg.HashKey, "env-key")
	}
	// Env should remain where flag is not specified
	if cfg.PollInterval != 3 {
		t.Errorf("PollInterval = %d, want %d", cfg.PollInterval, 3)
	}
	// Default values should remain where neither flag nor env is specified
	if cfg.ReportInterval != 10 {
		t.Errorf("ReportInterval = %d, want %d", cfg.ReportInterval, 10)
	}
}
