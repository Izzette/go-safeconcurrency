package workpool

import (
	"context"
	"errors"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/safeconcurrencyerrors"
	"github.com/Izzette/go-safeconcurrency/api/types"
	"github.com/Izzette/go-safeconcurrency/workpool/task"
)

// Submit is a helper function to submit a [types.Task] to a [types.Pool] and wait for the result.
// The result and error returned from the task are returned to the caller.
//
// # Context
//
// The provided [context.Context] is passed to the task when it is executed in the [types.Pool].
// It is advisable to use [context.WithDeadline] or [context.WithTimeout] to limit the time the task is allowed to run
// and ensure the Pool can be shared with other tasks.
// Alternatively, using [context.WithCancel] and deferring a call to the context.CancelFunc will stop the task from
// blocking the Pool if the caller is no longer interested in the result.
//
// # Warning
//
// ⚠️You must never attempt to submit tasks to a pool which has been closed, this will result in a panic!
//
// # Other
//
// If you need more control, use [task.WrapTaskBuffered] and publish to the [types.Pool.Requests] channel directly.
func Submit[ResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[ResourceT],
	tsk types.Task[ResourceT, ValueT],
) (ValueT, error) {
	var zero ValueT

	// select is not deterministic, and may still send tasks even if the context has been canceled.
	if err := context.Cause(ctx); err != nil {
		//nolint:wrapcheck
		return zero, err
	}

	// Wrap the task in a ValuelessTask to be able to submit it to the pool.
	valuelessTask, taskResults := task.WrapTask[ResourceT, ValueT](ctx, tsk)

	// Submit the task to the pool.
	// If context is canceled before the task is published it should return an error.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return zero, context.Cause(ctx)
	case pool.Requests() <- valuelessTask:
	}

	// Wait for the result or context cancellation, whichever comes first.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return zero, context.Cause(ctx)
	case result := <-taskResults.Results():
		// Before using the (possibly nil) result, check if the task produced an error.
		// We must drain the results channel to ensure that the err is set correctly.
		if err := taskResults.Drain(); err != nil {
			//nolint:wrapcheck
			return zero, err
		}

		return result, nil
	}
}

// SubmitMultiResultBuffered is a helper function to submit a [types.MultiResultTask] to a [types.Pool] and runs the
// callback for each result as it is produced.
// The same advisories as for [Submit] about context cancellation apply.
//
// # Callback
//
// If the callback produces an error, the task will be canceled and the error will be returned.
// If the special error [safeconcurrencyerrors.Stop] is be returned from the callback the task context will be canceled,
// the results channel drained, and the callback will not be called again.
//
// # Error Handling
// As both the callback and the task may return an error, the errors will be joined and returned, therefore always
// make sure to use [errors.Is] and [errors.As] to check for errors returned from this function.
//
// # Buffering
//
// The buffer size of the results channel is specified by the buffer parameter.
// It is recommended to avoid using a buffer size of 0, as this will block the worker until the result is received.
//
// # Other
//
// If you need more control, use [task.WrapMultiResultTaskBuffered] and publish to the [types.Pool.Requests] channe
// directly.
func SubmitMultiResultBuffered[ResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[ResourceT],
	tsk types.MultiResultTask[ResourceT, ValueT],
	buffer uint,
	callback types.TaskCallback[ValueT],
) error {
	// select is not deterministic, and may still send tasks even if the context has been canceled.
	if err := context.Cause(ctx); err != nil {
		//nolint:wrapcheck
		return err
	}

	// Ensure the context is canceled when the function returns, before the deference to drain.
	ctx, cancel := context.WithCancelCause(ctx)
	// The cancel func _must_ be called to ensure that a leak of contexts/goroutines does not occur.
	defer cancel(context.Canceled)

	// Wrap the task in a ValuelessTask to be able to submit it to the pool.
	valuelessTask, taskResults := task.WrapMultiResultTaskBuffered[ResourceT](ctx, tsk, buffer)

	// Submit the task to the pool.
	// If context is canceled before the task is published it should return an error.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return context.Cause(ctx)
	case pool.Requests() <- valuelessTask:
	}

	// Only drain the results channel if the task was submitted successfully.
	// We should always have already drained the results channel below, but this is a
	// precaution to avoid deadlocks in the case a panic occurs and recovered by the caller.
	defer func() {
		// We already have a deference to cancel, however this should be done before the task results are drained.
		// Remember that defered statements are executed in last-defered first-executed order (like a stack).
		cancel(context.Canceled)
		// We do not care about the error returned by Drain, if the error has not been returned it is because we are
		// already panicking.
		_ = taskResults.Drain()
	}()

	// Call the callback for each result as it is produced.
	callbackErr := callbackTaskResults(ctx, taskResults.Results(), callback)
	if callbackErr != nil {
		// The callback an error, which means we should cancel the task and stop processing results.
		cancel(fmt.Errorf("callback error %w", callbackErr))
	}
	taskErr := taskResults.Drain()

	// The special error(s) safeconcurrencyerrors.Stop ought not to be returned.
	if errors.Is(callbackErr, safeconcurrencyerrors.Stop) {
		callbackErr = nil
	}
	if errors.Is(taskErr, safeconcurrencyerrors.Stop) {
		taskErr = nil
	}

	return errors.Join(callbackErr, taskErr)
}

