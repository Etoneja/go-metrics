package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/etoneja/go-metrics/internal/common"
)

// ExampleBaseHandler_MetricBatchUpdateJSONHandler_basic demonstrates how to use the batch update endpoint.
//
// This example shows the complete workflow:
// 1. Preparing the JSON payload
// 2. Making the HTTP request
// 3. Handling the response
func ExampleBaseHandler_MetricBatchUpdateJSONHandler_basic() {

	storage := NewMemStorage()
	router := NewRouter(storage, "")
	server := httptest.NewServer(router)
	defer server.Close()

	metrics := []map[string]any{
		{
			"id":    "cpu_usage",
			"type":  "gauge",
			"value": 95.5,
		},
		{
			"id":    "memory_usage",
			"type":  "gauge",
			"value": 1024.0,
		},
		{
			"id":    "request_count",
			"type":  "counter",
			"delta": 42,
		},
	}
	jsonData, _ := json.Marshal(metrics)

	req, _ := http.NewRequest("POST", server.URL+"/updates/", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем выполнение запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var responseMetrics []map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseMetrics)
	if err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Updated Metrics Count: %d\n", len(responseMetrics))

	for _, metric := range responseMetrics {
		if metric["type"] == common.MetricTypeGauge {
			fmt.Printf("Metric: %s, Type: %s, Value: %v\n", metric["id"], metric["type"], metric["value"])
		} else {
			fmt.Printf("Metric: %s, Type: %s, Delta: %v\n", metric["id"], metric["type"], metric["delta"])
		}
	}

	// Output:
	// Status Code: 200
	// Content-Type: application/json
	// Updated Metrics Count: 3
	// Metric: cpu_usage, Type: gauge, Value: 95.5
	// Metric: memory_usage, Type: gauge, Value: 1024
	// Metric: request_count, Type: counter, Delta: 42

}

// ExampleMetricBatchUpdateJSONHandler_format shows the expected request format.
func ExampleBaseHandler_MetricBatchUpdateJSONHandler_format() {
	metrics := []map[string]any{
		{
			"id":    "temperature",
			"type":  "gauge",
			"value": 23.7,
		},
	}

	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Expected request format:")
	fmt.Println("POST /updates/")
	fmt.Println("Content-Type: application/json")
	fmt.Println("")
	fmt.Println(string(jsonData))

	// Output:
	// Expected request format:
	// POST /updates/
	// Content-Type: application/json
	//
	// [
	//   {
	//     "id": "temperature",
	//     "type": "gauge",
	//     "value": 23.7
	//   }
	// ]
}

// ExampleBaseHandler_MetricBatchUpdateJSONHandler_error demonstrates error handling.
func ExampleBaseHandler_MetricBatchUpdateJSONHandler_error() {
	handler := &BaseHandler{}
	httpHandler := handler.MetricBatchUpdateJSONHandler()

	server := httptest.NewServer(httpHandler)
	defer server.Close()

	invalidJSON := []byte(`{"invalid": "json"`)

	req, err := http.NewRequest("POST", server.URL, bytes.NewReader(invalidJSON))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Error Status Code: %d\n", resp.StatusCode)

	// Output:
	// Error Status Code: 400
}
