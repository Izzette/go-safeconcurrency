package pool

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

func TestSubmitSuccess(t *testing.T) {
	ctx := context.Background()
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	task := &mockTask{val: 42}
	val, err := Submit[any, int](ctx, p, task)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestSubmitTaskError(t *testing.T) {
	ctx := context.Background()
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	expectedErr := errors.New("task error")
	task := &mockTask{err: expectedErr}
	_, err := Submit[any, int](ctx, p, task)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestSubmitContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	task := &mockTask{val: 42}
	_, err := Submit[any, int](ctx, p, task)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

type mockTaskBlock struct {
	blockingChannel chan struct{}
	startedCond     *sync.Cond
	val             int
	err             error
}

func (t *mockTaskBlock) Execute(ctx context.Context, res interface{}) (int, error) {
	t.broadcastStarted()

	<-t.blockingChannel

	return t.val, t.err
}

func (t *mockTaskBlock) broadcastStarted() {
	t.startedCond.L.Lock()
	defer t.startedCond.L.Unlock()
	t.startedCond.Broadcast()
}

func TestSubmitContextCancelledDuringWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	block := make(chan struct{})
	defer close(block)
	startedCond := sync.NewCond(&sync.Mutex{})
	task := &mockTaskBlock{blockingChannel: block, startedCond: startedCond, val: 42, err: nil}

	startedCond.L.Lock()
	go func() {
		defer startedCond.L.Unlock()
		startedCond.Wait()
		cancel()
	}()

	_, err := Submit[any, int](ctx, p, task)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

type mockMultiResultTask2 struct {
	values []string
	err    error
}

func (t *mockMultiResultTask2) Execute(ctx context.Context, _ interface{}, h types.Handle[string]) error {
	for _, v := range t.values {
		if err := h.Publish(ctx, v); err != nil {
			return err
		}
	}

	return t.err
}

func TestSubmitMultiResultSuccess(t *testing.T) {
	ctx := context.Background()
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	expected := []string{"a", "b", "c"}
	task := &mockMultiResultTask2{values: expected}
	results, err := SubmitMultiResult[any, string](ctx, p, task)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestSubmitMultiResultTaskError(t *testing.T) {
	ctx := context.Background()
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	expectedErr := errors.New("task error")
	task := &mockMultiResultTask2{err: expectedErr}
	_, err := SubmitMultiResult[any, string](ctx, p, task)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

type mockMultiResultTaskBlocksOnContext struct {
	t              *testing.T
	values         []string
	publishCond    *sync.Cond
	publishedCount *atomic.Uint32
}

func (t *mockMultiResultTaskBlocksOnContext) Execute(ctx context.Context, _ interface{}, h types.Handle[string]) error {
	for _, v := range t.values {
		if err := h.Publish(ctx, v); err != nil {
			t.t.Fatalf("Failed to publish value: %v", err)
		}
		t.publishedCount.Add(1)
	}

	t.broadcastPublish()
	<-ctx.Done()

	return ctx.Err()
}

func (t *mockMultiResultTaskBlocksOnContext) broadcastPublish() {
	t.publishCond.L.Lock()
	defer t.publishCond.L.Unlock()
	t.publishCond.Broadcast()
}

func TestSubmitMultiResultContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	publishCond := sync.NewCond(&sync.Mutex{})
	task := &mockMultiResultTaskBlocksOnContext{
		t:              t,
		values:         []string{"a"},
		publishCond:    publishCond,
		publishedCount: &atomic.Uint32{},
	}

	publishCond.L.Lock()
	go func() {
		defer publishCond.L.Unlock()
		publishCond.Wait()
		cancel()
	}()

	_, err := SubmitMultiResult[any, string](ctx, p, task)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	if v := task.publishedCount.Load(); v != 1 {
		t.Errorf("Expected %d values to be published, got %d", len(task.values), v)
	}
}
