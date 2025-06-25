package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/etoneja/go-metrics/internal/models"
)

func performRequest(client HTTPDoer, endpoint string, metricModel *models.MetricModel, wg *sync.WaitGroup) error {
	defer wg.Done()

	url := buildURL(endpoint, "update/")

	rawData, err := json.Marshal(metricModel)
	if err != nil {
		return fmt.Errorf("unexpected error - failed to marshal metric: %w", err)
	}

	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(rawData)
	if err != nil {
		return fmt.Errorf("unexpected error - failed to write gzip: %w", err)
	}
	gz.Close()

	method := "POST"
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		log.Printf("http.NewRequest failed: method=%s, url=%s, err=%v", method, url, err)
		return fmt.Errorf("unexpected error - failed to create request: %w", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("failed to request %s: %v", url, err)
		return fmt.Errorf("unexpected error - failed to send request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("non-OK status code for %s: %d", url, resp.StatusCode)
		return errors.New("bad status")
	}

	log.Printf("Request to %s succeeded", url)
	return nil
}

type Reporter struct {
	stats         *Stats
	iteration     uint
	client        HTTPDoer
	endpoint      string
	sleepDuration time.Duration
}

func (r *Reporter) send(metrics []*models.MetricModel) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, semaphoreSize)

	for _, m := range metrics {
		semaphore <- struct{}{}
		wg.Add(1)

		go func(m *models.MetricModel) {
			defer func() {
				<-semaphore
			}()

			err := performRequest(r.client, r.endpoint, m, &wg)
			if err != nil {
				log.Printf("Error occurred sending metric %s: %v", m.ID, err)
			}
		}(m)
	}

	wg.Wait()
}

func (r *Reporter) report() {
	r.iteration++
	log.Println("Report - start iteration", r.iteration)
	if r.stats.getCounter() == 0 {
		log.Println("Stats not collected yet. Skip send")
		return
	}
	metrics := r.stats.dump()

	r.send(metrics)

	log.Println("Report - finish iteration", r.iteration)
}

func (r *Reporter) runRoutine() {
	for {
		r.report()
		time.Sleep(r.sleepDuration)
	}
}

func newReporter(stats *Stats, endpoint string, sleepDuration time.Duration) *Reporter {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &Reporter{
		stats:         stats,
		client:        client,
		endpoint:      endpoint,
		sleepDuration: sleepDuration,
	}
}