// SubmitMultiResult is a helper function to submit a [types.MultiResultTask] to a [types.Pool] and runs the
// callback for each result as it is produced.
// It is equivalent to calling [SubmitMultiResultBuffered] with a buffer size of 1.
// If you would like results buffering, use [SubmitMultiResultBuffered] instead.
// See [SubmitMultiResultBuffered] for more details.
func SubmitMultiResult[ResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[ResourceT],
	task types.MultiResultTask[ResourceT, ValueT],
	callback types.TaskCallback[ValueT],
) error {
	return SubmitMultiResultBuffered(ctx, pool, task, 1, callback)
}

// SubmitMultiResultCollectAll is a helper function to submit a [types.MultiResultTask] to a [types.Pool] and collects
// all results to a slice.
// It uses a buffer size of 1 for the results channel to minimize blocking the worker while reading the results.
// The results slice is initialized with an expected capacity of 0, this may not be very efficient if the task produces
// a lot of results.
// If the task returns an error, the results slice returned will be nil.
// SubmitMultiResultCollectAll useful for testing, debugging, and demonstration purposes where the performance
// difference is unimportant and the ability to consume the results as they are produced is not required.
// You should most likely use [SubmitMultiResult] instead.
func SubmitMultiResultCollectAll[ResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[ResourceT],
	task types.MultiResultTask[ResourceT, ValueT],
) ([]ValueT, error) {
	results := make([]ValueT, 0)
	if err := SubmitMultiResult(ctx, pool, task, func(ctx context.Context, value ValueT) error {
		results = append(results, value)

		return nil
	}); err != nil {
		return nil, err
	}

	return results, nil
}

// SubmitFunc is a helper function to submit a [types.TaskFunc] to a [types.Pool].
// The same advisories as for [Submit] about context cancellation apply.
//
// # Error Handling
// The error returned from the task is returned to the caller.
//
// # Other
//
// If you need more control, use [task.WrapTaskFunc] and publish to the [types.Pool.Requests] channel directly.
func SubmitFunc[ResourceT any](
	ctx context.Context,
	pool types.Pool[ResourceT],
	taskFunc types.TaskFunc[ResourceT],
) error {
	// Wrap the task in a ValuelessTask to be able to submit it to the pool.
	valuelessTask, taskResults := task.WrapTaskFunc[ResourceT](ctx, taskFunc)

	// Submit the task to the pool.
	// If context is canceled before the task is published it should return an error.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return context.Cause(ctx)
	case pool.Requests() <- valuelessTask:
	}

	// Wait for the result or context cancellation, whichever comes first.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return context.Cause(ctx)
	case <-taskResults.Results():
		// Check for an error returned by the task.
		if err := taskResults.Drain(); err != nil {
			//nolint:wrapcheck
			return err
		}

		return nil
	}
}

// callbackTaskResults consumes the results channel and calls the callback for each result.
func callbackTaskResults[ValueT any](
	ctx context.Context, resultsChan <-chan ValueT, callback types.TaskCallback[ValueT],
) error {
	for {
		select {
		case <-ctx.Done():
			//nolint:wrapcheck
			return context.Cause(ctx)
		case result, ok := <-resultsChan:
			if !ok {
				// The results channel is closed, we are done.
				return nil
			} else if err := callback(ctx, result); err != nil {
				// The callback returned an error, we should stop processing results.
				//nolint:wrapcheck
				return err
			}
		}
	}
}
