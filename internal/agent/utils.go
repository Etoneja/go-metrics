package agent

import (
	"strings"
)

func ensureEndpointProtocol(endpoint string, protocol string) string {
	if !strings.Contains(endpoint, "://") {
		endpoint = protocol + "://" + endpoint
	}
	return endpoint
}

func buildURL(endpoint string, parts ...string) string {
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}
	return endpoint + strings.Join(parts, "/")
}
