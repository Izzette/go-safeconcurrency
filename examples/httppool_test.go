package examples

import (
	"context"
	"fmt"
	"sync"
)

// Demonstrates how [types.WorkerPool] can be used to create a worker pool that executes HTTP requests concurrently
// using a shared [*http.Client].
// See [HttpPool] for more details on how to implement a custom pool wrapper for easily running like-tasks.
func Example_hTTPPool() {
	// Create a new HTTP pool with a concurrency of 5.
	httpPool := NewHttpPool(5)
	// Close the types.WorkerPool.
	// It's important to close the pool only after all tasks have been submitted.
	// This will also wait for the pool to finish processing all tasks.
	defer httpPool.Close()

	// We must start the pool before it will be able to process tasks.
	httpPool.Start()

	urls := []string{
		"https://httpbin.org/delay/1",    // Will take 1 second to respond
		"https://httpbin.org/status/500", // Will return a 500 error
		"httpxs:///mal_formatted_url",    // Should produce an error as the URL is malformed
		"http://300.300.300.300",         // Should produce an error as the IP is invalid

		// The following URLs are valid and should return a 200 OK response
		"https://www.google.com",
		"https://github.com",
		"https://stackoverflow.com",
		"https://go.dev",
	}

	// We'll request each URL in a separate goroutine and use a WaitGroup to wait for all of them to finish.
	// In the real world, instead of launching a goroutine for each URL, you'd be submitting tasks on the pool from a
	// handler for an HTTP request, for example.
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		// Tasks may be submitted and results consumed from multiple goroutines.
		go func(url string) {
			defer wg.Done()
			// httpPool.get() will submit the task to the pool and wait for the result.
			resp, err := httpPool.Get(context.Background(), url)
			if err != nil { // Check if the task produced an error
				fmt.Printf("[%v] error: %v\n", url, err)
				return
			}

			fmt.Printf("[%v] %v %v\n", url, resp.Proto, resp.Status)
		}(url)
	}

	// Wait for all tasks to finish.
	wg.Wait()
	// Unordered output:
	// [httpxs:///mal_formatted_url] error: Get "httpxs:///mal_formatted_url": unsupported protocol scheme "httpxs"
	// [http://300.300.300.300] error: Get "http://300.300.300.300": dial tcp: lookup 300.300.300.300: no such host
	// [https://github.com] HTTP/2.0 200 OK
	// [https://www.google.com] HTTP/2.0 200 OK
	// [https://go.dev] HTTP/2.0 200 OK
	// [https://stackoverflow.com] HTTP/2.0 200 OK
	// [https://httpbin.org/status/500] HTTP/2.0 500 Internal Server Error
	// [https://httpbin.org/delay/1] HTTP/2.0 200 OK
}
