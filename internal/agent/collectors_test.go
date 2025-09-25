package agent

import (
	"context"
	"testing"
)

func BenchmarkMemCollectorCollect(b *testing.B) {
	ctx := context.Background()
	collector := NewMemCollector()

	resultCh := make(chan Result, 1000)
	defer close(resultCh)

	go func() {
		for range resultCh {
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		collector.Collect(ctx, resultCh)
	}
}
