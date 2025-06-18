package server

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type config struct {
	ServerAddress   string `env:"ADDRESS"`
	StoreInterval   uint   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func PrepareConfig() *config {
	cfg := &config{}
	parseFlags(cfg)
	parseEnvOpts(cfg)
	return cfg
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address and port to start server")
	flag.UintVar(&cfg.StoreInterval, "i", 300, "store interval (seconds)")
	flag.StringVar(&cfg.FileStoragePath, "f", "data.json", "data dump file path")
	flag.BoolVar(&cfg.Restore, "r", false, "restore dump (bool)")
	flag.Parse()
}

func parseEnvOpts(cfg *config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
