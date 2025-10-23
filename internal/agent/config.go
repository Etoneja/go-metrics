package agent

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/logger"
	"go.uber.org/zap"
)

type config struct {
	ServerEndpoint string `env:"ADDRESS" json:"address"`
	ServerProtocol string `env:"PROTOCOL" json:"protocol"`
	PollInterval   uint   `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval uint   `env:"REPORT_INTERVAL" json:"report_interval"`
	HashKey        string `env:"KEY" json:"-"`
	RateLimit      uint   `env:"RATE_LIMIT" json:"-"`
	CryptoKey      string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigFile     string `env:"CONFIG" json:"-"`
	publicKey      *rsa.PublicKey
	localIP        net.IP
	mu             sync.RWMutex
}

func (c *config) getLocalIP() net.IP {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.localIP != nil {
		return c.localIP
	}

	ip, err := getOutboundIP(c.ServerEndpoint)
	if err != nil {
		logger.Get().Error("Failed to get outbound IP", zap.Error(err))
		return nil
	}
	c.localIP = ip
	return c.localIP
}

func (c *config) GetServerEndpoint() string {
	return c.ServerEndpoint
}

func (c *config) GetServerProtocol() string {
	return c.ServerProtocol
}

func (c *config) GetPublicKey() *rsa.PublicKey {
	return c.publicKey
}

func (c *config) GetHashKey() string {
	return c.HashKey
}

func (c *config) GetRateLimit() uint {
	return c.RateLimit
}

func PrepareConfig() (*config, error) {
	cfg := &config{
		ServerEndpoint: "localhost:8080",
		ServerProtocol: defaultServerEndpointProtocol,
		PollInterval:   2,
		ReportInterval: 10,
		HashKey:        "",
		RateLimit:      1,
		CryptoKey:      "",
		ConfigFile:     "",
	}
	parseFlags(cfg)

	err := common.LoadJSONConfig(cfg, cfg.ConfigFile)
	if err != nil {
		return nil, err
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	parseFlags(cfg)

	err = parseEnvOpts(cfg)
	if err != nil {
		return nil, err
	}
	err = validateConfig(cfg)
	if err != nil {
		return nil, err
	}

	publicKey, err := common.LoadPublicKey(cfg.CryptoKey)
	if err != nil {
		return nil, err
	}
	cfg.publicKey = publicKey

	return cfg, nil
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.ServerEndpoint, "a", cfg.ServerEndpoint, "address and port to send metrics")
	flag.StringVar(&cfg.ServerProtocol, "protocol", cfg.ServerProtocol, "server protocol to send metrics")
	flag.UintVar(&cfg.PollInterval, "p", cfg.PollInterval, "poll interval (seconds)")
	flag.UintVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "report interval (seconds)")
	flag.StringVar(&cfg.HashKey, "k", cfg.HashKey, "Hash key")
	flag.UintVar(&cfg.RateLimit, "l", cfg.RateLimit, "Rate limit ")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Crypto key")
	flag.StringVar(&cfg.ConfigFile, "c", "", "Config file path")
	flag.StringVar(&cfg.ConfigFile, "config", "", "Config file path")
	flag.Parse()
}

func parseEnvOpts(cfg *config) error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}
	return nil
}

func validateConfig(cfg *config) error {
	validProtocols := map[string]struct{}{
		"http": {},
		"grpc": {},
	}

	if _, valid := validProtocols[cfg.ServerProtocol]; !valid {
		return fmt.Errorf("invalid protocol '%s'", cfg.ServerProtocol)
	}
	return nil
}
