package bling

import (
	"context"
	"errors"
	"fmt"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/httpretry"
)

// withBlingReadRetry is shared by replay-safe GET requests. OAuth token
// exchanges are intentionally excluded because those POSTs may consume a
// one-time authorization code or refresh token.
func withBlingReadRetry[T any](
	ctx context.Context,
	method string,
	attempts int,
	wait func(context.Context, time.Duration) error,
	random func() float64,
	request func() (T, error),
) (T, error) {
	value, err := httpretry.Do(ctx, httpretry.Policy{
		Method:      method,
		MaxAttempts: attempts,
		Retryable:   retryableBlingError,
		RetryAfter: func(err error) time.Duration {
			var apiErr *BlingAPIError
			if errors.As(err, &apiErr) {
				return apiErr.RetryAfter
			}
			return 0
		},
		Wait:   wait,
		Random: random,
	}, request)
	if errors.Is(err, httpretry.ErrUnsafeMethod) {
		var zero T
		return zero, fmt.Errorf("%w: %v", ErrBlingAPIConfiguration, err)
	}
	return value, err
}
