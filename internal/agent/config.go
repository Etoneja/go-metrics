package agent

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type config struct {
	ServerEndpoint string `env:"ADDRESS"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	HashKey        string `env:"KEY"`
}

func normalizeConfig(cfg *config) {
	cfg.ServerEndpoint = ensureEndpointProtocol(cfg.ServerEndpoint, defaultServerEndpointProtocol)
}

func PrepareConfig() *config {
	cfg := &config{}
	parseFlags(cfg)
	parseEnvOpts(cfg)
	normalizeConfig(cfg)
	return cfg
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.ServerEndpoint, "a", "http://localhost:8080", "address and port to send metrics")
	flag.UintVar(&cfg.PollInterval, "p", 2, "poll interval (seconds)")
	flag.UintVar(&cfg.ReportInterval, "r", 10, "report interval (seconds)")
	flag.StringVar(&cfg.HashKey, "k", "", "Hash key")
	flag.Parse()
}

func parseEnvOpts(cfg *config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
