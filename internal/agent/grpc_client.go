package agent

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/etoneja/go-metrics/internal/models"
	"github.com/etoneja/go-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client proto.MetricsServiceClient
}

func NewGRPCClient(endpoint string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(endpoint,
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
	}, nil
}

func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *GRPCClient) PerformRequest(ctx context.Context, ip net.IP, metrics []models.MetricModel) error {
	grpcMetrics, err := models.MetricModelsToGRPC(metrics)
	if err != nil {
		return err
	}

	req := &proto.BatchUpdateRequest{
		Metrics: grpcMetrics,
	}

	if ip != nil {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", ip.String())
	}

	_, err = c.client.BatchUpdate(ctx, req)
	return err
}
