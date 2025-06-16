package server

import (
	"fmt"

	"github.com/etoneja/go-metrics/internal/common"
)

type storageHelper struct {
	store Storager
}

func (sh *storageHelper) getMetric(metricType string, metricName string) (any, error) {
	if metricType == common.MetricTypeGauge {
		value, ok, err := sh.store.GetGauge(metricName)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("not found: %w", errNotFound)
		}

		return value, nil
	}

	if metricType == common.MetricTypeCounter {
		value, ok, err := sh.store.GetCounter(metricName)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("not found: %w", errNotFound)
		}

		return value, nil
	}
	return nil, fmt.Errorf("bad metricType: %w", errValidation)
}

func (sh *storageHelper) listMetrics() (map[string]any, error) {
	gauges, err := sh.store.ListGauges()
	if err != nil {
		return nil, err
	}
	counters, err := sh.store.ListCounters()
	if err != nil {
		return nil, err
	}

	res := make(map[string]any)
	for k, v := range gauges {
		res[k] = v
	}
	for k, v := range counters {
		res[k] = v
	}

	return res, nil
}
