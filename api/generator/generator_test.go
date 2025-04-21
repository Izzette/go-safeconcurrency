package generator

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

type testRunner struct {
	values []int
	err    error
}

func (r *testRunner) Run(ctx context.Context, h types.Handle[int]) error {
	for _, v := range r.values {
		if err := h.Publish(ctx, v); err != nil {
			return err
		}
	}

	return r.err
}

func TestGeneratorSendsAllValues(t *testing.T) {
	expectedValues := []int{1, 2, 3}
	runner := &testRunner{values: expectedValues}
	gen := NewGeneratorBuffered[int](runner, uint(len(expectedValues)))
	ctx := context.Background()

	go gen.Start(ctx)

	var received []int
	for val := range gen.Results() {
		received = append(received, val)
	}

	if !reflect.DeepEqual(received, expectedValues) {
		t.Errorf("expected %v, got %v", expectedValues, received)
	}

	if err := gen.Wait(); err != nil {
		t.Errorf("unexpected error from Wait: %v", err)
	}
}

func TestGeneratorContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	runner := &testRunner{values: []int{1, 2, 3, 4, 5}}
	gen := NewGenerator[int](runner)

	go gen.Start(ctx)
	cancel()

	var count int
	for range gen.Results() {
		count++
	}

	if count > 0 {
		t.Logf("received %d results before cancellation", count)
	}

	err := gen.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGeneratorFatalError(t *testing.T) {
	expectedErr := errors.New("fatal error")
	runner := &testRunner{err: expectedErr}
	gen := NewGenerator[int](runner)
	ctx := context.Background()

	go gen.Start(ctx)

	err := gen.Wait()
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestGeneratorStartPanicsWhenCalledTwice(t *testing.T) {
	gen := NewGenerator[int](&testRunner{})
	ctx := context.Background()
	gen.Start(ctx)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when starting generator twice")
		}
	}()
	gen.Start(ctx)
}

type testRunnerDone struct {
	values []int
	err    error
	done   chan struct{}
}

func (r *testRunnerDone) Run(ctx context.Context, h types.Handle[int]) error {
	defer close(r.done)
	for _, v := range r.values {
		if err := h.Publish(ctx, v); err != nil {
			return err
		}
	}

	return r.err
}

func TestBufferedGenerator(t *testing.T) {
	values := []int{1, 2, 3}
	runner := &testRunnerDone{values: values, done: make(chan struct{})}
	gen := NewGeneratorBuffered[int](runner, uint(len(values)))
	ctx := context.Background()

	go gen.Start(ctx)

	// Wait for the generator to finish sending results
	<-runner.done

	if len(gen.Results()) != len(values) {
		t.Errorf("expected %d buffered results, got %d", len(values), len(gen.Results()))
	}

	var received []int
	for val := range gen.Results() {
		received = append(received, val)
	}

	if !reflect.DeepEqual(received, values) {
		t.Errorf("expected %v, got %v", values, received)
	}
	if err := gen.Wait(); err != nil {
		t.Errorf("unexpected error from generator: %v", err)
	}
}
