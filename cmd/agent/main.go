package main

import (
	"log"
	"time"
)

func main() {
	cfg := prepareConfig()

	log.Println("Service started")

	log.Println("ServerEndpoint", cfg.ServerEndpoint)
	log.Println("PollInterval", cfg.PollInterval)
	log.Println("ReportInterval", cfg.ReportInterval)

	stats := newStats()

	pollDuration := time.Second * time.Duration(cfg.PollInterval)
	poller := newPoller(stats, pollDuration)

	reportDuration := time.Second * time.Duration(cfg.ReportInterval)
	reporter := NewReporter(stats, cfg.ServerEndpoint, reportDuration)

	go poller.runRoutine()
	go reporter.runRoutine()

	select {}
}
