package bling

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestBlingReadRetryRejectsNonGETBeforeCallingRequest(t *testing.T) {
	calls := 0
	waits := 0
	_, err := withBlingReadRetry(
		context.Background(),
		http.MethodPost,
		3,
		func(context.Context, time.Duration) error {
			waits++
			return nil
		},
		func() float64 { return 0 },
		func() (string, error) {
			calls++
			return "", errors.New("temporary")
		},
	)
	if !errors.Is(err, ErrBlingAPIConfiguration) {
		t.Fatalf("non-GET retry error = %v, want configuration error", err)
	}
	if calls != 0 || waits != 0 {
		t.Fatalf("non-GET request executed %d calls and %d waits", calls, waits)
	}
}
