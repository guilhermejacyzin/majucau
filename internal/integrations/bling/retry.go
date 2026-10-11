package bling

import (
	"context"
	"errors"
	"time"
)

// withBlingReadRetry is shared by replay-safe GET requests. OAuth token
// exchanges are intentionally excluded because those POSTs may consume a
// one-time authorization code or refresh token.
func withBlingReadRetry[T any](
	ctx context.Context,
	attempts int,
	wait func(context.Context, time.Duration) error,
	delay func(int, time.Duration) time.Duration,
	request func() (T, error),
) (T, error) {
	if attempts < 1 {
		attempts = 1
	}
	if wait == nil {
		wait = waitBlingRetry
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		value, err := request()
		if err == nil {
			return value, nil
		}
		if attempt == attempts || !retryableBlingError(err) {
			var zero T
			return zero, err
		}
		var apiErr *BlingAPIError
		var retryAfter time.Duration
		if errors.As(err, &apiErr) {
			retryAfter = apiErr.RetryAfter
		}
		if err := wait(ctx, delay(attempt, retryAfter)); err != nil {
			var zero T
			return zero, err
		}
	}
	var zero T
	return zero, ErrBlingAPIUnavailable
}
