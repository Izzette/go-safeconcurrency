package main

import (
	"context"
	"fmt"

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
	// In order to use them, make a MultiResultTask like stringTask, which implements [types.MultiResultTask].
	// Then, submit it to the pool using [types.Pool.SubmitMultiResult], passing a callback that will be invoked in this
	// goroutine for each result published by the task.
	if err := pool.SubmitMultiResult[any, string](
		context.Background(), sharedpool, &stringTask{},
		func(_ctx context.Context, value string) error {
			// This function is called for each result published by the task.
			// It is _not_ called in the goroutine of the task, pool resources should not be used here.
			fmt.Printf("Result from stringTask: %s\n", value)
			return nil
		},
	); err != nil {
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
	if _, err := pool.SubmitMultiResultCollectAll[any, string](ctx, sharedpool, &stringTask{}); err != nil {
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
