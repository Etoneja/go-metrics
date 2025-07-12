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
	mu       sync.Mutex
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

	mockCli := &mockClient{}
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

		stats.collect(ctx)

		// stats collected
		r.report(ctx)

		assert.Equal(t, 1, len(mockCli.requests))

		req := mockCli.requests[0]
		assert.Equal(t, req.URL.Path, "/updates/")
		assert.Equal(t, http.MethodPost, req.Method)
	})

}
