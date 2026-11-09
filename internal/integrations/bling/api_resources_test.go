package bling

import (
	"context"
	"encoding/json"
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
		if got := r.Header.Get("enable-jwt"); got != "1" {
			t.Fatalf("enable-jwt = %q, want 1", got)
		}
		switch r.URL.Path {
		case "/Api/v3/contas/receber/321":
			_, _ = w.Write([]byte(`{"data":{"id":321,"situacao":1,"saldo":80.00}}`))
		case "/Api/v3/contas/pagar/123":
			_, _ = w.Write([]byte(`{"data":{"id":123,"categoria":{"id":9}}}`))
		case "/Api/v3/borderos/456":
			_, _ = w.Write([]byte(`{"data":{"id":456,"pagamentos":[{"valorPago":12.34}]}}`))
		case "/Api/v3/categorias/receitas-despesas/9":
			_, _ = w.Write([]byte(`{"data":{"id":9,"descricao":"CAPEX"}}`))
		case "/Api/v3/formas-pagamentos/11":
			_, _ = w.Write([]byte(`{"data":{"id":11,"descricao":"Cartao","destino":1,"taxas":{"aliquota":2.5,"valor":1.25,"prazo":30},"utilizaDiasUteis":true}}`))
		case "/Api/v3/caixas/789":
			_, _ = w.Write([]byte(`{"data":{"id":789,"categoria":{"id":9},"valor":18.25}}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}

	receivable, err := client.GetReceivable(context.Background(), 321)
	if err != nil || string(receivable) != `{"id":321,"situacao":1,"saldo":80.00}` {
		t.Fatalf("receivable = %s, err = %v", receivable, err)
	}
	payable, err := client.GetPayable(context.Background(), 123)
	if err != nil || string(payable) != `{"id":123,"categoria":{"id":9}}` {
		t.Fatalf("payable = %s, err = %v", payable, err)
	}
	bordero, err := client.GetBordero(context.Background(), 456)
	if err != nil || !strings.Contains(string(bordero), `"valorPago":12.34`) {
		t.Fatalf("bordero = %s, err = %v", bordero, err)
	}
	category, err := client.GetFinancialCategory(context.Background(), 9)
	if err != nil || !strings.Contains(string(category), `"descricao":"CAPEX"`) {
		t.Fatalf("category = %s, err = %v", category, err)
	}
	movement, err := client.GetCashMovement(context.Background(), 789)
	if err != nil || !strings.Contains(string(movement), `"categoria":{"id":9}`) {
		t.Fatalf("movement = %s, err = %v", movement, err)
	}
	method, err := client.GetPaymentMethod(context.Background(), 11)
	if err != nil || !strings.Contains(string(method), `"destino":1`) || !strings.Contains(string(method), `"prazo":30`) {
		t.Fatalf("payment method = %s, err = %v", method, err)
	}
}

func TestGetPayableRefreshesExpiredAccessTokenOnce(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet || r.URL.Path != "/Api/v3/contas/pagar/123" {
			t.Fatalf("request = %s %s, want GET payable detail", r.Method, r.URL.Path)
		}
		if calls == 1 {
			if got := r.Header.Get("Authorization"); got != "Bearer expired-token" {
				t.Errorf("first authorization = %q, want expired token", got)
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer refreshed-token" {
			t.Errorf("retry authorization = %q, want refreshed token", got)
		}
		_, _ = w.Write([]byte(`{"data":{"id":123,"saldo":250.00}}`))
	}))
	defer server.Close()

	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "expired-token")
	if err != nil {
		t.Fatal(err)
	}
	refreshCalls := 0
	client.SetUnauthorizedTokenRefresh(func(context.Context, string) (string, error) {
		refreshCalls++
		return "refreshed-token", nil
	})

	record, err := client.GetPayable(context.Background(), 123)
	if err != nil || string(record) != `{"id":123,"saldo":250.00}` {
		t.Fatalf("GetPayable() after refresh = %s, err = %v", record, err)
	}
	if calls != 2 || refreshCalls != 1 {
		t.Fatalf("requests=%d refreshes=%d, want 2/1", calls, refreshCalls)
	}
}

func TestGetPayableDoesNotRetryWhenTokenRefreshFails(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "expired-token")
	if err != nil {
		t.Fatal(err)
	}
	refreshErr := errors.New("refresh rejected")
	refreshCalls := 0
	client.SetUnauthorizedTokenRefresh(func(context.Context, string) (string, error) {
		refreshCalls++
		return "", refreshErr
	})

	_, err = client.GetPayable(context.Background(), 123)
	if !errors.Is(err, ErrBlingAPIUnauthorized) || !errors.Is(err, refreshErr) {
		t.Fatalf("GetPayable() error = %v, want authorization and refresh errors", err)
	}
	if calls != 1 || refreshCalls != 1 {
		t.Fatalf("requests=%d refreshes=%d, want 1/1", calls, refreshCalls)
	}
}

func TestListPaymentMethodsUsesBoundedPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/Api/v3/formas-pagamentos" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("pagina") != "2" || r.URL.Query().Get("limite") != "10" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":11,"descricao":"Cartao"}],"pagination":{"page":2,"limit":10,"hasNext":false}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListPaymentMethods(context.Background(), 2, 10)
	if err != nil || len(page.Records) != 1 || page.HasNext || !page.HasNextKnown {
		t.Fatalf("page = %+v, err = %v", page, err)
	}
}

func TestListFinancialAccountsUsesOfficialBoundedEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/Api/v3/contas-contabeis" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("pagina") != "1" || r.URL.Query().Get("limite") != "10" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":31,"descricao":"Itaú"},{"id":32,"descricao":"Mercado Pago"}],"pagination":{"page":1,"limit":10,"hasNext":false}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListFinancialAccounts(context.Background(), 1, 10)
	if err != nil || len(page.Records) != 2 || page.HasNext || !page.HasNextKnown {
		t.Fatalf("page = %+v, err = %v", page, err)
	}
}

func TestGetFinancialCategoryExplainsInsufficientReadScope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/Api/v3/categorias/receitas-despesas/9" {
			t.Fatalf("unexpected category request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"private provider response"}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetFinancialCategory(context.Background(), 9)
	if !errors.Is(err, ErrBlingFinancialCategoryScopeInsufficient) || !errors.Is(err, ErrBlingAPIScopeInsufficient) {
		t.Fatalf("category error = %v, want category-specific and generic scope errors", err)
	}
	if strings.Contains(err.Error(), "private provider response") {
		t.Fatal("category scope error exposed the provider response body")
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

func TestGetReceivableRejectsInvalidIDBeforeNetworkCall(t *testing.T) {
	doer := &errorHTTPDoer{err: errors.New("network should not be called")}
	client, err := NewBlingAPIClient(doer, "http://127.0.0.1/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{0, -1} {
		if _, err := client.GetReceivable(context.Background(), id); !errors.Is(err, ErrBlingAPIConfiguration) {
			t.Fatalf("id %d error = %v, want configuration error", id, err)
		}
	}
	if doer.calls != 0 {
		t.Fatalf("network calls = %d, want 0", doer.calls)
	}
}

func TestGetContactReadsOnlyTheRequestedNumericContact(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		gotPath = request.URL.Path
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"data":{"id":123,"nome":"Fornecedor teste"}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	contact, err := client.GetContact(context.Background(), 123)
	if err != nil {
		t.Fatalf("GetContact() error = %v", err)
	}
	if gotPath != "/Api/v3/contatos/123" {
		t.Fatalf("GetContact() path = %q", gotPath)
	}
	if !json.Valid(contact) || !strings.Contains(string(contact), `"nome":"Fornecedor teste"`) {
		t.Fatalf("GetContact() data = %s", contact)
	}
}

func TestGetContactRejectsInvalidIDBeforeNetworkCall(t *testing.T) {
	doer := &errorHTTPDoer{err: errors.New("network should not be called")}
	client, err := NewBlingAPIClient(doer, "http://127.0.0.1/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetContact(context.Background(), 0); !errors.Is(err, ErrBlingAPIConfiguration) {
		t.Fatalf("GetContact(0) error = %v", err)
	}
	if doer.calls != 0 {
		t.Fatalf("GetContact(0) network calls = %d, want 0", doer.calls)
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

func TestListFinancialCategoriesRespectsExplicitPaginationMetadata(t *testing.T) {
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

func TestWalkFinancialCategoryHierarchyReadsLeafToRootInBoundedCalls(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer access-token" || r.Header.Get("enable-jwt") != "1" {
			t.Errorf("unexpected request: method=%q authorization-present=%t enable-jwt=%q", r.Method, r.Header.Get("Authorization") != "", r.Header.Get("enable-jwt"))
		}
		switch r.URL.Path {
		case "/Api/v3/categorias/receitas-despesas/22":
			_, _ = w.Write([]byte(`{"data":{"id":22,"idCategoriaPai":11,"descricao":"Maquinário produtivo","tipo":1}}`))
		case "/Api/v3/categorias/receitas-despesas/11":
			_, _ = w.Write([]byte(`{"data":{"id":11,"idCategoriaPai":0,"descricao":"CAPEX","tipo":1}}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}

	var got []FinancialCategoryNode
	err = client.WalkFinancialCategoryHierarchy(context.Background(), 22, func(category FinancialCategoryNode) error {
		got = append(got, category)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(got) != 2 {
		t.Fatalf("calls=%d categories=%d, want 2", calls, len(got))
	}
	if got[0].ID != 22 || got[0].ParentID != 11 || got[0].Type != 1 || !strings.Contains(string(got[0].Payload), "Maquinário produtivo") {
		t.Fatalf("leaf = %+v", got[0])
	}
	if got[1].ID != 11 || got[1].ParentID != 0 || got[1].Type != 1 {
		t.Fatalf("root = %+v", got[1])
	}
}

func TestWalkFinancialCategoryHierarchyRejectsCyclesAndMissingParent(t *testing.T) {
	for _, test := range []struct {
		name string
		body map[string]string
	}{
		{
			name: "cycle",
			body: map[string]string{
				"1": `{"data":{"id":1,"idCategoriaPai":2,"descricao":"A","tipo":1}}`,
				"2": `{"data":{"id":2,"idCategoriaPai":1,"descricao":"B","tipo":1}}`,
			},
		},
		{
			name: "missing parent",
			body: map[string]string{
				"1": `{"data":{"id":1,"descricao":"CAPEX","tipo":1}}`,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id := strings.TrimPrefix(r.URL.Path, "/categorias/receitas-despesas/")
				body, ok := test.body[id]
				if !ok {
					t.Errorf("unexpected path %q", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()
			client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
			if err != nil {
				t.Fatal(err)
			}
			err = client.WalkFinancialCategoryHierarchy(context.Background(), 1, func(FinancialCategoryNode) error { return nil })
			if !errors.Is(err, ErrBlingAPISchemaMismatch) {
				t.Fatalf("error = %v, want schema mismatch", err)
			}
		})
	}
}

func TestWalkFinancialCategoryHierarchyStopsWhenVisitorFails(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"data":{"id":1,"idCategoriaPai":0,"descricao":"CAPEX","tipo":1}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("stop visitor")
	err = client.WalkFinancialCategoryHierarchy(context.Background(), 1, func(FinancialCategoryNode) error { return wantErr })
	if !errors.Is(err, wantErr) || calls != 1 {
		t.Fatalf("error=%v calls=%d, want visitor error and one request", err, calls)
	}
}

func TestListCashMovementsBuildsBoundedReadOnlyPageRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/Api/v3/caixas" {
			t.Fatalf("request = %s %s, want GET /Api/v3/caixas", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("pagina") != "2" || r.URL.Query().Get("limite") != "2" {
			t.Fatalf("pagination query = %s", r.URL.RawQuery)
		}
		if got := r.URL.Query().Get("situacao"); got != "R" {
			t.Fatalf("situation = %q, want R (registered)", got)
		}
		categoryIDs := r.URL.Query()["idsCategorias"]
		if len(categoryIDs) != 2 || categoryIDs[0] != "11" || categoryIDs[1] != "22" {
			t.Fatalf("category IDs = %v, want repeated idsCategorias=11&idsCategorias=22", categoryIDs)
		}
		if r.URL.Query().Get("dataInicial") != "2026-09-01" || r.URL.Query().Get("dataFinal") != "2026-09-30" {
			t.Fatalf("date query = %s", r.URL.RawQuery)
		}
		if got := r.URL.Query().Get("situacaoConciliacao"); got != "1" {
			t.Fatalf("reconciliation situation = %q, want 1 (reconciled)", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("authorization = %q", got)
		}
		if got := r.Header.Get("enable-jwt"); got != "1" {
			t.Fatalf("enable-jwt = %q, want 1", got)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":789}],"pagination":{"page":2,"limit":2,"total":5}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}

	page, err := client.ListCashMovements(context.Background(), CashMovementsFilter{
		Page: 2, MaxRecords: 2, DateFrom: "2026-09-01", DateTo: "2026-09-30", Situation: "R", ReconciliationSituation: 1, CategoryIDs: []int64{11, 22},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Records) != 1 || !strings.Contains(string(page.Records[0]), `"id":789`) || !page.HasNextKnown || !page.HasNext {
		t.Fatalf("page = %+v", page)
	}
}

func TestListCashMovementsRejectsInvalidReconciliationSituationBeforeNetworkCall(t *testing.T) {
	doer := &errorHTTPDoer{err: errors.New("network should not be called")}
	client, err := NewBlingAPIClient(doer, "http://127.0.0.1/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListCashMovements(context.Background(), CashMovementsFilter{ReconciliationSituation: 4}); !errors.Is(err, ErrBlingAPIConfiguration) {
		t.Fatalf("invalid reconciliation situation error = %v, want configuration error", err)
	}
	if doer.calls != 0 {
		t.Fatalf("network calls = %d, want 0", doer.calls)
	}
}

func TestListCashMovementsRejectsInvalidCategoryIDsBeforeNetworkCall(t *testing.T) {
	doer := &errorHTTPDoer{err: errors.New("network should not be called")}
	client, err := NewBlingAPIClient(doer, "http://127.0.0.1/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		ids  []int64
	}{
		{name: "zero", ids: []int64{0}},
		{name: "negative", ids: []int64{-1}},
		{name: "duplicate", ids: []int64{11, 11}},
		{name: "too many", ids: make([]int64, maxCashMovementCategoryIDs+1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := client.ListCashMovements(context.Background(), CashMovementsFilter{CategoryIDs: test.ids}); !errors.Is(err, ErrBlingAPIConfiguration) {
				t.Fatalf("category IDs %v error = %v, want configuration error", test.ids, err)
			}
		})
	}
	if doer.calls != 0 {
		t.Fatalf("network calls = %d, want 0", doer.calls)
	}
}

func TestListCashMovementsContinuesConservativelyWhenResponseOmitsPageSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":789}],"pagination":{"total":5}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListCashMovements(context.Background(), CashMovementsFilter{MaxRecords: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !page.HasNextKnown || !page.HasNext || page.Limit != 1 {
		t.Fatalf("cash movement response should use the requested documented page size and continue: %+v", page)
	}
}

func TestListCashMovementsUsesResponseLimitToConfirmShortPageEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limite") != "2" {
			t.Fatalf("limit = %q, want 2", r.URL.Query().Get("limite"))
		}
		_, _ = w.Write([]byte(`{"data":[{"id":789}],"pagination":{"limit":2,"total":1}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListCashMovements(context.Background(), CashMovementsFilter{Page: 1, MaxRecords: 2})
	if err != nil {
		t.Fatal(err)
	}
	if page.Limit != 2 || !page.HasNextKnown || page.HasNext {
		t.Fatalf("short cash movement page = %+v, want terminal page confirmed by response metadata", page)
	}
}

func TestListCashMovementsUsesDocumentedRequestedLimitWhenMetadataIsOmitted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limite") != "2" || r.URL.Query().Get("pagina") != "3" {
			t.Fatalf("query = %s, want pagina=3&limite=2", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":789}]}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "access-token")
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListCashMovements(context.Background(), CashMovementsFilter{Page: 3, MaxRecords: 2})
	if err != nil {
		t.Fatal(err)
	}
	if page.Limit != 2 || !page.HasNextKnown || page.HasNext {
		t.Fatalf("short caixa page = %+v, want terminal page based on documented requested limit", page)
	}
}

func TestListCashMovementsRejectsOversizedDateFilterBeforeNetworkCall(t *testing.T) {
	doer := &errorHTTPDoer{err: errors.New("network should not be called")}
	client, err := NewBlingAPIClient(doer, "http://127.0.0.1/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	filter := CashMovementsFilter{DateFrom: strings.Repeat("x", maxCashMovementDateFilterBytes+1)}
	if _, err := client.ListCashMovements(context.Background(), filter); !errors.Is(err, ErrBlingAPIConfiguration) {
		t.Fatalf("oversized date filter error = %v, want configuration error", err)
	}
	if doer.calls != 0 {
		t.Fatalf("network calls = %d, want 0", doer.calls)
	}
}

func TestListCashMovementsRejectsInvalidDatesBeforeNetworkCall(t *testing.T) {
	doer := &errorHTTPDoer{err: errors.New("network should not be called")}
	client, err := NewBlingAPIClient(doer, "http://127.0.0.1/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		filter CashMovementsFilter
	}{
		{name: "invalid start day", filter: CashMovementsFilter{DateFrom: "2026-02-30"}},
		{name: "invalid end month", filter: CashMovementsFilter{DateTo: "2026-13-01"}},
		{name: "non-canonical date", filter: CashMovementsFilter{DateFrom: "2026-9-01"}},
		{name: "reversed period", filter: CashMovementsFilter{DateFrom: "2026-10-01", DateTo: "2026-09-30"}},
		{name: "invalid situation", filter: CashMovementsFilter{Situation: "X"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := client.ListCashMovements(context.Background(), test.filter); !errors.Is(err, ErrBlingAPIConfiguration) {
				t.Fatalf("filter error = %v, want configuration error", err)
			}
		})
	}
	if doer.calls != 0 {
		t.Fatalf("network calls = %d, want 0", doer.calls)
	}
}

func TestGetCashMovementRejectsInvalidIDBeforeNetworkCall(t *testing.T) {
	doer := &errorHTTPDoer{err: errors.New("network should not be called")}
	client, err := NewBlingAPIClient(doer, "http://127.0.0.1/Api/v3", "access-token")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{0, -1} {
		if _, err := client.GetCashMovement(context.Background(), id); !errors.Is(err, ErrBlingAPIConfiguration) {
			t.Fatalf("id %d error = %v, want configuration error", id, err)
		}
	}
	if doer.calls != 0 {
		t.Fatalf("network calls = %d, want 0", doer.calls)
	}
}
