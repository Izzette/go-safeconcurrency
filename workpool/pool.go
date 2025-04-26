package workpool

import (
	"sync"
	"sync/atomic"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// New creates (but does not start) a basic implementation of [types.WorkerPool] with no requests buffering.
// If you would like requests buffering, use [NewBuffered] instead.
// It is equivalent to calling [NewBuffered] with a buffer size of 0.
func New[ResourceT any](resource ResourceT, concurrency int) types.WorkerPool[ResourceT] {
	return NewBuffered(resource, concurrency, 0)
}

// NewBuffered creates (but does not start) a basic implementation of [types.WorkerPool].
// It uses the specified pool resource (passed to each task), concurrency workers, and the specified buffer size for the
// requests channel.
// The concurrency argument must be greater than 0.
// Buffered pools may block tasks from executing until the previously submitted tasks are completed, even if the
// [context.Context] for the task is cancelled.
//
// The resource argument may be set to nil and ResourceT set to type any if a shared pool resource is not required.
func NewBuffered[ResourceT any](resource ResourceT, concurrency int, buffer uint) types.WorkerPool[ResourceT] {
	if concurrency <= 0 {
		panic("Worker pool must have at least one worker!")
	}

	pool := &workerPool[ResourceT]{
		resource:    resource,
		requests:    make(chan types.ValuelessTask[ResourceT], buffer),
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

// workerPool implements [types.WorkerPool].
type workerPool[ResourceT any] struct {
	resource    ResourceT
	requests    chan types.ValuelessTask[ResourceT]
	concurrency uint
	wg          *sync.WaitGroup
	started     *atomic.Bool
	closeOnce   *sync.Once
}

// Start implements [types.WorkerPool.Start].
// Starts the worker pool with the configured concurrency.
func (p *workerPool[ResourceT]) Start() {
	// Check if the pool has already been started.
	if p.started.Swap(true) {
		panic("attempt to start previously started worker pool")
	}

	// The WaitGroup is already populated for the number of workers.
	for i := uint(0); i < p.concurrency; i++ {
		go p.worker()
	}
}

// Close implements [types.WorkerPool.Close].
func (p *workerPool[ResourceT]) Close() {
	p.closeOnce.Do(p.closeRequests)
	if !p.started.Load() {
		return
	}
	p.wg.Wait()
}

// Requests implements [types.WorkerPool.Requests].
// ⚠️ DO NOT close this channel, instead it should be closed by [workerPool.Close].
func (p *workerPool[ResourceT]) Requests() chan<- types.ValuelessTask[ResourceT] {
	return p.requests
}

// worker is a goroutine that executes tasks from the requests channel.
func (p *workerPool[ResourceT]) worker() {
	defer p.wg.Done()

	for task := range p.requests {
		task.Execute(p.resource)
	}
}

// closeRequests closes the requests channel without synchronizing with [workerPool.closeOnce].
func (p *workerPool[ResourceT]) closeRequests() {
	close(p.requests)
}
