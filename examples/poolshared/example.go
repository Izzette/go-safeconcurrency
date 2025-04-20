package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Izzette/go-safeconcurrency/api/pool"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

func main() {
	// Create a new pool with no shared resource, no buffering, and a concurrency of 1.
	sharedpool := pool.NewPool[any](nil, 1)
	// Start the pool.
	sharedpool.Start()
	// Ensure the pool is closed when done.
	defer sharedpool.Close()

	// Submit a task that returns an int and handle the result.
	if value, err := pool.Submit[any, int](context.Background(), sharedpool, &intTask{}); err != nil {
		fmt.Printf("Error from intTask: %v\n", err)
	} else {
		fmt.Printf("Result from intTask: %d\n", value)
	}

	// This task doesn't return any result or error.
	pool.Submit[any, any](context.Background(), sharedpool, &logTask{"A message sent from the pool by logTask"})

	// Tasks that return multiple results are more complicated, but can provide significant performanc benefits in certain
	// cases.
	// In order to use them, make a MultiResultTask like stringTask, wrap the task in a ValuelessTask, than submit it to
	// the pool.
	valuelessTask, taskResults := pool.WrapMultiResultTask[any, string](&stringTask{})

	// Submit the task to the pool.
	// If context is canceled before the task is published it will instead return an error.
	ctx := context.Background()
	if err := sharedpool.Submit(ctx, valuelessTask); err != nil {
		fmt.Printf("Error submitting stringTask: %v\n", err)
		os.Exit(1)
	}
	// Only drain the results channel if the task was submitted successfully.
	// We should always have already drained the results channel below, but this is a precaution to avoid deadlocks in case
	// we want to return early due to an error encountered while processing the results.
	// Canceling the context can also be used to stop the pool from blocking.
	defer taskResults.Drain()

	// Wait for the result or context cancellation, whichever comes first.
	for value := range taskResults.Results() {
		fmt.Printf("Result from stringTask: %s\n", value)
	}
	if err := taskResults.Err(); err != nil {
		fmt.Printf("Error from stringTask: %v\n", err)
	}

	// This task won't run because the context is already canceled.
	// SubmitMultiResult is a helper function to submit a [types.MultiResultTask] to a [types.Pool] and wait for the
	// results.
	// It's not recommended to use this function in production code, but it's useful for testing, debugging, and
	// demonstration purposes where the performance difference is unimportant and the ability to consume the results as
	// they are produced is not required.
	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	if _, err := pool.SubmitMultiResult[any, string](ctx, sharedpool, &stringTask{}); err != nil {
		fmt.Printf("Error submitting stringTaskWontRun: %v\n", err)
	}
}

// intTask implements [types.Task].
// It returns 42 when executed.
type intTask struct{}

// Execute implements [types.Task.Execute].
func (t *intTask) Execute(ctx context.Context, _ any) (int, error) {
	return 42, nil
}

// stringTask implements [types.MultiResultTask].
// It publishes two strings to the provided handle.
type stringTask struct{}

// Execute implements [types.MultiResultTask.Execute].
func (t *stringTask) Execute(ctx context.Context, _ any, h types.Handle[string]) error {
	h.Publish(ctx, "hello")
	h.Publish(ctx, "world")
	return nil
}

// logTask implements [types.Task].
// It logs a message to the console.
type logTask struct {
	message string
}

// Execute implements [types.Task.Execute].
func (t *logTask) Execute(ctx context.Context, _ any) (any, error) {
	fmt.Println(t.message)
	return nil, nil
}
