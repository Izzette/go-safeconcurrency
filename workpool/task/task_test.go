package task

import (
	"context"
	"sync"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

type mockStreamingTask struct{ t *testing.T }

func (t *mockStreamingTask) Execute(ctx context.Context, res interface{}, h types.Emitter[string]) error {
	if err := h.Emit(ctx, "test"); err != nil {
		t.t.Errorf("Failed to publish result: %v", err)
	}

	return nil
}

func TestWrapStreamingTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bareTask, taskResult := WrapStreamingTask[interface{}, string](ctx, &mockStreamingTask{t}, 1)

	// Execute task synchronously
	bareTask.Execute(nil)

	results := make([]string, 0, 1)
	for val := range taskResult.Results() {
		results = append(results, val)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	} else if results[0] != "test" {
		t.Errorf("Expected 'test', got '%s'", results[0])
	}
	if err := taskResult.Drain(); err != nil {
		t.Errorf("Unexpected error from taskResult: %v", err)
	}
}

type mockTask struct {
	val int
	err error
}

func (t *mockTask) Execute(ctx context.Context, res interface{}) (int, error) {
	return t.val, t.err
}

func TestWrapTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bareTask, taskResult := WrapTask[interface{}, int](ctx, &mockTask{val: 42})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bareTask.Execute(nil)
	}()

	val, ok := <-taskResult.Results()
	if !ok {
		t.Error("Results channel closed unexpectedly")
	} else if val != 42 {
		t.Errorf("Unexpected result: %v", val)
	}
	if err := taskResult.Drain(); err != nil {
		t.Errorf("Unexpected error from taskResult: %v", err)
	}

	wg.Wait()
}

func TestWrapTaskFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bareTask, taskResult := WrapTaskFunc[interface{}](ctx, func(ctx context.Context, res interface{}) error {
		return nil
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bareTask.Execute(nil)
	}()

	val, ok := <-taskResult.Results()
	if !ok {
		t.Error("Results channel closed unexpectedly")
	} else if val != struct{}{} {
		t.Errorf("Unexpected result: %v", val)
	}
	if err := taskResult.Drain(); err != nil {
		t.Errorf("Unexpected error from taskResult: %v", err)
	}

	wg.Wait()
}
