## Examples

Here are some more practical examples demonstrating how to use SafeConcurrency's
key features.

#### Basic Generator Usage

**Files**: `generator.go` and `generator_test.go`

Demonstrates creating a generator that produces a sequence of integers.

_Here's an abridged version of the example_:
```go
// Example_generator shows generator usage
func Example_generator() {
    w := generator.NewGeneratorBuffered[int](&IteratorJob{EndExclusive: 10}, 1)
    ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
    w.Start(ctx)

    for value := range w.Results() {
        fmt.Println(value)
    }
}
```

Run with:
```bash
go test -v -run Example_generator
```

#### Simple Worker Pool

**File**: `easypool_test.go`

Shows basic worker pool usage with a shared resource.

_Here's an abridged version of the example_:
```go
p := workpool.New[int](42, 1)
defer p.Close()
p.Start()

workpool.SubmitFunc[int](context.Background(), p, func(_ context.Context, n int) error {
    fmt.Printf("value: %d\n", n)
    return nil
})
```

Run with:
```bash
go test -v -run Example_easyPool
```

#### HTTP Client Pool

**Files**: `httppool.go`, `httppool_test.go`

Implements a concurrent HTTP client pool handling multiple requests.

_Here's an abridged of the key method_:
```go
// Get makes concurrent HTTP requests
func (p *HttpPool) Get(ctx context.Context, url string) (*http.Response, error) {
    task := &HttpTask{Method: "GET", URL: url}
    return workpool.Submit[*http.Client, *http.Response](ctx, p.Pool, task)
}
```

Run with:
```bash
go test -v -run Example_hTTPPool
```

#### Heterogeneous Task Pool

**Files**: `sharedpool.go` and `sharedpool_test.go`

Demonstrates a pool handling multiple task types with different result patterns.

_Here's an abridged version of the example_:
```go
result, err := workpool.Submit[any, int](ctx, pool, &IntTask{42}) // Single-result task

// Submit a streaming task
err = workpool.SubmitStreaming[any, string](ctx, pool, task, func(ctx context.Context, result string) error {
    // Runs callback for each result
    fmt.Println(result)
    return nil
})

// Submit a function task
err = workpool.SubmitFunc[any](ctx, pool, func(ctx context.Context) error {
    // Runs in the worker pool
    fmt.Println("Function task executed")
    return nil
})
```

Run with:
```bash
go test -v -run Example_sharedPool
```

### Running Examples and Expected Outputs

If you're not familiar with Go's runnable examples, they are a way to include
provide executable code snippets in documentation.
Each example contains an `// Output:` comment or `// Unordered output:` comment
with the expected output.
The Go test tool can run these examples and check their output, if the output
matches the expected output, the test passes.

Here's how you can run all examples as Go tests:

```bash
# Run all examples
go test -v ./examples
```

On
[pkg.go.dev/github.com/Izzette/go-safeconcurrency/examples](https://pkg.go.dev/github.com/Izzette/go-safeconcurrency/examples#pkg-overview)
you can find the individual examples, their expected outputs, and more.
