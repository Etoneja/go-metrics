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
		{kind: common.MetricTypeGauge, name: "Alloc", value: common.AnyToString(s.memStats.Alloc)},
		{kind: common.MetricTypeGauge, name: "BuckHashSys", value: common.AnyToString(s.memStats.BuckHashSys)},
		{kind: common.MetricTypeGauge, name: "Frees", value: common.AnyToString(s.memStats.Frees)},
		{kind: common.MetricTypeGauge, name: "GCCPUFraction", value: common.AnyToString(s.memStats.GCCPUFraction)},
		{kind: common.MetricTypeGauge, name: "GCSys", value: common.AnyToString(s.memStats.GCSys)},
		{kind: common.MetricTypeGauge, name: "HeapAlloc", value: common.AnyToString(s.memStats.HeapAlloc)},
		{kind: common.MetricTypeGauge, name: "HeapIdle", value: common.AnyToString(s.memStats.HeapIdle)},
		{kind: common.MetricTypeGauge, name: "HeapInuse", value: common.AnyToString(s.memStats.HeapInuse)},
		{kind: common.MetricTypeGauge, name: "HeapObjects", value: common.AnyToString(s.memStats.HeapObjects)},
		{kind: common.MetricTypeGauge, name: "HeapReleased", value: common.AnyToString(s.memStats.HeapReleased)},
		{kind: common.MetricTypeGauge, name: "HeapSys", value: common.AnyToString(s.memStats.HeapSys)},
		{kind: common.MetricTypeGauge, name: "LastGC", value: common.AnyToString(s.memStats.LastGC)},
		{kind: common.MetricTypeGauge, name: "Lookups", value: common.AnyToString(s.memStats.Lookups)},
		{kind: common.MetricTypeGauge, name: "MCacheInuse", value: common.AnyToString(s.memStats.MCacheInuse)},
		{kind: common.MetricTypeGauge, name: "MCacheSys", value: common.AnyToString(s.memStats.MCacheSys)},
		{kind: common.MetricTypeGauge, name: "MSpanInuse", value: common.AnyToString(s.memStats.MSpanInuse)},
		{kind: common.MetricTypeGauge, name: "MSpanSys", value: common.AnyToString(s.memStats.MSpanSys)},
		{kind: common.MetricTypeGauge, name: "Mallocs", value: common.AnyToString(s.memStats.Mallocs)},
		{kind: common.MetricTypeGauge, name: "NextGC", value: common.AnyToString(s.memStats.NextGC)},
		{kind: common.MetricTypeGauge, name: "NumForcedGC", value: common.AnyToString(s.memStats.NumForcedGC)},
		{kind: common.MetricTypeGauge, name: "NumGC", value: common.AnyToString(s.memStats.NumGC)},
		{kind: common.MetricTypeGauge, name: "OtherSys", value: common.AnyToString(s.memStats.OtherSys)},
		{kind: common.MetricTypeGauge, name: "PauseTotalNs", value: common.AnyToString(s.memStats.PauseTotalNs)},
		{kind: common.MetricTypeGauge, name: "StackInuse", value: common.AnyToString(s.memStats.StackInuse)},
		{kind: common.MetricTypeGauge, name: "StackSys", value: common.AnyToString(s.memStats.StackSys)},
		{kind: common.MetricTypeGauge, name: "Sys", value: common.AnyToString(s.memStats.Sys)},
		{kind: common.MetricTypeGauge, name: "TotalAlloc", value: common.AnyToString(s.memStats.TotalAlloc)},

		{kind: common.MetricTypeGauge, name: "RandomValue", value: common.AnyToString(s.RandomValue)},

		// NOTE: Should we send global counter of diff since last report?
		{kind: common.MetricTypeCounter, name: "PollCount", value: common.AnyToString(s.PollCount)},
	}
	return metrics
}
