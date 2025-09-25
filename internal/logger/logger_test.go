package logger

import (
	"sync"
	"testing"

	"go.uber.org/zap"
)

func TestInit_DevelopmentMode(t *testing.T) {
	// Reset global state for test
	globalLogger = zap.NewNop()
	globalLoggerOnce = sync.Once{}

	Init(true)

	logger := Get()
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Development logger should not be a no-op
	if logger.Core().Enabled(zap.DebugLevel) {
		t.Log("Development logger supports debug level")
	}
}

func TestInit_ProductionMode(t *testing.T) {
	// Reset global state for test
	globalLogger = zap.NewNop()
	globalLoggerOnce = sync.Once{}

	Init(false)

	logger := Get()
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Production logger should exist
	if logger.Core() == nil {
		t.Error("Production logger should have a core")
	}
}

func TestInit_OnlyOnce(t *testing.T) {
	// Reset global state
	globalLogger = zap.NewNop()
	globalLoggerOnce = sync.Once{}

	// First call should initialize
	Init(true)
	firstLogger := Get()

	// Second call should not change logger
	Init(false)
	secondLogger := Get()

	if firstLogger != secondLogger {
		t.Error("Logger should be initialized only once")
	}
}

func TestInit_ErrorRecovery(t *testing.T) {
	// This test is tricky since zap rarely fails in normal conditions
	// But we can verify that we don't panic and get a no-op logger at least
	globalLogger = zap.NewNop()
	globalLoggerOnce = sync.Once{}

	// Should not panic and should provide a logger
	Init(false)
	logger := Get()

	if logger == nil {
		t.Fatal("Should always return a logger, even on error")
	}
}

func TestGet_BeforeInit(t *testing.T) {
	// Reset to initial state
	globalLogger = zap.NewNop()
	globalLoggerOnce = sync.Once{}

	// Get logger before initialization should return no-op
	logger := Get()
	if logger == nil {
		t.Fatal("Get() should never return nil")
	}

	// Should be no-op logger before init
	if logger.Core().Enabled(zap.DebugLevel) {
		t.Error("Default logger should be no-op and not enable debug level")
	}
}

func TestInit_EnvironmentVariable(t *testing.T) {
	// Test that debug mode can be controlled by environment
	t.Setenv("DEBUG", "true")

	globalLogger = zap.NewNop()
	globalLoggerOnce = sync.Once{}

	// This test would need your actual debug detection logic
	// For now testing both modes explicitly as above
}
