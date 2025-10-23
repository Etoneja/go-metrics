package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMetricClient struct {
	mu             sync.Mutex
	sendBatchCalls []struct {
		ctx     context.Context
		metrics []models.MetricModel
	}
	closeCalls     int
	sendBatchError error
}

func (m *mockMetricClient) SendBatch(ctx context.Context, metrics []models.MetricModel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sendBatchCalls = append(m.sendBatchCalls, struct {
		ctx     context.Context
		metrics []models.MetricModel
	}{ctx, metrics})

	return m.sendBatchError
}

func (m *mockMetricClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closeCalls++
	return nil
}

type testCollector struct {
	metrics []models.MetricModel
	err     error
	delay   time.Duration
}

func (t *testCollector) Collect(ctx context.Context, resultCh chan<- Result) {
	if t.delay > 0 {
		select {
		case <-time.After(t.delay):
		case <-ctx.Done():
			return
		}
	}

	if t.err != nil {
		select {
		case <-ctx.Done():
			return
		case resultCh <- Result{err: t.err}:
		}
	} else {
		for _, metric := range t.metrics {
			select {
			case <-ctx.Done():
				return
			case resultCh <- Result{metric: metric}:
			}
		}
	}
}

func testMetrics() []models.MetricModel {
	return []models.MetricModel{
		{ID: "test1", MType: "gauge", Value: common.Float64Ptr(1.23)},
		{ID: "test2", MType: "counter", Delta: common.Int64Ptr(42)},
	}
}

func TestReporter_WithRealStats(t *testing.T) {
	t.Run("report with collected metrics", func(t *testing.T) {
		mockClient := &mockMetricClient{}

		stats := &Stats{
			mu: &sync.RWMutex{},
			collectors: []Collecter{
				&testCollector{metrics: testMetrics()},
			},
		}

		cfg := &config{ReportInterval: 10}
		reporter := newReporter(stats, cfg, mockClient)

		err := stats.collect(context.Background())
		require.NoError(t, err)

		reporter.report(context.Background())

		assert.Equal(t, uint(1), reporter.iteration)
		assert.Equal(t, 1, len(mockClient.sendBatchCalls))
		assert.Equal(t, testMetrics(), mockClient.sendBatchCalls[0].metrics)
	})

	t.Run("report without collection - no metrics", func(t *testing.T) {
		mockClient := &mockMetricClient{}

		stats := &Stats{
			mu:         &sync.RWMutex{},
			collectors: []Collecter{&testCollector{metrics: testMetrics()}},
		}

		reporter := newReporter(stats, &config{ReportInterval: 10}, mockClient)

		reporter.report(context.Background())

		assert.Equal(t, uint(1), reporter.iteration)
		assert.Equal(t, 0, len(mockClient.sendBatchCalls))
	})

	t.Run("stats collection error", func(t *testing.T) {
		mockClient := &mockMetricClient{}

		stats := &Stats{
			mu:         &sync.RWMutex{},
			collectors: []Collecter{&testCollector{err: assert.AnError}},
		}

		reporter := newReporter(stats, &config{ReportInterval: 10}, mockClient)

		err := stats.collect(context.Background())
		require.Error(t, err)

		metrics := stats.GetMetrics()
		assert.Empty(t, metrics)

		reporter.report(context.Background())

		assert.Equal(t, uint(1), reporter.iteration)
		assert.Equal(t, 0, len(mockClient.sendBatchCalls))
	})
}

func TestReporter_WithRealCollectors(t *testing.T) {
	t.Run("using real anyCollector", func(t *testing.T) {
		mockClient := &mockMetricClient{}

		stats := &Stats{
			mu:         &sync.RWMutex{},
			collectors: []Collecter{NewAnyCollector()},
		}

		reporter := newReporter(stats, &config{ReportInterval: 10}, mockClient)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := stats.collect(ctx)
		require.NoError(t, err)

		metrics := stats.GetMetrics()
		require.NotEmpty(t, metrics)

		metricNames := make(map[string]bool)
		for _, metric := range metrics {
			metricNames[metric.ID] = true
		}

		assert.True(t, metricNames["RandomValue"])
		assert.True(t, metricNames["PollCount"])

		reporter.report(context.Background())

		assert.Equal(t, uint(1), reporter.iteration)
		assert.Equal(t, 1, len(mockClient.sendBatchCalls))
		assert.Equal(t, metrics, mockClient.sendBatchCalls[0].metrics)
	})

	t.Run("using real memCollector", func(t *testing.T) {
		mockClient := &mockMetricClient{}

		stats := &Stats{
			mu:         &sync.RWMutex{},
			collectors: []Collecter{NewMemCollector()},
		}

		reporter := newReporter(stats, &config{ReportInterval: 10}, mockClient)

		err := stats.collect(context.Background())
		require.NoError(t, err)

		metrics := stats.GetMetrics()
		require.NotEmpty(t, metrics)

		hasMemMetric := false
		for _, metric := range metrics {
			if metric.ID == "Alloc" || metric.ID == "HeapAlloc" {
				hasMemMetric = true
				break
			}
		}
		assert.True(t, hasMemMetric)

		reporter.report(context.Background())

		assert.Equal(t, 1, len(mockClient.sendBatchCalls))
	})
}

func TestReporter_RunRoutineWithRealStats(t *testing.T) {
	t.Run("periodic reporting with real stats", func(t *testing.T) {
		mockClient := &mockMetricClient{}

		stats := &Stats{
			mu:         &sync.RWMutex{},
			collectors: []Collecter{NewAnyCollector()},
		}

		cfg := &config{ReportInterval: 1} // 1 second
		reporter := newReporter(stats, cfg, mockClient)

		err := stats.collect(context.Background())
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- reporter.runRoutine(ctx)
		}()

		time.Sleep(2200 * time.Millisecond)
		cancel()

		err = <-errCh
		require.Error(t, err)

		assert.GreaterOrEqual(t, len(mockClient.sendBatchCalls), 2)
		assert.GreaterOrEqual(t, reporter.iteration, uint(2))
	})
}

func TestReporter_Integration(t *testing.T) {
	t.Run("full flow with all real collectors", func(t *testing.T) {
		mockClient := &mockMetricClient{}

		stats := newStats()

		cfg := &config{ReportInterval: 10}
		reporter := newReporter(stats, cfg, mockClient)

		err := stats.collect(context.Background())
		require.NoError(t, err)

		initialMetrics := stats.GetMetrics()
		require.NotEmpty(t, initialMetrics)

		reporter.report(context.Background())

		assert.Equal(t, 1, len(mockClient.sendBatchCalls))
		sentMetrics := mockClient.sendBatchCalls[0].metrics
		assert.Equal(t, len(initialMetrics), len(sentMetrics))

		initialMap := make(map[string]models.MetricModel)
		for _, m := range initialMetrics {
			initialMap[m.ID] = m
		}

		for _, sent := range sentMetrics {
			initial, exists := initialMap[sent.ID]
			assert.True(t, exists)
			assert.Equal(t, initial.MType, sent.MType)
		}
	})
}
