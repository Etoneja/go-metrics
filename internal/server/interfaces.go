package server

type Storager interface {
	GetGauge(key string) (float64, bool, error)
	ListGauges() (map[string]float64, error)
	SetGauge(key string, value float64) error
	GetCounter(key string) (int64, bool, error)
	ListCounters() (map[string]int64, error)
	IncrementCounter(key string, value int64) error
}
