package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	mu       *sync.Mutex
	requests []*http.Request
	doFunc   func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req)

	if m.doFunc != nil {
		return m.doFunc(req)
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("fake")),
	}, nil
}

func (m *mockClient) Close() {}

func TestReporter_report(t *testing.T) {
	fakeReportInterval := time.Duration(time.Millisecond)
	fakeEndpoint := "http://fake.com/"
	stats := newStats()

	mockCli := &mockClient{mu: &sync.Mutex{}}
	t.Run("test report", func(t *testing.T) {
		r := &Reporter{
			stats:          stats,
			client:         mockCli,
			endpoint:       fakeEndpoint,
			reportInterval: fakeReportInterval,
		}
		assert.Equal(t, uint(0), r.iteration)

		ctx := context.Background()

		// stats not collected
		r.report(ctx)

		assert.Equal(t, uint(1), r.iteration)
		assert.Equal(t, 0, len(mockCli.requests))

		err := stats.collect(ctx)
		if err != nil {
			t.Fatalf("Unexpected err: %v", err)
		}

		// stats collected
		r.report(ctx)

		assert.Equal(t, 1, len(mockCli.requests))

		req := mockCli.requests[0]
		assert.Equal(t, req.URL.Path, "/updates/")
		assert.Equal(t, http.MethodPost, req.Method)
	})

}

func TestNewReporter(t *testing.T) {
	stats := &Stats{}
	endpoint := "http://localhost:8080"
	reportInterval := 10 * time.Second
	rateLimit := uint(5)
	hashKey := "test-key"

	cfg := &config{
		ServerEndpoint: endpoint,
		ServerProtocol: "http",
		ReportInterval: 10,
		RateLimit:      rateLimit,
		HashKey:        hashKey,
	}

	reporter, err := newReporter(stats, cfg, nil)
	if err != nil {
		t.Fatalf("Expected reporter instance, got error: %v", err)
	}

	if reporter == nil {
		t.Fatal("Expected reporter instance, got nil")
	}

	if reporter.stats != stats {
		t.Error("Stats not set correctly")
	}

	if reporter.endpoint != endpoint {
		t.Errorf("Expected endpoint %s, got %s", endpoint, reporter.endpoint)
	}

	if reporter.reportInterval != reportInterval {
		t.Errorf("Expected report interval %v, got %v", reportInterval, reporter.reportInterval)
	}

	if reporter.rateLimit != rateLimit {
		t.Errorf("Expected rate limit %d, got %d", rateLimit, reporter.rateLimit)
	}

	if reporter.hashKey != hashKey {
		t.Errorf("Expected hash key %s, got %s", hashKey, reporter.hashKey)
	}

	if reporter.client == nil {
		t.Error("Client should be initialized")
	}
}

func TestPerformRequest_Success(t *testing.T) {
	client := &mockClient{mu: &sync.Mutex{}}
	metrics := []models.MetricModel{
		*models.NewMetricModel("test", common.MetricTypeGauge, 0, float64(1.0)),
	}

	err := performRequest(context.Background(), client, "http://test", nil, "", nil, metrics)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPerformRequest_WithHash(t *testing.T) {
	client := &mockClient{mu: &sync.Mutex{}}
	metrics := []models.MetricModel{
		*models.NewMetricModel("test", common.MetricTypeGauge, 0, float64(1.0)),
	}

	err := performRequest(context.Background(), client, "http://test", nil, "secret", nil, metrics)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(client.requests) == 0 {
		t.Fatal("No requests made")
	}

	hash := client.requests[0].Header.Get(common.HashHeaderKey)
	if hash == "" {
		t.Error("Hash header not set")
	}
}

func TestPerformRequest_WithEncryption(t *testing.T) {
	client := &mockClient{mu: &sync.Mutex{}}

	// Генерируем нормальный RSA ключ (2048 бит вместо 512)
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	metrics := []models.MetricModel{
		*models.NewMetricModel("test", common.MetricTypeGauge, 0, float64(1.0)),
	}

	err = performRequest(context.Background(), client, "http://test", &privKey.PublicKey, "", nil, metrics)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(client.requests) == 0 {
		t.Fatal("No requests made")
	}

	encryptedFlag := client.requests[0].Header.Get("X-Encrypted")
	if encryptedFlag != "true" {
		t.Error("X-Encrypted header not set")
	}
}

func TestPerformRequest_5xxRetry(t *testing.T) {
	originalBackoff := common.DefaultBackoffSchedule
	common.DefaultBackoffSchedule = []time.Duration{1 * time.Millisecond, 1 * time.Millisecond}
	defer func() { common.DefaultBackoffSchedule = originalBackoff }()

	attempts := 0
	client := &mockClient{
		mu: &sync.Mutex{},
		doFunc: func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts < 3 {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		},
	}

	metrics := []models.MetricModel{
		*models.NewMetricModel("test", common.MetricTypeGauge, 0, float64(1.0)),
	}

	err := performRequest(context.Background(), client, "http://test", nil, "", nil, metrics)
	if err != nil {
		t.Errorf("Unexpected error after retry: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestPerformRequest_4xxError(t *testing.T) {
	client := &mockClient{
		mu: &sync.Mutex{},
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		},
	}

	metrics := []models.MetricModel{
		*models.NewMetricModel("test", common.MetricTypeGauge, 0, float64(1.0)),
	}

	err := performRequest(context.Background(), client, "http://test", nil, "", nil, metrics)
	if err == nil {
		t.Error("Expected error for 4xx status")
	}
}

func TestPerformRequest_NetworkErrorWithRetry(t *testing.T) {
	originalBackoff := common.DefaultBackoffSchedule
	common.DefaultBackoffSchedule = []time.Duration{1 * time.Millisecond, 1 * time.Millisecond}
	defer func() { common.DefaultBackoffSchedule = originalBackoff }()

	attempts := 0
	client := &mockClient{
		mu: &sync.Mutex{},
		doFunc: func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts < 2 {
				return nil, errors.New("network error")
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		},
	}

	metrics := []models.MetricModel{
		*models.NewMetricModel("test", common.MetricTypeGauge, 0, float64(1.0)),
	}

	err := performRequest(context.Background(), client, "http://test", nil, "", nil, metrics)
	if err != nil {
		t.Errorf("Unexpected error after retry: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestPerformRequest_AllAttemptsFail(t *testing.T) {
	originalBackoff := common.DefaultBackoffSchedule
	common.DefaultBackoffSchedule = []time.Duration{1 * time.Millisecond, 1 * time.Millisecond}
	defer func() { common.DefaultBackoffSchedule = originalBackoff }()

	client := &mockClient{
		mu: &sync.Mutex{},
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("persistent network error")
		},
	}

	metrics := []models.MetricModel{
		*models.NewMetricModel("test", common.MetricTypeGauge, 0, float64(1.0)),
	}

	err := performRequest(context.Background(), client, "http://test", nil, "", nil, metrics)
	if err == nil {
		t.Error("Expected error after all attempts failed")
	}
}
