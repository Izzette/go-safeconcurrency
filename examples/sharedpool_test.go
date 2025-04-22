package examples

import (
	"context"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/workpool"
)

// Demonstrates how [types.Pool] can be used to create a worker pool that executes heterogeneous tasks concurrently.
// Here we execute both [types.Task] ([IntTask] and [LogTask]) and [types.MultiResultTask] ([StringTask]).
// The pool is created with no shared resource, no buffering, and a concurrency of 1 using [workpool.NewPoolBuffered].
// Tasks are submitted with [workpool.Submit] and [workpool.SubmitMultiResult].
func Example_sharedPool() {
	// Create a new pool with no shared resource, no buffering, and a concurrency of 1.
	sharedpool := workpool.NewPoolBuffered[any](nil, 1, 0)
	// Close the types.Pool.
	// It's important to close the pool only after all tasks have been submitted.
	// This will also wait for the pool to finish processing all tasks.
	defer sharedpool.Close()

	// We must start the pool before it will be able to process tasks.
	sharedpool.Start()

	// Submit a task that returns an int and handle the result.
	value, err := workpool.Submit[any, int](context.Background(), sharedpool, &IntTask{42})
	if err != nil {
		fmt.Printf("Error from IntTask: %v\n", err)
	} else {
		fmt.Printf("Result from IntTask: %d\n", value)
	}

	// This task doesn't return any result or error.
	workpool.Submit[any, any](context.Background(), sharedpool, &LogTask{"A message sent from the pool by LogTask"})

	// Tasks can also stream multiple results.
	// StringTask implements types.MultiResultTask, submit it to the pool using workpool.SubmitMultiResult, passing a callback
	// that will be invoked for each result.
	err = workpool.SubmitMultiResult[any, string](
		context.Background(), sharedpool, &StringTask{[]string{"hello", "world"}},
		func(_ctx context.Context, value string) error {
			// This function is called for each result published by the task.
			// It is _not_ called in the goroutine of the task, pool resources should not be used here.
			fmt.Printf("Result from StringTask: %s\n", value)
			return nil
		},
	)
	if err != nil {
		fmt.Printf("Error from StringTask: %v\n", err)
	}

	// This task won't run because the context is already canceled.
	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	_, err = workpool.SubmitMultiResultCollectAll[any, string](ctx, sharedpool, &StringTask{})
	if err != nil {
		fmt.Printf("Error submitting StringTask: %v\n", err)
	}
	// Output:
	// Result from IntTask: 42
	// A message sent from the pool by LogTask
	// Result from StringTask: hello
	// Result from StringTask: world
	// Error submitting StringTask: context canceled
}
