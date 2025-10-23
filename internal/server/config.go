package server

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/etoneja/go-metrics/internal/common"
)

type config struct {
	ServerAddress     string `env:"ADDRESS" json:"address"`
	ServerGRPCAddress string `env:"GRPC_ADDRESS" json:"grpc_address"`
	StoreInterval     uint   `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath   string `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore           bool   `env:"RESTORE" json:"restore"`
	DatabaseDSN       string `env:"DATABASE_DSN" json:"database_dsn"`
	HashKey           string `env:"KEY" json:"-"`
	CryptoKey         string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigFile        string `env:"CONFIG" json:"-"`
	TrustedSubnet     string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	privateKey        *rsa.PrivateKey
}

func (c *config) GetPrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

func PrepareConfig() (*config, error) {
	cfg := &config{
		ServerAddress:     "localhost:8080",
		ServerGRPCAddress: "",
		StoreInterval:     300,
		FileStoragePath:   "data.json",
		Restore:           false,
		DatabaseDSN:       "",
		HashKey:           "",
		CryptoKey:         "",
		ConfigFile:        "",
		TrustedSubnet:     "",
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

	privateKey, err := common.LoadPrivateKey(cfg.CryptoKey)
	if err != nil {
		return nil, err
	}
	cfg.privateKey = privateKey

	return cfg, nil
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "address and port to start http server")
	flag.StringVar(&cfg.ServerGRPCAddress, "g", cfg.ServerGRPCAddress, "address and port to start grpc server")
	flag.UintVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "store interval (seconds)")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "data dump file path")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "restore dump (bool)")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database DSN")
	flag.StringVar(&cfg.HashKey, "k", cfg.HashKey, "Hash key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Crypto key")
	flag.StringVar(&cfg.ConfigFile, "c", "", "Config file path")
	flag.StringVar(&cfg.ConfigFile, "config", "", "Config file path")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Trusted subnet")
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
	if cfg.TrustedSubnet != "" {
		_, _, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return fmt.Errorf("invalid trusted subnet format '%s': %w", cfg.TrustedSubnet, err)
		}
	}
	return nil
}
