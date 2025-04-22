package workpool

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/results"
)

// WrapTask wraps a [types.Task] so that it can be executed in a [types.Pool] and returns a channel for
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
func WrapTask[PoolResourceT any, ValueT any](
	task types.Task[PoolResourceT, ValueT],
) (types.ValuelessTask[PoolResourceT], types.TaskResult[ValueT]) {
	res := make(chan ValueT, 1)
	// taskResult.err and wrappedTask.err must point to the same error variable.
	var err error
	taskResult := &taskResult[ValueT]{
		results: res,
		err:     &err,
	}
	wrappedTask := &taskWrapper[PoolResourceT, ValueT]{task, res, &err}

	return wrappedTask, taskResult
}

// WrapMultiResultTaskBuffered wraps a [types.MultiResultTask] so that it can be executed in a [types.Pool] and returns
// a channel for results.
// The buffer size of the results channel is specified by the buffer parameter.
// It is recommended avoid using a buffer size of 0, as this will block the worker until the result is received.
//
// It is recommended not to use this wrapper directly, but rather use the [SubmitMultiResultBuffered] helper function.
// The [SubmitMultiResultBuffered] helper will wrap the [types.MultiResultTask], submit it to the pool, and call the
// callback for each result as it is produced.
// The [SubmitMultiResultBuffered] helper implements error handling correctly and is less error prone than using
// this wrapper and calling [types.Pool.Submit] directly.
func WrapMultiResultTaskBuffered[PoolResourceT any, ValueT any](
	task types.MultiResultTask[PoolResourceT, ValueT],
	buffer uint,
) (types.ValuelessTask[PoolResourceT], types.TaskResult[ValueT]) {
	res := make(chan ValueT, buffer)
	handle := results.NewHandle(res)
	var err error
	taskResult := &taskResult[ValueT]{
		results: res,
		err:     &err,
	}
	wrappedTask := multiResultTaskWrapper[PoolResourceT, ValueT]{
		task:   task,
		handle: handle,
		err:    &err,
	}

	return wrappedTask, taskResult
}

// WrapMultiResultTask wraps a [types.MultiResultTask] with a buffer size of 1.
// This is equivalent to calling [WrapMultiResultTaskBuffered] with a buffer size of 1.
func WrapMultiResultTask[PoolResourceT any, ValueT any](
	task types.MultiResultTask[PoolResourceT, ValueT],
) (types.ValuelessTask[PoolResourceT], types.TaskResult[ValueT]) {
	return WrapMultiResultTaskBuffered(task, 1)
}

// WrapTaskFunc wraps a [types.TaskFunc] so that it can be executed in a [types.Pool] and returns a [types.TaskResult]
// for execution monitoring and error propagation.
// It is not recommended to use this wrapper directly, but rather use the [SubmitFunc] helper function.
func WrapTaskFunc[PoolResourceT any](f types.TaskFunc[PoolResourceT]) (
	types.ValuelessTask[PoolResourceT], types.TaskResult[struct{}],
) {
	return WrapTask[PoolResourceT, struct{}](taskFuncWrapper[PoolResourceT]{f})
}

// taskResult implements [types.TaskResult].
// It is used to wrap the results channel and error handling for tasks that return results.
type taskResult[ValueT any] struct {
	results <-chan ValueT
	err     *error
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

	// err should only be used after the results channel is closed.
	return *tr.err
}

// multiResultTaskWrapper is a wrapper for a [types.MultiResultTask] implementing [types.ValuelessTask].
type multiResultTaskWrapper[PoolResourceT any, ValueT any] struct {
	task   types.MultiResultTask[PoolResourceT, ValueT]
	handle types.Handle[ValueT]
	err    *error
}

// Execute implements [types.MultiResultTask.Execute].
func (t multiResultTaskWrapper[PoolResourceT, ValueT]) Execute(ctx context.Context, resource PoolResourceT) {
	defer t.handle.Close()
	// We must not overwrite the error pointer, but instead store the error at the address of the pointer.
	*t.err = t.task.Execute(ctx, resource, t.handle)
}

// taskWrapper is a wrapper for a [types.Task] implementing [types.ValuelessTask].
type taskWrapper[PoolResourceT any, ValueT any] struct {
	task types.Task[PoolResourceT, ValueT]
	r    chan<- ValueT
	err  *error
}

// Execute implements [types.ValuelessTask.Execute].
func (t taskWrapper[PoolResourceT, ValueT]) Execute(
	ctx context.Context,
	resource PoolResourceT,
) {
	defer close(t.r)
	// taskWithHandleWrapper will close the handle for us when the task is done.
	value, err := t.task.Execute(ctx, resource)
	// We must not overwrite the error pointer, but instead store the error at the address of the pointer.
	*t.err = err

	// As the results channel is always buffered, we can publish the result without blocking the worker.
	t.r <- value
}

// taskFunc is a function that implements [types.Task].
type taskFuncWrapper[PoolResourceT any] struct {
	f types.TaskFunc[PoolResourceT]
}

// Execute implements [types.Task.Execute].
func (t taskFuncWrapper[PoolResourceT]) Execute(
	ctx context.Context,
	resource PoolResourceT,
) (struct{}, error) {
	return struct{}{}, t.f(ctx, resource)
}
