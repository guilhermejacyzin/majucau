// Package httpretry provides the shared, bounded retry policy for HTTP calls
// whose replay safety has been confirmed by the owning integration.
package httpretry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	DefaultMaxAttempts = 3
	DefaultBaseDelay   = 250 * time.Millisecond
	DefaultJitter      = 100 * time.Millisecond
	MaxRetryAfter      = 30 * time.Second
)

var (
	ErrUnsafeMethod  = errors.New("automatic retry is restricted to replay-safe GET requests")
	ErrInvalidPolicy = errors.New("HTTP retry policy is invalid")
)

type retryCounterKey struct{}

// Counter records replay-safe HTTP retries for one integration sync batch.
// It is intentionally separate from request/page counts and contains no
// provider response data.
type Counter struct {
	retries atomic.Int32
}

func NewCounter() *Counter {
	return &Counter{}
}

// WithCounter attaches a retry counter to an operation context. The HTTP
// retry helper increments it only when a retry request is actually issued.
func WithCounter(ctx context.Context, counter *Counter) context.Context {
	if ctx == nil || counter == nil {
		return ctx
	}
	return context.WithValue(ctx, retryCounterKey{}, counter)
}

func (c *Counter) Count() int32 {
	if c == nil {
		return 0
	}
	return c.retries.Load()
}

func countRetry(ctx context.Context) {
	if ctx == nil {
		return
	}
	if counter, ok := ctx.Value(retryCounterKey{}).(*Counter); ok && counter != nil {
		counter.retries.Add(1)
	}
}

// Policy classifies provider-specific transient failures while this package
// enforces the shared limits, backoff, cancellation, and GET-only default.
type Policy struct {
	Method      string
	MaxAttempts int
	Retryable   func(error) bool
	RetryAfter  func(error) time.Duration
	Wait        func(context.Context, time.Duration) error
	Random      func() float64
}

// Do retries a replay-safe GET request at most three times total. Callers must
// classify only temporary transport errors, 408, 429, and 5xx responses as
// retryable. State-changing requests are rejected; supporting one requires a
// separately reviewed idempotency contract and must not use this helper.
func Do[T any](ctx context.Context, policy Policy, request func() (T, error)) (T, error) {
	var zero T
	if ctx == nil || request == nil {
		return zero, ErrInvalidPolicy
	}
	if policy.Method != http.MethodGet {
		return zero, fmt.Errorf("%w: %s", ErrUnsafeMethod, policy.Method)
	}

	attempts := policy.MaxAttempts
	if attempts < 1 {
		attempts = DefaultMaxAttempts
	} else if attempts > DefaultMaxAttempts {
		attempts = DefaultMaxAttempts
	}

	wait := policy.Wait
	if wait == nil {
		wait = Wait
	}
	random := policy.Random
	if random == nil {
		random = rand.Float64
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		if attempt > 1 {
			countRetry(ctx)
		}
		value, err := request()
		if err == nil {
			return value, nil
		}
		if attempt == attempts || policy.Retryable == nil || !policy.Retryable(err) {
			return zero, err
		}

		var retryAfter time.Duration
		if policy.RetryAfter != nil {
			retryAfter = policy.RetryAfter(err)
		}
		if err := wait(ctx, delay(attempt, retryAfter, random)); err != nil {
			return zero, err
		}
	}
	return zero, errors.New("HTTP retry attempts exhausted unexpectedly")
}

// Wait pauses for delay unless the caller cancels the operation first.
func Wait(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// ParseRetryAfter parses delta-seconds or an HTTP date and caps the result at
// MaxRetryAfter. Invalid or expired values return zero so the caller can use
// its normal backoff.
func ParseRetryAfter(value string, now time.Time) time.Duration {
	trimmed := strings.TrimSpace(value)
	seconds, err := strconv.ParseInt(trimmed, 10, 64)
	if err == nil {
		if seconds < 0 {
			return 0
		}
		maxSeconds := int64(MaxRetryAfter / time.Second)
		if seconds >= maxSeconds {
			return MaxRetryAfter
		}
		return time.Duration(seconds) * time.Second
	}
	if isDecimal(trimmed) {
		return MaxRetryAfter
	}

	when, err := http.ParseTime(trimmed)
	if err != nil || !when.After(now) {
		return 0
	}
	delay := when.Sub(now)
	if delay > MaxRetryAfter {
		return MaxRetryAfter
	}
	return delay
}

func delay(attempt int, retryAfter time.Duration, random func() float64) time.Duration {
	if retryAfter > 0 {
		if retryAfter > MaxRetryAfter {
			return MaxRetryAfter
		}
		return retryAfter
	}

	backoff := DefaultBaseDelay
	for step := 1; step < attempt; step++ {
		backoff *= 2
	}
	jitter := random()
	if math.IsNaN(jitter) || jitter < 0 {
		jitter = 0
	} else if jitter > 1 {
		jitter = 1
	}
	return backoff + time.Duration(jitter*float64(DefaultJitter))
}

func isDecimal(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
