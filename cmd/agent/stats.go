package main

import (
	"math/rand"
	"runtime"
	"sync"
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
		{kind: metricTypeGauge, name: "Alloc", value: anyToString(s.memStats.Alloc)},
		{kind: metricTypeGauge, name: "BuckHashSys", value: anyToString(s.memStats.BuckHashSys)},
		{kind: metricTypeGauge, name: "Frees", value: anyToString(s.memStats.Frees)},
		{kind: metricTypeGauge, name: "GCCPUFraction", value: anyToString(s.memStats.GCCPUFraction)},
		{kind: metricTypeGauge, name: "GCSys", value: anyToString(s.memStats.GCSys)},
		{kind: metricTypeGauge, name: "HeapAlloc", value: anyToString(s.memStats.HeapAlloc)},
		{kind: metricTypeGauge, name: "HeapIdle", value: anyToString(s.memStats.HeapIdle)},
		{kind: metricTypeGauge, name: "HeapInuse", value: anyToString(s.memStats.HeapInuse)},
		{kind: metricTypeGauge, name: "HeapObjects", value: anyToString(s.memStats.HeapObjects)},
		{kind: metricTypeGauge, name: "HeapReleased", value: anyToString(s.memStats.HeapReleased)},
		{kind: metricTypeGauge, name: "HeapSys", value: anyToString(s.memStats.HeapSys)},
		{kind: metricTypeGauge, name: "LastGC", value: anyToString(s.memStats.LastGC)},
		{kind: metricTypeGauge, name: "Lookups", value: anyToString(s.memStats.Lookups)},
		{kind: metricTypeGauge, name: "MCacheInuse", value: anyToString(s.memStats.MCacheInuse)},
		{kind: metricTypeGauge, name: "MCacheSys", value: anyToString(s.memStats.MCacheSys)},
		{kind: metricTypeGauge, name: "MSpanInuse", value: anyToString(s.memStats.MSpanInuse)},
		{kind: metricTypeGauge, name: "MSpanSys", value: anyToString(s.memStats.MSpanSys)},
		{kind: metricTypeGauge, name: "Mallocs", value: anyToString(s.memStats.Mallocs)},
		{kind: metricTypeGauge, name: "NextGC", value: anyToString(s.memStats.NextGC)},
		{kind: metricTypeGauge, name: "NumForcedGC", value: anyToString(s.memStats.NumForcedGC)},
		{kind: metricTypeGauge, name: "NumGC", value: anyToString(s.memStats.NumGC)},
		{kind: metricTypeGauge, name: "OtherSys", value: anyToString(s.memStats.OtherSys)},
		{kind: metricTypeGauge, name: "PauseTotalNs", value: anyToString(s.memStats.PauseTotalNs)},
		{kind: metricTypeGauge, name: "StackInuse", value: anyToString(s.memStats.StackInuse)},
		{kind: metricTypeGauge, name: "StackSys", value: anyToString(s.memStats.StackSys)},
		{kind: metricTypeGauge, name: "Sys", value: anyToString(s.memStats.Sys)},
		{kind: metricTypeGauge, name: "TotalAlloc", value: anyToString(s.memStats.TotalAlloc)},

		{kind: metricTypeGauge, name: "RandomValue", value: anyToString(s.RandomValue)},

		// NOTE: Should we send global counter of diff since last report?
		{kind: metricTypeCounter, name: "PollCount", value: anyToString(s.PollCount)},
	}
	return metrics
}
