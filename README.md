# SafeConcurrency for Go

<img align="left" width="250" height="250" alt="SafeConcurrency Logo" src="doc/assets/images/logo-500x500.png">

SafeConcurrency is a Go library designed to simplify the management of concurrent tasks, providing a safe and structured way to produce and consume results.
It enforces best practices for context propagation, error handling, and resource cleanup.

⚠️ **Warning**: This library is in early development and may not be suitable for production use yet. ⚠️
This project is still in early development, and the API may change as we refine the design and functionality.
Expect new features and improvements in future releases, generators are just the beginning.

![Go Version](https://img.shields.io/badge/go-1.19-blue) ![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Generator Pattern**: Safely produce values from concurrent operations via channel-based results
  - Adheres to Go best-practice: ["Do not communicate by sharing memory; instead, share memory by communicating"](https://go.dev/blog/codelab-share)
- **Context Integration**: Built-in support for context cancellation and deadlines
- **Error Handling**: Gracefully handle errors with recoverable and fatal error reporting
- **Thread-Safe**: All APIs are designed for concurrent use with goroutines
- **Flexible Buffering**: Configurable result channel buffering for different throughput needs
- **Worker Pools**: Manage concurrent task execution with configurable concurrency and buffering

### Planned Features

- **Parallel Mapping**: Support for parallel mapping of input values to output results
- **Pipeline Support**: Create pipelines of generators for complex workflows
- **Event Loops**: Support for event loops for handling asynchronous events

## Usage

### Basic Generator Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Izzette/go-safeconcurrency/api/generator"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

type Counter struct{}

func (r *Counter) Run(ctx context.Context, h types.Handle[int]) error {
	for i := 0; i < 5; i++ {
		// Stop if context is cancelled
		if err := h.Publish(ctx, i); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func main() {
	gen := generator.NewGeneratorBuffered[int](&Counter{}, 3)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go gen.Start(ctx)

	for result := range gen.Results() {
		if val, err := result.Get(); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Received: %d\n", val)
		}
	}

	if err := gen.Wait(); err != nil {
		fmt.Printf("Final error: %v\n", err)
	}
}
```

See [examples](examples) for more detailed examples.

### Key Components

1. **Runner Interface**
   Implement your concurrent logic by creating a type that satisfies `types.Runner`:
   ```go
   type MyRunner struct{}

   func (r *MyRunner) Run(ctx context.Context, h types.Handle[Output]) error {
       // Your concurrent logic here
       h.Publish(ctx, value)
       h.Error(ctx, recoverableErr)
       return fatalErr
   }
   ```

2. **Generator**
   Create and manage concurrent execution:
   ```go
   gen := generator.NewGenerator[Output](&MyRunner{})
   gen.Start(ctx)
   ```

3. **Result Handling**
   Consume results safely from the channel:
   ```go
   for result := range gen.Results() {
       val, err := result.Get()
       // Handle value/error
   }
   ```

4. **Worker Pools**
   Create and manage worker pools:
   ```go
   pool := pool.NewPool[ResourceType](resource, concurrency)
   pool.Start()
   defer pool.Close()

   // Submit tasks
   pool.Submit(ctx, task)
   ```

## Documentation

Full API documentation is available on [GoDoc](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency).

## Contributing

We welcome contributions! Please follow these guidelines:

1. Set up development environment:
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   pre-commit install
   ```

2. Ensure all tests pass:
   ```bash
   pre-commit run --all-files
   go test -v ./...
   ```

3. Add tests for new features
4. Update documentation accordingly
5. Open a pull request with a clear description

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
