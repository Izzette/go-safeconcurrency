package workpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

type countingTask struct {
	count *atomic.Int32
	wg    *sync.WaitGroup
}

func (t *countingTask) Execute(ctx context.Context, res interface{}) {
	defer t.wg.Done()
	t.count.Add(1)
}

func TestPoolConcurrency(t *testing.T) {
	const concurrency = 3
	p := NewPool[any](nil, concurrency)
	p.Start()
	defer p.Close()

	var count atomic.Int32
	wg := &sync.WaitGroup{}
	task := &countingTask{count: &count, wg: wg}

	// Submit more tasks than concurrency
	for i := 0; i < concurrency*2; i++ {
		wg.Add(1)
		if err := p.Submit(context.Background(), task); err != nil {
			t.Fatalf("Failed to submit task: %v\n", err)
		}
	}

	wg.Wait()

	if count.Load() != concurrency*2 {
		t.Errorf("Expected at least %d concurrent executions, got %d",
			concurrency*2, count.Load())
	}
}

func TestPoolSubmitAfterClose(t *testing.T) {
	p := NewPool[any](nil, 1)
	p.Start()
	p.Close()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when submitting to closed pool")
		}
	}()
	if err := p.Submit(context.Background(), &countingTask{}); err != nil {
		t.Errorf("Expected panic, got error: %v", err)
	}
}

func TestPoolContextCancel(t *testing.T) {
	p := NewPoolBuffered[any](nil, 1, 0)
	p.Start()
	defer p.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.Submit(ctx, &countingTask{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context canceled error, got %v", err)
	}
}
