package models

import (
	"testing"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMetricModelToGRPC(t *testing.T) {
	tests := []struct {
		name      string
		appMetric MetricModel
		wantErr   bool
	}{
		{
			name: "valid gauge",
			appMetric: MetricModel{
				ID:    "gauge1",
				MType: common.MetricTypeGauge,
				Value: float64Ptr(1.23),
			},
			wantErr: false,
		},
		{
			name: "valid counter",
			appMetric: MetricModel{
				ID:    "counter1",
				MType: common.MetricTypeCounter,
				Delta: int64Ptr(42),
			},
			wantErr: false,
		},
		{
			name: "gauge with nil value",
			appMetric: MetricModel{
				ID:    "gauge1",
				MType: common.MetricTypeGauge,
				Value: nil,
			},
			wantErr: true,
		},
		{
			name: "counter with nil delta",
			appMetric: MetricModel{
				ID:    "counter1",
				MType: common.MetricTypeCounter,
				Delta: nil,
			},
			wantErr: true,
		},
		{
			name: "unknown metric type",
			appMetric: MetricModel{
				ID:    "unknown1",
				MType: "unknown_type",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcMetric, err := MetricModelToGRPC(tt.appMetric)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if grpcMetric.Id != tt.appMetric.ID {
				t.Errorf("Expected ID %s, got %s", tt.appMetric.ID, grpcMetric.Id)
			}
			if grpcMetric.Type != tt.appMetric.MType {
				t.Errorf("Expected Type %s, got %s", tt.appMetric.MType, grpcMetric.Type)
			}
		})
	}
}

func TestMetricModelFromGRPC(t *testing.T) {
	tests := []struct {
		name       string
		grpcMetric *proto.Metric
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name: "valid gauge",
			grpcMetric: &proto.Metric{
				Id:    "gauge1",
				Type:  common.MetricTypeGauge,
				Value: float64Ptr(1.23),
			},
			wantErr:  false,
			wantCode: codes.OK,
		},
		{
			name: "gauge with nil value",
			grpcMetric: &proto.Metric{
				Id:   "gauge1",
				Type: common.MetricTypeGauge,
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "unknown metric type",
			grpcMetric: &proto.Metric{
				Id:   "unknown1",
				Type: "unknown_type",
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appMetric, err := MetricModelFromGRPC(tt.grpcMetric)

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
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if appMetric.ID != tt.grpcMetric.Id {
				t.Errorf("Expected ID %s, got %s", tt.grpcMetric.Id, appMetric.ID)
			}
			if appMetric.MType != tt.grpcMetric.Type {
				t.Errorf("Expected Type %s, got %s", tt.grpcMetric.Type, appMetric.MType)
			}
		})
	}
}

func float64Ptr(f float64) *float64 { return &f }
func int64Ptr(i int64) *int64       { return &i }
