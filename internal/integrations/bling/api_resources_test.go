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

func TestGetPayableAndBorderoReadBoundedRawDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer access-token" {
			t.Fatalf("authorization header missing")
		}
		switch r.URL.Path {
		case "/Api/v3/contas/pagar/123":
			_, _ = w.Write([]byte(`{"data":{"id":123,"categoria":{"id":9}}}`))
		case "/Api/v3/borderos/456":
			_, _ = w.Write([]byte(`{"data":{"id":456,"pagamentos":[{"valorPago":12.34}]}}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}

	payable, err := client.GetPayable(context.Background(), 123)
	if err != nil || string(payable) != `{"id":123,"categoria":{"id":9}}` {
		t.Fatalf("payable = %s, err = %v", payable, err)
	}
	bordero, err := client.GetBordero(context.Background(), 456)
	if err != nil || !strings.Contains(string(bordero), `"valorPago":12.34`) {
		t.Fatalf("bordero = %s, err = %v", bordero, err)
	}
}

func TestGetPayableRejectsInvalidIDBeforeNetworkCall(t *testing.T) {
	doer := &errorHTTPDoer{err: errors.New("network should not be called")}
	client, err := NewBlingAPIClient(doer, "http://127.0.0.1/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{0, -1} {
		if _, err := client.GetPayable(context.Background(), id); !errors.Is(err, ErrBlingAPIConfiguration) {
			t.Fatalf("id %d error = %v, want configuration error", id, err)
		}
	}
	if doer.calls != 0 {
		t.Fatalf("network calls = %d, want 0", doer.calls)
	}
}

func TestGetPayableRetriesTransientFailureAndCapsRetryAfter(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "45")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"id":123}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	var waited time.Duration
	client.wait = func(_ context.Context, delay time.Duration) error { waited = delay; return nil }
	if _, err := client.GetPayable(context.Background(), 123); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || waited != defaultBlingMaxRetryAfter {
		t.Fatalf("calls = %d, waited = %s", calls, waited)
	}
}

func TestGetPayableRejectsInvalidAndOversizedDetail(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
		want error
	}{
		{name: "missing data object", body: `{"data":[]}`, want: ErrBlingAPISchemaMismatch},
		{name: "missing data", body: `{"pagination":{"total":1}}`, want: ErrBlingAPISchemaMismatch},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(test.body)) }))
			defer server.Close()
			client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.GetPayable(context.Background(), 1)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"data":{"id":123}}`)) }))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	client.maxResponseBytes = 10
	if _, err := client.GetPayable(context.Background(), 1); !errors.Is(err, ErrBlingAPIInvalidResponse) {
		t.Fatalf("oversized response error = %v", err)
	}
}

func TestListFinancialCategoriesRequiresExplicitPaginationMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Api/v3/categorias/receitas-despesas" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("pagina") != "2" || r.URL.Query().Get("limite") != "10" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":7,"descricao":"CAPEX"}],"pagination":{"page":2,"limit":10,"hasNext":false}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListFinancialCategories(context.Background(), 2, 10)
	if err != nil || len(page.Records) != 1 || page.HasNext || !page.HasNextKnown {
		t.Fatalf("page = %+v, err = %v", page, err)
	}
}
