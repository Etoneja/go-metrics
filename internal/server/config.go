package server

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type config struct {
	ServerAddress string `env:"ADDRESS"`
}

func PrepareConfig() *config {
	cfg := &config{}
	parseFlags(cfg)
	parseEnvOpts(cfg)
	return cfg
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address and port to start server")
	flag.Parse()
}

func parseEnvOpts(cfg *config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
