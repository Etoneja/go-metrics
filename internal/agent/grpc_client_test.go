package agent

import (
	"context"
	"net"
	"testing"

	"github.com/etoneja/go-metrics/internal/models"
	"github.com/etoneja/go-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockMetricsServiceClient struct {
	batchUpdateFunc func(ctx context.Context, in *proto.BatchUpdateRequest, opts ...grpc.CallOption) (*proto.BatchUpdateResponse, error)
}

func (m *mockMetricsServiceClient) BatchUpdate(ctx context.Context, in *proto.BatchUpdateRequest, opts ...grpc.CallOption) (*proto.BatchUpdateResponse, error) {
	return m.batchUpdateFunc(ctx, in, opts...)
}

func (m *mockMetricsServiceClient) Ping(ctx context.Context, in *proto.PingRequest, opts ...grpc.CallOption) (*proto.PingResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *mockMetricsServiceClient) ListMetrics(ctx context.Context, in *proto.ListMetricsRequest, opts ...grpc.CallOption) (*proto.ListMetricsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func TestGRPCClient_PerformRequest(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []models.MetricModel
		ip       net.IP
		mockFunc func(ctx context.Context, in *proto.BatchUpdateRequest, opts ...grpc.CallOption) (*proto.BatchUpdateResponse, error)
		wantErr  bool
		checkIP  bool
	}{
		{
			name: "successful request",
			metrics: []models.MetricModel{
				{ID: "test1", MType: "gauge", Value: float64Ptr(1.23)},
			},
			ip: net.ParseIP("192.168.1.1"),
			mockFunc: func(ctx context.Context, in *proto.BatchUpdateRequest, opts ...grpc.CallOption) (*proto.BatchUpdateResponse, error) {
				return &proto.BatchUpdateResponse{}, nil
			},
			wantErr: false,
			checkIP: true,
		},
		{
			name: "server error",
			metrics: []models.MetricModel{
				{ID: "test2", MType: "counter", Delta: int64Ptr(42)},
			},
			ip: nil,
			mockFunc: func(ctx context.Context, in *proto.BatchUpdateRequest, opts ...grpc.CallOption) (*proto.BatchUpdateResponse, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
			wantErr: true,
			checkIP: false,
		},
		{
			name: "with ip in metadata",
			metrics: []models.MetricModel{
				{ID: "test3", MType: "gauge", Value: float64Ptr(3.14)},
			},
			ip: net.ParseIP("10.0.0.1"),
			mockFunc: func(ctx context.Context, in *proto.BatchUpdateRequest, opts ...grpc.CallOption) (*proto.BatchUpdateResponse, error) {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					return nil, status.Error(codes.Internal, "no metadata")
				}
				ipHeaders := md.Get("x-real-ip")
				if len(ipHeaders) == 0 || ipHeaders[0] != "10.0.0.1" {
					return nil, status.Error(codes.Internal, "ip not found in metadata")
				}
				return &proto.BatchUpdateResponse{}, nil
			},
			wantErr: false,
			checkIP: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &GRPCClient{
				client: &mockMetricsServiceClient{
					batchUpdateFunc: tt.mockFunc,
				},
			}

			err := client.PerformRequest(context.Background(), tt.ip, tt.metrics)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
