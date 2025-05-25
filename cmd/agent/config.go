package main

import "flag"

type config struct {
	ServerEnpoint  string
	PollInterval   uint
	ReportInterval uint
}

func normalizeConfig(cfg *config) {
	cfg.ServerEnpoint = ensureEndpointProtocol(cfg.ServerEnpoint, defaultServerEndpointProtocol)
}

func prepareConfig() *config {
	cfg := &config{}
	parseFlags(cfg)
	normalizeConfig(cfg)
	return cfg
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.ServerEnpoint, "a", "http://localhost:8080", "address and port to send metrics")
	flag.UintVar(&cfg.PollInterval, "p", 2, "poll interval (seconds)")
	flag.UintVar(&cfg.ReportInterval, "r", 10, "report interval (seconds)")
	flag.Parse()
}
