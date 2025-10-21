package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockInvoker struct {
	responses []error
	callCount int
}

func (m *mockInvoker) invoke(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	if m.callCount >= len(m.responses) {
		return status.Error(codes.Internal, "unexpected call")
	}
	err := m.responses[m.callCount]
	m.callCount++
	return err
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "unavailable error",
			err:      status.Error(codes.Unavailable, "service unavailable"),
			expected: true,
		},
		{
			name:     "deadline exceeded",
			err:      status.Error(codes.DeadlineExceeded, "timeout"),
			expected: true,
		},
		{
			name:     "internal error",
			err:      status.Error(codes.Internal, "internal error"),
			expected: true,
		},
		{
			name:     "invalid argument",
			err:      status.Error(codes.InvalidArgument, "bad request"),
			expected: false,
		},
		{
			name:     "permission denied",
			err:      status.Error(codes.PermissionDenied, "forbidden"),
			expected: false,
		},
		{
			name:     "non-status error",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRetry(tt.err)
			if result != tt.expected {
				t.Errorf("shouldRetry(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestRetryInterceptor_SuccessAfterRetries(t *testing.T) {
	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.Unavailable, "try again"),
			status.Error(codes.Internal, "internal"),
			nil,
		},
	}

	originalBackoff := common.DefaultBackoffSchedule
	common.DefaultBackoffSchedule = []time.Duration{1 * time.Millisecond, 1 * time.Millisecond, 1 * time.Millisecond}
	defer func() { common.DefaultBackoffSchedule = originalBackoff }()

	err := retryInterceptor(
		context.Background(),
		"testMethod",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if mock.callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", mock.callCount)
	}
}

func TestRetryInterceptor_AllRetriesFailed(t *testing.T) {
	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.Unavailable, "fail 1"),
			status.Error(codes.Internal, "fail 2"),
			status.Error(codes.Unknown, "fail 3"),
		},
	}

	originalBackoff := common.DefaultBackoffSchedule
	common.DefaultBackoffSchedule = []time.Duration{1 * time.Millisecond, 1 * time.Millisecond, 1 * time.Millisecond}
	defer func() { common.DefaultBackoffSchedule = originalBackoff }()

	err := retryInterceptor(
		context.Background(),
		"testMethod",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if mock.callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", mock.callCount)
	}
}

func TestRetryInterceptor_NonRetriableError(t *testing.T) {
	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.InvalidArgument, "bad request"),
		},
	}

	err := retryInterceptor(
		context.Background(),
		"testMethod",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	if mock.callCount != 1 {
		t.Errorf("Expected 1 call for non-retriable error, got %d", mock.callCount)
	}

	if err == nil {
		t.Error("Expected error, got nil")
	}
}
