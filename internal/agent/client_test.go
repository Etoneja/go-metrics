package agent

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewBaseClient(t *testing.T) {
	client := NewBaseClient()

	if client == nil {
		t.Fatal("Expected client instance, got nil")
	}

	if client.client == nil {
		t.Fatal("Expected http client instance, got nil")
	}

	if client.client.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", client.client.Timeout)
	}
}

type MockHTTPDoer struct {
	doFunc func(*http.Request) (*http.Response, error)
}

func (m *MockHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func (m *MockHTTPDoer) Close() {}

func TestBaseClient_Do(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	baseClient := &BaseClient{client: &http.Client{}}
	req, _ := http.NewRequest("GET", server.URL, nil)

	resp, err := baseClient.Do(req)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestBaseClient_Close(t *testing.T) {
	baseClient := &BaseClient{client: &http.Client{}}
	baseClient.Close()
}

func TestNewConcurrentLimitedClient(t *testing.T) {
	mockClient := &MockHTTPDoer{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200}, nil
		},
	}
	rateLimit := uint(5)

	client := NewConcurrentLimitedClient(mockClient, rateLimit)

	if client == nil {
		t.Fatal("Expected client instance, got nil")
	}

	if client.client != mockClient {
		t.Error("Client not set correctly")
	}

	if cap(client.semaphore) != 5 {
		t.Errorf("Expected semaphore capacity 5, got %d", cap(client.semaphore))
	}

	if client.mu == nil {
		t.Fatal("Mutex should be initialized")
	}
}

func TestConcurrentLimitedClient_Do(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockClient := &MockHTTPDoer{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(nil),
				}, nil
			},
		}

		limitedClient := NewConcurrentLimitedClient(mockClient, 2)
		defer limitedClient.Close()

		req := httptest.NewRequest("GET", "http://example.com", nil)
		resp, err := limitedClient.Do(req)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("error from underlying client", func(t *testing.T) {
		expectedErr := fmt.Errorf("network error")
		mockClient := &MockHTTPDoer{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, expectedErr
			},
		}

		limitedClient := NewConcurrentLimitedClient(mockClient, 2)
		defer limitedClient.Close()

		req := httptest.NewRequest("GET", "http://example.com", nil)
		resp, err := limitedClient.Do(req)

		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}

		if resp != nil && resp.Body != nil {
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()
		}
	})

	t.Run("request after client closed", func(t *testing.T) {
		mockClient := &MockHTTPDoer{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(nil),
				}, nil
			},
		}

		limitedClient := NewConcurrentLimitedClient(mockClient, 2)
		limitedClient.Close()

		req := httptest.NewRequest("GET", "http://example.com", nil)
		resp, err := limitedClient.Do(req)

		if err == nil || err.Error() != "client is closed" {
			t.Errorf("Expected 'client is closed' error, got %v", err)
		}

		if resp != nil && resp.Body != nil {
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()
		}
	})
}
