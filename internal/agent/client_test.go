package agent

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockHTTPDoer struct {
	doFunc func(*http.Request) (*http.Response, error)
}

func (m *MockHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func (m *MockHTTPDoer) Close() {}

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
		defer resp.Body.Close()

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
			defer resp.Body.Close()
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
			defer resp.Body.Close()
		}
	})
}
