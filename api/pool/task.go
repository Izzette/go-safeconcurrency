package pool

import (
	"context"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/results"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

// TaskWrapper wraps a [types.Task] so that it can be executed in a [types.Pool] and returns a channel for
// results.
// This channel has a buffer size of 1 and will not block the worker when publishing the result.
// The results channel is always written to, even if the task returns an error.
// If the [context.Context] passed to [types.Pool.Submit] is canceled before the task is finished, the task will
// not produce any results and the result channel will be closed.
//
// It is recommended not to use this wrapper directly, but rather use the [Submit] helper function.
// The [Submit] helper will wrap the [types.Task], submit it to the pool, and wait for the result.
// The [Submit] helper implements error handling correctly and is less error prone than using this wrapper and calling
// [types.Pool.Submit] directly.
func TaskWrapper[PoolResourceT any, ValueT any](
	task types.Task[PoolResourceT, ValueT],
) (types.ValuelessTask[PoolResourceT], types.TaskResult[ValueT]) {
	res := make(chan ValueT, 1)
	errPtr := &atomic.Pointer[error]{}
	taskResult := &taskResult[ValueT]{
		results: res,
		err:     errPtr,
	}
	wrappedTask := &taskWrapper[PoolResourceT, ValueT]{task, res, errPtr}

	return wrappedTask, taskResult
}

// WrapMultiResultTaskBuffered wraps a [types.MultiResultTask] so that it can be executed in a [types.Pool] and returns
// a channel for results.
// The buffer size of the results channel is specified by the buffer parameter.
// It is recommended avoid using a buffer size of 0, as this will block the worker until the result is received.
//
// It is recommended to defer a call to [types.TaskResult.Drain] on the results channel immediately after the task is
// submitted to avoid blocking the worker pool if all results are not consumed.
// Alternatively, canceling the [context.Context] passed to the task will also stop the [types.Pool] from blocking.
// If the [context.Context] passed to [types.Pool.Submit] is canceled before the task is finished, the task will
// not produce any results and the result channel will be closed.
//
// For example:
//
//	// Wrap the MultiResultTask in a ValuelessTask with WrapMultiResultTask.
//	valuelessTask, taskResult := pool.WrapMultiResultTaskBuffered(multiResultTask, 1)
//
//	// Submit the task to the pool, check for errors (context cancellation).
//	if err := pool.Submit(ctx, valuelessTask); err != nil {
//		return err
//	}
//	// Defer a call to .Drain() to avoid blocking the worker pool in the case that all results are not consumed.
//	defer taskResult.Drain()
//
//	// Consume the results channel.
//	for value := range taskResult.Results() {
//		Do(value)
//	}
//
//	// If the task returns an error, it will be available in .Err() after the results channel is closed.
//	if err := taskResult.Err(); err != nil {
//		return err
//	}
//
// There exists a [SubmitMultiResult] helper function that wraps the task and submits it to the pool, waiting for the
// results and returning them as a slice.
// However, this does not provide a mechanism to process the results as they are produced, thus making it only as
// effective as using [types.Task] and returning a single sclice as the result.
// [SubmitMultiResult] is however useful for testing, debugging, and demonstration purposes where the performance
// difference is unimportant.
func WrapMultiResultTaskBuffered[PoolResourceT any, ValueT any](
	task types.MultiResultTask[PoolResourceT, ValueT],
	buffer uint,
) (types.ValuelessTask[PoolResourceT], types.TaskResult[ValueT]) {
	res := make(chan ValueT, buffer)
	handle := results.NewHandle(res)
	errPtr := &atomic.Pointer[error]{}
	taskResult := &taskResult[ValueT]{
		results: res,
		err:     errPtr,
	}
	wrappedTask := multiResultTaskWrapper[PoolResourceT, ValueT]{
		task:   task,
		handle: handle,
		err:    errPtr,
	}

	return wrappedTask, taskResult
}

// WrapMultiResultTask wraps a [types.MultiResultTask] with no results buffering.
// As there is no buffering, this will block the worker until all the results are received.
// If you would like results buffering, use [WrapMultiResultTaskBuffered] instead.
// This is equivalent to calling [WrapMultiResultTaskBuffered] with a buffer size of 0.
func WrapMultiResultTask[PoolResourceT any, ValueT any](
	task types.MultiResultTask[PoolResourceT, ValueT],
) (types.ValuelessTask[PoolResourceT], types.TaskResult[ValueT]) {
	return WrapMultiResultTaskBuffered(task, 0)
}

// taskResult implements [types.TaskResult].
// It is used to wrap the results channel and error handling for tasks that return results.
type taskResult[ValueT any] struct {
	results <-chan ValueT
	err     *atomic.Pointer[error]
}

// Results implements [types.TaskResult.Results].
func (tr *taskResult[ValueT]) Results() <-chan ValueT {
	return tr.results
}

// Drain implements [types.TaskResult.Drain].
func (tr *taskResult[ValueT]) Drain() error {
	for range tr.results {
		// drain the results channel
	}

	return tr.Err()
}

// Err implements [types.TaskResult.Err].
func (tr *taskResult[ValueT]) Err() error {
	if err := tr.err.Load(); err != nil {
		return *err
	}

	return nil
}

// multiResultTaskWrapper is a wrapper for a [types.MultiResultTask] implementing [types.ValuelessTask].
type multiResultTaskWrapper[PoolResourceT any, ValueT any] struct {
	task   types.MultiResultTask[PoolResourceT, ValueT]
	handle types.Handle[ValueT]
	err    *atomic.Pointer[error]
}

// Execute implements [types.MultiResultTask.Execute].
func (t multiResultTaskWrapper[PoolResourceT, ValueT]) Execute(ctx context.Context, resource PoolResourceT) {
	defer t.handle.Close()
	if err := t.task.Execute(ctx, resource, t.handle); err != nil {
		t.err.Store(&err)
	}
}

// taskWrapper is a wrapper for a [types.Task] implementing [types.ValuelessTask].
type taskWrapper[PoolResourceT any, ValueT any] struct {
	task types.Task[PoolResourceT, ValueT]
	r    chan ValueT
	err  *atomic.Pointer[error]
}

// Execute implements [types.ValuelessTask.Execute].
func (t taskWrapper[PoolResourceT, ValueT]) Execute(
	ctx context.Context,
	resource PoolResourceT,
) {
	defer close(t.r)
	// taskWithHandleWrapper will close the handle for us when the task is done.
	value, err := t.task.Execute(ctx, resource)
	// As the results channel is always buffered, we can publish the result without blocking the worker.
	t.r <- value
	if err != nil {
		t.err.Store(&err)
	}
}
