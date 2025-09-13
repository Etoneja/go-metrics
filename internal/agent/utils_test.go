package agent

import (
	"testing"
)

func TestEnsureEndpointProtocol(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		protocol string
		expected string
	}{
		{
			name:     "endpoint without protocol",
			endpoint: "example.com",
			protocol: "https",
			expected: "https://example.com",
		},
		{
			name:     "endpoint with existing protocol",
			endpoint: "http://example.com",
			protocol: "https",
			expected: "http://example.com",
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			protocol: "https",
			expected: "https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureEndpointProtocol(tt.endpoint, tt.protocol)
			if result != tt.expected {
				t.Errorf("ensureEndpointProtocol(%q, %q) = %q, expected %q", 
					tt.endpoint, tt.protocol, result, tt.expected)
			}
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		parts    []string
		expected string
	}{
		{
			name:     "endpoint without trailing slash",
			endpoint: "https://example.com",
			parts:    []string{"api", "v1", "users"},
			expected: "https://example.com/api/v1/users",
		},
		{
			name:     "endpoint with trailing slash",
			endpoint: "https://example.com/",
			parts:    []string{"api", "v1", "users"},
			expected: "https://example.com/api/v1/users",
		},
		{
			name:     "no parts provided",
			endpoint: "https://example.com",
			parts:    []string{},
			expected: "https://example.com/",
		},
		{
			name:     "empty parts",
			endpoint: "https://example.com",
			parts:    []string{"", "", ""},
			expected: "https://example.com///",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildURL(tt.endpoint, tt.parts...)
			if result != tt.expected {
				t.Errorf("buildURL(%q, %v) = %q, expected %q", 
					tt.endpoint, tt.parts, result, tt.expected)
			}
		})
	}
}
