package common

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

var DefaultBackoffSchedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func AnyToString(v any) string {
	switch v := v.(type) {
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func GetBackoffTicker(ctx context.Context, backoffSchedule []time.Duration) <-chan struct{} {
	ticker := make(chan struct{})

	go func() {
		defer close(ticker)

		select {
		case ticker <- struct{}{}:
		case <-ctx.Done():
			return
		}

		for _, backoffDuration := range backoffSchedule {
			select {
			case <-time.After(backoffDuration):
			case <-ctx.Done():
				return
			}

			select {
			case ticker <- struct{}{}:
			case <-ctx.Done():
				return
			}
		}

	}()

	return ticker
}
