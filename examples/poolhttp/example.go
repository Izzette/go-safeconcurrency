package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Izzette/go-safeconcurrency/api/pool"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

func main() {
	httpPool := newHttpPool(5)
	// Close the types.Pool.
	// It's important to close the pool only after all tasks have been submitted.
	// This will also wait for the pool to finish processing all tasks.
	defer httpPool.Close()

	// As the pool is created with a concurrency of 5, it will run up to 5 tasks concurrently.
	// Here we are going to submit 8 tasks, so we will see that the pool will run 5 tasks concurrently and then
	// wait for the first 5 to finish before starting the next 3.
	urls := []string{
		"https://httpbin.org/delay/2",    // Will take 2 seconds to respond
		"https://httpbin.org/status/500", // Will return a 500 error
		"httpxs:///mal_formatted_url",    // Should produce an error as the URL is malformed
		"http://300.300.300.300",         // Should produce an error as the IP is invalid

		// The following URLs are valid and should return a 200 OK response
		"https://www.google.com",
		"https://github.com",
		"https://stackoverflow.com",
		"https://go.dev",
	}

	// We must start the pool before it will be able to process tasks.
	httpPool.Start()

	// We'll request each URL in a separate goroutine and use a WaitGroup to wait for all of them to finish.
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		// Tasks may be submitted and results consumed from multiple goroutines, just make sure not to submit tasks to a
		// closed pool.
		go func(url string) {
			defer wg.Done()

			// The context used when submitting the task will be passed to the task when it is executed.
			// This can be used to cancel the task or to set a timeout.
			ctx := context.Background()

			// httpPool.get() will submit the task to the pool and wait for the result.
			resp, err := httpPool.get(ctx, url)
			if err != nil { // Check if the task produced an error
				fmt.Printf("Error [%v]: %v\n", url, err)
				return
			}

			// We'll close the response body and print the status line.
			defer resp.Body.Close()
			fmt.Printf("[%v] %v %v\n", url, resp.Proto, resp.Status)
		}(url)
	}

	// Wait for all tasks to finish.
	wg.Wait()
}

// httpTask implements [types.Task].
// It fetches a URL and returns the HTTP response.
type httpTask struct {
	// The URL to fetch.
	url string
}

// Execute implements [types.Task.Execute].
func (t *httpTask) Execute(ctx context.Context, client *http.Client) (*http.Response, error) {
	// Create a new HTTP GET request with the provided context and URL.
	req, err := http.NewRequestWithContext(ctx, "GET", t.url, nil)
	if err != nil {
		return nil, err
	}

	// Send the request using the HTTP client.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// httpPool is a worker pool that executes HTTP requests concurrently.
type httpPool struct {
	types.Pool[*http.Client]
}

// newHttpPool creates a new HTTP pool with the specified concurrency.
func newHttpPool(concurrency int) *httpPool {
	// The pool is created with a resource of type *http.Client, which is used to make HTTP requests.
	client := &http.Client{}
	p := pool.NewPool[*http.Client](client, concurrency)
	return &httpPool{Pool: p}
}

// get makes an HTTP GET request to the provided URL using the HTTP pool and returns the response.
func (p *httpPool) get(ctx context.Context, url string) (*http.Response, error) {
	// Create a new HTTP task with the URL to fetch.
	task := &httpTask{url: url}

	return pool.Submit[*http.Client, *http.Response](ctx, p.Pool, task)
}
