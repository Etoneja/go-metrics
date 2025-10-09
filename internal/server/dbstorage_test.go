package server

import (
	"errors"
	"net"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsDBRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "network error",
			err:      &net.DNSError{},
			expected: true,
		},
		{
			name:     "pg error",
			err:      &pgconn.PgError{},
			expected: true,
		},
		{
			name:     "regular error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDBRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isDBRetryableError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
