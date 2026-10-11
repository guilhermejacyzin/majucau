package httpretry

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestDoRetriesReplaySafeGETWithinSharedLimit(t *testing.T) {
	transient := errors.New("temporary")
	calls := 0
	var waits []time.Duration
	value, err := Do(context.Background(), Policy{
		Method:    http.MethodGet,
		Retryable: func(err error) bool { return errors.Is(err, transient) },
		Wait: func(_ context.Context, delay time.Duration) error {
			waits = append(waits, delay)
			return nil
		},
		Random: func() float64 { return 0.5 },
	}, func() (string, error) {
		calls++
		if calls < 3 {
			return "", transient
		}
		return "ok", nil
	})
	if err != nil || value != "ok" {
		t.Fatalf("Do() = %q, %v; want ok, nil", value, err)
	}
	if calls != DefaultMaxAttempts {
		t.Fatalf("calls = %d, want %d", calls, DefaultMaxAttempts)
	}
	if len(waits) != 2 || waits[0] != 300*time.Millisecond || waits[1] != 550*time.Millisecond {
		t.Fatalf("waits = %v, want 300ms and 550ms", waits)
	}
}

func TestDoCapsAttemptsAndRetryAfter(t *testing.T) {
	transient := errors.New("rate limited")
	calls := 0
	var waits []time.Duration
	_, err := Do(context.Background(), Policy{
		Method:      http.MethodGet,
		MaxAttempts: 10,
		Retryable:   func(error) bool { return true },
		RetryAfter:  func(error) time.Duration { return time.Minute },
		Wait: func(_ context.Context, delay time.Duration) error {
			waits = append(waits, delay)
			return nil
		},
	}, func() (int, error) {
		calls++
		return 0, transient
	})
	if !errors.Is(err, transient) {
		t.Fatalf("Do() error = %v, want last transient error", err)
	}
	if calls != DefaultMaxAttempts || len(waits) != DefaultMaxAttempts-1 {
		t.Fatalf("calls=%d waits=%d; want %d attempts and %d waits", calls, len(waits), DefaultMaxAttempts, DefaultMaxAttempts-1)
	}
	for _, wait := range waits {
		if wait != MaxRetryAfter {
			t.Errorf("wait = %v, want capped %v", wait, MaxRetryAfter)
		}
	}
}

func TestDoRejectsStateChangingMethodsBeforeRequest(t *testing.T) {
	calls := 0
	_, err := Do(context.Background(), Policy{Method: http.MethodPost}, func() (string, error) {
		calls++
		return "", errors.New("must not run")
	})
	if !errors.Is(err, ErrUnsafeMethod) {
		t.Fatalf("Do() error = %v, want ErrUnsafeMethod", err)
	}
	if calls != 0 {
		t.Fatalf("request called %d times, want 0", calls)
	}
}

func TestDoDoesNotRetryPermanentError(t *testing.T) {
	permanent := errors.New("unauthorized")
	calls, waits := 0, 0
	_, err := Do(context.Background(), Policy{
		Method:    http.MethodGet,
		Retryable: func(error) bool { return false },
		Wait: func(context.Context, time.Duration) error {
			waits++
			return nil
		},
	}, func() (int, error) {
		calls++
		return 0, permanent
	})
	if !errors.Is(err, permanent) || calls != 1 || waits != 0 {
		t.Fatalf("Do() = %v, calls=%d waits=%d; want permanent error, one call, no waits", err, calls, waits)
	}
}

func TestDoStopsForCancelledContextBeforeRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	_, err := Do(ctx, Policy{Method: http.MethodGet}, func() (int, error) {
		calls++
		return 1, nil
	})
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatalf("Do() = %v, calls=%d; want cancellation without request", err, calls)
	}
}

func TestDoStopsWhenWaitIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	transient := errors.New("temporary")
	calls := 0
	_, err := Do(ctx, Policy{
		Method:    http.MethodGet,
		Retryable: func(error) bool { return true },
		Wait: func(context.Context, time.Duration) error {
			cancel()
			return context.Canceled
		},
	}, func() (int, error) {
		calls++
		return 0, transient
	})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("Do() = %v, calls=%d; want cancellation after one request", err, calls)
	}
}

func TestParseRetryAfterSupportsSecondsAndHTTPDateWithLimit(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	if got := ParseRetryAfter("7", now); got != 7*time.Second {
		t.Fatalf("seconds = %v, want 7s", got)
	}
	if got := ParseRetryAfter(now.Add(15*time.Second).Format(http.TimeFormat), now); got != 15*time.Second {
		t.Fatalf("HTTP date = %v, want 15s", got)
	}
	if got := ParseRetryAfter("999999999999999999999999", now); got != MaxRetryAfter {
		t.Fatalf("oversized decimal = %v, want %v", got, MaxRetryAfter)
	}
	if got := ParseRetryAfter("not-a-date", now); got != 0 {
		t.Fatalf("invalid value = %v, want 0", got)
	}
}

func TestDoRejectsMissingContextOrRequest(t *testing.T) {
	if _, err := Do[string](nil, Policy{Method: http.MethodGet}, func() (string, error) { return "", nil }); !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("nil context error = %v, want ErrInvalidPolicy", err)
	}
	if _, err := Do[string](context.Background(), Policy{Method: http.MethodGet}, nil); !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("nil request error = %v, want ErrInvalidPolicy", err)
	}
}
