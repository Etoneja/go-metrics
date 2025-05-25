package main

import (
	"log"
	"time"
)

func main() {
	cfg := prepareConfig()

	log.Println("Service started")
	stats := newStats()

	pollDuration := time.Second * time.Duration(cfg.PollInterval)
	poller := newPoller(stats, pollDuration)

	reportDuration := time.Second * time.Duration(cfg.ReportInterval)
	reporter := NewReporter(stats, cfg.ServerEnpoint, reportDuration)

	go poller.runRoutine()
	go reporter.runRoutine()

	select {}
}
