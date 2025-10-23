package agent

import (
	"crypto/rsa"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockConfig struct {
	serverEndpoint string
	serverProtocol string
	hashKey        string
	rateLimit      uint
	publicKey      *rsa.PublicKey
	localIP        net.IP
}

func (m *mockConfig) GetServerEndpoint() string    { return m.serverEndpoint }
func (m *mockConfig) GetServerProtocol() string    { return m.serverProtocol }
func (m *mockConfig) GetHashKey() string           { return m.hashKey }
func (m *mockConfig) GetRateLimit() uint           { return m.rateLimit }
func (m *mockConfig) GetPublicKey() *rsa.PublicKey { return m.publicKey }
func (m *mockConfig) getLocalIP() net.IP           { return m.localIP }

func TestNewMetricClient_HTTP(t *testing.T) {
	cfg := &mockConfig{
		serverEndpoint: "localhost:8080",
		serverProtocol: "http",
	}

	client, err := NewMetricClient(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewMetricClient_GRPC(t *testing.T) {
	cfg := &mockConfig{
		serverEndpoint: "localhost:9090",
		serverProtocol: "grpc",
	}

	client, err := NewMetricClient(cfg)

	if err != nil {
		assert.NotContains(t, err.Error(), "unsupported protocol")
	} else {
		assert.NotNil(t, client)
	}
}

func TestNewMetricClient_UnsupportedProtocol(t *testing.T) {
	cfg := &mockConfig{
		serverEndpoint: "localhost:8080",
		serverProtocol: "ftp",
	}

	client, err := NewMetricClient(cfg)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "unsupported protocol")
}
