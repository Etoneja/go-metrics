package agent

import "net/http"

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
