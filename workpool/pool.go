package workpool

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// NewPool creates (but does not start) a basic implementation of [types.Pool] with no requests buffering.
// If you would like requests buffering, use [NewPoolBuffered] instead.
// It is equivalent to calling [NewPoolBuffered] with a buffer size of 0.
func NewPool[ResourceT any](resource ResourceT, concurrency int) types.Pool[ResourceT] {
	return NewPoolBuffered(resource, concurrency, 0)
}

// NewPoolBuffered creates (but does not start) a basic implementation of [types.Pool].
// It uses the specified pool resource (passed to each task), concurrency workers, and the specified buffer size for the
// requests channel (used by [types.Pool.Submit] to queue tasks).
//
// The resource argument may be set to nil and ResourceT set to type any if a shared pool resource is not required.
func NewPoolBuffered[ResourceT any](resource ResourceT, concurrency int, buffer uint) types.Pool[ResourceT] {
	if concurrency <= 0 {
		panic("Worker pool must have at least one worker!")
	}

	pool := &pool[ResourceT]{
		resource:    resource,
		requests:    make(chan contextualTask[ResourceT], buffer),
		concurrency: uint(concurrency),
		wg:          &sync.WaitGroup{},
		started:     &atomic.Bool{},
		closeOnce:   &sync.Once{},
	}
	// We will run concurrency workers when Start() is called.
	// This WaitGroup must be pre-populated in the case that Wait() is called in another goroutine before Start().
	pool.wg.Add(concurrency)

	return pool
}

// contextualTask is a wrapper for [types.ValuelessTask] that adds a [context.Context] to the task.
type contextualTask[ResourceT any] struct {
	types.ValuelessTask[ResourceT]
	getContext func() context.Context
}

// pool implements [types.Pool].
type pool[ResourceT any] struct {
	resource    ResourceT
	requests    chan contextualTask[ResourceT]
	concurrency uint
	wg          *sync.WaitGroup
	started     *atomic.Bool
	closeOnce   *sync.Once
}

// Start implements [types.Pool.Start].
// Starts the worker pool with the configured concurrency.
func (p *pool[ResourceT]) Start() {
	// Check if the pool has already been started.
	if p.started.Swap(true) {
		panic("attempt to start previously started pool.Pool")
	}

	// The WaitGroup is already populated for the number of workers.
	for i := uint(0); i < p.concurrency; i++ {
		go p.worker()
	}
}

// Close implements [types.Pool.Close].
func (p *pool[ResourceT]) Close() {
	p.closeOnce.Do(p.closeRequests)
	if !p.started.Load() {
		return
	}
	p.wg.Wait()
}

// Submit implements [types.Pool.Submit].
func (p *pool[ResourceT]) Submit(ctx context.Context, task types.ValuelessTask[ResourceT]) error {
	// select is not deterministic, and may still send tasks even if the context has been canceled.
	if err := context.Cause(ctx); err != nil {
		//nolint:wrapcheck
		return err
	}

	ctxTask := contextualTask[ResourceT]{
		ValuelessTask: task,
		getContext:    func() context.Context { return ctx },
	}

	// Submit the task to the requests channel.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return context.Cause(ctx)
	case p.requests <- ctxTask:
		return nil
	}
}

// worker is a goroutine that executes tasks from the requests channel.
func (p *pool[ResourceT]) worker() {
	defer p.wg.Done()

	for task := range p.requests {
		task.Execute(task.getContext(), p.resource)
	}
}

// closeRequests closes the requests channel without synchronizing with [pool.closeOnce].
func (p *pool[ResourceT]) closeRequests() {
	close(p.requests)
}
