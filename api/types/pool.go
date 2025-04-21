package types

import "context"

// Pool is a worker pool that executes arbitrary tasks concurrently.
// PoolResourceT is the type of resource used by the pool (ex. API client), it may be set to type any if a shared pool
// resource is not required.
type Pool[PoolResourceT any] interface {
	// Start initializes the pool and prepares it for task execution.
	Start()

	// Close stops the pool, and waits for all tasks to complete.
	// If the pool was never started, it will be closed immediately.
	// It is safe to call .Close() multiple times.
	Close()

	// Submit submits a task to the Pool for execution.
	// The provided context.Context is used passed to of the task when it is executed in the Pool.
	// It is advisable to use context.WithDeadline or context.WithTimeout to limit the time the task is allowed to run
	// and ensure the Pool can be shared with other tasks.
	// Alternatively using context.WithCancel and deferring a call to the context.CancelFunc will stop the task from
	// blocking the Pool if the caller is no longer interested in the result.
	// ⚠️You must never attempt to submit tasks to a pool which has been closed, this will result in a panic!
	Submit(context.Context, ValuelessTask[PoolResourceT]) error
}

// Task is a simple task representing a unit of work that can be execution in a [Pool].
// Task produces a single value result, and an error.
// Task may be wrapped in a [ValuelessTask] with [github.com/Izzette/go-safeconcurrency/api/pool.TaskWrapper].
// PoolResourceT is the same type as the resource used by the [Pool].
// ValueT is the type of value(s) produced by the task.
// If you do not need to return a value but only an error, you can simply set ValueT to any and return nil from the
// task.
type Task[PoolResourceT any, ValueT any] interface {
	// Execute runs the task with the pool resource and returns the result.
	Execute(context.Context, PoolResourceT) (ValueT, error)
}

// MultiResultTask represents a unit of work that can be executed in a [Pool].
// MultiResultTask may be wrapped in a [ValuelessTask] with
// [github.com/Izzette/go-safeconcurrency/api/pool.WrapMultiResultTask].
// PoolResourceT is the same type as the resource used by the [Pool].
// ValueT is the type of value(s) produced by the task.
type MultiResultTask[PoolResourceT any, ValueT any] interface {
	// Execute runs the task with the pool resource and publishes results to the provided handle.
	Execute(context.Context, PoolResourceT, Handle[ValueT]) error
}

// ValuelessTask is a minimal version of Task that does not produce results.
// Instances of [types.MultiValueTask] and [Task] are wrapped in a [ValuelessTask] in order to allow [Pool] to run tasks
// that produce different results.
// See [github.com/Izzette/go-safeconcurrency/api/pool.WrapMultiResultTask] and
// [github.com/Izzette/go-safeconcurrency/api/pool.TaskWrapper] for info on this functionality.
// You are not expected to implement this interface directly, but rather use the provided wrappers.
type ValuelessTask[PoolResourceT any] interface {
	// Execute runs the task with the pool resource.
	Execute(context.Context, PoolResourceT)
}

// TaskResult is the result of a task execution.
// It is used to propagate results and recoverable errors from the task to the caller.
type TaskResult[ValueT any] interface {
	// Results returns the results channel for the task.
	// This channel may return zero or more results.
	// The channel is closed when the task is finished or the context is canceled.
	Results() <-chan ValueT

	// Drain waits the task to complete and drains the results channel, the results are not returned.
	// It returns the error produced by the task, if any.
	// It is safe to call this method as many times as needed to access the error and will never block after the first call
	// to Drain returns.
	Drain() error
}

// TaskCallback is a callback used by [github.com/Izzette/go-safeconcurrency/api/pool.SubmitMultiResult] to process the
// results of a [types.MultiValueTask] as they are produced.
type TaskCallback[ValueT any] func(context.Context, ValueT) error
