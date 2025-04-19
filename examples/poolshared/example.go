package main

import (
	"context"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/pool"
	"github.com/Izzette/go-safeconcurrency/api/results"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

func main() {
	sharedpool := pool.NewPool[any](nil, 1)
	sharedpool.Start()
	defer sharedpool.Close()

	bareInt, intCh := pool.TaskWithReturn[any, int](&intTask{})
	sharedpool.Submit(context.Background(), bareInt)
	if value, err := (<-intCh).Get(); err != nil {
		fmt.Printf("Error from intTask: %v\n", err)
	} else {
		fmt.Printf("Result from intTask: %d\n", value)
	}

	bareStr, strCh := pool.TaskWithHandle[any, string](&stringTask{})
	sharedpool.Submit(context.Background(), bareStr)
	for strResult := range strCh {
		value, err := strResult.Get()
		if err != nil {
			fmt.Printf("Error from stringTask: %v\n", err)
			continue
		}

		fmt.Printf("Result from stringTask: %s\n", value)
	}

	// This task doesn't return a result, so no need to wrap it with pool.TaskWithReturn or pool.TaskWithHandle.
	sharedpool.Submit(context.Background(), &logTask{"A message sent from the pool by logTask"})

	// This task won't run because the context is already canceled.
	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	bareStrWontRun, strChWontRun := pool.TaskWithHandle[any, string](&stringTask{})
	if err := sharedpool.Submit(ctx, bareStrWontRun); err != nil {
		fmt.Printf("Error submitting stringTaskWontRun: %v\n", err)
	} else {
		// It's important to consume all results from the channel, if you're not sure if you will consume all results
		// call results.DrainResultChannel() to avoid a deadlock.
		// Only drain or attempt to consume from the results channel if the task was submitted successfully.
		defer results.DrainResultChannel(strChWontRun)
	}
}

type intTask struct{}

func (t *intTask) Execute(ctx context.Context, _ interface{}) (int, error) {
	return 42, nil
}

type stringTask struct{}

func (t *stringTask) Execute(ctx context.Context, _ interface{}, h types.Handle[string]) {
	h.Publish(ctx, "hello")
	h.Publish(ctx, "world")
}

type logTask struct {
	message string
}

func (t *logTask) Execute(ctx context.Context, _ interface{}) {
	fmt.Println(t.message)
}
