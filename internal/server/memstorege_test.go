package server

import (
	"context"
	"testing"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
)

func TestMemStorage_NewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	if storage == nil {
		t.Error("NewMemStorage returned nil")
	}
}

func TestMemStorage_GetGauge_NotFound(t *testing.T) {
	storage := NewMemStorage()
	_, err := storage.GetGauge(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent gauge")
	}
}

func TestMemStorage_SetGauge(t *testing.T) {
	storage := NewMemStorage()
	key := "test_gauge"
	value := 42.5

	result, err := storage.SetGauge(context.Background(), key, value)
	if err != nil {
		t.Errorf("SetGauge failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %f, got %f", value, result)
	}
}

func TestMemStorage_GetCounter_NotFound(t *testing.T) {
	storage := NewMemStorage()
	_, err := storage.GetCounter(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent counter")
	}
}

func TestMemStorage_IncrementCounter(t *testing.T) {
	storage := NewMemStorage()
	key := "test_counter"
	value := int64(10)

	result, err := storage.IncrementCounter(context.Background(), key, value)
	if err != nil {
		t.Errorf("IncrementCounter failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %d, got %d", value, result)
	}
}

func TestMemStorage_IncrementCounter_Existing(t *testing.T) {
	storage := NewMemStorage()
	key := "test_counter"

	// First increment
	storage.IncrementCounter(context.Background(), key, 5)

	// Second increment
	result, err := storage.IncrementCounter(context.Background(), key, 3)
	if err != nil {
		t.Errorf("IncrementCounter failed: %v", err)
	}
	if result != 8 {
		t.Errorf("Expected 8, got %d", result)
	}
}

func TestMemStorage_GetAll(t *testing.T) {
	storage := NewMemStorage()

	// Add some test data
	storage.SetGauge(context.Background(), "gauge1", 1.0)
	storage.IncrementCounter(context.Background(), "counter1", 1)

	metrics, err := storage.GetAll(context.Background())
	if err != nil {
		t.Errorf("GetAll failed: %v", err)
	}
	if len(metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(metrics))
	}
}

func TestMemStorage_Ping(t *testing.T) {
	storage := NewMemStorage()
	err := storage.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestMemStorage_Ping_ContextCancelled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := storage.Ping(ctx)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestMemStorage_ContextCancelled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := storage.GetGauge(ctx, "test")
	if err == nil {
		t.Error("Expected error for cancelled context in GetGauge")
	}

	_, err = storage.SetGauge(ctx, "test", 1.0)
	if err == nil {
		t.Error("Expected error for cancelled context in SetGauge")
	}

	_, err = storage.GetCounter(ctx, "test")
	if err == nil {
		t.Error("Expected error for cancelled context in GetCounter")
	}

	_, err = storage.IncrementCounter(ctx, "test", 1)
	if err == nil {
		t.Error("Expected error for cancelled context in IncrementCounter")
	}

	_, err = storage.GetAll(ctx)
	if err == nil {
		t.Error("Expected error for cancelled context in GetAll")
	}
}

func TestMemStorage_ShutDown(t *testing.T) {
	storage := NewMemStorage()
	storage.syncDump = true
	// Should not panic
	storage.ShutDown()
}

func TestMemStorage_BatchUpdate(t *testing.T) {
	storage := NewMemStorage()

	metrics := []models.MetricModel{
		*models.NewMetricModel("counter1", common.MetricTypeCounter, 5, 0),
		*models.NewMetricModel("gauge1", common.MetricTypeGauge, 0, 10.5),
	}

	result, err := storage.BatchUpdate(context.Background(), metrics)
	if err != nil {
		t.Errorf("BatchUpdate failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}
}

func TestMemStorage_BatchUpdate_InvalidType(t *testing.T) {
	storage := NewMemStorage()

	metrics := []models.MetricModel{
		{ID: "test", MType: "invalid"},
	}

	_, err := storage.BatchUpdate(context.Background(), metrics)
	if err == nil {
		t.Error("Expected error for invalid metric type")
	}
}

func TestNewMemStorageFromStorageConfig(t *testing.T) {
	config := &StorageConfig{
		StoreInterval:   5,
		FileStoragePath: "",
		Restore:         false,
	}

	storage := NewMemStorageFromStorageConfig(config)
	if storage == nil {
		t.Error("NewMemStorageFromStorageConfig returned nil")
	}
}

func TestMemStorage_GetSetIntegration(t *testing.T) {
	storage := NewMemStorage()

	// Test integration of set and get
	expectedGauge := 99.9
	expectedCounter := int64(42)

	storage.SetGauge(context.Background(), "test_gauge", expectedGauge)
	storage.IncrementCounter(context.Background(), "test_counter", expectedCounter)

	gaugeVal, err := storage.GetGauge(context.Background(), "test_gauge")
	if err != nil {
		t.Errorf("GetGauge failed: %v", err)
	}
	if gaugeVal != expectedGauge {
		t.Errorf("Expected gauge %f, got %f", expectedGauge, gaugeVal)
	}

	counterVal, err := storage.GetCounter(context.Background(), "test_counter")
	if err != nil {
		t.Errorf("GetCounter failed: %v", err)
	}
	if counterVal != expectedCounter {
		t.Errorf("Expected counter %d, got %d", expectedCounter, counterVal)
	}
}
