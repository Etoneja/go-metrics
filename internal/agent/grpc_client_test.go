// agent/test_helpers.go
package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateTestMetrics() []models.MetricModel {
	return []models.MetricModel{
		{ID: "cpu_usage", MType: "gauge", Value: common.Float64Ptr(95.5)},
		{ID: "request_count", MType: "counter", Delta: common.Int64Ptr(100)},
	}
}

func TestGRPCClient_SendBatch_EmptyMetrics(t *testing.T) {
	cfg := &mockConfig{
		serverEndpoint: "localhost:9090",
	}

	client := &GRPCClient{
		cfg: cfg,
	}

	ctx := context.Background()
	err := client.SendBatch(ctx, []models.MetricModel{})

	assert.NoError(t, err)
}

func TestGRPCClient_Close_NilConnection(t *testing.T) {
	cfg := &mockConfig{
		serverEndpoint: "localhost:9090",
	}

	client := &GRPCClient{
		cfg: cfg,
		// conn = nil
	}

	err := client.Close()
	assert.NoError(t, err)
}

func TestShouldRetry_RetryableErrors(t *testing.T) {
	retryableCodes := []codes.Code{
		codes.DeadlineExceeded,
		codes.Unavailable,
		codes.ResourceExhausted,
		codes.Internal,
		codes.Unknown,
	}

	for _, code := range retryableCodes {
		t.Run(code.String(), func(t *testing.T) {
			err := status.Error(code, "test error")
			result := shouldRetry(err)
			assert.True(t, result, "code %s should be retryable", code)
		})
	}
}

func TestShouldRetry_NonRetryableErrors(t *testing.T) {
	nonRetryableCodes := []codes.Code{
		codes.OK,
		codes.Canceled,
		codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.FailedPrecondition,
		codes.Aborted,
		codes.OutOfRange,
		codes.Unimplemented,
		codes.Unauthenticated,
	}

	for _, code := range nonRetryableCodes {
		t.Run(code.String(), func(t *testing.T) {
			err := status.Error(code, "test error")
			result := shouldRetry(err)
			assert.False(t, result, "code %s should not be retryable", code)
		})
	}
}

func TestShouldRetry_NonStatusError(t *testing.T) {
	err := errors.New("regular error")
	result := shouldRetry(err)
	assert.False(t, result)
}

func TestShouldRetry_NilError(t *testing.T) {
	result := shouldRetry(nil)
	assert.False(t, result)
}

func TestRetryInterceptor_BackoffSchedule(t *testing.T) {

	originalSchedule := common.DefaultBackoffSchedule

	common.DefaultBackoffSchedule = []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
	}
	defer func() {
		common.DefaultBackoffSchedule = originalSchedule
	}()

	attempts := 0
	mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		return status.Error(codes.Unavailable, "service unavailable")
	}

	start := time.Now()
	err := retryInterceptor(
		context.Background(),
		"testMethod",
		nil, nil, nil,
		mockInvoker,
	)

	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all gRPC attempts failed")

	assert.Equal(t, 2, attempts)

	expectedMinTime := 10 * time.Millisecond
	assert.GreaterOrEqual(t, elapsed, expectedMinTime)
}

func TestRetryInterceptor_SuccessOnRetry(t *testing.T) {
	originalSchedule := common.DefaultBackoffSchedule
	common.DefaultBackoffSchedule = []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
	}
	defer func() {
		common.DefaultBackoffSchedule = originalSchedule
	}()

	attempts := 0
	mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		if attempts < 2 {
			return status.Error(codes.Unavailable, "service unavailable")
		}
		return nil
	}

	err := retryInterceptor(
		context.Background(),
		"testMethod",
		nil, nil, nil,
		mockInvoker,
	)

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetryInterceptor_NonRetryableError(t *testing.T) {
	attempts := 0
	mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		return status.Error(codes.InvalidArgument, "bad request")
	}

	err := retryInterceptor(
		context.Background(),
		"testMethod",
		nil, nil, nil,
		mockInvoker,
	)

	assert.Error(t, err)
	assert.Equal(t, 1, attempts)
}
