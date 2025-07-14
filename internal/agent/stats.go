package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/etoneja/go-metrics/internal/models"
)

func newStats() *Stats {
	return &Stats{
		mu: &sync.RWMutex{},
		collectors: []Collecter{
			NewAnyCollector(),
			NewMemCollector(),
			NewPSCollector(),
		}}
}

type Stats struct {
	mu         *sync.RWMutex
	collectors []Collecter
	metrics    []models.MetricModel
}

func (s *Stats) collect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	resultCh := make(chan Result)

	var wg sync.WaitGroup

	for _, col := range s.collectors {
		wg.Add(1)
		col := col
		go func() {
			defer wg.Done()
			col.Collect(ctx, resultCh)
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var errs []error
	var metrics []models.MetricModel

Loop:
	for {
		select {
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			break Loop
		case res, ok := <-resultCh:
			if !ok {
				break Loop
			}
			if res.err != nil {
				errs = append(errs, res.err)
				continue
			}
			metrics = append(metrics, res.metric)
		}
	}

	if ctx.Err() != nil {
		errs = append(errs, fmt.Errorf("collection interrupted: %w", ctx.Err()))
	}

	if len(errs) > 0 {
		s.metrics = nil
		return errors.Join(errs...)
	}

	s.metrics = metrics

	return nil
}

func (s *Stats) GetMetrics() []models.MetricModel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metricsCopy := make([]models.MetricModel, len(s.metrics))
	copy(metricsCopy, s.metrics)

	return metricsCopy
}
