package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPoller(t *testing.T) {
	stats := &Stats{}
	pollInterval := 2 * time.Second

	poller := newPoller(stats, pollInterval)

	if poller == nil {
		t.Fatal("Expected poller instance, got nil")
	}

	if poller.stats != stats {
		t.Error("Stats not set correctly")
	}

	if poller.pollInterval != pollInterval {
		t.Errorf("Expected poll interval %v, got %v", pollInterval, poller.pollInterval)
	}
}

func TestPoller_runRoutine_ContextCancel(t *testing.T) {
	poller := &Poller{
		stats:        &Stats{},
		pollInterval: time.Hour,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := poller.runRoutine(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context canceled, got %v", err)
	}
}

func TestPoller_poll(t *testing.T) {
	fakePollInterval := time.Duration(time.Millisecond)
	type fields struct {
		stats *Stats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "base poll",
			fields: fields{
				stats: newStats(),
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				stats:        tt.fields.stats,
				pollInterval: fakePollInterval,
			}

			assert.Equal(t, 0, len(tt.fields.stats.metrics))
			assert.Equal(t, uint(0), p.iteration)

			p.poll(ctx)

			assert.NotEqual(t, 0, len(tt.fields.stats.metrics))
			assert.Equal(t, uint(1), p.iteration)
		})
	}
}
