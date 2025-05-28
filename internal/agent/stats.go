package agent

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/etoneja/go-metrics/internal/common"
)

type metric struct {
	kind  string
	name  string
	value any
}

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

func (s *Stats) dump() []metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := []metric{
		{kind: common.MetricTypeGauge, name: "Alloc", value: s.memStats.Alloc},
		{kind: common.MetricTypeGauge, name: "BuckHashSys", value: s.memStats.BuckHashSys},
		{kind: common.MetricTypeGauge, name: "Frees", value: s.memStats.Frees},
		{kind: common.MetricTypeGauge, name: "GCCPUFraction", value: s.memStats.GCCPUFraction},
		{kind: common.MetricTypeGauge, name: "GCSys", value: s.memStats.GCSys},
		{kind: common.MetricTypeGauge, name: "HeapAlloc", value: s.memStats.HeapAlloc},
		{kind: common.MetricTypeGauge, name: "HeapIdle", value: s.memStats.HeapIdle},
		{kind: common.MetricTypeGauge, name: "HeapInuse", value: s.memStats.HeapInuse},
		{kind: common.MetricTypeGauge, name: "HeapObjects", value: s.memStats.HeapObjects},
		{kind: common.MetricTypeGauge, name: "HeapReleased", value: s.memStats.HeapReleased},
		{kind: common.MetricTypeGauge, name: "HeapSys", value: s.memStats.HeapSys},
		{kind: common.MetricTypeGauge, name: "LastGC", value: s.memStats.LastGC},
		{kind: common.MetricTypeGauge, name: "Lookups", value: s.memStats.Lookups},
		{kind: common.MetricTypeGauge, name: "MCacheInuse", value: s.memStats.MCacheInuse},
		{kind: common.MetricTypeGauge, name: "MCacheSys", value: s.memStats.MCacheSys},
		{kind: common.MetricTypeGauge, name: "MSpanInuse", value: s.memStats.MSpanInuse},
		{kind: common.MetricTypeGauge, name: "MSpanSys", value: s.memStats.MSpanSys},
		{kind: common.MetricTypeGauge, name: "Mallocs", value: s.memStats.Mallocs},
		{kind: common.MetricTypeGauge, name: "NextGC", value: s.memStats.NextGC},
		{kind: common.MetricTypeGauge, name: "NumForcedGC", value: s.memStats.NumForcedGC},
		{kind: common.MetricTypeGauge, name: "NumGC", value: s.memStats.NumGC},
		{kind: common.MetricTypeGauge, name: "OtherSys", value: s.memStats.OtherSys},
		{kind: common.MetricTypeGauge, name: "PauseTotalNs", value: s.memStats.PauseTotalNs},
		{kind: common.MetricTypeGauge, name: "StackInuse", value: s.memStats.StackInuse},
		{kind: common.MetricTypeGauge, name: "StackSys", value: s.memStats.StackSys},
		{kind: common.MetricTypeGauge, name: "Sys", value: s.memStats.Sys},
		{kind: common.MetricTypeGauge, name: "TotalAlloc", value: s.memStats.TotalAlloc},

		{kind: common.MetricTypeGauge, name: "RandomValue", value: s.RandomValue},

		// NOTE: Should we send global counter of diff since last report?
		{kind: common.MetricTypeCounter, name: "PollCount", value: s.PollCount},
	}
	return metrics
}
