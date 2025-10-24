package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

type mockTrustedHandler struct {
	called bool
}

func (m *mockTrustedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
	w.WriteHeader(http.StatusOK)
}

func TestTrustedIPMiddleware(t *testing.T) {
	logger := zap.NewNop()
	bmw := &BaseMiddleware{logger: logger}

	tests := []struct {
		name          string
		allowedSubnet string
		realIP        string
		wantStatus    int
		wantCalled    bool
	}{
		{
			name:          "empty subnet allows all",
			allowedSubnet: "",
			realIP:        "192.168.1.1",
			wantStatus:    http.StatusOK,
			wantCalled:    true,
		},
		{
			name:          "valid ip in subnet",
			allowedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.100",
			wantStatus:    http.StatusOK,
			wantCalled:    true,
		},
		{
			name:          "missing x-real-ip header",
			allowedSubnet: "192.168.1.0/24",
			realIP:        "",
			wantStatus:    http.StatusForbidden,
			wantCalled:    false,
		},
		{
			name:          "invalid ip format",
			allowedSubnet: "192.168.1.0/24",
			realIP:        "invalid-ip",
			wantStatus:    http.StatusForbidden,
			wantCalled:    false,
		},
		{
			name:          "ip not in subnet",
			allowedSubnet: "192.168.1.0/24",
			realIP:        "10.0.0.1",
			wantStatus:    http.StatusForbidden,
			wantCalled:    false,
		},
		{
			name:          "invalid subnet format",
			allowedSubnet: "invalid-subnet",
			realIP:        "192.168.1.1",
			wantStatus:    http.StatusInternalServerError,
			wantCalled:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := &mockTrustedHandler{}
			middleware := bmw.TrustedIPMiddleware(tt.allowedSubnet)
			handler := middleware(mockHandler)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if mockHandler.called != tt.wantCalled {
				t.Errorf("Expected handler called %v, got %v", tt.wantCalled, mockHandler.called)
			}
		})
	}
}

func TestTrustedIPMiddleware_IPv6(t *testing.T) {
	logger := zap.NewNop()
	bmw := &BaseMiddleware{logger: logger}

	mockHandler := &mockTrustedHandler{}
	middleware := bmw.TrustedIPMiddleware("2001:db8::/32")
	handler := middleware(mockHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "2001:db8::1")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for IPv6, got %d", rr.Code)
	}

	if !mockHandler.called {
		t.Error("Expected handler to be called for IPv6")
	}
}
