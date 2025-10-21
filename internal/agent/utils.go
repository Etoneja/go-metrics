package agent

import (
	"log"
	"net"
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

func getOutboundIP(endpoint string) (net.IP, error) {
	conn, err := net.Dial("udp", endpoint)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close UDP connection: %v", err)
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
