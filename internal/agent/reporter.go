package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/etoneja/go-metrics/internal/models"
)

func performRequest(client HTTPDoer, endpoint string, metricModel *models.MetricModel, wg *sync.WaitGroup) {
	defer wg.Done()

	url := buildURL(endpoint, "update/")

	rawData, err := json.Marshal(metricModel)
	if err != nil {
		log.Fatalf("Unexpected error - failed to marshal metric, err=%v", err)
	}

	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(rawData)
	if err != nil {
		log.Fatal(err)
	}
	gz.Close()

	method := "POST"
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		log.Fatalf("http.NewRequest failed: method=%s, url=%s, err=%v", method, url, err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("failed to request %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("non-OK status code for %s: %d", url, resp.StatusCode)
		return
	}

	log.Printf("Request to %s succeeded", url)
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

			performRequest(r.client, r.endpoint, m, &wg)
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
