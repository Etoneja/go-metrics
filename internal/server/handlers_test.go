package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/etoneja/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBaseHandler_writeHTML(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	handler := &BaseHandler{
		store:  nil,
		logger: logger,
	}

	recorder := httptest.NewRecorder()
	handler.writeHTML(recorder, "test content")

	if recorder.Body.String() != "test content" {
		t.Errorf("Expected 'test content', got '%s'", recorder.Body.String())
	}

	handler.writeHTML(httptest.NewRecorder(), "another test")
}

func TestMetricUpdateHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name   string
		store  Storager
		uri    string
		method string
		want   want
	}{
		{
			name:   "bad url - 404",
			store:  NewMemStorage(),
			uri:    "/gauge/fake/123",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:   "bad method - 405 on get",
			store:  NewMemStorage(),
			uri:    "/update/gauge/fake/123",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "bad method - 405 on delete",
			store:  NewMemStorage(),
			uri:    "/update/gauge/fake/123",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "bad url - too few args",
			store:  NewMemStorage(),
			uri:    "/update/gauge/123",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:   "bad url - too many args",
			store:  NewMemStorage(),
			uri:    "/update/gauge/fake/123/wtf",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:   "bad metric type",
			store:  NewMemStorage(),
			uri:    "/update/faketype/fake/wtf",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "bad metric value for gauge",
			store:  NewMemStorage(),
			uri:    "/update/gauge/fake/wtf",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "bad metric value for counter",
			store:  NewMemStorage(),
			uri:    "/update/gauge/fake/wtf",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "success gauge",
			store:  NewMemStorage(),
			uri:    "/update/gauge/fake/123.123",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "success counter",
			store:  NewMemStorage(),
			uri:    "/update/counter/fake/123",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			server := httptest.NewServer(NewRouter(tt.store, cfg))
			defer server.Close()

			req, err := http.NewRequest(tt.method, server.URL+tt.uri, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

		})
	}

}

func TestMetricGetHandler(t *testing.T) {

	type want struct {
		statusCode int
		contains   string
	}

	tests := []struct {
		name    string
		prepare func(Storager)
		uri     string
		method  string
		want    want
	}{
		{
			name: "get gauge - success",
			prepare: func(s Storager) {
				_, err := s.SetGauge(context.Background(), "test_metric", 123.45)
				if err != nil {
					t.Fatalf("SetGauge failed: %v", err)
				}
			},
			uri:    "/value/gauge/test_metric",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
				contains:   "123.45",
			},
		},
		{
			name: "get counter - success",
			prepare: func(s Storager) {
				_, err := s.IncrementCounter(context.Background(), "test_metric", 42)
				if err != nil {
					t.Fatalf("IncrementCounter failed: %v", err)
				}
			},
			uri:    "/value/counter/test_metric",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusOK,
				contains:   "42",
			},
		},
		{
			name:    "metric not found",
			prepare: func(s Storager) {},
			uri:     "/value/gauge/nonexistent",
			method:  http.MethodGet,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStorage()
			if tt.prepare != nil {
				tt.prepare(store)
			}

			cfg := &config{}
			server := httptest.NewServer(NewRouter(store, cfg))
			defer server.Close()

			req, err := http.NewRequest(tt.method, server.URL+tt.uri, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.contains != "" {
				body, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(body), tt.want.contains)
			}
		})
	}
}

