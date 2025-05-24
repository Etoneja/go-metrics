package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Agent started")

	log.Println("ServerEndpoint", defaultServerEndpoint)
	log.Println("PollInterval", defaultPollInterval)
	log.Println("ReportInterval", defaultReportInterval)

	stats := newStats()

	pollDuration := time.Second * time.Duration(defaultPollInterval)
	poller := newPoller(stats, pollDuration)

	reportDuration := time.Second * time.Duration(defaultReportInterval)
	reporter := NewReporter(stats, defaultServerEndpoint, reportDuration)

	go poller.runRoutine()
	go reporter.runRoutine()

	select {}
}
