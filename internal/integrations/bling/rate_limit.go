package bling

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

// Bling documents a shared account limit of three requests per second. A
// 400 ms minimum interval is intentionally conservative and leaves margin
// below that ceiling for clock and scheduling variation.
const defaultBlingRequestInterval = 400 * time.Millisecond

// The Bling integration map records an additional 20 requests per minute
// limit for the OAuth token endpoint. A 3.1 s interval leaves a small margin.
const defaultBlingTokenRequestInterval = 3100 * time.Millisecond

// blingRequestPacer serializes request starts for one adapter session. The
// one-slot queue remains held while the HTTP doer waits and starts its
// request, so two concurrent callers cannot reserve a slot and then reach the
// transport out of order. Waiting callers can still leave promptly on cancel.
// It tracks both account-wide and token-specific spacing.
type blingRequestPacer struct {
	slot          chan struct{}
	interval      time.Duration
	nextRequest   time.Time
	nextTokenCall time.Time
	now           func() time.Time
	sleep         func(context.Context, time.Duration) error
}

func newBlingRequestPacer() *blingRequestPacer {
	slot := make(chan struct{}, 1)
	slot <- struct{}{}
	return &blingRequestPacer{
		slot:     slot,
		interval: defaultBlingRequestInterval,
		now:      time.Now,
		sleep:    waitForBlingRequestInterval,
	}
}

type rateLimitedBlingHTTPDoer struct {
	next  HTTPDoer
	pacer *blingRequestPacer
}

func (d *rateLimitedBlingHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("Bling request is missing")
	}
	if d == nil || d.next == nil || d.pacer == nil {
		return nil, ErrBlingAPIConfiguration
	}
	if err := req.Context().Err(); err != nil {
		return nil, err
	}

	p := d.pacer
	if p.slot == nil {
		return nil, ErrBlingAPIConfiguration
	}
	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	case <-p.slot:
	}
	defer func() { p.slot <- struct{}{} }()
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	interval := p.interval
	if interval <= 0 {
		interval = defaultBlingRequestInterval
	}
	tokenRequest := isBlingTokenRequest(req)
	now := p.now
	if now == nil {
		now = time.Now
	}
	current := now()
	availableAt := p.nextRequest
	if tokenRequest && p.nextTokenCall.After(availableAt) {
		availableAt = p.nextTokenCall
	}
	if delay := availableAt.Sub(current); delay > 0 {
		sleep := p.sleep
		if sleep == nil {
			sleep = waitForBlingRequestInterval
		}
		if err := sleep(req.Context(), delay); err != nil {
			return nil, err
		}
	}
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	started := now()
	response, err := d.next.Do(req)
	// Even a transport failure consumes a request slot. Record from the time
	// the attempt started, while still holding the shared queue slot.
	p.nextRequest = started.Add(interval)
	if tokenRequest {
		p.nextTokenCall = started.Add(defaultBlingTokenRequestInterval)
	}
	return response, err
}

func waitForBlingRequestInterval(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isBlingTokenRequest(req *http.Request) bool {
	return req != nil && req.URL != nil && strings.HasSuffix(strings.TrimRight(req.URL.Path, "/"), "/oauth/token")
}

func withBlingRequestPacing(client HTTPDoer) HTTPDoer {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	if _, alreadyLimited := client.(*rateLimitedBlingHTTPDoer); alreadyLimited {
		return client
	}
	return &rateLimitedBlingHTTPDoer{next: client, pacer: newBlingRequestPacer()}
}
