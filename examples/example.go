package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Izzette/go-safeconcurrency/api/generator"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

type myJob struct{}

func (r *myJob) Run(ctx context.Context, h types.Handle[int]) error {
	for i := 0; i < 10; i += 1 {
		if cancelation := h.Publish(ctx, i); cancelation != nil {
			return cancelation
		}
	}

	return nil
}

func main() {
	w := generator.NewGeneratorBuffered[int](&myJob{}, 1)

	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	w.Start(ctx)
	defer func() {
		cancelFunc()
		if err := w.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Generator produced a fatal error: %q\n", err)
		}
	}()

	for result := range w.Results() {
		if value, err := result.Get(); err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] Generator produced a recoverable error: %q\n", err)
		} else {
			if _, err = fmt.Fprintf(os.Stdout, "%d\n", value); err != nil {
				panic(err)
			}
		}
	}
}
