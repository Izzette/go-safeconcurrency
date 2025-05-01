package examples_test

import (
	"context"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/examples"
	"github.com/Izzette/go-safeconcurrency/workpool"
)

// Demonstrates how [types.WorkerPool] can be used to create a worker pool that executes heterogeneous tasks
// concurrently.
// Here we execute both [types.Task] ([IntTask] and [LogTask]) and [types.StreamingTask] ([StringTask]).
// The pool is created with no shared resource, no buffering, and a concurrency of 1 using [workpool.NewBuffered].
// Tasks are submitted with [workpool.Submit] and [workpool.SubmitStreaming].
func Example_sharedPool() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new pool with no shared resource, no buffering, and a concurrency of 1.
	sharedpool := workpool.NewBuffered[any](nil, 1, 0)
	// Close the types.WorkerPool.
	// It's important to close the pool only after all tasks have been submitted.
	// This will also wait for the pool to finish processing all tasks.
	defer sharedpool.Close()

	// We must start the pool before it will be able to process tasks.
	sharedpool.Start()

	// Submit a task that returns an int and handle the result.
	value, err := workpool.Submit[any, int](context.Background(), sharedpool, &examples.IntTask{42})
	if err != nil {
		fmt.Printf("Error from IntTask: %v\n", err)
	} else {
		fmt.Printf("Result from IntTask: %d\n", value)
	}

	// This task doesn't return any result or error.
	workpool.Submit[any, any](ctx, sharedpool, &examples.LogTask{"A message sent from the pool by LogTask"})

	// Tasks can also stream multiple results.
	// StringTask implements types.StreamingTask, submit it to the pool using workpool.SubmitStreaming, passing a callback
	// that will be invoked for each result.
	err = workpool.SubmitStreaming[any, string](
		context.Background(), sharedpool, &examples.StringTask{[]string{"hello", "world"}},
		func(_ctx context.Context, value string) error {
			// This function is called for each result emitted by the task.
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
	_, err = workpool.SubmitStreamingCollectAll[any, string](ctx, sharedpool, &examples.StringTask{})
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
