package main

const (
	metricTypeGauge   string = "gauge"
	metricTypeCounter string = "counter"
)

const defaultPollInterval int = 2
const defaultReportInterval int = 10
const defaultServerEndpoint string = "http://localhost:8080"

const semaphoreSize int = 10

const maxRandNum int = 1_000_000
