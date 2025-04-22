package results

import (
	"context"
	"errors"
	"testing"
)

func TestHandlePublishSendsResult(t *testing.T) {
	resultsChan := make(chan int, 1)
	h := NewHandle[int](resultsChan)

	ctx := context.Background()
	value := 42
	err := h.Publish(ctx, value)
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	select {
	case val := <-resultsChan:
		if val != value {
			t.Errorf("expected %d, got %d", value, val)
		}
	default:
		t.Error("no result was sent to the channel")
	}
}

func TestHandlePublishAfterContextCancel(t *testing.T) {
	resultsChan := make(chan int, 1)
	h := NewHandle[int](resultsChan)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := h.Publish(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestHandleCloseClosesChannel(t *testing.T) {
	resultsChan := make(chan int)
	h := NewHandle[int](resultsChan)
	h.Close()

	_, ok := <-resultsChan
	if ok {
		t.Error("Expected channel to be closed after Close()")
	}
}
