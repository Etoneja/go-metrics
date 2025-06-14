package main

import (
	"net/http"

	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/etoneja/go-metrics/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := server.PrepareConfig()

	logger.Init(false)
	defer logger.Sync()

	store := server.NewMemStorage()
	router := server.NewRouter(store)

	logger.Get().Info("Server started",
		zap.String("ServerAddress", cfg.ServerAddress),
	)

	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		logger.Get().Fatal("Server failed",
			zap.Error(err),
		)
	}

}
