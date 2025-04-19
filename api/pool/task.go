package pool

import (
	"context"

	"github.com/Izzette/go-safeconcurrency/api/results"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

type taskWithHandleWrapper[PoolResourceT any, ValueT any] struct {
	task   types.Task[PoolResourceT, ValueT]
	handle types.Handle[ValueT]
}

func (t taskWithHandleWrapper[PoolResourceT, ValueT]) Execute(ctx context.Context, resource PoolResourceT) {
	defer t.handle.Close()
	t.task.Execute(ctx, resource, t.handle)
}

// TaskWithHandle wraps a types.Task so that it can be executed in a pool and returns a channel for results.
// This is equivalent to TaskWithHandleBuffered with a buffer size of 1.
func TaskWithHandle[PoolResourceT any, ValueT any](
	task types.Task[PoolResourceT, ValueT],
) (types.BareTask[PoolResourceT], <-chan types.Result[ValueT]) {
	return TaskWithHandleBuffered(task, 0)
}

// TaskWithHandleBuffered wraps a types.Task so that it can be executed in a pool and returns a channel for results.
// The buffer size of the results channel is specified by the buffer parameter.
// It is recommended to not use a buffer size of 0, as this will block the worker until the result is received.
func TaskWithHandleBuffered[PoolResourceT any, ValueT any](
	task types.Task[PoolResourceT, ValueT],
	buffer uint,
) (types.BareTask[PoolResourceT], <-chan types.Result[ValueT]) {
	r := make(chan types.Result[ValueT], buffer)
	handle := results.NewHandle(r)

	return taskWithHandleWrapper[PoolResourceT, ValueT]{task: task, handle: handle}, r
}

type taskWithReturnWrapper[PoolResourceT any, ValueT any] struct {
	task types.SingleTask[PoolResourceT, ValueT]
}

func (t taskWithReturnWrapper[PoolResourceT, ValueT]) Execute(
	ctx context.Context,
	resource PoolResourceT,
	handle types.Handle[ValueT],
) {
	// taskWithHandleWrapper will close the handle for us when the task is done.
	value, err := t.task.Execute(ctx, resource)

	// Publishing the result may return a context error if the task context is cancelled.
	if err != nil {
		_ = handle.Error(ctx, err)

		return
	}
	_ = handle.Publish(ctx, value)
}

// TaskWithReturn wraps a types.SingleTask so that it can be executed in a pool and returns a channel for results.
// This channel will always be unbuffered and will block until the result is received.
func TaskWithReturn[PoolResourceT any, ValueT any](
	task types.SingleTask[PoolResourceT, ValueT],
) (types.BareTask[PoolResourceT], <-chan types.Result[ValueT]) {
	wrappedTask := &taskWithReturnWrapper[PoolResourceT, ValueT]{task}

	return TaskWithHandle[PoolResourceT, ValueT](wrappedTask)
}
