package workpool

import (
	"context"
	"runtime"
	"sync"
	"testing"
)

// benchTask is a simple [types.Task] that does nothing.
type benchTask struct{}

// Execute implements the [types.Task.Execute] interface.
func (t *benchTask) Execute(ctx context.Context, _ any) (struct{}, error) {
	return struct{}{}, nil
}

// BenchmarkSubmitRTT measures the performance of submitting tasks to a worker pool
// There is no backlog of tasks, so the performance of the worker pool is hindered in terms of maximum throughput.
// However, we are measuring both the time to submit tasks and to execute them (effectively the round-trip time).
func BenchmarkSubmitRTT(b *testing.B) {
	pool := NewPool[any](nil, 1)
	defer pool.Close() // Ensure Close is called even if Start panics
	pool.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task := &benchTask{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Submit[any, struct{}](ctx, pool, task)
		if err != nil {
			b.Fatal(err)
		}
	}
	// Stop timer before pool.Close() and context.CancelFunc() to avoid measuring cleanup time.
	b.StopTimer()
}

// wgDoneTask implements [types.ValuelessTask].
// It calls [sync.WaitGroup.Done] when executed.
type wgDoneTask struct{}

// Execute implements the [types.ValuelessTask.Execute] interface.
func (*wgDoneTask) Execute(wg *sync.WaitGroup) {
	wg.Done()
}

// BenchmarkPoolThroughput measures the maximum throughput of the worker pool using all available cores.
// It calls [types.Pool.Submit] directly with a [typse.ValuelessTask], rather than using the [Submit] family of helper
// functions for maximum throughput.
// It runs using a work pool of concurrency workers equal to the number of CPUs available and waits for their completion
// using a [*sync.WaitGroup].
// This can actually be slower when 1 CPU is used, as the overhead of locking and conditions for the underlying requests
// channel increases with the number of workers.
// This overhead can actually be higher than that of processing the task itself, and thus [BenchmarkPoolThroughput1]
// generally shows higher throughput, depending on OS and CPU architecture.
// In the real world, for tasks that actually take some time to execute, this overhead on the channel is negligible.
func BenchmarkPoolThroughput(b *testing.B) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}
	wg.Add(b.N)
	pool := NewPoolBuffered[*sync.WaitGroup](wg, numCPU, 16*uint(numCPU))
	defer pool.Close() // Ensure Close is called even if Start panics
	pool.Start()

	task := &wgDoneTask{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Requests() <- task
	}
	wg.Wait()
	// Stop timer before pool.Close() and context.CancelFunc() to avoid measuring cleanup time.
	b.StopTimer()
}

// BenchmarkPoolThroughput1 is a variant of [BenchmarkPoolThroughput] that uses a single worker.
func BenchmarkPoolThroughput1(b *testing.B) {
	wg := &sync.WaitGroup{}
	pool := NewPoolBuffered[*sync.WaitGroup](wg, 1, 16)
	defer pool.Close() // Ensure Close is called even if Start panics
	pool.Start()

	task := &wgDoneTask{}

	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Requests() <- task
	}
	wg.Wait()
	// Stop timer before pool.Close() and context.CancelFunc() to avoid measuring cleanup time.
	b.StopTimer()
}
