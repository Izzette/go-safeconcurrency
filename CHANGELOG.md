# Changelog

## [0.5.0](https://github.com/Izzette/go-safeconcurrency/compare/v0.4.0...v0.5.0) (2025-05-01)


### Features

* **eventloop:** improve snapshot interface for eventloop ([688be86](https://github.com/Izzette/go-safeconcurrency/commit/688be86b4f256f78c659d98653e31436d3a417c5))
* **eventloop:** improve snapshot interface for eventloop ([eb7bfd2](https://github.com/Izzette/go-safeconcurrency/commit/eb7bfd2d72bdae51b9a74a051e744d77f9a85332))

## [0.4.0](https://github.com/Izzette/go-safeconcurrency/compare/v0.3.2...v0.4.0) (2025-04-27)


### Features

* **eventloop:** Add event loops ([777aa4d](https://github.com/Izzette/go-safeconcurrency/commit/777aa4d0af24544c063d6035c46efbcf86cdb02e))
  * Interface type `types.EventLoop`, `types.Event`, and
    `types.StateSnapshot`
  * Type alias `types.GenerationID`
  * Factory functions for `types.EventLoop`: `eventloop.New`,
    `eventloop.NewBuffered`
  * Helpers `eventloop.WaitForGeneration` and `eventloop.SendAndWait`
  * Helper to watch to eventloop state snapshots `eventloop.WatchState`
    and `eventloop.WatchStateFunc`
  * Basic example in `examples/eventloop_test.go` and
    `examples/ecommerce.go`
  * Advanced example with fake e-commerce server in
    `examples/ecommerceeventloop_test.go` and the `examples/ecommerce`
    module
  * Updated documentation
* Remove redundant "task" from workpool/task ([467bccc](https://github.com/Izzette/go-safeconcurrency/commit/467bccc5a5a0867cb5c172595022aafe417d13ac))
  * Rename `task.WrapTask` → `task.Wrap`
  * Rename `task.WrapStreamingTask` → `task.WrapStreaming`
  * Rename `task.WrapTaskFunc` → `task.WrapFunc`
* correct typo in GoDoc comment for workpool/util.go ([b666099](https://github.com/Izzette/go-safeconcurrency/commit/b666099d9a9ef103f645499f54de39183fd606fe))
* move `types.ResultCallback` to the `workpool` package ([bbc5f24](https://github.com/Izzette/go-safeconcurrency/commit/bbc5f242c6600020068de554bfa185e38cf9181c))
* remove context from types.ValuelessTask ([29d9dc3](https://github.com/Izzette/go-safeconcurrency/commit/29d9dc32ce642ff0d739c74a98df0fdf157fd5eb))
* rename `generator.NewGenerator*` → `generator.New*` ([4845164](https://github.com/Izzette/go-safeconcurrency/commit/48451649a670be4acf7b120867fe416ff76f6c0b))
* Rename many APIs ([d40498e](https://github.com/Izzette/go-safeconcurrency/commit/d40498ea96cb8abdc43383c265a3f7969441c830))
  * `types.Runner` → `types.Producer`
  * `types.Handle` → `types.Emitter`
    * `.Publish` → `.Emit`
  * `types.Pool` → `types.WorkerPool`
  * `types.MultiResultTask` → `types.StreamingTask`
    * `workpool.SubmitMultiResult*` → `workpool.SubmitStreaming*`:
      These functions no longer return joined errors, and prefer returning
      the callback error.
    * `task.WrapMultiResultTaskBuffered` → `task.WrapStreamingTask`
    * Removed `task.WrapMultiResultTask`
  * `types.TaskCallback` → `types.ResultCallback`
  * `workpool.NewPool*` → `workpool.New*`

## [0.3.2](https://github.com/Izzette/go-safeconcurrency/compare/v0.3.1...v0.3.2) (2025-04-22)


### Features

* add package overview go-doc comment ([691f2ae](https://github.com/Izzette/go-safeconcurrency/commit/691f2ae890bc3b243b717e1d09b4116ef85cfdc7))
* de-clutter GoDoc for work pools ([0edd18f](https://github.com/Izzette/go-safeconcurrency/commit/0edd18f1ebc17fea90a5d294b1e4e0237e97e046))
* move task wrappers out of workpool and into workpool/task ([bbf3849](https://github.com/Izzette/go-safeconcurrency/commit/bbf384979adaf47e11dbce49e62dc6ce5d90441e))
* rename PoolResourceT to ResourceT for brevity ([eb847c8](https://github.com/Izzette/go-safeconcurrency/commit/eb847c880b799d5df06358e28f241e30d2dba4f0))

## [0.3.1](https://github.com/Izzette/go-safeconcurrency/compare/v0.3.0...v0.3.1) (2025-04-22)


### Features

* rename packages ([eb8be36](https://github.com/Izzette/go-safeconcurrency/commit/eb8be3607ecc2069c68002f711e20ba70bc5c6dc))
  * Base package renamed to `safeconcurrency` by adding a `package.go`,
    this only effects the display module name for the pkg.go.dev site.
  * Moved implementations down from `api/`.
  * Renamed `pool` to `workpool`.

## [0.3.0](https://github.com/Izzette/go-safeconcurrency/compare/v0.2.1...v0.3.0) (2025-04-21)


### Features

* add `pool.SubmitFunc` and improve examples ([9188e81](https://github.com/Izzette/go-safeconcurrency/commit/9188e8186dda06027aa3a92784ed44e6e4c62845))
  * Improve the examples using Go testing runnable examples
  * Add a `pool.SubmitFunc` to ease the simplest of cases (and the
    documented examples)

* Improve `SubmitMultiResult` using callbacks ([3a19ddc](https://github.com/Izzette/go-safeconcurrency/commit/3a19ddc507b162254face64bb833c0272d0a9973))
  * `SubmitMultiResult` now takes a callback, and with `SubmitMultiResultBuffered`
    it has a configurable buffer.
  * Upgrade to Go version 1.20+ for `errors.Join` and `context.WithCancelCause`.
  * Remove `types.TaskResult.Err()`, as `.Drain()` should always be called
    instead.

* use `context.Cause` everywhere ([3ec064a](https://github.com/Izzette/go-safeconcurrency/commit/3ec064ab44d4204dc5619456b8d6ab24dbf09a77))
  * Replace usage of `context.Context.Err()` with `context.Cause()`.
  * Add forbidigo linter to ensure `context.Context.Err` is never referenced.


## [0.2.1](https://github.com/Izzette/go-safeconcurrency/compare/v0.2.0...v0.2.1) (2025-04-20)


### Features

* Refactor stuff and add pool utilities ([4e57c22](https://github.com/Izzette/go-safeconcurrency/commit/4e57c22adf1107de244f02fd9e8987c92afefb18))
  * Generator API Overhaul:
    * `Generator.Results()` now returns `<-chan T` instead of `<-chan Result[T]`
    * Removed `Result[T]` interface and error recovery via `result.Get()`
    * Errors are now exclusively propagated via `Generator.Wait()`
    * Updated `Runner` interface to remove error publishing methods
  * Pool Utilities:
    * Added `Submit[PoolResourceT, ValueT]` helper for single-result task submission
    * Added `SubmitMultiResult[PoolResourceT, ValueT]` helper for multi-result task collection
    * Introduced `TaskResult[T]` interface with `Results()`, `Drain()`, and `Err()` methods
  * Resource Management:
    * Added `Handle.Close()` method for explicit resource cleanup in generators/pools

## [0.2.0](https://github.com/Izzette/go-safeconcurrency/compare/v0.1.2...v0.2.0) (2025-04-19)


### Features

* add worker pool ([d6417ee](https://github.com/Izzette/go-safeconcurrency/commit/d6417eeb7c06614be7ea6e08af3c9b0b62373cd5))

## [0.1.2](https://github.com/Izzette/go-safeconcurrency/compare/v0.1.1...v0.1.2) (2025-04-12)

### Bug Fixes

* update golang to 1.19 to allow for sync/atomic.Bool ([1946283](https://github.com/Izzette/go-safeconcurrency/commit/19462831e1bc61752d491d887078358b36716a75))
