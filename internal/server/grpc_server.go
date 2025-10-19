package server

import (
	"fmt"
	"net"

	"github.com/etoneja/go-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func StartGRPCServer(store Storager, logger *zap.Logger, cfg *config, serverErrChan chan<- error) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(logger),
			TrustedSubnetInterceptor(cfg.TrustedSubnet, logger),
		),
	)

	grpcMetricsServer := NewGRPCServer(store, logger)

	proto.RegisterMetricsServiceServer(grpcServer, grpcMetricsServer)

	listener, err := net.Listen("tcp", cfg.ServerGRPCAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen gRPC on %s: %w", cfg.ServerGRPCAddress, err)
	}

	go func() {
		logger.Info("gRPC server starting", zap.String("addr", cfg.ServerGRPCAddress))
		err := grpcServer.Serve(listener)
		if err != nil && err != grpc.ErrServerStopped {
			logger.Error("gRPC server failed", zap.Error(err))
			serverErrChan <- err
		}
	}()

	return grpcServer, nil
}
