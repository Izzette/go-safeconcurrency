package pool

import (
	"context"
	"errors"
	"fmt"

	"github.com/Izzette/go-safeconcurrency/api/results"
	"github.com/Izzette/go-safeconcurrency/api/types"
)

// Submit is a helper function to submit a [types.Task] to a [types.Pool] and wait for the result.
func Submit[PoolResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[PoolResourceT],
	task types.Task[PoolResourceT, ValueT],
) (ValueT, error) {
	var zero ValueT

	// Wrap the task in a ValuelessTask to be able to submit it to the pool.
	valuelessTask, taskResults := TaskWrapper[PoolResourceT, ValueT](task)

	// Submit the task to the pool.
	// If context is canceled before the task is published it will instead return an error.
	if err := pool.Submit(ctx, valuelessTask); err != nil {
		//nolint:wrapcheck
		return zero, err
	}

	// Wait for the result or context cancellation, whichever comes first.
	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return zero, ctx.Err()
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
//
// # Callback
//
// If the callback produces an error, the task will be canceled and the error will be returned.
// If the special error [results.Stop] is be returned from the callback the task context will be canceled, the results
// channel drained, and the callback will not be called again.
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
// If you need more control, use [WrapMultiResultTaskBuffered] and call [types.Pool.Submit] directly.
func SubmitMultiResultBuffered[PoolResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[PoolResourceT],
	task types.MultiResultTask[PoolResourceT, ValueT],
	buffer uint,
	callback types.TaskCallback[ValueT],
) error {
	// Wrap the task in a ValuelessTask to be able to submit it to the pool.
	valuelessTask, taskResults := WrapMultiResultTaskBuffered[PoolResourceT](task, buffer)

	// Only drain the results channel if the task was submitted successfully.
	// We should always have already drained the results channel below, but this is a
	// precaution to avoid deadlocks in the case a panic occurs and recovered by the caller.
	submitted := false
	defer func() {
		if submitted {
			_ = taskResults.Drain()
		}
	}()

	// Ensure the context is canceled when the function returns, before the deference to drain.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	// Submit the task to the pool.
	// If context is canceled before the task is published it will instead return an error.
	if err := pool.Submit(ctx, valuelessTask); err != nil {
		//nolint:wrapcheck
		return err
	}
	submitted = true

	// Call the callback for each result as it is produced.
	callbackErr := callbackTaskResults(ctx, taskResults.Results(), callback)
	if callbackErr != nil {
		// The callback an error, which means we should cancel the task and stop processing results.
		cancel(fmt.Errorf("callback error %w caused %w", callbackErr, context.Canceled))
	}
	taskErr := taskResults.Drain()

	// The special error results.Stop ought not to be returned.
	if err := errors.Join(callbackErr, taskErr); !errors.Is(err, results.Stop) {
		return err
	}

	return nil
}

// SubmitMultiResult is a helper function to submit a [types.MultiResultTask] to a [types.Pool] and runs the
// callback for each result as it is produced.
// It is equivalent to calling [SubmitMultiResultBuffered] with a buffer size of 1.
// If you would like results buffering, use [SubmitMultiResultBuffered] instead.
// See [SubmitMultiResultBuffered] for more details.
func SubmitMultiResult[PoolResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[PoolResourceT],
	task types.MultiResultTask[PoolResourceT, ValueT],
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
func SubmitMultiResultCollectAll[PoolResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[PoolResourceT],
	task types.MultiResultTask[PoolResourceT, ValueT],
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

// callbackTaskResults consumes the results channel and calls the callback for each result.
func callbackTaskResults[ValueT any](
	ctx context.Context, resultsChan <-chan ValueT, callback types.TaskCallback[ValueT],
) error {
	for {
		select {
		case <-ctx.Done():
			//nolint:wrapcheck
			return ctx.Err()
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
