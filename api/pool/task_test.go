package pool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

type mockTask struct{ t *testing.T }

func (t *mockTask) Execute(ctx context.Context, res interface{}, h types.Handle[string]) {
	if err := h.Publish(ctx, "test"); err != nil {
		t.t.Errorf("Failed to publish result: %v", err)
	}
}

func TestTaskWithHandle(t *testing.T) {
	bareTask, ch := TaskWithHandleBuffered[interface{}, string](&mockTask{t}, 1)
	ctx := context.Background()

	// Execute task synchronously
	bareTask.Execute(ctx, nil)

	select {
	case res := <-ch:
		val, err := res.Get()
		if err != nil || val != "test" {
			t.Errorf("Unexpected result: %v, %v", val, err)
		}
	default:
		t.Error("No result received")
	}
}

type mockSingleTask struct{}

func (t *mockSingleTask) Execute(ctx context.Context, res interface{}) (int, error) {
	return 42, nil
}

func TestTaskWithReturn(t *testing.T) {
	bareTask, ch := TaskWithReturn[interface{}, int](&mockSingleTask{})
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bareTask.Execute(ctx, nil)
	}()

	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()
	select {
	case res := <-ch:
		val, err := res.Get()
		if err != nil || val != 42 {
			t.Errorf("Unexpected result: %v, %v", val, err)
		}
	case <-timer.C:
		t.Error("No result received")
	}

	wg.Wait()
}
