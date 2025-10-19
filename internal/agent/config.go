package agent

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/etoneja/go-metrics/internal/common"
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
}

func PrepareConfig() (*config, error) {
	cfg := &config{
		ServerEndpoint: "localhost:8080",
		ServerProtocol: "http",
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
	}

	if _, valid := validProtocols[cfg.ServerProtocol]; !valid {
		return fmt.Errorf("invalid protocol '%s'", cfg.ServerProtocol)
	}
	return nil
}
