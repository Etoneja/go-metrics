package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
)

func performRequest(ctx context.Context, client HTTPDoer, endpoint string, hashKey string, metrics []models.MetricModel) error {

	url := buildURL(endpoint, "updates/")

	rawData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("unexpected error - failed to marshal metrics: %w", err)
	}

	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(rawData)
	if err != nil {
		return fmt.Errorf("unexpected error - failed to write gzip: %w", err)
	}
	if err = gz.Close(); err != nil {
		return fmt.Errorf("failed to close gzip: %w", err)
	}

	method := "POST"
	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		log.Printf("http.NewRequest failed: method=%s, url=%s, err=%v", method, url, err)
		return fmt.Errorf("unexpected error - failed to create request: %w", err)
	}

	if hashKey != "" {
		hash := common.ComputeHash(hashKey, buf.Bytes())
		req.Header.Set(common.HashHeaderKey, hash)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	var reqErr error

	backoffSchedule := common.DefaultBackoffSchedule
	backoffTicker := common.GetBackoffTicker(ctx, backoffSchedule)
	attemptNum := 0
	for range backoffTicker {
		attemptNum++
		attemptString := fmt.Sprintf("[%d/%d]", attemptNum, len(backoffSchedule)+1)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("%s failed to request %s: %v", attemptString, url, err)
			reqErr = fmt.Errorf("failed to send request: %w", err)
			continue
		}

		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			log.Printf("discard body error: %v", err)
		}

		if err := resp.Body.Close(); err != nil {
			log.Printf("close body error: %v", err)
		}

		if resp.StatusCode/100 == 2 {
			log.Printf("%s Request to %s succeeded", attemptString, url)
			return nil
		} else if resp.StatusCode/100 == 5 {
			continue
		}

		log.Printf("%s bad status code for %s: %d", attemptString, url, resp.StatusCode)
		return errors.New("bad status")
	}

	return reqErr

}

type Reporter struct {
	stats          *Stats
	iteration      uint
	client         HTTPDoer
	endpoint       string
	reportInterval time.Duration
	rateLimit      uint
	hashKey        string
}

func (r *Reporter) report(ctx context.Context) {
	r.iteration++
	log.Println("Report - start iteration", r.iteration)
	metrics := r.stats.GetMetrics()

	if len(metrics) == 0 {
		log.Println("No metrics. Skip report")
		return
	}

	err := performRequest(ctx, r.client, r.endpoint, r.hashKey, metrics)
	if err != nil {
		log.Printf("Error occurred sending metrcs %v", err)
	}

	log.Println("Report - finish iteration", r.iteration)
}

func (r *Reporter) stop() {
	r.client.Close()
}

func (r *Reporter) runRoutine(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(r.reportInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.stop()
			return ctx.Err()
		case <-ticker.C:
			r.report(ctx)
		}
	}
}

func newReporter(stats *Stats, endpoint string, reportInterval time.Duration, rateLimit uint, hashKey string) *Reporter {
	var client HTTPDoer
	client = NewBaseClient()
	if rateLimit > 0 {
		client = NewConcurrentLimitedClient(client, rateLimit)
	}
	return &Reporter{
		stats:          stats,
		client:         client,
		endpoint:       endpoint,
		reportInterval: reportInterval,
		rateLimit:      rateLimit,
		hashKey:        hashKey,
	}
}
