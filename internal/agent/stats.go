package agent

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
)

func newStats() *Stats {
	return &Stats{
		memStats: &runtime.MemStats{},
	}
}

type Stats struct {
	mu          sync.RWMutex
	memStats    *runtime.MemStats
	RandomValue int
	PollCount   int
}

func (s *Stats) getCounter() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.PollCount
}

func (s *Stats) collect() {
	s.mu.Lock()
	defer s.mu.Unlock()

	runtime.ReadMemStats(s.memStats)
	s.RandomValue = rand.Intn(maxRandNum)
	s.PollCount++
}

func (s *Stats) dump() []*models.MetricModel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := []*models.MetricModel{
		models.NewMetricModel("Alloc", common.MetricTypeGauge, 0, float64(s.memStats.Alloc)),
		models.NewMetricModel("BuckHashSys", common.MetricTypeGauge, 0, float64(s.memStats.BuckHashSys)),
		models.NewMetricModel("Frees", common.MetricTypeGauge, 0, float64(s.memStats.Frees)),
		models.NewMetricModel("GCCPUFraction", common.MetricTypeGauge, 0, s.memStats.GCCPUFraction),
		models.NewMetricModel("GCSys", common.MetricTypeGauge, 0, float64(s.memStats.GCSys)),
		models.NewMetricModel("HeapAlloc", common.MetricTypeGauge, 0, float64(s.memStats.HeapAlloc)),
		models.NewMetricModel("HeapIdle", common.MetricTypeGauge, 0, float64(s.memStats.HeapIdle)),
		models.NewMetricModel("HeapInuse", common.MetricTypeGauge, 0, float64(s.memStats.HeapInuse)),
		models.NewMetricModel("HeapObjects", common.MetricTypeGauge, 0, float64(s.memStats.HeapObjects)),
		models.NewMetricModel("HeapReleased", common.MetricTypeGauge, 0, float64(s.memStats.HeapReleased)),
		models.NewMetricModel("HeapSys", common.MetricTypeGauge, 0, float64(s.memStats.HeapSys)),
		models.NewMetricModel("LastGC", common.MetricTypeGauge, 0, float64(s.memStats.LastGC)),
		models.NewMetricModel("Lookups", common.MetricTypeGauge, 0, float64(s.memStats.Lookups)),
		models.NewMetricModel("MCacheInuse", common.MetricTypeGauge, 0, float64(s.memStats.MCacheInuse)),
		models.NewMetricModel("MCacheSys", common.MetricTypeGauge, 0, float64(s.memStats.MCacheSys)),
		models.NewMetricModel("MSpanInuse", common.MetricTypeGauge, 0, float64(s.memStats.MSpanInuse)),
		models.NewMetricModel("MSpanSys", common.MetricTypeGauge, 0, float64(s.memStats.MSpanSys)),
		models.NewMetricModel("Mallocs", common.MetricTypeGauge, 0, float64(s.memStats.Mallocs)),
		models.NewMetricModel("NextGC", common.MetricTypeGauge, 0, float64(s.memStats.NextGC)),
		models.NewMetricModel("NumForcedGC", common.MetricTypeGauge, 0, float64(s.memStats.NumForcedGC)),
		models.NewMetricModel("NumGC", common.MetricTypeGauge, 0, float64(s.memStats.NumGC)),
		models.NewMetricModel("OtherSys", common.MetricTypeGauge, 0, float64(s.memStats.OtherSys)),
		models.NewMetricModel("PauseTotalNs", common.MetricTypeGauge, 0, float64(s.memStats.PauseTotalNs)),
		models.NewMetricModel("StackInuse", common.MetricTypeGauge, 0, float64(s.memStats.StackInuse)),
		models.NewMetricModel("StackSys", common.MetricTypeGauge, 0, float64(s.memStats.StackSys)),
		models.NewMetricModel("Sys", common.MetricTypeGauge, 0, float64(s.memStats.Sys)),
		models.NewMetricModel("TotalAlloc", common.MetricTypeGauge, 0, float64(s.memStats.TotalAlloc)),

		models.NewMetricModel("RandomValue", common.MetricTypeGauge, 0, float64(s.RandomValue)),

		models.NewMetricModel("PollCount", common.MetricTypeCounter, int64(s.PollCount), 0),
	}
	return metrics
}
