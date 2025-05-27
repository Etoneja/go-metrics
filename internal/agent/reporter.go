package agent

import (
	"log"
	"net/http"
	"sync"
	"time"
)

func performRequest(client HTTPClient, url string, wg *sync.WaitGroup) {
	defer wg.Done()

	method := "POST"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatalf("http.NewRequest failed: method=%s, url=%s, err=%v", method, url, err)
	}

	req.Header.Set("Content-Type", "text/plain")

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
	client        HTTPClient
	endpoint      string
	sleepDuration time.Duration
}

func (r *Reporter) send(metrics []metric) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, semaphoreSize)

	for _, m := range metrics {
		semaphore <- struct{}{}
		wg.Add(1)

		url := buildURL(r.endpoint, "update", m.kind, m.name, m.value)
		go func(url string) {
			defer func() {
				<-semaphore
			}()

			performRequest(r.client, url, &wg)
		}(url)
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
