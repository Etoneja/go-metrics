package server

import (
	"context"
	"errors"
	"testing"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/etoneja/go-metrics/internal/proto"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockStore struct {
	pingFunc        func(ctx context.Context) error
	batchUpdateFunc func(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error)
	getAllFunc      func(ctx context.Context) ([]models.MetricModel, error)
}

func (m *mockStore) GetGauge(ctx context.Context, key string) (float64, error) { return 0, nil }
func (m *mockStore) SetGauge(ctx context.Context, key string, value float64) (float64, error) {
	return 0, nil
}
func (m *mockStore) GetCounter(ctx context.Context, key string) (int64, error) { return 0, nil }
func (m *mockStore) IncrementCounter(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}
func (m *mockStore) ShutDown() {}

func (m *mockStore) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func (m *mockStore) BatchUpdate(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error) {
	if m.batchUpdateFunc != nil {
		return m.batchUpdateFunc(ctx, metrics)
	}
	return metrics, nil
}

func (m *mockStore) GetAll(ctx context.Context) ([]models.MetricModel, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx)
	}
	return []models.MetricModel{}, nil
}

func TestGRPCServer_Ping(t *testing.T) {
	tests := []struct {
		name        string
		pingFunc    func(ctx context.Context) error
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "successful ping",
			pingFunc: func(ctx context.Context) error {
				return nil
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "ping failed",
			pingFunc: func(ctx context.Context) error {
				return errors.New("connection failed")
			},
			wantSuccess: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			store := &mockStore{pingFunc: tt.pingFunc}
			server := NewGRPCServer(store, logger)

			resp, err := server.Ping(context.Background(), &proto.PingRequest{})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if resp != nil && resp.Success {
					t.Error("Expected success false when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.Success != tt.wantSuccess {
					t.Errorf("Expected success %v, got %v", tt.wantSuccess, resp.Success)
				}
			}
		})
	}
}

func TestGRPCServer_BatchUpdate(t *testing.T) {
	tests := []struct {
		name            string
		requestMetrics  []*proto.Metric
		batchUpdateFunc func(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error)
		wantErr         bool
		wantCode        codes.Code
	}{
		{
			name: "successful batch update",
			requestMetrics: []*proto.Metric{
				{Id: "gauge1", Type: common.MetricTypeGauge, Value: common.Float64Ptr(1.23)},
				{Id: "counter1", Type: common.MetricTypeCounter, Delta: common.Int64Ptr(42)},
			},
			batchUpdateFunc: func(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error) {
				return metrics, nil
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "invalid metric in request",
			requestMetrics: []*proto.Metric{
				{Id: "invalid", Type: "unknown_type"},
			},
			batchUpdateFunc: nil,
			wantErr:         true,
			wantCode:        codes.InvalidArgument,
		},
		{
			name: "store batch update failed",
			requestMetrics: []*proto.Metric{
				{Id: "gauge1", Type: common.MetricTypeGauge, Value: common.Float64Ptr(1.23)},
			},
			batchUpdateFunc: func(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error) {
				return nil, errors.New("storage error")
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
		{
			name: "conversion error on response",
			requestMetrics: []*proto.Metric{
				{Id: "gauge1", Type: common.MetricTypeGauge, Value: common.Float64Ptr(1.23)},
			},
			batchUpdateFunc: func(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error) {
				// Return metric with invalid type to cause conversion error
				return []models.MetricModel{
					{ID: "invalid", MType: "unknown_type"},
				}, nil
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			store := &mockStore{batchUpdateFunc: tt.batchUpdateFunc}
			server := NewGRPCServer(store, logger)

			req := &proto.BatchUpdateRequest{Metrics: tt.requestMetrics}
			resp, err := server.BatchUpdate(context.Background(), req)

			if tt.wantErr {
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
					t.Fatalf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Fatal("Expected response, got nil")
				}
				if len(resp.Metrics) != len(tt.requestMetrics) {
					t.Errorf("Expected %d metrics in response, got %d", len(tt.requestMetrics), len(resp.Metrics))
				}
			}
		})
	}
}

func TestGRPCServer_ListMetrics(t *testing.T) {
	tests := []struct {
		name       string
		getAllFunc func(ctx context.Context) ([]models.MetricModel, error)
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name: "successful list metrics",
			getAllFunc: func(ctx context.Context) ([]models.MetricModel, error) {
				return []models.MetricModel{
					{ID: "gauge1", MType: common.MetricTypeGauge, Value: common.Float64Ptr(1.23)},
					{ID: "counter1", MType: common.MetricTypeCounter, Delta: common.Int64Ptr(42)},
				}, nil
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "store get all failed",
			getAllFunc: func(ctx context.Context) ([]models.MetricModel, error) {
				return nil, errors.New("storage error")
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
		{
			name: "empty metrics list",
			getAllFunc: func(ctx context.Context) ([]models.MetricModel, error) {
				return []models.MetricModel{}, nil
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "conversion error on response",
			getAllFunc: func(ctx context.Context) ([]models.MetricModel, error) {
				return []models.MetricModel{
					{ID: "invalid", MType: "unknown_type"},
				}, nil
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			store := &mockStore{getAllFunc: tt.getAllFunc}
			server := NewGRPCServer(store, logger)

			resp, err := server.ListMetrics(context.Background(), &proto.ListMetricsRequest{})

			if tt.wantErr {
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
				if resp == nil {
					t.Error("Expected response, got nil")
				}
			}
		})
	}
}
