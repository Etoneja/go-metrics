package server

import (
	"context"

	"github.com/etoneja/go-metrics/internal/models"
)

type Storager interface {
	GetGauge(ctx context.Context, key string) (float64, error)
	SetGauge(ctx context.Context, key string, value float64) (float64, error)
	GetCounter(ctx context.Context, key string) (int64, error)
	IncrementCounter(ctx context.Context, key string, value int64) (int64, error)
	GetAll(ctx context.Context) ([]models.MetricModel, error)

	BatchUpdate(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error)
	Ping(ctx context.Context) error
	ShutDown()
}
