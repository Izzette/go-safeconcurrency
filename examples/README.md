# SafeConcurrency Examples

## Basic Generator

`generator/example.go` demonstrates:
- Creating a buffered generator
- Context deadline handling
- Safe result channel draining

## HTTP Worker Pool

`poolhttp/example.go` shows:
- Creating a pool of HTTP workers
- Handling concurrent requests
- Graceful error handling
- Response processing

## Shared Resource Pool

`poolshared/example.go` illustrates:
- Heterogeneous task types in a pool
- Mixing result-producing and void tasks
- Context cancellation handling
- Shared resource usage patterns

---

To run examples:
```bash
go run examples/<example-dir>/example.go
```
