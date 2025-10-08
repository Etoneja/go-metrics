package agent

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	mu       *sync.Mutex
	requests []*http.Request
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req)

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
		ReportInterval: 10,
		RateLimit:      rateLimit,
		HashKey:        hashKey,
	}

	reporter := newReporter(stats, cfg, nil)

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
