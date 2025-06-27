package server

import (
	"context"

	"github.com/etoneja/go-metrics/internal/models"
)

type Storager interface {
	GetGauge(key string) (float64, bool, error)
	SetGauge(key string, value float64) (float64, error)
	GetCounter(key string) (int64, bool, error)
	IncrementCounter(key string, value int64) (int64, error)
	GetAll() (*[]models.MetricModel, error)

	BatchUpdate(*[]models.MetricModel) (*[]models.MetricModel, error)
	Ping(ctx context.Context) error
	ShutDown()
}
