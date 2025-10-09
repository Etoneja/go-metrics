package common

import (
	"encoding/json"
	"os"
	"testing"
)

type TestConfig struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func TestLoadJSONConfig_EmptyPath(t *testing.T) {
	cfg := &TestConfig{}

	err := LoadJSONConfig(cfg, "")
	if err != nil {
		t.Errorf("Expected no error for empty path, got %v", err)
	}
}

func TestLoadJSONConfig_FileNotExists(t *testing.T) {
	cfg := &TestConfig{}

	err := LoadJSONConfig(cfg, "/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoadJSONConfig_InvalidJSON(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_invalid*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(`{"name": "test", "port": }`); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg := &TestConfig{}
	err = LoadJSONConfig(cfg, tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadJSONConfig_ValidJSON(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_valid*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	expectedConfig := TestConfig{Name: "test", Port: 8080}
	jsonData, err := json.Marshal(expectedConfig)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(tmpfile.Name(), jsonData, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &TestConfig{}
	err = LoadJSONConfig(cfg, tmpfile.Name())
	if err != nil {
		t.Errorf("Unexpected error for valid JSON: %v", err)
	}

	if cfg.Name != expectedConfig.Name || cfg.Port != expectedConfig.Port {
		t.Errorf("Config not loaded correctly. Expected %+v, got %+v", expectedConfig, cfg)
	}
}
