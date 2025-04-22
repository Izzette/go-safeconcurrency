package examples

import (
	"context"
	"net/http"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/workpool"
)

// HttpTask implements [types.Task].
// It fetches a URL and returns the HTTP response.
type HttpTask struct {
	// The HTTP Method to use.
	Method string

	// The URL to fetch.
	URL string
}

// Execute implements [types.Task.Execute].
func (t *HttpTask) Execute(ctx context.Context, client *http.Client) (*http.Response, error) {
	// Create a new HTTP GET request with the provided context and URL.
	req, err := http.NewRequestWithContext(ctx, t.Method, t.URL, nil)
	if err != nil {
		return nil, err
	}

	// Send the request using the HTTP client.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return resp, nil
}

// HttpPool is a worker pool that executes HTTP requests concurrently.
type HttpPool struct {
	types.Pool[*http.Client]
}

// NewHttpPool creates a new HTTP pool with the specified concurrency.
func NewHttpPool(concurrency int) *HttpPool {
	// The pool is created with a resource of type *http.Client, which is used to make HTTP requests.
	client := &http.Client{}
	p := workpool.NewPool[*http.Client](client, concurrency)
	return &HttpPool{Pool: p}
}

// Get makes an HTTP GET request to the provided URL using the HTTP pool and returns the response.
func (p *HttpPool) Get(ctx context.Context, url string) (*http.Response, error) {
	// Create a new HTTP task with the URL to fetch.
	task := &HttpTask{Method: "GET", URL: url}

	return workpool.Submit[*http.Client, *http.Response](ctx, p.Pool, task)
}
