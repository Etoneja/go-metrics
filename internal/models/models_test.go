package models

import (
	"encoding/json"
	"testing"

	"github.com/etoneja/go-metrics/internal/common"
)

func TestNewMetricModel(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mtype    string
		delta    int64
		value    float64
		expected *MetricModel
	}{
		{
			name:  "counter_metric",
			id:    "test_counter",
			mtype: common.MetricTypeCounter,
			delta: 100,
			value: 0,
			expected: &MetricModel{
				ID:    "test_counter",
				MType: common.MetricTypeCounter,
				Delta: func() *int64 { v := int64(100); return &v }(),
				Value: func() *float64 { v := float64(0); return &v }(),
			},
		},
		{
			name:  "gauge_metric",
			id:    "test_gauge",
			mtype: common.MetricTypeGauge,
			delta: 0,
			value: 3.14,
			expected: &MetricModel{
				ID:    "test_gauge",
				MType: common.MetricTypeGauge,
				Delta: func() *int64 { v := int64(0); return &v }(),
				Value: func() *float64 { v := 3.14; return &v }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMetricModel(tt.id, tt.mtype, tt.delta, tt.value)

			if result.ID != tt.expected.ID {
				t.Errorf("ID = %s, want %s", result.ID, tt.expected.ID)
			}
			if result.MType != tt.expected.MType {
				t.Errorf("MType = %s, want %s", result.MType, tt.expected.MType)
			}
			if *result.Delta != *tt.expected.Delta {
				t.Errorf("Delta = %d, want %d", *result.Delta, *tt.expected.Delta)
			}
			if *result.Value != *tt.expected.Value {
				t.Errorf("Value = %f, want %f", *result.Value, *tt.expected.Value)
			}
		})
	}
}

func TestMetricModel_MarshalJSON_Counter(t *testing.T) {
	metric := &MetricModel{
		ID:    "test_counter",
		MType: common.MetricTypeCounter,
		Delta: func() *int64 { v := int64(42); return &v }(),
		Value: func() *float64 { v := 3.14; return &v }(), // This should be nil after marshal
	}

	data, err := json.Marshal(metric)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify Value is omitted and Delta is present
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled["value"] != nil {
		t.Error("Value field should be omitted for counter metrics")
	}
	if unmarshaled["delta"] == nil {
		t.Error("Delta field should be present for counter metrics")
	}
	if unmarshaled["delta"].(float64) != 42 {
		t.Errorf("Delta value = %f, want %d", unmarshaled["delta"].(float64), 42)
	}
}

func TestMetricModel_MarshalJSON_Gauge(t *testing.T) {
	metric := &MetricModel{
		ID:    "test_gauge",
		MType: common.MetricTypeGauge,
		Delta: func() *int64 { v := int64(42); return &v }(), // This should be nil after marshal
		Value: func() *float64 { v := 3.14; return &v }(),
	}

	data, err := json.Marshal(metric)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify Delta is omitted and Value is present
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled["delta"] != nil {
		t.Error("Delta field should be omitted for gauge metrics")
	}
	if unmarshaled["value"] == nil {
		t.Error("Value field should be present for gauge metrics")
	}
	if unmarshaled["value"].(float64) != 3.14 {
		t.Errorf("Value = %f, want %f", unmarshaled["value"].(float64), 3.14)
	}
}

func TestMetricModel_UnmarshalJSON_Counter(t *testing.T) {
	jsonData := `{"id":"test_counter","type":"counter","delta":42}`

	var metric MetricModel
	err := json.Unmarshal([]byte(jsonData), &metric)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if metric.ID != "test_counter" {
		t.Errorf("ID = %s, want %s", metric.ID, "test_counter")
	}
	if metric.MType != common.MetricTypeCounter {
		t.Errorf("MType = %s, want %s", metric.MType, common.MetricTypeCounter)
	}
	if metric.Delta == nil || *metric.Delta != 42 {
		t.Errorf("Delta = %v, want %d", metric.Delta, 42)
	}
	if metric.Value != nil {
		t.Error("Value should be nil for counter metrics")
	}
}

func TestMetricModel_UnmarshalJSON_Gauge(t *testing.T) {
	jsonData := `{"id":"test_gauge","type":"gauge","value":3.14}`

	var metric MetricModel
	err := json.Unmarshal([]byte(jsonData), &metric)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if metric.ID != "test_gauge" {
		t.Errorf("ID = %s, want %s", metric.ID, "test_gauge")
	}
	if metric.MType != common.MetricTypeGauge {
		t.Errorf("MType = %s, want %s", metric.MType, common.MetricTypeGauge)
	}
	if metric.Value == nil || *metric.Value != 3.14 {
		t.Errorf("Value = %v, want %f", metric.Value, 3.14)
	}
	if metric.Delta != nil {
		t.Error("Delta should be nil for gauge metrics")
	}
}

func TestMetricModel_UnmarshalJSON_InvalidCounter(t *testing.T) {
	jsonData := `{"id":"test_counter","type":"counter"}` // Missing delta

	var metric MetricModel
	err := json.Unmarshal([]byte(jsonData), &metric)
	if err == nil {
		t.Error("Expected error for counter without delta")
	}
}

func TestMetricModel_UnmarshalJSON_InvalidGauge(t *testing.T) {
	jsonData := `{"id":"test_gauge","type":"gauge"}` // Missing value

	var metric MetricModel
	err := json.Unmarshal([]byte(jsonData), &metric)
	if err == nil {
		t.Error("Expected error for gauge without value")
	}
}

func TestMetricModel_UnmarshalJSON_UnknownType(t *testing.T) {
	jsonData := `{"id":"test_unknown","type":"unknown_type","value":1.0}`

	var metric MetricModel
	err := json.Unmarshal([]byte(jsonData), &metric)
	if err == nil {
		t.Error("Expected error for unknown metric type")
	}
}

func TestMetricModel_RoundTrip(t *testing.T) {
	original := &MetricModel{
		ID:    "test_roundtrip",
		MType: common.MetricTypeGauge,
		Value: func() *float64 { v := 123.456; return &v }(),
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back
	var restored MetricModel
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify data integrity
	if original.ID != restored.ID {
		t.Errorf("ID mismatch: %s != %s", original.ID, restored.ID)
	}
	if original.MType != restored.MType {
		t.Errorf("MType mismatch: %s != %s", original.MType, restored.MType)
	}
	if *original.Value != *restored.Value {
		t.Errorf("Value mismatch: %f != %f", *original.Value, *restored.Value)
	}
}
