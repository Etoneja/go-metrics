package server

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockLoggingGRPCHandler struct {
	resp  interface{}
	err   error
	delay time.Duration
}

func (m *mockLoggingGRPCHandler) handle(ctx context.Context, req interface{}) (interface{}, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.resp, m.err
}

func TestLoggingInterceptor(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		handlerResp interface{}
		handlerErr  error
		wantCode    codes.Code
	}{
		{
			name:        "successful request",
			method:      "/service.Method",
			handlerResp: "response-data",
			handlerErr:  nil,
			wantCode:    codes.OK,
		},
		{
			name:        "request with error",
			method:      "/service.Method",
			handlerResp: nil,
			handlerErr:  status.Error(codes.InvalidArgument, "bad request"),
			wantCode:    codes.InvalidArgument,
		},
		{
			name:        "internal server error",
			method:      "/service.Method",
			handlerResp: nil,
			handlerErr:  status.Error(codes.Internal, "internal error"),
			wantCode:    codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			interceptor := loggingInterceptor(logger)

			mockHandler := &mockLoggingGRPCHandler{
				resp: tt.handlerResp,
				err:  tt.handlerErr,
			}

			info := &grpc.UnaryServerInfo{FullMethod: tt.method}

			resp, err := interceptor(context.Background(), "test-request", info, mockHandler.handle)

			if tt.handlerErr != nil {
				if err == nil {
					t.Error("Expected error, got nil")
				} else {
					st, ok := status.FromError(err)
					if !ok {
						t.Errorf("Expected status error, got %v", err)
					} else if st.Code() != tt.wantCode {
						t.Errorf("Expected code %v, got %v", tt.wantCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp != tt.handlerResp {
					t.Errorf("Expected response %v, got %v", tt.handlerResp, resp)
				}
			}
		})
	}
}

func TestLoggingInterceptor_Duration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := loggingInterceptor(logger)

	mockHandler := &mockLoggingGRPCHandler{
		resp:  "response",
		err:   nil,
		delay: 10 * time.Millisecond,
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/service.Method"}

	start := time.Now()
	_, err := interceptor(context.Background(), "test-request", info, mockHandler.handle)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", duration)
	}
}
