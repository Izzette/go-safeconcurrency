package pool

import (
	"context"

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
		if err := taskResults.Err(); err != nil {
			//nolint:wrapcheck
			return zero, err
		}

		return result, nil
	}
}

// SubmitMultiResult is a helper function to submit a [types.MultiResultTask] to a [types.Pool] and wait for the
// results.
// It uses a buffer size of 1 for the results channel, to avoid blocking the worker while reading the results.
// The results slice is initialized with an expected capacity of 0, this may not be very efficient if the task produces
// a lot of results.
// If the task returns an error, the results slice returned will be nil.
//
// If you need more control, use [WrapMultiResultTaskBuffered] and call [types.Pool.Submit] directly.
// [SubmitMultiResult] useful for testing, debugging, and demonstration purposes where the performance difference is
// unimportant and the ability to consume the results as they are produced is not required.
func SubmitMultiResult[PoolResourceT any, ValueT any](
	ctx context.Context,
	pool types.Pool[PoolResourceT],
	task types.MultiResultTask[PoolResourceT, ValueT],
) ([]ValueT, error) {
	// Wrap the task in a ValuelessTask to be able to submit it to the pool.
	valuelessTask, taskResults := WrapMultiResultTaskBuffered[PoolResourceT](task, 1)

	// Submit the task to the pool.
	// If context is canceled before the task is published it will instead return an error.
	if err := pool.Submit(ctx, valuelessTask); err != nil {
		//nolint:wrapcheck
		return nil, err
	}
	// Only drain the results channel if the task was submitted successfully.
	// We should always have already drained the results channel below, but this is a
	// precaution to avoid deadlocks.
	defer func() {
		_ = taskResults.Drain()
	}()

	results, err := collectTaskResults(ctx, taskResults.Results())
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	} else if err := taskResults.Err(); err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	return results, nil
}

// collectTaskResults collects the results from the taskResults channel until it is closed or the context is canceled.
func collectTaskResults[ValueT any](ctx context.Context, resultsChan <-chan ValueT) ([]ValueT, error) {
	results := make([]ValueT, 0)
	for {
		select {
		case <-ctx.Done():
			//nolint:wrapcheck
			return nil, ctx.Err()
		case result, ok := <-resultsChan:
			if !ok {
				// The results channel is closed, we are done.
				return results, nil
			}
			results = append(results, result)
		}
	}
}
