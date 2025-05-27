package server

import (
	"fmt"
	"strconv"

	"github.com/etoneja/go-metrics/internal/common"
)

type storageHelper struct {
	store Storager
}

func (sh *storageHelper) addMetric(metricType string, metricName string, metricValue string) error {

	if metricName == "" {
		return fmt.Errorf("bad metricName: %w", validationError)
	}

	if metricType == common.MetricTypeGauge {
		num, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("bad metricValue: %w", validationError)
		}
		err = sh.store.SetGauge(metricName, num)
		if err != nil {
			return err
		}
		return nil
	}

	if metricType == common.MetricTypeCounter {
		num, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("bad metricValue: %w", validationError)
		}
		err = sh.store.IncrementCounter(metricName, num)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("bad metricType: %w", validationError)
}

func (sh *storageHelper) getMetric(metricType string, metricName string) (any, error) {
	if metricType == common.MetricTypeGauge {
		value, ok, err := sh.store.GetGauge(metricName)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("not found: %w", notFoundError)
		}

		return value, nil
	}

	if metricType == common.MetricTypeCounter {
		value, ok, err := sh.store.GetCounter(metricName)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("not found: %w", notFoundError)
		}

		return value, nil
	}
	return nil, fmt.Errorf("bad metricType: %w", validationError)
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
