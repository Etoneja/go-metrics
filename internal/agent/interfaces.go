package agent

import (
	"context"
	"crypto/rsa"
	"net"

	"github.com/etoneja/go-metrics/internal/models"
)

type Collecter interface {
	Collect(ctx context.Context, resultCh chan<- Result)
}

type MetricClienter interface {
	SendBatch(ctx context.Context, metrics []models.MetricModel) error
	Close() error
}

type Configer interface {
	GetServerEndpoint() string
	GetServerProtocol() string
	GetHashKey() string
	GetRateLimit() uint
	GetPublicKey() *rsa.PublicKey
	getLocalIP() net.IP
}
