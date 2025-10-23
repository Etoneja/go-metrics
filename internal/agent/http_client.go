package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
)

type httpMetricClient struct {
	cfg    Configer
	client *http.Client

	semaphore chan struct{}
	mu        sync.Mutex
	closed    bool
}

func (c *httpMetricClient) SendBatch(ctx context.Context, metrics []models.MetricModel) error {
	if len(metrics) == 0 {
		return nil
	}
	return c.doRequest(ctx, "POST", "updates/", metrics)
}

func (c *httpMetricClient) doRequest(ctx context.Context, method, path string, metrics []models.MetricModel) error {
	req, buf, err := c.prepareRequest(ctx, method, path, metrics)
	if err != nil {
		return err
	}

	if c.semaphore != nil {
		c.semaphore <- struct{}{}
		defer func() { <-c.semaphore }()
	}

	return c.executeWithRetry(req, buf)
}

func (c *httpMetricClient) prepareRequest(ctx context.Context, method, path string, metrics []models.MetricModel) (*http.Request, *bytes.Buffer, error) {
	endpoint := ensureEndpointProtocol(c.cfg.GetServerEndpoint(), c.cfg.GetServerProtocol())
	url := buildURL(endpoint, path)

	rawData, err := json.Marshal(metrics)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal metrics: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(rawData); err != nil {
		return nil, nil, fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, nil, fmt.Errorf("gzip close: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	c.applyHeadersAndEncryption(req, &buf)

	return req, &buf, nil
}

func (c *httpMetricClient) applyHeadersAndEncryption(req *http.Request, buf *bytes.Buffer) {
	if c.cfg.getLocalIP() != nil {
		req.Header.Set("X-Real-IP", c.cfg.getLocalIP().String())
	}

	if c.cfg.GetPublicKey() != nil {
		encryptedData, err := common.EncryptHybrid(c.cfg.GetPublicKey(), buf.Bytes())
		if err != nil {
			log.Printf("Encryption failed: %v", err)
			return
		}

		buf.Reset()
		buf.Write(encryptedData)
		req.Body = io.NopCloser(buf)
		req.ContentLength = int64(buf.Len())
		req.Header.Set("X-Encrypted", "true")
	}

	if c.cfg.GetHashKey() != "" {
		hash := common.ComputeHash(c.cfg.GetHashKey(), buf.Bytes())
		req.Header.Set(common.HashHeaderKey, hash)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
}

func (c *httpMetricClient) executeWithRetry(req *http.Request, buf *bytes.Buffer) error {
	backoffSchedule := common.DefaultBackoffSchedule

	for attempt, backoff := range backoffSchedule {
		if buf != nil && req.Body != nil {
			req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
		}

		resp, err := c.client.Do(req)
		if err != nil {
			log.Printf("Attempt %d failed: %v", attempt+1, err)
			time.Sleep(backoff)
			continue
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode/100 == 2 {
			log.Printf("Request succeeded after %d attempt(s)", attempt+1)
			return nil
		} else if resp.StatusCode/100 == 5 {
			log.Printf("Attempt %d: server error %d, retrying", attempt+1, resp.StatusCode)
			time.Sleep(backoff)
			continue
		}

		return fmt.Errorf("http %d", resp.StatusCode)
	}

	return fmt.Errorf("max retries exceeded")
}

func (c *httpMetricClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	if c.semaphore != nil {
		close(c.semaphore)
	}
	return nil
}

func NewHTTPMetricClient(cfg Configer) *httpMetricClient {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &httpMetricClient{
		cfg:       cfg,
		client:    client,
		semaphore: makeSemaphore(cfg.GetRateLimit()),
	}
}

func makeSemaphore(rateLimit uint) chan struct{} {
	if rateLimit > 0 {
		return make(chan struct{}, rateLimit)
	}
	return nil
}
