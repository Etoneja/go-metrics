package agent

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type BaseClient struct {
	client *http.Client
}

func (b *BaseClient) Close() {}

func (b *BaseClient) Do(req *http.Request) (*http.Response, error) {
	return b.client.Do(req)
}

func NewBaseClient() *BaseClient {
	return &BaseClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type ConcurrentLimitedClient struct {
	client    HTTPDoer
	semaphore chan struct{}
	closed    atomic.Bool
}

func NewConcurrentLimitedClient(client HTTPDoer, rateLimit uint) *ConcurrentLimitedClient {
	return &ConcurrentLimitedClient{
		client:    client,
		semaphore: make(chan struct{}, rateLimit),
	}
}

func (c *ConcurrentLimitedClient) Do(req *http.Request) (*http.Response, error) {
	if c.closed.Load() {
		return nil, fmt.Errorf("client is closed")
	}
	c.semaphore <- struct{}{}
	defer func() {
		<-c.semaphore
	}()
	return c.client.Do(req)
}

func (c *ConcurrentLimitedClient) Close() {
	if c.closed.Swap(true) {
		return
	}

	c.client.Close()
	close(c.semaphore)
}