func TestMetricListHandler(t *testing.T) {
	type want struct {
		statusCode int
		contains   []string
	}

	tests := []struct {
		name    string
		prepare func(Storager)
		want    want
	}{
		{
			name: "success with metrics",
			prepare: func(s Storager) {
				if _, err := s.SetGauge(context.Background(), "gauge1", 123.45); err != nil {
					t.Fatalf("SetGauge gauge1 failed: %v", err)
				}
				if _, err := s.IncrementCounter(context.Background(), "counter1", 42); err != nil {
					t.Fatalf("IncrementCounter counter1 failed: %v", err)
				}
				if _, err := s.SetGauge(context.Background(), "gauge2", 67.89); err != nil {
					t.Fatalf("SetGauge gauge2 failed: %v", err)
				}
			},
			want: want{
				statusCode: http.StatusOK,
				contains: []string{
					"<html><body><pre>",
					"gauge1[gauge]=123.45",
					"counter1[counter]=42",
					"gauge2[gauge]=67.89",
					"</pre></body></html>",
				},
			},
		},
		{
			name:    "success empty storage",
			prepare: func(s Storager) {},
			want: want{
				statusCode: http.StatusOK,
				contains: []string{
					"<html><body><pre>",
					"</pre></body></html>",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStorage()
			if tt.prepare != nil {
				tt.prepare(store)
			}

			cfg := &config{}
			server := httptest.NewServer(NewRouter(store, cfg))
			defer server.Close()

			req, err := http.NewRequest("GET", server.URL+"/", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.statusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				bodyStr := string(body)

				for _, expected := range tt.want.contains {
					assert.Contains(t, bodyStr, expected)
				}

				assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestMetricUpdateJSONHandler(t *testing.T) {
	type want struct {
		statusCode int
		contains   string
	}

	tests := []struct {
		name   string
		method string
		uri    string
		body   string
		want   want
	}{
		{
			name:   "success update gauge",
			method: http.MethodPost,
			uri:    "/update/",
			body:   `{"id":"temperature","type":"gauge","value":23.5}`,
			want: want{
				statusCode: http.StatusOK,
				contains:   `"value":23.5`,
			},
		},
		{
			name:   "success update counter",
			method: http.MethodPost,
			uri:    "/update/",
			body:   `{"id":"requests","type":"counter","delta":1}`,
			want: want{
				statusCode: http.StatusOK,
				contains:   `"delta":1`,
			},
		},
		{
			name:   "missing metric id",
			method: http.MethodPost,
			uri:    "/update/",
			body:   `{"type":"gauge","value":23.5}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "missing gauge value",
			method: http.MethodPost,
			uri:    "/update/",
			body:   `{"id":"temp","type":"gauge"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "missing counter delta",
			method: http.MethodPost,
			uri:    "/update/",
			body:   `{"id":"count","type":"counter"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "invalid metric type",
			method: http.MethodPost,
			uri:    "/update/",
			body:   `{"id":"test","type":"invalid","value":1.0}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "invalid json",
			method: http.MethodPost,
			uri:    "/update/",
			body:   `invalid json`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "wrong method - get",
			method: http.MethodGet,
			uri:    "/update/",
			body:   `{"id":"test","type":"gauge","value":1.0}`,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStorage()
			cfg := &config{}
			server := httptest.NewServer(NewRouter(store, cfg))
			defer server.Close()

			req, err := http.NewRequest(tt.method, server.URL+tt.uri, strings.NewReader(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.statusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(body), tt.want.contains)
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

				if strings.Contains(tt.body, "gauge") {
					value, _ := store.GetGauge(context.Background(), "temperature")
					assert.Equal(t, 23.5, value)
				} else if strings.Contains(tt.body, "counter") {
					value, _ := store.GetCounter(context.Background(), "requests")
					assert.Equal(t, int64(1), value)
				}
			}
		})
	}
}

func TestMetricGetJSONHandler(t *testing.T) {
	type want struct {
		statusCode int
		contains   string
	}

	tests := []struct {
		name    string
		prepare func(Storager)
		body    string
		want    want
	}{
		{
			name: "success get gauge",
			prepare: func(s Storager) {
				_, err := s.SetGauge(context.Background(), "temperature", 23.5)
				if err != nil {
					t.Fatalf("SetGauge failed: %v", err)
				}
			},
			body: `{"id":"temperature","type":"gauge"}`,
			want: want{
				statusCode: http.StatusOK,
				contains:   `"value":23.5`,
			},
		},
		{
			name: "success get counter",
			prepare: func(s Storager) {
				_, err := s.IncrementCounter(context.Background(), "requests", 42)
				if err != nil {
					t.Fatalf("IncrementCounter failed: %v", err)
				}
			},
			body: `{"id":"requests","type":"counter"}`,
			want: want{
				statusCode: http.StatusOK,
				contains:   `"delta":42`,
			},
		},
		{
			name: "gauge not found",
			prepare: func(s Storager) {
				// не добавляем метрику
			},
			body: `{"id":"nonexistent","type":"gauge"}`,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "counter not found",
			prepare: func(s Storager) {
				// не добавляем метрику
			},
			body: `{"id":"nonexistent","type":"counter"}`,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:    "missing metric id",
			prepare: func(s Storager) {},
			body:    `{"type":"gauge"}`,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:    "missing metric type",
			prepare: func(s Storager) {},
			body:    `{"id":"test"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "invalid metric type",
			prepare: func(s Storager) {},
			body:    `{"id":"test","type":"invalid"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "invalid json",
			prepare: func(s Storager) {},
			body:    `invalid json`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "empty body",
			prepare: func(s Storager) {},
			body:    ``,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStorage()
			if tt.prepare != nil {
				tt.prepare(store)
			}

			cfg := &config{}
			server := httptest.NewServer(NewRouter(store, cfg))
			defer server.Close()

			req, err := http.NewRequest("POST", server.URL+"/value/", strings.NewReader(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.statusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				bodyStr := string(body)
				assert.Contains(t, bodyStr, tt.want.contains)
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

				var metric models.MetricModel
				err := json.Unmarshal(body, &metric)
				assert.NoError(t, err)
				assert.Equal(t, tt.body[strings.Index(tt.body, `"id":"`)+6:strings.Index(tt.body, `","`)], metric.ID)
			}
		})
	}
}

func TestPingHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name    string
		prepare func(Storager)
		want    want
	}{
		{
			name:    "ping success",
			prepare: func(s Storager) {},
			want: want{
				statusCode: http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStorage()
			if tt.prepare != nil {
				tt.prepare(store)
			}
			cfg := &config{}
			server := httptest.NewServer(NewRouter(store, cfg))
			defer server.Close()

			req, err := http.NewRequest("GET", server.URL+"/ping", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}
