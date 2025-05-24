package main

type Storager interface {
	GetGauge(key string) (float64, bool, error)
	SetGauge(key string, value float64) error
	GetCounter(key string) (int64, bool, error)
	IncrementCounter(key string, value int64) error
}
