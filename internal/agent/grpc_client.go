package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/etoneja/go-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client proto.MetricsServiceClient
	cfg    Configer
}

func NewGRPCMetricClient(cfg Configer) (*GRPCClient, error) {
	conn, err := grpc.NewClient(cfg.GetServerEndpoint(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 10 * time.Second,
		}),
		grpc.WithUnaryInterceptor(retryInterceptor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: proto.NewMetricsServiceClient(conn),
		cfg:    cfg,
	}, nil
}

func (c *GRPCClient) SendBatch(ctx context.Context, metrics []models.MetricModel) error {
	if len(metrics) == 0 {
		return nil
	}

	grpcMetrics, err := models.MetricModelsToGRPC(metrics)
	if err != nil {
		return err
	}

	req := &proto.BatchUpdateRequest{
		Metrics: grpcMetrics,
	}

	if ip := c.cfg.getLocalIP(); ip != nil {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", ip.String())
	}

	_, err = c.client.BatchUpdate(ctx, req)
	return err
}

func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func retryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	backoffSchedule := common.DefaultBackoffSchedule
	attemptNum := 0

	for _, backoff := range backoffSchedule {
		attemptNum++
		attemptString := fmt.Sprintf("[%d/%d]", attemptNum, len(backoffSchedule)+1)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logger.Get().Info("gRPC request succeeded",
				zap.String("attempt", attemptString),
				zap.String("method", method),
			)
			return nil
		}

		if shouldRetry(err) {
			logger.Get().Warn("gRPC request failed, retrying",
				zap.String("attempt", attemptString),
				zap.Error(err),
			)
			time.Sleep(backoff)
			continue
		}

		logger.Get().Error("gRPC request failed",
			zap.String("attempt", attemptString),
			zap.Error(err),
		)
		return err
	}

	return fmt.Errorf("all gRPC attempts failed")
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.DeadlineExceeded, codes.Unavailable, codes.ResourceExhausted, codes.Internal, codes.Unknown:
		return true
	default:
		return false
	}
}
