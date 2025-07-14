package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
