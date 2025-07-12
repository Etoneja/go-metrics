package agent

import (
	"context"
	"net/http"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
	Close()
}

type Collecter interface {
	Collect(ctx context.Context, resultCh chan<- Result)
}
