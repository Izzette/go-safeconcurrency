# Changelog

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
