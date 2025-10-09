package agent

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/etoneja/go-metrics/internal/common"
)

type config struct {
	ServerEndpoint string `env:"ADDRESS" json:"address"`
	PollInterval   uint   `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval uint   `env:"REPORT_INTERVAL" json:"report_interval"`
	HashKey        string `env:"KEY" json:"-"`
	RateLimit      uint   `env:"RATE_LIMIT" json:"-"`
	CryptoKey      string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigFile     string `env:"CONFIG" json:"-"`
}

func normalizeConfig(cfg *config) {
	cfg.ServerEndpoint = ensureEndpointProtocol(cfg.ServerEndpoint, defaultServerEndpointProtocol)
}

func PrepareConfig() (*config, error) {
	cfg := &config{
		ServerEndpoint: "http://localhost:8080",
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
	normalizeConfig(cfg)
	return cfg, nil
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.ServerEndpoint, "a", cfg.ServerEndpoint, "address and port to send metrics")
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
