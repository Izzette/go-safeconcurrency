package workpool

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/safeconcurrencyerrors"
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
	results, err := SubmitMultiResultCollectAll[any, string](ctx, p, task)
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
	_, err := SubmitMultiResultCollectAll[any, string](ctx, p, task)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestSubmitMultiResultEarlyContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately to simulate early cancellation.
	cancel()

	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	task := &mockMultiResultTask{t}
	results, err := SubmitMultiResultCollectAll[any, string](ctx, p, task)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
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

	return context.Cause(ctx)
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

	_, err := SubmitMultiResultCollectAll[any, string](ctx, p, task)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	if v := task.publishedCount.Load(); v != 1 {
		t.Errorf("Expected %d values to be published, got %d", len(task.values), v)
	}
}

func TestSubmitMultiResultStop(t *testing.T) {
	ctx := context.Background()
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	task := &mockMultiResultTask2{values: []string{"a", "b", "c"}}
	values := make([]string, 0)
	err := SubmitMultiResultBuffered[any, string](ctx, p, task, 0, func(ctx context.Context, value string) error {
		values = append(values, value)
		if value == "b" {
			return safeconcurrencyerrors.Stop
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := []string{"a", "b"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Expected %v, got %v", expected, values)
	}
}

func TestSubmitMultiResultBothErrors(t *testing.T) {
	ctx := context.Background()
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	taskErr := errors.New("task error")
	callbackErr := errors.New("callback error")
	task := &mockMultiResultTask2{values: []string{"a", "b", "c"}, err: taskErr}
	values := make([]string, 0)
	err := SubmitMultiResult[any, string](ctx, p, task, func(ctx context.Context, value string) error {
		values = append(values, value)
		// Ensure the callback error is only after all the values are processed.
		if value == "c" {
			return callbackErr
		}

		return nil
	})
	if err == nil {
		t.Fatal("Expected non-nil error")
	}
	// Check if the error is a join error and contains both task and callback errors.
	if !errors.Is(err, taskErr) || !errors.Is(err, callbackErr) {
		t.Errorf("Expected task error %v and callback error %v, got %v", taskErr, callbackErr, err)
	}

	// Make sure that returning any error will not continue to process results.
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Expected %v, got %v", expected, values)
	}
}

func TestSubmitMultiResultContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	task := &mockMultiResultTask2{values: []string{"a", "b", "c"}}
	values := make([]string, 0)
	err := SubmitMultiResultBuffered[any, string](ctx, p, task, 0, func(ctx context.Context, value string) error {
		values = append(values, value)
		if value == "b" {
			// Cancel the context to simulate a cancelation in-flight.
			cancel()
			// Do not return an error here, as we want to test the behavior of the context cancelation.
			return nil
		}

		return nil
	})
	if err == nil {
		t.Fatal("Expected non-nil error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	expected := []string{"a", "b"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Expected %v, got %v", expected, values)
	}
}

func TestSubmitFunc(t *testing.T) {
	ctx := context.Background()
	p := NewPool[any](nil, 1)
	p.Start()
	defer p.Close()

	ran := false
	anError := errors.New("an error")
	err := SubmitFunc[any](ctx, p, func(ctx context.Context, res interface{}) error {
		ran = true

		return anError
	})
	if !ran {
		t.Errorf("Expected task to be run, but it was not")
	}
	if !errors.Is(err, anError) {
		t.Errorf("Expected error %v, got %v", anError, err)
	}
}
