package main

import (
	"log"

	"github.com/etoneja/go-metrics/internal/agent"
)

func main() {
	cfg := agent.PrepareConfig()

	log.Println("Service started")

	log.Println("ServerEndpoint", cfg.ServerEndpoint)
	log.Println("PollInterval", cfg.PollInterval)
	log.Println("ReportInterval", cfg.ReportInterval)

	service := agent.NewService(cfg)
	service.Run()
}
