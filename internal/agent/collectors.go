package agent

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type Result struct {
	metric models.MetricModel
	err    error
}

func sendResult(ctx context.Context, res Result, resultCh chan<- Result) {
	select {
	case <-ctx.Done():
		return
	case resultCh <- res:
	}
}

type memCollector struct {
	memStats *runtime.MemStats
}

func NewMemCollector() *memCollector {
	return &memCollector{memStats: &runtime.MemStats{}}
}

func (m *memCollector) Collect(ctx context.Context, resultCh chan<- Result) {
	runtime.ReadMemStats(m.memStats)
	metrics := []*models.MetricModel{
		models.NewMetricModel("Alloc", common.MetricTypeGauge, 0, float64(m.memStats.Alloc)),
		models.NewMetricModel("BuckHashSys", common.MetricTypeGauge, 0, float64(m.memStats.BuckHashSys)),
		models.NewMetricModel("Frees", common.MetricTypeGauge, 0, float64(m.memStats.Frees)),
		models.NewMetricModel("GCCPUFraction", common.MetricTypeGauge, 0, m.memStats.GCCPUFraction),
		models.NewMetricModel("GCSys", common.MetricTypeGauge, 0, float64(m.memStats.GCSys)),
		models.NewMetricModel("HeapAlloc", common.MetricTypeGauge, 0, float64(m.memStats.HeapAlloc)),
		models.NewMetricModel("HeapIdle", common.MetricTypeGauge, 0, float64(m.memStats.HeapIdle)),
		models.NewMetricModel("HeapInuse", common.MetricTypeGauge, 0, float64(m.memStats.HeapInuse)),
		models.NewMetricModel("HeapObjects", common.MetricTypeGauge, 0, float64(m.memStats.HeapObjects)),
		models.NewMetricModel("HeapReleased", common.MetricTypeGauge, 0, float64(m.memStats.HeapReleased)),
		models.NewMetricModel("HeapSys", common.MetricTypeGauge, 0, float64(m.memStats.HeapSys)),
		models.NewMetricModel("LastGC", common.MetricTypeGauge, 0, float64(m.memStats.LastGC)),
		models.NewMetricModel("Lookups", common.MetricTypeGauge, 0, float64(m.memStats.Lookups)),
		models.NewMetricModel("MCacheInuse", common.MetricTypeGauge, 0, float64(m.memStats.MCacheInuse)),
		models.NewMetricModel("MCacheSys", common.MetricTypeGauge, 0, float64(m.memStats.MCacheSys)),
		models.NewMetricModel("MSpanInuse", common.MetricTypeGauge, 0, float64(m.memStats.MSpanInuse)),
		models.NewMetricModel("MSpanSys", common.MetricTypeGauge, 0, float64(m.memStats.MSpanSys)),
		models.NewMetricModel("Mallocs", common.MetricTypeGauge, 0, float64(m.memStats.Mallocs)),
		models.NewMetricModel("NextGC", common.MetricTypeGauge, 0, float64(m.memStats.NextGC)),
		models.NewMetricModel("NumForcedGC", common.MetricTypeGauge, 0, float64(m.memStats.NumForcedGC)),
		models.NewMetricModel("NumGC", common.MetricTypeGauge, 0, float64(m.memStats.NumGC)),
		models.NewMetricModel("OtherSys", common.MetricTypeGauge, 0, float64(m.memStats.OtherSys)),
		models.NewMetricModel("PauseTotalNs", common.MetricTypeGauge, 0, float64(m.memStats.PauseTotalNs)),
		models.NewMetricModel("StackInuse", common.MetricTypeGauge, 0, float64(m.memStats.StackInuse)),
		models.NewMetricModel("StackSys", common.MetricTypeGauge, 0, float64(m.memStats.StackSys)),
		models.NewMetricModel("Sys", common.MetricTypeGauge, 0, float64(m.memStats.Sys)),
		models.NewMetricModel("TotalAlloc", common.MetricTypeGauge, 0, float64(m.memStats.TotalAlloc)),
	}
	for _, metric := range metrics {
		sendResult(ctx, Result{metric: *metric}, resultCh)
	}
}

type psCollector struct{}

func NewPSCollector() *psCollector {
	return &psCollector{}
}

func (p *psCollector) Collect(ctx context.Context, resultCh chan<- Result) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		sendResult(ctx, Result{err: err}, resultCh)
	}
	metrics := []*models.MetricModel{
		models.NewMetricModel("TotalMemory", common.MetricTypeGauge, 0, float64(vmStat.Total)),
		models.NewMetricModel("FreeMemory", common.MetricTypeGauge, 0, float64(vmStat.Free)),
	}
	cpuPercent, err := cpu.Percent(1*time.Second, true)
	if err != nil {
		sendResult(ctx, Result{err: err}, resultCh)
	}
	for i, percent := range cpuPercent {
		metrics = append(metrics, models.NewMetricModel(
			fmt.Sprintf("CPUutilization%d", i), common.MetricTypeGauge, 0, percent))
	}
	for _, metric := range metrics {
		sendResult(ctx, Result{metric: *metric}, resultCh)
	}
}

type anyCollector struct {
	pollCount int
}

func NewAnyCollector() *anyCollector {
	return &anyCollector{}
}

func (a *anyCollector) Collect(ctx context.Context, resultCh chan<- Result) {
	a.pollCount++
	randomValue := rand.Intn(maxRandNum)
	metrics := []*models.MetricModel{
		models.NewMetricModel("RandomValue", common.MetricTypeGauge, 0, float64(randomValue)),
		models.NewMetricModel("PollCount", common.MetricTypeCounter, int64(a.pollCount), 0),
	}
	for _, metric := range metrics {
		sendResult(ctx, Result{metric: *metric}, resultCh)
	}
}
