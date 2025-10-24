package server

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockTrustedGRPCHandler struct {
	called bool
	resp   interface{}
	err    error
}

func (m *mockTrustedGRPCHandler) handle(ctx context.Context, req interface{}) (interface{}, error) {
	m.called = true
	return m.resp, m.err
}

func TestTrustedSubnetInterceptor(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name          string
		allowedSubnet string
		metadata      metadata.MD
		wantCalled    bool
		wantCode      codes.Code
	}{
		{
			name:          "empty subnet allows all",
			allowedSubnet: "",
			metadata:      metadata.New(map[string]string{"x-real-ip": "192.168.1.1"}),
			wantCalled:    true,
			wantCode:      codes.OK,
		},
		{
			name:          "valid ip in subnet",
			allowedSubnet: "192.168.1.0/24",
			metadata:      metadata.New(map[string]string{"x-real-ip": "192.168.1.100"}),
			wantCalled:    true,
			wantCode:      codes.OK,
		},
		{
			name:          "no metadata",
			allowedSubnet: "192.168.1.0/24",
			metadata:      nil,
			wantCalled:    false,
			wantCode:      codes.PermissionDenied,
		},
		{
			name:          "missing x-real-ip header",
			allowedSubnet: "192.168.1.0/24",
			metadata:      metadata.New(map[string]string{}),
			wantCalled:    false,
			wantCode:      codes.PermissionDenied,
		},
		{
			name:          "invalid ip format",
			allowedSubnet: "192.168.1.0/24",
			metadata:      metadata.New(map[string]string{"x-real-ip": "invalid-ip"}),
			wantCalled:    false,
			wantCode:      codes.PermissionDenied,
		},
		{
			name:          "ip not in subnet",
			allowedSubnet: "192.168.1.0/24",
			metadata:      metadata.New(map[string]string{"x-real-ip": "10.0.0.1"}),
			wantCalled:    false,
			wantCode:      codes.PermissionDenied,
		},
		{
			name:          "invalid subnet format",
			allowedSubnet: "invalid-subnet",
			metadata:      metadata.New(map[string]string{"x-real-ip": "192.168.1.1"}),
			wantCalled:    false,
			wantCode:      codes.Internal,
		},
		{
			name:          "ipv6 in subnet",
			allowedSubnet: "2001:db8::/32",
			metadata:      metadata.New(map[string]string{"x-real-ip": "2001:db8::1"}),
			wantCalled:    true,
			wantCode:      codes.OK,
		},
		{
			name:          "ipv6 not in subnet",
			allowedSubnet: "2001:db8::/32",
			metadata:      metadata.New(map[string]string{"x-real-ip": "2001:db9::1"}),
			wantCalled:    false,
			wantCode:      codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := &mockTrustedGRPCHandler{resp: "test-response"}
			interceptor := TrustedSubnetInterceptor(tt.allowedSubnet, logger)

			var ctx context.Context
			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(context.Background(), tt.metadata)
			} else {
				ctx = context.Background()
			}

			resp, err := interceptor(ctx, "test-request", &grpc.UnaryServerInfo{}, mockHandler.handle)

			if tt.wantCalled {
				if !mockHandler.called {
					t.Error("Expected handler to be called")
				}
				if resp != "test-response" {
					t.Errorf("Expected response 'test-response', got %v", resp)
				}
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			} else {
				if mockHandler.called {
					t.Error("Expected handler not to be called")
				}
				if err == nil {
					t.Error("Expected error, got nil")
				} else {
					st, ok := status.FromError(err)
					if !ok {
						t.Errorf("Expected gRPC status error, got %v", err)
					} else if st.Code() != tt.wantCode {
						t.Errorf("Expected code %v, got %v", tt.wantCode, st.Code())
					}
				}
			}
		})
	}
}
