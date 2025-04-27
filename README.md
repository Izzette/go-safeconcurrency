# SafeConcurrency for Go

<img align="left" width="250" height="250" alt="SafeConcurrency Logo" src="doc/assets/images/logo-500x500.png">

SafeConcurrency is a Go library designed to simplify the management of concurrent tasks, providing a safe and structured way to produce and consume results.
It enforces best practices for context propagation, error handling, and resource cleanup.

⚠️ **Warning**: This library is in early development and may not be suitable for production use yet.

The API will change frequently as we refine the design and functionality.
Expect new features and improvements in future releases, generators and work pools are just the beginning.

![Go Version](https://img.shields.io/badge/go-1.20-blue) [![Go Reference](https://pkg.go.dev/badge/github.com/Izzette/go-safeconcurrency.svg)](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency) [![Go Report Card](https://goreportcard.com/badge/github.com/Izzette/go-safeconcurrency)](https://goreportcard.com/report/github.com/Izzette/go-safeconcurrency) ![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Generator Pattern**: Safely produce values from concurrent operations via channel-based results
  - Adheres to Go best-practice: “[Do not communicate by sharing memory; instead, share memory by communicating](https://go.dev/blog/codelab-share)”
- **Context Integration**: Built-in support for context cancellation and deadlines
- **Error Handling**: Gracefully handle errors from concurrent operations
- **Concurrency-Safe**: All APIs are designed for concurrent use from different goroutines
  - Very few mutexes are used, instead synchronizing using channels and atomic operations
- **Flexible Buffering**: Configurable request and result channel buffering for different throughput, synchronization,
  and back-pressure needs
- **Worker Pools**: Operate in a pool of workers to manage shared resources across heterogeneous tasks called from
  different goroutines (perfect for API clients or database connections)
- **Event Loops**: Support for event loops for handling events in a sequential manner
   - Atomic state snapshots with generation tracking
   - Event hooks for monitoring and customization

### Planned Features

- **Parallel Mapping**: Support for parallel mapping of input values to output results
- **Pipeline Support**: Create pipelines of generators for complex workflows

## Usage

### Key Components

1. **Generators**
   Implement your concurrent logic by creating a type that satisfies `types.Producer`:
   ```go
   type MyProducer struct{}

   func (r *MyProducer) Run(ctx context.Context, h types.Emitter[Output]) error {
       // Your concurrent logic here
       h.Emit(ctx, value)
       return fatalErr
   }
   ```

   Create and manage concurrent execution:
   ```go
   gen := generator.New[Output](&MyProducer{})
   gen.Start(ctx)
   ```

   Consume results safely from the channel:
   ```go
   for val := range gen.Results() {
       // Handle value
   }
   ```

2. **Worker Pools**
   Implement your task logic by creating a type that satisfies `types.Task`:
   ```go
   type task struct{}

   func (t *task) Execute(ctx context.Context, resource ResourceType) (Output, error) {
       // Your task logic here
       return result, nil
   }
   ```

   Create and manage worker pools:
   ```go
   mypool := workpool.New[ResourceType](resource, concurrency)
   mypool.Start()
   defer mypool.Close()
   ```

   Submit tasks to the worker pool and receive results:
   ```go
   // Submit tasks
   val, err := workpool.Submit[ResourceType, Output](ctx, mypool, &task{})
   // Handle result
   ```

3. **Event Loops**
   Implement your event logic by creating a type that satisfies `types.Event`:
   ```go
   type RequestEvent struct {}

   func (e *RequestEvent) Dispatch(gen types.GenerationID, s *AppState) {
       fmt.Printf("Processing request #%d\n", gen)
       s.Requests++
   }
   ```

   Create a state type to hold your application state:
   ```go
   type AppState struct { Requests int }
   ```

   Create and manage the event loop:
   ```go
   el := eventloop.New[AppState](&AppState{})
   defer el.Close()
   el.Start()
   ```

   Send events to the loop:
   ```go
   gen, err := el.Send(ctx, &RequestEvent{})
   if err != nil {
       panic(err)
   }
   ```

   Wait for the event to be processed and get a snapshot of the state:
   ```go
   snap, err := eventloop.WaitFor(ctx, el, gen)
   if err != nil {
       panic(err)
   }
   fmt.Printf("Current requests: %d\n", snap.State().Requests)
   ```

See the [examples](examples) directory for more detailed examples, or interact
with them in the browser on
[pkg.go.dev](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency/examples).

## Documentation

Full API documentation is available on [GoDoc](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency).

- For types and interfaces, see [api/types](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency/api/types).
- For creating generators, see [generator](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency/generator).
- For creating worker pools and tasks, see [workpool](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency/workpool).
- For creating event loops, see [eventloop](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency/eventloop).
- Examples can be interacted with in the browser at [examples](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency/examples).

## Contributing

We welcome contributions! Please follow these guidelines:

1. Install pre-requisites:
   - Go 1.20 or later
   - Python 3.9 or later (for pre-commit)
   - pre-commit (https://pre-commit.com/)
   - Make (GNU Make recommended: https://www.gnu.org/software/make/)
   - Golangci-lint (https://golangci-lint.run/welcome/install/#local-installation):
     - ```shell
       go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
       ```

2. Set up development environment:
   ```shell
   # Install python virtual environment for pre-commit hooks
   pre-commit install
   ```

3. Ensure all tests pass:
   ```bash
   # Run all tests and linters
   make all
   ```

4. Add tests for new features, make sure to check the test coverage:
   ```bash
   # Run tests with coverage
   make test-unit
   ```
   Use `go tool cover -html tmp/coverage/cover.out` to view the coverage report in your browser.

5. Update documentation accordingly.
   Use Godoc comments for public types and functions: https://go.dev/blog/godoc

6. Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit titles.
   This is required for our automated release process, [Release Please](https://github.com/googleapis/release-please).

7. Open a pull request with a clear description of the changes and why they are needed.
   Include the CHANGELOG entry you would like to see in the release, it doesn't need to be perfect: we can refine it together.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
