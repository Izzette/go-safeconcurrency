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

type mockTask struct {
	val int
	err error
}

func (t *mockTask) Execute(ctx context.Context, res interface{}) (int, error) {
	return t.val, t.err
}

func TestSubmitSuccess(t *testing.T) {
	ctx := context.Background()
	p := New[any](nil, 1)
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
	p := New[any](nil, 1)
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

	p := New[any](nil, 1)
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
	p := New[any](nil, 1)
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

type mockStreamingTask2 struct {
	values []string
	err    error
}

func (t *mockStreamingTask2) Execute(ctx context.Context, _ interface{}, h types.Emitter[string]) error {
	for _, v := range t.values {
		if err := h.Emit(ctx, v); err != nil {
			return err
		}
	}

	return t.err
}

func TestSubmitStreamingSuccess(t *testing.T) {
	ctx := context.Background()
	p := New[any](nil, 1)
	p.Start()
	defer p.Close()

	expected := []string{"a", "b", "c"}
	task := &mockStreamingTask2{values: expected}
	results, err := SubmitStreamingCollectAll[any, string](ctx, p, task)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("Expected %v, got %v", expected, results)
	}
}

func TestSubmitStreamingError(t *testing.T) {
	ctx := context.Background()
	p := New[any](nil, 1)
	p.Start()
	defer p.Close()

	expectedErr := errors.New("task error")
	task := &mockStreamingTask2{err: expectedErr}
	_, err := SubmitStreamingCollectAll[any, string](ctx, p, task)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

type mockStreamingTask struct{ t *testing.T }

func (t *mockStreamingTask) Execute(ctx context.Context, res interface{}, h types.Emitter[string]) error {
	if err := h.Emit(ctx, "test"); err != nil {
		t.t.Errorf("Failed to emit result: %v", err)
	}

	return nil
}

func TestSubmitStreamingEarlyCtxCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately to simulate early cancellation.
	cancel()

	p := New[any](nil, 1)
	p.Start()
	defer p.Close()

	task := &mockStreamingTask{t}
	results, err := SubmitStreamingCollectAll[any, string](ctx, p, task)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
	}
}

type mockStreamingBlocksOnContext struct {
	t         *testing.T
	values    []string
	emitCond  *sync.Cond
	emitCount *atomic.Uint32
}

func (t *mockStreamingBlocksOnContext) Execute(ctx context.Context, _ interface{}, h types.Emitter[string]) error {
	for _, v := range t.values {
		if err := h.Emit(ctx, v); err != nil {
			t.t.Fatalf("Failed to emit value: %v", err)
		}
		t.emitCount.Add(1)
	}

	t.broadcastEmit()
	<-ctx.Done()

	return context.Cause(ctx)
}

func (t *mockStreamingBlocksOnContext) broadcastEmit() {
	t.emitCond.L.Lock()
	defer t.emitCond.L.Unlock()
	t.emitCond.Broadcast()
}

func TestSubmitStreamingCtxCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := New[any](nil, 1)
	p.Start()
	defer p.Close()

	emitCond := sync.NewCond(&sync.Mutex{})
	task := &mockStreamingBlocksOnContext{
		t:         t,
		values:    []string{"a"},
		emitCond:  emitCond,
		emitCount: &atomic.Uint32{},
	}

	emitCond.L.Lock()
	go func() {
		defer emitCond.L.Unlock()
		emitCond.Wait()
		cancel()
	}()

	_, err := SubmitStreamingCollectAll[any, string](ctx, p, task)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	if v := task.emitCount.Load(); v != 1 {
		t.Errorf("Expected %d values to be emitted, got %d", len(task.values), v)
	}
}

func TestSubmitStreamingStop(t *testing.T) {
	ctx := context.Background()
	p := New[any](nil, 1)
	p.Start()
	defer p.Close()

	task := &mockStreamingTask2{values: []string{"a", "b", "c"}}
	values := make([]string, 0)
	err := SubmitStreamingBuffered[any, string](ctx, p, task, 0, func(ctx context.Context, value string) error {
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

func TestSubmitStreamingBothErrors(t *testing.T) {
	ctx := context.Background()
	p := New[any](nil, 1)
	p.Start()
	defer p.Close()

	taskErr := errors.New("task error")
	callbackErr := errors.New("callback error")
	task := &mockStreamingTask2{values: []string{"a", "b", "c"}, err: taskErr}
	values := make([]string, 0)
	err := SubmitStreaming[any, string](ctx, p, task, func(ctx context.Context, value string) error {
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
	// Check if the error is the callback error, not the task error.
	if !errors.Is(err, callbackErr) {
		t.Errorf("Expected callback error %v, got %v", callbackErr, err)
	}
	if errors.Is(err, taskErr) {
		t.Errorf("Did not expected task error %v, but it was returned as part of %v", taskErr, err)
	}

	// Make sure that returning any error will not continue to process results.
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(values, expected) {
		t.Errorf("Expected %v, got %v", expected, values)
	}
}

func TestSubmitStreamingCtxCancel2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := New[any](nil, 1)
	p.Start()
	defer p.Close()

	task := &mockStreamingTask2{values: []string{"a", "b", "c"}}
	values := make([]string, 0)
	err := SubmitStreamingBuffered[any, string](ctx, p, task, 0, func(ctx context.Context, value string) error {
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
	p := New[any](nil, 1)
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
