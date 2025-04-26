package generator

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

type testProducer struct {
	values []int
	err    error
}

func (r *testProducer) Run(ctx context.Context, h types.Emitter[int]) error {
	for _, v := range r.values {
		if err := h.Emit(ctx, v); err != nil {
			return err
		}
	}

	return r.err
}

func TestGeneratorSendsAllValues(t *testing.T) {
	expectedValues := []int{1, 2, 3}
	producer := &testProducer{values: expectedValues}
	gen := NewGeneratorBuffered[int](producer, uint(len(expectedValues)))
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
	producer := &testProducer{values: []int{1, 2, 3, 4, 5}}
	gen := NewGenerator[int](producer)

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
	producer := &testProducer{err: expectedErr}
	gen := NewGenerator[int](producer)
	ctx := context.Background()

	go gen.Start(ctx)

	err := gen.Wait()
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestGeneratorStartPanicsWhenCalledTwice(t *testing.T) {
	gen := NewGenerator[int](&testProducer{})
	ctx := context.Background()
	gen.Start(ctx)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when starting generator twice")
		}
	}()
	gen.Start(ctx)
}

type testProducerDone struct {
	values []int
	err    error
	done   chan struct{}
}

func (r *testProducerDone) Run(ctx context.Context, h types.Emitter[int]) error {
	defer close(r.done)
	for _, v := range r.values {
		if err := h.Emit(ctx, v); err != nil {
			return err
		}
	}

	return r.err
}

func TestBufferedGenerator(t *testing.T) {
	values := []int{1, 2, 3}
	producer := &testProducerDone{values: values, done: make(chan struct{})}
	gen := NewGeneratorBuffered[int](producer, uint(len(values)))
	ctx := context.Background()

	go gen.Start(ctx)

	// Wait for the generator to finish sending results
	<-producer.done

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
