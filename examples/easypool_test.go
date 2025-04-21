package examples

import (
	"context"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/pool"
)

// Simple example of how [pool.NewPool] can be used to create a [types.Pool] worker pool that executes tasks
// concurrently.
// Here we're submitting a function to exexcute in the pool with [pool.SubmitFunc].
func Example_easyPool() {
	poolResource := 42
	p := pool.NewPool[int](poolResource, 1)
	defer p.Close()
	p.Start()

	err := pool.SubmitFunc[int](context.Background(), p, func(_ctx context.Context, resource int) error {
		// This is the function that will be executed in the pool and use the pool resource.
		fmt.Printf("value: %v\n", resource)
		return nil
	})
	fmt.Printf("err: %v\n", err)

	// Output:
	// value: 42
	// err: <nil>
}
