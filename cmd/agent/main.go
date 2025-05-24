package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Service started")
	stats := newStats()

	pollDuration := time.Second * time.Duration(defaultPollInternal)
	poller := newPoller(stats, pollDuration)

	reportDuration := time.Second * time.Duration(defaultReportInternal)
	reporter := NewReporter(stats, defaultServerEndpoint, reportDuration)

	go poller.runRoutine()
	go reporter.runRoutine()

	select {}
}
