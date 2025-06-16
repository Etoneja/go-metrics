package server

import "sync"

type MemStorage struct {
	gaugeMu sync.RWMutex
	gauge   map[string]float64

	counterMu sync.RWMutex
	counter   map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (ms *MemStorage) GetGauge(key string) (float64, bool, error) {
	ms.gaugeMu.RLock()
	defer ms.gaugeMu.RUnlock()
	val, ok := ms.gauge[key]
	return val, ok, nil
}

func (ms *MemStorage) ListGauges() (map[string]float64, error) {
	ms.gaugeMu.RLock()
	defer ms.gaugeMu.RUnlock()
	res := make(map[string]float64)
	for k, v := range ms.gauge {
		res[k] = v
	}
	return res, nil
}

func (ms *MemStorage) SetGauge(key string, value float64) (float64, error) {
	ms.gaugeMu.Lock()
	defer ms.gaugeMu.Unlock()
	ms.gauge[key] = value
	return value, nil
}

func (ms *MemStorage) GetCounter(key string) (int64, bool, error) {
	ms.counterMu.RLock()
	defer ms.counterMu.RUnlock()
	val, ok := ms.counter[key]
	return val, ok, nil
}

func (ms *MemStorage) ListCounters() (map[string]int64, error) {
	ms.counterMu.RLock()
	defer ms.counterMu.RUnlock()
	res := make(map[string]int64)
	for k, v := range ms.counter {
		res[k] = v
	}
	return res, nil
}

func (ms *MemStorage) IncrementCounter(key string, value int64) (int64, error) {
	ms.counterMu.Lock()
	defer ms.counterMu.Unlock()

	val, ok := ms.counter[key]

	if ok {
		value += val
	}
	ms.counter[key] = value
	return value, nil
}
