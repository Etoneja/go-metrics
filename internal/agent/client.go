package agent

import (
	"net/http"
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
}

func NewConcurrentLimitedClient(client HTTPDoer, rateLimit uint) *ConcurrentLimitedClient {
	return &ConcurrentLimitedClient{
		client:    client,
		semaphore: make(chan struct{}, rateLimit),
	}
}

func (c *ConcurrentLimitedClient) Do(req *http.Request) (*http.Response, error) {
	c.semaphore <- struct{}{}
	defer func() {
		<-c.semaphore
	}()
	return c.client.Do(req)
}

func (c *ConcurrentLimitedClient) Close() {
	c.client.Close()
	close(c.semaphore)
}
