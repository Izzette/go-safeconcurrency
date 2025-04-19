package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Izzette/go-safeconcurrency/api/pool"
	"github.com/Izzette/go-safeconcurrency/api/results"
)

func main() {
	httpPool := newHttpPool(5)
	// Close the pool. it's important to close the pool only after all tasks have been submitted.
	// This will also wait for the pool to finish processing all tasks.
	defer httpPool.Close()

	// As the pool is created with a concurrency of 5, it will run up to 5 tasks concurrently.
	// Here we are going to submit 8 tasks, so we will see that the pool will run 5 tasks concurrently and then
	// wait for the first 5 to finish before starting the next 3.
	urls := []string{
		"https://httpbin.org/status/500", // Will return a 500 error
		"https://httpbin.org/delay/2",    // Will take 2 seconds to respond
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

			// The context used when submitting the task does not need to be the same as the one used to start the pool.
			// It should represent the amount of time the caller should wait to submit the task and get a result.
			// If this context is canceled, but the task is already submitted it will continue.
			ctx := context.Background()
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

type httpTask struct {
	// The URL to fetch.
	url string
}

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
	pool.Pool[*http.Client]
}

// newHttpPool creates a new HTTP pool with the specified concurrency.
func newHttpPool(concurrency int) *httpPool {
	// The pool is created with a resource of type *http.Client, which is used to make HTTP requests.
	client := &http.Client{}
	p := pool.NewPool[*http.Client](client, concurrency)
	return &httpPool{Pool: *p}
}

// get makes an HTTP GET request to the provided URL using the HTTP pool and returns the response.
func (p *httpPool) get(ctx context.Context, url string) (*http.Response, error) {
	// Create a new HTTP task with the URL to fetch.
	task := &httpTask{url: url}
	// Wrap the task in a BareTask to be able to submit it to the pool.
	bareTask, resCh := pool.TaskWithReturn[*http.Client, *http.Response](task)
	// Submit the task to the pool.
	// If context is canceled before the task is published it will instead return an error.
	if err := p.Submit(ctx, bareTask); err != nil {
		return nil, err
	}

	// Ensure to drain the result channel
	// This is important to avoid permanent blocking on the result channel
	// Only drain or attempt to consume from the results channel if the task was submitted successfully.
	defer results.DrainResultChannel(resCh)

	// Wait for the result
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resCh:
		return result.Get()
	}
}
