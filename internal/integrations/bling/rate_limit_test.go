package bling

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type rateLimitTestDoer struct {
	calls int
}

func (d *rateLimitTestDoer) Do(*http.Request) (*http.Response, error) {
	d.calls++
	return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
}

type concurrentRateLimitTestDoer struct {
	mu      sync.Mutex
	now     func() time.Time
	started []time.Time
}

type blockingRateLimitTestDoer struct {
	mu      sync.Mutex
	calls   int
	started chan struct{}
	release chan struct{}
}

func (d *blockingRateLimitTestDoer) Do(req *http.Request) (*http.Response, error) {
	d.mu.Lock()
	d.calls++
	if d.calls == 1 {
		close(d.started)
	}
	d.mu.Unlock()
	select {
	case <-d.release:
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}
}

func (d *concurrentRateLimitTestDoer) Do(*http.Request) (*http.Response, error) {
	d.mu.Lock()
	d.started = append(d.started, d.now())
	d.mu.Unlock()
	return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
}

func newRateLimitTestPacer(current *time.Time, waits *[]time.Duration) *blingRequestPacer {
	pacer := newBlingRequestPacer()
	pacer.now = func() time.Time { return *current }
	pacer.sleep = func(_ context.Context, delay time.Duration) error {
		*waits = append(*waits, delay)
		*current = current.Add(delay)
		return nil
	}
	return pacer
}

