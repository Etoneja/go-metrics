package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			fakeHashKey := ""
			server := httptest.NewServer(NewRouter(tt.store, fakeHashKey))
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
