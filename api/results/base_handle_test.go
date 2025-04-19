package results

import (
	"context"
	"errors"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

func TestHandlePublishSendsResult(t *testing.T) {
	resultsChan := make(chan types.Result[int], 1)
	h := NewHandle[int](resultsChan)

	ctx := context.Background()
	value := 42
	err := h.Publish(ctx, value)
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	select {
	case result := <-resultsChan:
		val, err := result.Get()
		if err != nil {
			t.Errorf("result.Get() returned error: %v", err)
		}
		if val != value {
			t.Errorf("expected %d, got %d", value, val)
		}
	default:
		t.Error("no result was sent to the channel")
	}
}

func TestHandleErrorSendsError(t *testing.T) {
	resultsChan := make(chan types.Result[int], 1)
	h := NewHandle[int](resultsChan)

	ctx := context.Background()
	expectedErr := errors.New("test error")
	err := h.Error(ctx, expectedErr)
	if err != nil {
		t.Fatalf("Error returned error: %v", err)
	}

	select {
	case result := <-resultsChan:
		_, actualErr := result.Get()
		if !errors.Is(actualErr, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, actualErr)
		}
	default:
		t.Error("no error was sent to the channel")
	}
}

func TestHandlePublishAfterContextCancel(t *testing.T) {
	resultsChan := make(chan types.Result[int], 1)
	h := NewHandle[int](resultsChan)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := h.Publish(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
