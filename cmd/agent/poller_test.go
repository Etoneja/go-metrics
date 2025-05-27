package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoller_poll(t *testing.T) {
	fakeDuration := time.Duration(time.Millisecond)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				stats:         tt.fields.stats,
				sleepDuration: fakeDuration,
			}

			assert.Equal(t, 0, tt.fields.stats.PollCount)
			assert.Equal(t, uint(0), p.iteration)

			p.poll()

			assert.Equal(t, 1, tt.fields.stats.PollCount)
			assert.Equal(t, uint(1), p.iteration)
		})
	}
}
