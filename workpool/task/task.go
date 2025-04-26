package task

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/results"
)

// WrapTask wraps a [types.Task] so that it can be executed in a [types.WorkerPool] and returns a channel for
// results.
// This channel has a buffer size of 1 and will not block the worker when publishing the result.
// The results channel is always written to, even if the task returns an error.
// The provided context is passed to the task when it is executed in the [types.WorkerPool].
//
// It is recommended not to use this wrapper directly, but rather use the [workpool.Submit] helper function.
// The [workpool.Submit] helper will wrap the [types.Task], submit it to the pool, and wait for the result.
// The [workpool.Submit] helper implements error handling correctly and is less error prone than using this wrapper and
// sending to the [types.WorkerPool.Requests] channel directly.
func WrapTask[ResourceT any, ValueT any](
	ctx context.Context,
	task types.Task[ResourceT, ValueT],
) (types.ValuelessTask[ResourceT], types.TaskResult[ValueT]) {
	res := make(chan ValueT, 1)
	// taskResult.err and wrappedTask.err must point to the same error variable.
	var err error
	taskResult := &taskResult[ValueT]{
		results: res,
		err:     &err,
	}
	wrappedTask := &taskWrapper[ResourceT, ValueT]{
		ctx:  ctx,
		task: task,
		r:    res,
		err:  &err,
	}

	return wrappedTask, taskResult
}

// WrapStreamingTask wraps a [types.StreamingTask] so that it can be executed in a [types.WorkerPool] and returns
// a channel for results.
// The buffer size of the results channel is specified by the buffer parameter.
// It is recommended avoid using a buffer size of 0, as this will block the worker until the result is received.
//
// It is recommended not to use this wrapper directly, but rather use the
// [github.com/Izzette/go-safeconcurrency/workpool.SubmitStreamingBuffered] helper function.
// This helper will wrap the [types.StreamingTask], send it to the pool requests channel, and call the callback for each
// result as it is produced.
// This helper implements error handling correctly and is less error prone than using this wrapper and sending to the
// [types.WorkerPool.Requests] channel directly.
func WrapStreamingTask[ResourceT any, ValueT any](
	ctx context.Context,
	task types.StreamingTask[ResourceT, ValueT],
	buffer uint,
) (types.ValuelessTask[ResourceT], types.TaskResult[ValueT]) {
	res := make(chan ValueT, buffer)
	emitter := results.NewEmitter(res)
	var err error
	taskResult := &taskResult[ValueT]{
		results: res,
		err:     &err,
	}
	wrappedTask := streamingTaskWrapper[ResourceT, ValueT]{
		ctx:     ctx,
		task:    task,
		emitter: emitter,
		err:     &err,
	}

	return wrappedTask, taskResult
}

// WrapTaskFunc wraps a [types.TaskFunc] so that it can be executed in a [types.WorkerPool] and returns a
// [types.TaskResult] for execution monitoring and error propagation.
// It is not recommended to use this wrapper directly, but rather use the
// [github.com/Izzette/go-safeconcurrency/workpool.SubmitFunc] helper function.
// This [workpool.SubmitFunc] helper will wrap the [types.TaskFunc], submit it to the pool, and wait for the result.
// The [workpool.SubmitFunc] helper implements error handling correctly and is less error prone than using this wrapper
// and sending to the [types.WorkerPool.Requests] channel directly.
func WrapTaskFunc[ResourceT any](ctx context.Context, f types.TaskFunc[ResourceT]) (
	types.ValuelessTask[ResourceT], types.TaskResult[struct{}],
) {
	return WrapTask[ResourceT, struct{}](ctx, taskFuncWrapper[ResourceT]{f})
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

// streamingTaskWrapper is a wrapper for a [types.StreamingTask] implementing [types.ValuelessTask].
type streamingTaskWrapper[ResourceT any, ValueT any] struct {
	//nolint:containedctx
	ctx     context.Context
	task    types.StreamingTask[ResourceT, ValueT]
	emitter types.Emitter[ValueT]
	err     *error
}

// Execute implements [types.StreamingTask.Execute].
func (t streamingTaskWrapper[ResourceT, ValueT]) Execute(resource ResourceT) {
	defer t.emitter.Close()
	// We must not overwrite the error pointer, but instead store the error at the address of the pointer.
	*t.err = t.task.Execute(t.ctx, resource, t.emitter)
}

// taskWrapper is a wrapper for a [types.Task] implementing [types.ValuelessTask].
type taskWrapper[ResourceT any, ValueT any] struct {
	//nolint:containedctx
	ctx  context.Context
	task types.Task[ResourceT, ValueT]
	r    chan<- ValueT
	err  *error
}

// Execute implements [types.ValuelessTask.Execute].
func (t taskWrapper[ResourceT, ValueT]) Execute(
	resource ResourceT,
) {
	defer close(t.r)
	// streamingTaskWrapper will close the emitter for us when the task is done.
	value, err := t.task.Execute(t.ctx, resource)
	// We must not overwrite the error pointer, but instead store the error at the address of the pointer.
	*t.err = err

	// As the results channel is always buffered, we can publish the result without blocking the worker.
	t.r <- value
}

// taskFunc is a function that implements [types.Task].
type taskFuncWrapper[ResourceT any] struct {
	f types.TaskFunc[ResourceT]
}

// Execute implements [types.Task.Execute].
func (t taskFuncWrapper[ResourceT]) Execute(
	//nolint:containedctx
	ctx context.Context,
	resource ResourceT,
) (struct{}, error) {
	return struct{}{}, t.f(ctx, resource)
}
