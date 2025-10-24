package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/server"
	"github.com/etoneja/go-metrics/internal/version"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	version.Print()

	logger.Init(false)

	cfg, err := server.PrepareConfig()
	if err != nil {
		logger.Get().Fatal("Failed prepare config", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	store := server.NewStorageFromConfig(cfg)

	logger.Get().Info("Starting server",
		zap.String("ServerAddress", cfg.ServerAddress),
		zap.String("ServerGRPCAddress", cfg.ServerGRPCAddress),
		zap.Uint("StoreInterval", cfg.StoreInterval),
		zap.String("FileStoragePath", cfg.FileStoragePath),
		zap.Bool("Restore", cfg.Restore),
		zap.String("CryptoKey", cfg.CryptoKey),
		zap.String("ConfigFile", cfg.ConfigFile),
		zap.String("TrustedSubnet", cfg.TrustedSubnet),
	)

	// create http
	router := server.NewRouter(store, cfg)
	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	serverErrChan := make(chan error, 2)

	// start http
	go func() {
		logger.Get().Info("HTTP server starting", zap.String("addr", cfg.ServerAddress))
		err := srv.ListenAndServe()
		if err != nil {
			logger.Get().Error("Server failed",
				zap.Error(err),
			)
			serverErrChan <- err
		}
	}()

	// start grpc
	var grpcServer *grpc.Server
	if cfg.ServerGRPCAddress != "" {
		grpcServer, err = server.StartGRPCServer(store, logger.Get(), cfg, serverErrChan)
		if err != nil {
			logger.Get().Fatal("Failed to start gRPC server", zap.Error(err))
		}
	} else {
		logger.Get().Info("gRPC server is disabled (no address configured)")
	}

	select {
	case <-ctx.Done():
		logger.Get().Info("Received shutdown signal")
	case err := <-serverErrChan:
		logger.Get().Info("Server error",
			zap.Error(err),
		)
	}

	logger.Get().Info("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// shutdown http
	logger.Get().Info("Stopping HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Get().Error("HTTP server shutdown error", zap.Error(err))
	}

	// shutdown grpc
	server.StopGRPCServer(grpcServer, shutdownCtx)

	store.ShutDown()
	logger.Get().Info("Server(s) stopped")

}
