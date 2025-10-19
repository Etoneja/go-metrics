package server

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func loggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		logger.Debug("gRPC request started",
			zap.String("method", info.FullMethod),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := status.Code(err)

		logger.Info("gRPC request completed",
			zap.String("method", info.FullMethod),
			zap.String("duration", duration.String()),
			zap.String("status", statusCode.String()),
			zap.Error(err),
		)

		return resp, err
	}
}
