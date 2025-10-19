package models

import (
	"fmt"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MetricModelsToGRPC(appMetrics []MetricModel) ([]*proto.Metric, error) {
	grpcMetrics := make([]*proto.Metric, 0, len(appMetrics))

	for _, appMetric := range appMetrics {
		grpcMetric, err := MetricModelToGRPC(appMetric)
		if err != nil {
			return nil, err
		}
		grpcMetrics = append(grpcMetrics, grpcMetric)
	}

	return grpcMetrics, nil
}

func MetricModelsFromGRPC(grpcMetrics []*proto.Metric) ([]MetricModel, error) {
	appMetrics := make([]MetricModel, 0, len(grpcMetrics))

	for _, grpcMetric := range grpcMetrics {
		appMetric, err := MetricModelFromGRPC(grpcMetric)
		if err != nil {
			return nil, err
		}
		appMetrics = append(appMetrics, appMetric)
	}

	return appMetrics, nil
}

func MetricModelToGRPC(appMetric MetricModel) (*proto.Metric, error) {
	grpcMetric := &proto.Metric{
		Id:   appMetric.ID,
		Type: appMetric.MType,
	}

	switch appMetric.MType {
	case common.MetricTypeCounter:
		if appMetric.Delta == nil {
			return nil, fmt.Errorf("counter metric %s has nil delta", appMetric.ID)
		}
		grpcMetric.Delta = appMetric.Delta

	case common.MetricTypeGauge:
		if appMetric.Value == nil {
			return nil, fmt.Errorf("gauge metric %s has nil value", appMetric.ID)
		}
		grpcMetric.Value = appMetric.Value

	default:
		return nil, fmt.Errorf("unknown metric type: %s", appMetric.MType)
	}

	return grpcMetric, nil
}

func MetricModelFromGRPC(grpcMetric *proto.Metric) (MetricModel, error) {
	var appMetric MetricModel

	switch grpcMetric.Type {
	case common.MetricTypeGauge:
		if grpcMetric.Value == nil {
			return MetricModel{}, status.Error(codes.InvalidArgument,
				fmt.Sprintf("gauge metric %s has nil value", grpcMetric.Id))
		}
		appMetric = *NewMetricModel(grpcMetric.Id, grpcMetric.Type, 0, *grpcMetric.Value)

	case common.MetricTypeCounter:
		if grpcMetric.Delta == nil {
			return MetricModel{}, status.Error(codes.InvalidArgument,
				fmt.Sprintf("counter metric %s has nil delta", grpcMetric.Id))
		}
		appMetric = *NewMetricModel(grpcMetric.Id, grpcMetric.Type, *grpcMetric.Delta, 0)

	default:
		return MetricModel{}, status.Error(codes.InvalidArgument,
			fmt.Sprintf("unknown metric type: %s", grpcMetric.Type))
	}

	return appMetric, nil
}
