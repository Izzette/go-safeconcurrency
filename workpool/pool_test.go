package workpool

import (
	"sync"
	"sync/atomic"
	"testing"
)

type countingTask struct {
	count *atomic.Int32
	wg    *sync.WaitGroup
}

func (t *countingTask) Execute(res interface{}) {
	defer t.wg.Done()
	t.count.Add(1)
}

func TestPoolConcurrency(t *testing.T) {
	const concurrency = 3
	p := New[any](nil, concurrency)
	defer p.Close()
	p.Start()

	var count atomic.Int32
	wg := &sync.WaitGroup{}
	task := &countingTask{count: &count, wg: wg}

	// Submit more tasks than concurrency
	for i := 0; i < concurrency*2; i++ {
		wg.Add(1)
		p.Requests() <- task
	}

	wg.Wait()

	if count.Load() != concurrency*2 {
		t.Errorf("Expected at least %d concurrent executions, got %d",
			concurrency*2, count.Load())
	}
}

func TestPoolSubmitAfterClose(t *testing.T) {
	p := New[any](nil, 1)
	defer p.Close() // Ensure Close is called even if Start panics
	p.Start()
	p.Close()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when submitting to closed pool")
		}
	}()
	p.Requests() <- &countingTask{}
}
