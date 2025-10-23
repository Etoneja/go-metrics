package server

import (
	"context"
	"fmt"
	"net"

	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func createGRPCServer(logger *zap.Logger, cfg *config) *grpc.Server {
	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(logger),
			TrustedSubnetInterceptor(cfg.TrustedSubnet, logger),
		),
	)
}

func registerServices(server *grpc.Server, store Storager, logger *zap.Logger) {
	grpcMetricsServer := NewGRPCServer(store, logger)
	proto.RegisterMetricsServiceServer(server, grpcMetricsServer)
}

func startServing(server *grpc.Server, addr string, logger *zap.Logger, serverErrChan chan<- error) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen gRPC on %s: %w", addr, err)
	}

	go func() {
		logger.Info("gRPC server starting", zap.String("addr", addr))
		err := server.Serve(listener)
		if err != nil && err != grpc.ErrServerStopped {
			logger.Error("gRPC server failed", zap.Error(err))
			serverErrChan <- err
		}
	}()

	return nil
}

func StartGRPCServer(store Storager, logger *zap.Logger, cfg *config, serverErrChan chan<- error) (*grpc.Server, error) {
	grpcServer := createGRPCServer(logger, cfg)
	registerServices(grpcServer, store, logger)

	if err := startServing(grpcServer, cfg.ServerGRPCAddress, logger, serverErrChan); err != nil {
		return nil, err
	}

	return grpcServer, nil
}

func StopGRPCServer(grpcServer *grpc.Server, shutdownCtx context.Context) {
	logger.Get().Info("Stopping gRPC server...")
	if grpcServer != nil {
		grpcStopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(grpcStopped)
		}()

		select {
		case <-grpcStopped:
			logger.Get().Info("gRPC server stopped gracefully")
		case <-shutdownCtx.Done():
			logger.Get().Warn("gRPC server forced to stop")
			grpcServer.Stop()
		}
	} else {
		logger.Get().Info("gRPC server was not started")
	}
}
