package main

import "flag"

type config struct {
	ServerAddress string
}

func prepareConfig() *config {
	cfg := &config{}
	parseFlags(cfg)
	return cfg
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address and port to start server")
	flag.Parse()
}
