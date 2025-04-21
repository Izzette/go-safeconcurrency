package pool

import (
	"context"
	"sync"
	"testing"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

type mockMultiResultTask struct{ t *testing.T }

func (t *mockMultiResultTask) Execute(ctx context.Context, res interface{}, h types.Handle[string]) error {
	if err := h.Publish(ctx, "test"); err != nil {
		t.t.Errorf("Failed to publish result: %v", err)
	}

	return nil
}

func TestWrapMultiResultTask(t *testing.T) {
	bareTask, taskResult := WrapMultiResultTaskBuffered[interface{}, string](&mockMultiResultTask{t}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Execute task synchronously
	bareTask.Execute(ctx, nil)

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

func TestTaskWrapper(t *testing.T) {
	bareTask, taskResult := TaskWrapper[interface{}, int](&mockTask{val: 42})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bareTask.Execute(ctx, nil)
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
