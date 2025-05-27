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
	value string
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
		{kind: common.MetricTypeGauge, name: "Alloc", value: anyToString(s.memStats.Alloc)},
		{kind: common.MetricTypeGauge, name: "BuckHashSys", value: anyToString(s.memStats.BuckHashSys)},
		{kind: common.MetricTypeGauge, name: "Frees", value: anyToString(s.memStats.Frees)},
		{kind: common.MetricTypeGauge, name: "GCCPUFraction", value: anyToString(s.memStats.GCCPUFraction)},
		{kind: common.MetricTypeGauge, name: "GCSys", value: anyToString(s.memStats.GCSys)},
		{kind: common.MetricTypeGauge, name: "HeapAlloc", value: anyToString(s.memStats.HeapAlloc)},
		{kind: common.MetricTypeGauge, name: "HeapIdle", value: anyToString(s.memStats.HeapIdle)},
		{kind: common.MetricTypeGauge, name: "HeapInuse", value: anyToString(s.memStats.HeapInuse)},
		{kind: common.MetricTypeGauge, name: "HeapObjects", value: anyToString(s.memStats.HeapObjects)},
		{kind: common.MetricTypeGauge, name: "HeapReleased", value: anyToString(s.memStats.HeapReleased)},
		{kind: common.MetricTypeGauge, name: "HeapSys", value: anyToString(s.memStats.HeapSys)},
		{kind: common.MetricTypeGauge, name: "LastGC", value: anyToString(s.memStats.LastGC)},
		{kind: common.MetricTypeGauge, name: "Lookups", value: anyToString(s.memStats.Lookups)},
		{kind: common.MetricTypeGauge, name: "MCacheInuse", value: anyToString(s.memStats.MCacheInuse)},
		{kind: common.MetricTypeGauge, name: "MCacheSys", value: anyToString(s.memStats.MCacheSys)},
		{kind: common.MetricTypeGauge, name: "MSpanInuse", value: anyToString(s.memStats.MSpanInuse)},
		{kind: common.MetricTypeGauge, name: "MSpanSys", value: anyToString(s.memStats.MSpanSys)},
		{kind: common.MetricTypeGauge, name: "Mallocs", value: anyToString(s.memStats.Mallocs)},
		{kind: common.MetricTypeGauge, name: "NextGC", value: anyToString(s.memStats.NextGC)},
		{kind: common.MetricTypeGauge, name: "NumForcedGC", value: anyToString(s.memStats.NumForcedGC)},
		{kind: common.MetricTypeGauge, name: "NumGC", value: anyToString(s.memStats.NumGC)},
		{kind: common.MetricTypeGauge, name: "OtherSys", value: anyToString(s.memStats.OtherSys)},
		{kind: common.MetricTypeGauge, name: "PauseTotalNs", value: anyToString(s.memStats.PauseTotalNs)},
		{kind: common.MetricTypeGauge, name: "StackInuse", value: anyToString(s.memStats.StackInuse)},
		{kind: common.MetricTypeGauge, name: "StackSys", value: anyToString(s.memStats.StackSys)},
		{kind: common.MetricTypeGauge, name: "Sys", value: anyToString(s.memStats.Sys)},
		{kind: common.MetricTypeGauge, name: "TotalAlloc", value: anyToString(s.memStats.TotalAlloc)},

		{kind: common.MetricTypeGauge, name: "RandomValue", value: anyToString(s.RandomValue)},

		// NOTE: Should we send global counter of diff since last report?
		{kind: common.MetricTypeCounter, name: "PollCount", value: anyToString(s.PollCount)},
	}
	return metrics
}
