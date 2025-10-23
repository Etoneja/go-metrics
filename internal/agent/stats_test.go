package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStats_GetMetrics_EmptyInitially(t *testing.T) {
	stats := newStats()
	metrics := stats.GetMetrics()

	assert.Empty(t, metrics)
}

func TestStats_Collect_Success(t *testing.T) {
	stats := newStats()
	ctx := context.Background()

	err := stats.collect(ctx)

	assert.NoError(t, err)
	metrics := stats.GetMetrics()
	assert.NotEmpty(t, metrics)

	hasRandomValue := false
	hasPollCount := false
	hasMemMetric := false

	for _, metric := range metrics {
		switch metric.ID {
		case "RandomValue":
			hasRandomValue = true
		case "PollCount":
			hasPollCount = true
		case "Alloc":
			hasMemMetric = true
		}
	}

	assert.True(t, hasRandomValue)
	assert.True(t, hasPollCount)
	assert.True(t, hasMemMetric)
}

func TestStats_Collect_ContextCancel(t *testing.T) {
	stats := newStats()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := stats.collect(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	metrics := stats.GetMetrics()
	assert.Empty(t, metrics)
}

func TestStats_GetMetrics_ThreadSafe(t *testing.T) {
	stats := newStats()
	ctx := context.Background()

	go func() {
		stats.collect(ctx)
	}()

	assert.NotPanics(t, func() {
		_ = stats.GetMetrics()
	})
}

func TestStats_Collect_MultipleCalls(t *testing.T) {
	stats := newStats()
	ctx := context.Background()

	err1 := stats.collect(ctx)
	assert.NoError(t, err1)
	metrics1 := stats.GetMetrics()

	err2 := stats.collect(ctx)
	assert.NoError(t, err2)
	metrics2 := stats.GetMetrics()

	assert.NotEqual(t, metrics1, metrics2)

	var pollCount1, pollCount2 int64
	for _, m := range metrics1 {
		if m.ID == "PollCount" && m.Delta != nil {
			pollCount1 = *m.Delta
		}
	}
	for _, m := range metrics2 {
		if m.ID == "PollCount" && m.Delta != nil {
			pollCount2 = *m.Delta
		}
	}

	assert.Greater(t, pollCount2, pollCount1)
}
