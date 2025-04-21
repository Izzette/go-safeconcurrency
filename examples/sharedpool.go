package examples

import (
	"context"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/types"
)

// IntTask implements [types.Task].
// It returns 42 when executed.
type IntTask struct {
	Value int
}

// Execute implements [types.Task.Execute].
func (t *IntTask) Execute(ctx context.Context, _ any) (int, error) {
	return t.Value, nil
}

// StringTask implements [types.MultiResultTask].
// It publishes two strings to the provided handle.
type StringTask struct {
	Strings []string
}

// Execute implements [types.MultiResultTask.Execute].
func (t *StringTask) Execute(ctx context.Context, _ any, h types.Handle[string]) error {
	for _, s := range t.Strings {
		if err := h.Publish(ctx, s); err != nil {
			// The context is canceled, we should stop publishing results.
			return err
		}
	}
	return nil
}

// LogTask implements [types.Task].
// It logs a message to the console.
type LogTask struct {
	Message string
}

// Execute implements [types.Task.Execute].
func (t *LogTask) Execute(ctx context.Context, _ any) (any, error) {
	fmt.Println(t.Message)
	return nil, nil
}
