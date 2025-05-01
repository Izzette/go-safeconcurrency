package safeconcurrency_test

import (
	"context"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/eventloop"
	"github.com/Izzette/go-safeconcurrency/eventloop/snapshot"
	"github.com/Izzette/go-safeconcurrency/generator"
	"github.com/Izzette/go-safeconcurrency/workpool"
)

type MyProducer struct{}

type Output struct {
	// Some output data
	data int
}

func (r *MyProducer) Run(ctx context.Context, h types.Emitter[Output]) error {
	// Your concurrent logic here
	value := Output{data: 42} // Example output
	err := h.Emit(ctx, value)

	return err
}

func Example_generator() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gen := generator.New[Output](&MyProducer{})
	gen.Start(ctx)

	for val := range gen.Results() {
		fmt.Println(val)
	}
	// Output:
	// {42}
}

type Task struct{}

type ResourceType struct {
	// Some resource data
	data string
}

func (t *Task) Execute(ctx context.Context, resource ResourceType) (Output, error) {
	// Your task logic here
	return Output{len(resource.data)}, nil
}

func Example_workerPool() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resource := ResourceType{data: "example data"}
	concurrency := 5
	mypool := workpool.New[ResourceType](resource, concurrency)
	mypool.Start()
	defer mypool.Close()

	val, err := workpool.Submit[ResourceType, Output](ctx, mypool, &Task{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Task result: %d\n", val.data)
	// Output:
	// Task result: 12
}

type RequestEvent struct{}

func (e *RequestEvent) Dispatch(gen types.GenerationID, s *AppState) *AppState {
	//nolint:forbidigo
	fmt.Printf("Processing request #%d\n", gen)
	s.Requests++

	return s
}

type AppState struct{ Requests int }

func (s *AppState) Copy() *AppState {
	return snapshot.CopyPtr(s)
}

func Example_eventLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	snap := snapshot.NewCopyable(&AppState{})
	el := eventloop.New[*AppState](snap)
	defer el.Close()
	el.Start()

	gen, err := el.Send(ctx, &RequestEvent{})
	if err != nil {
		panic(err)
	}

	snap, err = eventloop.WaitForGeneration(ctx, el, gen)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Current requests: %d\n", snap.State().Requests)
	// Output:
	// Processing request #1
	// Current requests: 1
}
