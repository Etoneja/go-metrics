package agent

import (
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

func TestReporter_report(t *testing.T) {
	fakeDuration := time.Duration(time.Millisecond)
	fakeEndpoint := "http://fake.com/"
	stats := newStats()

	mockCli := &mockClient{}
	t.Run("test report", func(t *testing.T) {
		r := &Reporter{
			stats:         stats,
			client:        mockCli,
			endpoint:      fakeEndpoint,
			sleepDuration: fakeDuration,
		}
		assert.Equal(t, uint(0), r.iteration)

		// stats not collected
		r.report()

		assert.Equal(t, uint(1), r.iteration)
		assert.Equal(t, 0, len(mockCli.requests))

		stats.collect()

		// stats collected
		r.report()

		metrics := r.stats.dump()

		assert.Equal(t, len(metrics), len(mockCli.requests))
		for _, req := range mockCli.requests {
			assert.Equal(t, req.URL.Path, "/update/")
			assert.Equal(t, http.MethodPost, req.Method)
		}
	})

}
