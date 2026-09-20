package bling

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestListReceivablesKeepsRawRecordsAndBuildsReadOnlyRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/Api/v3/contas/receber" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("authorization = %q", got)
		}
		if got := r.URL.Query().Get("pagina"); got != "2" {
			t.Fatalf("pagina = %q", got)
		}
		if got := r.URL.Query().Get("limite"); got != "2" {
			t.Fatalf("limite = %q", got)
		}
		if got := r.URL.Query().Get("dataRecebimentoInicial"); got != "2026-09-01" {
			t.Fatalf("dataRecebimentoInicial = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":123,"situacao":"pago"},{"id":124}],"pagination":{"page":2,"limit":2,"total":5}}`))
	}))
	defer server.Close()

	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListReceivables(context.Background(), ReceivablesFilter{Page: 2, Limit: 2, ReceivedDateFrom: "2026-09-01"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Page != 2 || page.Limit != 2 || page.Total == nil || *page.Total != 5 || !page.HasNext {
		t.Fatalf("unexpected pagination: %+v", page)
	}
	if len(page.Records) != 2 || !strings.Contains(string(page.Records[0]), `"id":123`) {
		t.Fatalf("records were not preserved: %+v", page.Records)
	}
}

func TestListPayablesUsesPayableEndpointAndPaymentFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Api/v3/contas/pagar" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("dataPagamentoInicial"); got != "2026-09-01" {
			t.Fatalf("dataPagamentoInicial = %q", got)
		}
		if got := r.URL.Query().Get("dataPagamentoFinal"); got != "2026-09-30" {
			t.Fatalf("dataPagamentoFinal = %q", got)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":77,"situacao":"pago"}],"pagination":{"page":1,"limit":1,"total":1}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListPayables(context.Background(), ReceivablesFilter{PaymentDateFrom: "2026-09-01", PaymentDateTo: "2026-09-30", Limit: 1})
	if err != nil || len(page.Records) != 1 || !strings.Contains(string(page.Records[0]), `"id":77`) {
		t.Fatalf("unexpected payable page: %+v, err=%v", page, err)
	}
}

func TestListReceivablesClassifiesSafeTransportErrors(t *testing.T) {
	for _, test := range []struct {
		name      string
		status    int
		header    string
		want      error
		wantRetry bool
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, want: ErrBlingAPIUnauthorized},
		{name: "rate limited", status: http.StatusTooManyRequests, header: "7", want: ErrBlingAPIRateLimited, wantRetry: true},
		{name: "server unavailable", status: http.StatusBadGateway, want: ErrBlingAPIUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.header != "" {
					w.Header().Set("Retry-After", test.header)
				}
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(`{"secret":"must not escape"}`))
			}))
			defer server.Close()
			client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
			if err != nil {
				t.Fatal(err)
			}
			client.wait = func(context.Context, time.Duration) error { return nil }
			_, err = client.ListReceivables(context.Background(), ReceivablesFilter{})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, errors.Is(%v) is false", err, test.want)
			}
			if strings.Contains(err.Error(), "must not escape") || strings.Contains(err.Error(), "access-token") {
				t.Fatalf("error leaked response or token: %v", err)
			}
			if test.wantRetry {
				var apiErr *BlingAPIError
				if !errors.As(err, &apiErr) || apiErr.RetryAfter != 7*1e9 {
					t.Fatalf("retry metadata missing: %v", err)
				}
			}
		})
	}
}

func TestListReceivablesRetriesTransientFailuresWithJitter(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":321}],"pagination":{"page":1,"limit":1,"total":1}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	var waits []time.Duration
	client.wait = func(_ context.Context, delay time.Duration) error { waits = append(waits, delay); return nil }
	client.random = func() float64 { return 0.5 }
	page, err := client.ListReceivables(context.Background(), ReceivablesFilter{Limit: 1})
	if err != nil || len(page.Records) != 1 {
		t.Fatalf("page = %+v, err = %v", page, err)
	}
	if attempts != 3 || len(waits) != 2 || waits[0] != 300*time.Millisecond || waits[1] != 550*time.Millisecond {
		t.Fatalf("attempts=%d waits=%v", attempts, waits)
	}
}

func TestListReceivablesStopsAfterRetryBudget(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	client.wait = func(context.Context, time.Duration) error { return nil }
	_, err = client.ListReceivables(context.Background(), ReceivablesFilter{})
	if !errors.Is(err, ErrBlingAPIRateLimited) {
		t.Fatalf("error = %v", err)
	}
	if attempts != defaultBlingMaxAttempts {
		t.Fatalf("attempts = %d, want %d", attempts, defaultBlingMaxAttempts)
	}
}

func TestRetryAfterAcceptsSecondsAndHTTPDate(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	if got := retryAfterAt("7", now); got != 7*time.Second {
		t.Fatalf("seconds retry-after = %v", got)
	}
	future := now.Add(15 * time.Second).Format(http.TimeFormat)
	if got := retryAfterAt(future, now); got != 15*time.Second {
		t.Fatalf("date retry-after = %v", got)
	}
	if got := retryAfterAt(now.Add(-time.Second).Format(http.TimeFormat), now); got != 0 {
		t.Fatalf("past date retry-after = %v", got)
	}
	if got := retryAfterAt("not-a-date", now); got != 0 {
		t.Fatalf("invalid retry-after = %v", got)
	}
}

func TestListReceivablesRejectsUnhomologatedShapeAndOversizedBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing data", body: `{"pagination":{"total":1}}`},
		{name: "data is object", body: `{"data":{"id":1}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()
			client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.ListReceivables(context.Background(), ReceivablesFilter{})
			if !errors.Is(err, ErrBlingAPISchemaMismatch) {
				t.Fatalf("error = %v", err)
			}
		})
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":123}]}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	client.maxResponseBytes = 5
	_, err = client.ListReceivables(context.Background(), ReceivablesFilter{})
	if !errors.Is(err, ErrBlingAPIInvalidResponse) {
		t.Fatalf("oversized error = %v", err)
	}
}

func TestNewBlingAPIClientRejectsNonTLSNonLoopback(t *testing.T) {
	if _, err := NewBlingAPIClient(nil, "http://bling.example.test/Api/v3", "token"); !errors.Is(err, ErrBlingAPIConfiguration) {
		t.Fatalf("error = %v", err)
	}
}
