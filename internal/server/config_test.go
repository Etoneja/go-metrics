package server

import (
	"flag"
	"os"
	"testing"
)

// TestPrepareConfig_EnvPrecedence tests that environment variables override flag values
func TestPrepareConfig_EnvPrecedence(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		envVars        map[string]string
		expectedConfig config
	}{
		{
			name: "env_overrides_flags",
			args: []string{"-a", "flag-value:8080", "-i", "100", "-f", "flag-file.json", "-r", "true", "-d", "flag-dsn", "-k", "flag-key"},
			envVars: map[string]string{
				"ADDRESS":           "env-value:8080",
				"STORE_INTERVAL":    "200",
				"FILE_STORAGE_PATH": "env-file.json",
				"RESTORE":           "false",
				"DATABASE_DSN":      "env-dsn",
				"KEY":               "env-key",
			},
			expectedConfig: config{
				ServerAddress:   "env-value:8080", // ENV overrides flag
				StoreInterval:   200,              // ENV overrides flag
				FileStoragePath: "env-file.json",  // ENV overrides flag
				Restore:         false,            // ENV overrides flag
				DatabaseDSN:     "env-dsn",        // ENV overrides flag
				HashKey:         "env-key",        // ENV overrides flag
			},
		},
		{
			name:    "flags_when_no_env",
			args:    []string{"-a", "flag-value:8080", "-i", "100", "-f", "flag-file.json", "-r", "-d", "flag-dsn", "-k", "flag-key"},
			envVars: map[string]string{}, // Empty env
			expectedConfig: config{
				ServerAddress:   "flag-value:8080", // Flag value used
				StoreInterval:   100,               // Flag value used
				FileStoragePath: "flag-file.json",  // Flag value used
				Restore:         true,              // Flag value used
				DatabaseDSN:     "flag-dsn",        // Flag value used
				HashKey:         "flag-key",        // Flag value used
			},
		},
		{
			name:    "defaults_when_no_flags_no_env",
			args:    []string{},          // No flags
			envVars: map[string]string{}, // No env
			expectedConfig: config{
				ServerAddress:   "localhost:8080", // Default value
				StoreInterval:   300,              // Default value
				FileStoragePath: "data.json",      // Default value
				Restore:         false,            // Default value
				DatabaseDSN:     "",               // Default value
				HashKey:         "",               // Default value
			},
		},
		{
			name: "mixed_env_and_flags",
			args: []string{"-a", "flag-value:8080", "-i", "100"}, // Some flags
			envVars: map[string]string{
				"FILE_STORAGE_PATH": "env-file.json", // Some env
				"KEY":               "env-key",       // Some env
			},
			expectedConfig: config{
				ServerAddress:   "flag-value:8080", // Flag (no env override)
				StoreInterval:   100,               // Flag (no env override)
				FileStoragePath: "env-file.json",   // ENV overrides default
				Restore:         false,             // Default (no flag, no env)
				DatabaseDSN:     "",                // Default (no flag, no env)
				HashKey:         "env-key",         // ENV overrides default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set test arguments
			os.Args = append([]string{"test"}, tt.args...)

			// Reset flags for clean test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Get config - env should override flags
			cfg, err := PrepareConfig()
			if err != nil {
				t.Fatalf("PrepareConfig failed: %v", err)
			}

			// Verify env precedence
			if cfg.ServerAddress != tt.expectedConfig.ServerAddress {
				t.Errorf("ServerAddress = %s, want %s", cfg.ServerAddress, tt.expectedConfig.ServerAddress)
			}
			if cfg.StoreInterval != tt.expectedConfig.StoreInterval {
				t.Errorf("StoreInterval = %d, want %d", cfg.StoreInterval, tt.expectedConfig.StoreInterval)
			}
			if cfg.FileStoragePath != tt.expectedConfig.FileStoragePath {
				t.Errorf("FileStoragePath = %s, want %s", cfg.FileStoragePath, tt.expectedConfig.FileStoragePath)
			}
			if cfg.Restore != tt.expectedConfig.Restore {
				t.Errorf("Restore = %t, want %t", cfg.Restore, tt.expectedConfig.Restore)
			}
			if cfg.DatabaseDSN != tt.expectedConfig.DatabaseDSN {
				t.Errorf("DatabaseDSN = %s, want %s", cfg.DatabaseDSN, tt.expectedConfig.DatabaseDSN)
			}
			if cfg.HashKey != tt.expectedConfig.HashKey {
				t.Errorf("HashKey = %s, want %s", cfg.HashKey, tt.expectedConfig.HashKey)
			}
		})
	}
}

// TestPrepareConfig_BoolEnv tests boolean environment variable parsing
func TestPrepareConfig_BoolEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"bool_true", "true", true},
		{"bool_false", "false", false},
		{"bool_1", "1", true},
		{"bool_0", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset environment and flags
			os.Args = []string{"test"}
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			t.Setenv("RESTORE", tt.envValue)

			cfg, err := PrepareConfig()
			if err != nil {
				t.Fatalf("PrepareConfig failed: %v", err)
			}

			if cfg.Restore != tt.expected {
				t.Errorf("Restore = %t, want %t for env value %s", cfg.Restore, tt.expected, tt.envValue)
			}
		})
	}
}
