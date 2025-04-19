package pool

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

type contextualTask[PoolResourceT any] struct {
	types.BareTask[PoolResourceT]
	getContext func() context.Context
}

// Pool represents a worker pool that can execute tasks concurrently.
type Pool[PoolResourceT any] struct {
	resource    PoolResourceT
	requests    chan contextualTask[PoolResourceT]
	concurrency uint
	wg          *sync.WaitGroup
	started     *atomic.Bool
}

// NewPool creates a new worker pool with the specified pool resource (passed to each task) and concurrency workers.
// It is equivalent to NewPoolBuffered with a buffer size of 0.
func NewPool[PoolResourceT any](resource PoolResourceT, concurrency int) *Pool[PoolResourceT] {
	return NewPoolBuffered(resource, concurrency, 0)
}

// NewPoolBuffered creates a new worker pool with the specified pool resource (passed to each task), concurrency
// workers, and the specified buffer size for the requests channel.
func NewPoolBuffered[PoolResourceT any](resource PoolResourceT, concurrency int, buffer uint) *Pool[PoolResourceT] {
	if concurrency <= 0 {
		panic("Worker pool must have at least one worker!")
	}

	pool := &Pool[PoolResourceT]{
		resource:    resource,
		requests:    make(chan contextualTask[PoolResourceT], buffer),
		concurrency: uint(concurrency),
		wg:          &sync.WaitGroup{},
		started:     &atomic.Bool{},
	}
	// We will run concurrency workers when Start() is called.
	// This WaitGroup must be pre-populated in the case that Wait() is called in another goroutine before Start().
	pool.wg.Add(concurrency)

	return pool
}

// Start implements github.com/Izzette/go-safeconcurrency/api/types/Pool.Start.
// Start the worker pool with the configured concurrency.
func (p *Pool[PoolResourceT]) Start() {
	// Check if the pool has already been started.
	if p.started.Swap(true) {
		panic("attempt to start previously started pool.Pool")
	}

	// The WaitGroup is already populated for the number of workers.
	for i := uint(0); i < p.concurrency; i++ {
		go p.worker()
	}
}

// Close implements github.com/Izzette/go-safeconcurrency/api/types/Pool.Close.
// Close stops the pool, and waits for all tasks to complete.
// It is safe to call Close() multiple times.
func (p *Pool[PoolResourceT]) Close() {
	close(p.requests)
	if !p.started.Load() {
		return
	}
	p.wg.Wait()
}

// Submit implements github.com/Izzette/go-safeconcurrency/api/types/Pool.Submit.
// The provided context is used passed to of the task when it is executed in the pool.
// ⚠️You must never attempt to submit tasks to a pool which has been closed, this will result in a panic!
func (p *Pool[PoolResourceT]) Submit(ctx context.Context, task types.BareTask[PoolResourceT]) error {
	// select is not deterministic, and may still send tasks even if the context has been canceled.
	if ctx.Err() != nil {
		//nolint:wrapcheck
		return ctx.Err()
	}

	ctxTask := contextualTask[PoolResourceT]{
		BareTask:   task,
		getContext: func() context.Context { return ctx },
	}

	// Submit the task to the requests channel.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return ctx.Err()
	case p.requests <- ctxTask:
		return nil
	}
}

// worker is a goroutine that executes tasks from the requests channel.
func (p *Pool[PoolResourceT]) worker() {
	defer p.wg.Done()

	for task := range p.requests {
		task.Execute(task.getContext(), p.resource)
	}
}
