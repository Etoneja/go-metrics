package common

import (
	"context"
	"testing"
	"time"
)

func TestAnyToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"int", 42, "42"},
		{"int32", int32(42), "42"},
		{"uint32", uint32(42), "42"},
		{"float64", 3.14, "3.14"},
		{"float32", float32(3.14), "3.14"},
		{"string", "hello", "hello"},
		{"bool", true, "true"},
		{"nil", nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnyToString(tt.input)
			if result != tt.expected {
				t.Errorf("AnyToString(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetBackoffTicker(t *testing.T) {
	ctx := context.Background()
	schedule := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
	}

	ticker := GetBackoffTicker(ctx, schedule)

	count := 0
	for range ticker {
		count++
		if count >= 3 {
			break
		}
	}

	if count != 3 {
		t.Errorf("Expected 3 ticks, got %d", count)
	}
}

func TestGetBackoffTicker_WithCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	schedule := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
	}

	ticker := GetBackoffTicker(ctx, schedule)

	<-ticker
	cancel()

	_, ok := <-ticker
	if ok {
		t.Error("Ticker channel should be closed after context cancellation")
	}
}

func TestComputeHash(t *testing.T) {
	key := "secret"
	data := []byte("hello world")

	hash := СomputeHash(key, data)

	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	hash2 := СomputeHash(key, data)
	if hash != hash2 {
		t.Error("Hashes should be identical for same input")
	}

	hash3 := СomputeHash(key, []byte("different"))
	if hash == hash3 {
		t.Error("Hashes should be different for different input")
	}
}

func TestCompareHashes(t *testing.T) {
	key := "secret"
	data := []byte("test data")

	hash1 := СomputeHash(key, data)
	hash2 := СomputeHash(key, data)

	if !CompareHashes(hash1, hash2) {
		t.Error("Identical hashes should compare equal")
	}

	hash3 := СomputeHash(key, []byte("different"))
	if CompareHashes(hash1, hash3) {
		t.Error("Different hashes should not compare equal")
	}

	if CompareHashes("invalid", hash1) {
		t.Error("Invalid hashes should not compare equal")
	}

	if CompareHashes(hash1, "invalid") {
		t.Error("Invalid hashes should not compare equal")
	}
}
