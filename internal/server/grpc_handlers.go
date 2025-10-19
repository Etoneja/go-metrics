package server

import (
	"context"

	"github.com/etoneja/go-metrics/internal/models"
	"github.com/etoneja/go-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	proto.UnimplementedMetricsServiceServer
	store  Storager
	logger *zap.Logger
}

func NewGRPCServer(store Storager, logger *zap.Logger) *GRPCServer {
	return &GRPCServer{
		store:  store,
		logger: logger,
	}
}

func (s *GRPCServer) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	err := s.store.Ping(ctx)
	if err != nil {
		s.logger.Error("gRPC Ping failed", zap.Error(err))
		return &proto.PingResponse{Success: false}, err
	}

	return &proto.PingResponse{Success: true}, nil
}

func (s *GRPCServer) BatchUpdate(ctx context.Context, req *proto.BatchUpdateRequest) (*proto.BatchUpdateResponse, error) {
	metricModels, err := models.MetricModelsFromGRPC(req.Metrics)
	if err != nil {
		return nil, err
	}

	updatedMetrics, err := s.store.BatchUpdate(ctx, metricModels)
	if err != nil {
		s.logger.Error("gRPC BatchUpdate failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	responseMetrics, err := models.MetricModelsToGRPC(updatedMetrics)
	if err != nil {
		s.logger.Error("failed to convert metrics for response", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.BatchUpdateResponse{
		Metrics: responseMetrics,
	}, nil
}

func (s *GRPCServer) ListMetrics(ctx context.Context, req *proto.ListMetricsRequest) (*proto.ListMetricsResponse, error) {
	metrics, err := s.store.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to get metrics", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get metrics")
	}

	grpcMetrics, err := models.MetricModelsToGRPC(metrics)
	if err != nil {
		s.logger.Error("failed to convert metrics for response", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.ListMetricsResponse{
		Metrics: grpcMetrics,
	}, nil
}
