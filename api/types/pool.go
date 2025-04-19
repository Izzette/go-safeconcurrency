package types

import "context"

// Task represents a unit of work that can be executed in a Pool.
// PoolResourceT is the type of resource used by the pool (ex. API client), it may be of type interface{} and equal to
// nil if this is not required.
// ValueT is the type of value(s) produced by the task.
type Task[PoolResourceT any, ValueT any] interface {
	// Execute runs the task with the pool resource and publishes results to the provided handle.
	Execute(context.Context, PoolResourceT, Handle[ValueT])
}

// SingleTask is a simplified version of Task that produces a single result.
type SingleTask[PoolResourceT any, ValueT any] interface {
	// Execute runs the task with the pool resource and returns the result.
	Execute(context.Context, PoolResourceT) (ValueT, error)
}

// BareTask is a minimal version of Task that does not publish results.
// Tasks and SingleTasks are wrapped in a BareTask in order to allow Pools to run tasks that produce different results.
type BareTask[PoolResourceT any] interface {
	Execute(context.Context, PoolResourceT)
}

// Pool is a worker pool that executes arbitrary tasks concurrently.
type Pool[PoolResourceT any] interface {
	// Start initializes the pool and prepares it for task execution.
	Start()

	// Close stops the pool, and waits for all tasks to complete.
	// It is safe to call Close() multiple times.
	Close()

	// Submit submits a task to the pool for execution.
	// The provided context is used passed to of the task when it is executed in the pool.
	// ⚠️You must never attempt to submit tasks to a pool which has been closed, this will result in a panic!
	Submit(context.Context, BareTask[PoolResourceT]) error
}
