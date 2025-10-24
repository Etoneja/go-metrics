package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestHTTPMetricClient_SendBatch_Success(t *testing.T) {
	receivedRequests := 0
	var lastRequest *http.Request
	var lastBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequests++
		lastRequest = r

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		gz, _ := gzip.NewReader(bytes.NewReader(body))
		decompressed, _ := io.ReadAll(gz)
		gz.Close()

		lastBody = decompressed
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &mockConfig{
		serverEndpoint: strings.TrimPrefix(server.URL, "http://"),
		serverProtocol: "http",
	}
	client := NewHTTPMetricClient(cfg)

	metrics := []models.MetricModel{
		{ID: "test1", MType: "gauge", Value: common.Float64Ptr(1.23)},
		{ID: "test2", MType: "counter", Delta: common.Int64Ptr(42)},
	}

	ctx := context.Background()
	err := client.SendBatch(ctx, metrics)

	assert.NoError(t, err)
	assert.Equal(t, 1, receivedRequests)
	assert.Equal(t, "POST", lastRequest.Method)
	assert.Equal(t, "/updates/", lastRequest.URL.Path)
	assert.Equal(t, "gzip", lastRequest.Header.Get("Content-Encoding"))
	assert.Equal(t, "application/json", lastRequest.Header.Get("Content-Type"))

	var receivedMetrics []models.MetricModel
	err = json.Unmarshal(lastBody, &receivedMetrics)
	assert.NoError(t, err)
	assert.Len(t, receivedMetrics, 2)
}

func TestHTTPMetricClient_SendBatch_WithHash(t *testing.T) {
	var lastHash string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastHash = r.Header.Get(common.HashHeaderKey)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &mockConfig{
		serverEndpoint: strings.TrimPrefix(server.URL, "http://"),
		serverProtocol: "http",
		hashKey:        "test-key",
	}
	client := NewHTTPMetricClient(cfg)

	metrics := []models.MetricModel{
		{ID: "test", MType: "gauge", Value: common.Float64Ptr(1.23)},
	}

	ctx := context.Background()
	err := client.SendBatch(ctx, metrics)

	assert.NoError(t, err)
	assert.NotEmpty(t, lastHash)
}

func TestHTTPMetricClient_SendBatch_WithRealIP(t *testing.T) {
	var lastRealIP string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastRealIP = r.Header.Get("X-Real-IP")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testIP := net.ParseIP("192.168.1.1")
	cfg := &mockConfig{
		serverEndpoint: strings.TrimPrefix(server.URL, "http://"),
		serverProtocol: "http",
		localIP:        testIP,
	}
	client := NewHTTPMetricClient(cfg)

	metrics := []models.MetricModel{
		{ID: "test", MType: "gauge", Value: common.Float64Ptr(1.23)},
	}

	ctx := context.Background()
	err := client.SendBatch(ctx, metrics)

	assert.NoError(t, err)
	assert.Equal(t, testIP.String(), lastRealIP)
}

func TestHTTPMetricClient_SendBatch_RetryOnServerError(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &mockConfig{
		serverEndpoint: strings.TrimPrefix(server.URL, "http://"),
		serverProtocol: "http",
	}
	client := NewHTTPMetricClient(cfg)

	originalBackoff := common.DefaultBackoffSchedule
	common.DefaultBackoffSchedule = []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
	}
	defer func() { common.DefaultBackoffSchedule = originalBackoff }()

	metrics := []models.MetricModel{
		{ID: "test", MType: "gauge", Value: common.Float64Ptr(1.23)},
	}

	ctx := context.Background()
	err := client.SendBatch(ctx, metrics)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestHTTPMetricClient_SendBatch_EmptyMetrics(t *testing.T) {
	cfg := &mockConfig{
		serverEndpoint: "localhost:8080",
		serverProtocol: "http",
	}
	client := NewHTTPMetricClient(cfg)

	ctx := context.Background()
	err := client.SendBatch(ctx, []models.MetricModel{})

	assert.NoError(t, err)
}

func TestHTTPMetricClient_Close(t *testing.T) {
	cfg := &mockConfig{
		serverEndpoint: "localhost:8080",
		serverProtocol: "http",
		rateLimit:      2,
	}
	client := NewHTTPMetricClient(cfg)

	assert.NotNil(t, client.semaphore)
	assert.Equal(t, 2, cap(client.semaphore))

	err := client.Close()
	assert.NoError(t, err)
	assert.True(t, client.closed)

	err = client.Close()
	assert.NoError(t, err)
}