func makeRateLimitTestRequest(t *testing.T, method, target string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(method, target, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestRateLimitedBlingHTTPDoerPacesRequests(t *testing.T) {
	current := time.Unix(1_800_000_000, 0)
	var waits []time.Duration
	underlying := &rateLimitTestDoer{}
	doer := &rateLimitedBlingHTTPDoer{next: underlying, pacer: newRateLimitTestPacer(&current, &waits)}

	for range 2 {
		response, err := doer.Do(makeRateLimitTestRequest(t, http.MethodGet, "https://api.bling.com.br/Api/v3/contas/receber"))
		if err != nil {
			t.Fatalf("Do() error = %v", err)
		}
		_ = response.Body.Close()
	}

	if underlying.calls != 2 {
		t.Fatalf("underlying calls = %d, want 2", underlying.calls)
	}
	if len(waits) != 1 || waits[0] != defaultBlingRequestInterval {
		t.Fatalf("waits = %v, want one %s wait", waits, defaultBlingRequestInterval)
	}
}

func TestRateLimitedBlingHTTPDoerSharesOAuthAndAPIIntervals(t *testing.T) {
	current := time.Unix(1_800_000_000, 0)
	var waits []time.Duration
	underlying := &rateLimitTestDoer{}
	doer := &rateLimitedBlingHTTPDoer{next: underlying, pacer: newRateLimitTestPacer(&current, &waits)}

	requests := []*http.Request{
		makeRateLimitTestRequest(t, http.MethodPost, "https://api.bling.com.br/Api/v3/oauth/token"),
		makeRateLimitTestRequest(t, http.MethodGet, "https://api.bling.com.br/Api/v3/contas/receber"),
		makeRateLimitTestRequest(t, http.MethodPost, "https://api.bling.com.br/Api/v3/oauth/token"),
	}
	for _, request := range requests {
		response, err := doer.Do(request)
		if err != nil {
			t.Fatalf("Do() error = %v", err)
		}
		_ = response.Body.Close()
	}

	if underlying.calls != len(requests) {
		t.Fatalf("underlying calls = %d, want %d", underlying.calls, len(requests))
	}
	want := []time.Duration{defaultBlingRequestInterval, defaultBlingTokenRequestInterval - defaultBlingRequestInterval}
	if len(waits) != len(want) {
		t.Fatalf("waits = %v, want %v", waits, want)
	}
	for i := range want {
		if waits[i] != want[i] {
			t.Fatalf("waits = %v, want %v", waits, want)
		}
	}
}

func TestRateLimitedBlingHTTPDoerSpacesConcurrentRequests(t *testing.T) {
	current := time.Unix(1_800_000_000, 0)
	var waits []time.Duration
	pacer := newRateLimitTestPacer(&current, &waits)
	underlying := &concurrentRateLimitTestDoer{now: func() time.Time { return current }}
	doer := &rateLimitedBlingHTTPDoer{next: underlying, pacer: pacer}
	const callers = 8
	var wg sync.WaitGroup
	errCh := make(chan error, callers)
	requests := make([]*http.Request, 0, callers)
	for range callers {
		requests = append(requests, makeRateLimitTestRequest(t, http.MethodGet, "https://api.bling.com.br/Api/v3/contas/receber"))
	}
	for _, request := range requests {
		wg.Add(1)
		go func(request *http.Request) {
			defer wg.Done()
			response, err := doer.Do(request)
			if err != nil {
				errCh <- err
				return
			}
			_ = response.Body.Close()
		}(request)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("Do() error = %v", err)
	}

	underlying.mu.Lock()
	defer underlying.mu.Unlock()
	if len(underlying.started) != callers {
		t.Fatalf("transport starts = %d, want %d", len(underlying.started), callers)
	}
	for i := 1; i < len(underlying.started); i++ {
		if gap := underlying.started[i].Sub(underlying.started[i-1]); gap < defaultBlingRequestInterval {
			t.Fatalf("concurrent requests started %s apart; want at least %s", gap, defaultBlingRequestInterval)
		}
	}
}

func TestBlingRequestPacerStopsWhenContextIsCancelledDuringWait(t *testing.T) {
	current := time.Unix(1_800_000_000, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var waits []time.Duration
	pacer := newRateLimitTestPacer(&current, &waits)
	underlying := &rateLimitTestDoer{}
	doer := &rateLimitedBlingHTTPDoer{next: underlying, pacer: pacer}

	if _, err := doer.Do(makeRateLimitTestRequest(t, http.MethodGet, "https://api.bling.com.br/Api/v3/contas/receber")); err != nil {
		t.Fatalf("first Do() error = %v", err)
	}
	pacer.sleep = func(waitCtx context.Context, _ time.Duration) error {
		cancel()
		<-waitCtx.Done()
		return waitCtx.Err()
	}
	request := makeRateLimitTestRequest(t, http.MethodGet, "https://api.bling.com.br/Api/v3/contas/receber")
	request = request.WithContext(ctx)
	if _, err := doer.Do(request); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Do() error = %v, want context.Canceled", err)
	}
	if underlying.calls != 1 {
		t.Fatalf("underlying calls = %d, want canceled request not to reach transport", underlying.calls)
	}
}

func TestBlingRequestPacerCancelsCallWaitingForBusySlot(t *testing.T) {
	underlying := &blockingRateLimitTestDoer{started: make(chan struct{}), release: make(chan struct{})}
	doer := &rateLimitedBlingHTTPDoer{next: underlying, pacer: newBlingRequestPacer()}
	firstResult := make(chan error, 1)
	firstRequest := makeRateLimitTestRequest(t, http.MethodGet, "https://api.bling.com.br/Api/v3/contas/receber")
	go func() {
		response, err := doer.Do(firstRequest)
		if response != nil {
			_ = response.Body.Close()
		}
		firstResult <- err
	}()
	select {
	case <-underlying.started:
	case <-time.After(time.Second):
		t.Fatal("first request did not reach transport")
	}

	ctx, cancel := context.WithCancel(context.Background())
	request := makeRateLimitTestRequest(t, http.MethodGet, "https://api.bling.com.br/Api/v3/contas/receber").WithContext(ctx)
	secondResult := make(chan error, 1)
	secondAttempting := make(chan struct{})
	go func() {
		close(secondAttempting)
		response, err := doer.Do(request)
		if response != nil {
			_ = response.Body.Close()
		}
		secondResult <- err
	}()
	<-secondAttempting
	time.Sleep(10 * time.Millisecond)
	cancel()
	var secondErr error
	secondReturned := false
	select {
	case secondErr = <-secondResult:
		secondReturned = true
	case <-time.After(100 * time.Millisecond):
	}

	close(underlying.release)
	if !secondReturned {
		select {
		case secondErr = <-secondResult:
		case <-time.After(time.Second):
			t.Fatal("waiting request did not finish after releasing the active request")
		}
		t.Error("canceled request remained blocked behind the active request")
	}
	if !errors.Is(secondErr, context.Canceled) {
		t.Errorf("waiting request error = %v, want context.Canceled", secondErr)
	}
	select {
	case err := <-firstResult:
		if err != nil {
			t.Fatalf("first request error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("first request did not finish")
	}
	underlying.mu.Lock()
	defer underlying.mu.Unlock()
	if underlying.calls != 1 {
		t.Fatalf("transport calls = %d, want canceled waiting request not to reach transport", underlying.calls)
	}
}

func TestWithBlingRequestPacingReusesSharedLimiter(t *testing.T) {
	underlying := &rateLimitTestDoer{}
	limited := withBlingRequestPacing(underlying)
	if again := withBlingRequestPacing(limited); again != limited {
		t.Fatal("request pacing wrapper was nested instead of reused")
	}
}
