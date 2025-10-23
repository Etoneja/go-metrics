package agent

import (
	"net"
	"strings"

	"github.com/etoneja/go-metrics/internal/logger"
	"go.uber.org/zap"
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
			logger.Get().Error("Failed to close UDP connection", zap.Error(err))
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
