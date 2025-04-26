package types

import "context"

// WorkerPool is a worker pool that executes arbitrary tasks concurrently.
// ResourceT is the type of resource used by the pool (ex. API client), it may be set to type any if a shared pool
// resource is not required.
// The resource is shared between all workers and all tasks.
// If you would like to use a separate resource for each concurrently running task, use a [sync.Pool] as your shared
// resource and [sync.Pool.Put] enough resources for the number of workers.
type WorkerPool[ResourceT any] interface {
	// Start initializes the pool and prepares it for task execution.
	Start()

	// Close stops the pool, and waits for all tasks to complete.
	// If the pool was never started, it will be closed immediately.
	// It is safe to call .Close() multiple times.
	Close()

	// Requests returns the channel used to submit tasks to the pool.
	// DO NOT close this channel, instead it should be closed by calling .Close().
	Requests() chan<- ValuelessTask[ResourceT]
}

// Task is a simple task representing a unit of work that can be execution in a [WorkerPool].
// Task produces a single value result, and an error.
// The task may be submitted to a [WorkerPool] using [github.com/Izzette/go-safeconcurrency/workpool.Submit].
// ResourceT is the same type as the resource used by the [WorkerPool].
// ValueT is the type of value(s) produced by the task.
// If you do not need to return a value but only an error, you can simply set ValueT to any and return nil from the
// task.
type Task[ResourceT any, ValueT any] interface {
	// Execute runs the task with the pool resource and returns the result.
	Execute(context.Context, ResourceT) (ValueT, error)
}

// StreamingTask represents a unit of work that can be executed in a [WorkerPool].
// StreamingTask can be submitted to a [WorkerPool] using
// [github.com/Izzette/go-safeconcurrency/workpool.SubmitStreaming].
// ResourceT is the same type as the resource used by the [WorkerPool].
// ValueT is the type of value(s) produced by the task.
type StreamingTask[ResourceT any, ValueT any] interface {
	// Execute runs the task with the pool resource and publishes results to the provided emitter.
	Execute(context.Context, ResourceT, Emitter[ValueT]) error
}

// ValuelessTask is a minimal version of Task that does not produce results.
// Instances of [types.StreamingTask] and [Task] are wrapped in a [ValuelessTask] in order to allow [WorkerPool] to run
// tasks that produce different results.
// See [github.com/Izzette/go-safeconcurrency/workpool.Submit] and
// [github.com/Izzette/go-safeconcurrency/workpool.SubmitStreaming] for information on this functionality.
// You are not expected to implement this interface directly, but rather use the provided helpers to wrap and submit
// tasks.
type ValuelessTask[ResourceT any] interface {
	// Execute runs the task with the pool resource.
	Execute(ResourceT)
}

// TaskFunc can be passed to [github.com/Izzette/go-safeconcurrency/workpool.SubmitFunc] to execute the function as a
// task.
type TaskFunc[ResourceT any] func(context.Context, ResourceT) error

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

// ResultCallback is a callback used by [github.com/Izzette/go-safeconcurrency/workpool.SubmitStreaming] to process the
// results of a [types.StreamingTask] as they are produced.
// If the callback returns any error, the task context will be canceled and the callback will not be called with new
// results.
// If the error is the special [github.com/Izzette/go-safeconcurrency/api/safeconcurrencyerrors.Stop] error, the task
// will be stopped as normal, however the error will not be returned to the caller (simulating break).
type ResultCallback[ValueT any] func(context.Context, ValueT) error
