package bling

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	database "majucau.local/financial-intelligence/database/gen"
)

func TestDailyReceivablesAPIIsNormalizedAndFailedRefreshKeepsPreviousSnapshot(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL receivable API integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	var databaseName string
	if err := pool.QueryRow(ctx, `SELECT current_database()`).Scan(&databaseName); err != nil {
		t.Fatalf("identify PostgreSQL database: %v", err)
	}
	if databaseName != "majucau_test" {
		t.Skip("receivable API integration test only runs against the dedicated majucau_test database")
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ('BLING', 'CONNECTED', 'api-receivables-normalizer-ci')
		ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("prepare Bling connection: %v", err)
	}
	connection, err := database.New(pool).GetIntegrationConnectionByProvider(ctx, "BLING")
	if err != nil || !connection.ID.Valid {
		t.Fatalf("load Bling connection: %v", err)
	}
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	dateOffset := int(time.Now().UnixNano() % int64(36500))
	uniqueDate := time.Date(2100, time.January, 1, 0, 0, 0, 0, location).AddDate(0, 0, dateOffset)
	firstRunAt := time.Date(uniqueDate.Year(), uniqueDate.Month(), uniqueDate.Day(), 9, 30, 0, 0, location)
	firstCorrelation, err := dailyBlingCorrelationID(firstRunAt)
	if err != nil {
		t.Fatal(err)
	}
	failedCorrelation, err := dailyBlingCorrelationID(firstRunAt.Add(24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	baseID := time.Now().UnixNano()
	firstSourceID := strconv.FormatInt(baseID, 10)
	secondSourceID := strconv.FormatInt(baseID+1, 10)
	contactSourceID := strconv.FormatInt(baseID+2, 10)
	failSecondRecord := false
	publicationTransactionOpenDuringRequest := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("enable-jwt") != "1" {
			t.Errorf("unexpected method/header for %s", r.URL.Path)
		}
		var openTransactions int
		if err := pool.QueryRow(r.Context(), `
			SELECT COUNT(*) FROM pg_stat_activity
			WHERE datname = current_database() AND state = 'idle in transaction'
			  AND pid <> pg_backend_pid()`).Scan(&openTransactions); err != nil {
			t.Errorf("check publication transactions during Bling request: %v", err)
		} else if openTransactions > 0 {
			publicationTransactionOpenDuringRequest = true
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/Api/v3/contas/receber":
			records := fmt.Sprintf(`{"id":%s,"situacao":1,"vencimento":"2026-11-20","dataEmissao":"2026-10-01","valor":100.25,"contato":{"id":%s,"nome":"Cliente de teste"}}`, firstSourceID, contactSourceID)
			if failSecondRecord {
				records += fmt.Sprintf(`,{"id":%s,"situacao":1,"vencimento":"2026-11-21","valor":50,"contato":{"id":%s,"nome":"Outro cliente"}}`, secondSourceID, contactSourceID)
			}
			_, _ = fmt.Fprintf(w, `{"data":[%s]}`, records)
		case "/Api/v3/contas/receber/" + firstSourceID:
			_, _ = fmt.Fprintf(w, `{"data":{"id":%s,"situacao":1,"vencimento":"2026-11-20","dataEmissao":"2026-10-01","competencia":"2026-10-01","numeroDocumento":"CR-teste","valor":100.25,"saldo":80.10,"contato":{"id":%s}}}`, firstSourceID, contactSourceID)
		case "/Api/v3/contas/receber/" + secondSourceID:
			if !failSecondRecord {
				t.Errorf("unexpected detail request for second test record")
			}
			// The mismatched contact must fail before any new snapshot is published.
			_, _ = fmt.Fprintf(w, `{"data":{"id":%s,"situacao":1,"vencimento":"2026-11-21","valor":50,"saldo":50,"contato":{"id":%s}}}`, secondSourceID, strconv.FormatInt(baseID+3, 10))
		default:
			t.Errorf("unexpected Bling API path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "fake-access-token")
	if err != nil {
		t.Fatal(err)
	}
	service := NewReceivablesAPISyncService(pool, client)
	first, err := service.SyncWithCorrelationID(ctx, ReceivablesFilter{}, pgUUID(firstCorrelation))
	if err != nil || first.Status != "SUCCESS" {
		t.Fatalf("first daily receivable sync = %+v, err=%v", first, err)
	}
	if publicationTransactionOpenDuringRequest {
		t.Fatal("Bling API was called while a PostgreSQL publication transaction was open")
	}
	var stage, status, sourceID, businessType, customerName string
	var grossAmount, openBalance string
	if err := pool.QueryRow(ctx, `
		SELECT b.data_stage, b.status, r.source_id, r.business_type, c.name,
		       r.gross_amount::text, r.open_balance::text
		  FROM integration_sync_batches b
		  JOIN receivable_sync_snapshot_rows snapshot ON snapshot.sync_batch_id = b.id
		  JOIN receivables r ON r.id = snapshot.receivable_id
		  JOIN contacts c ON c.id = r.customer_id
		 WHERE b.id = $1::uuid`, first.BatchID).Scan(&stage, &status, &sourceID, &businessType, &customerName, &grossAmount, &openBalance); err != nil {
		t.Fatalf("read normalized daily receivable: %v", err)
	}
	if stage != "NORMALIZED" || status != "SUCCESS" || sourceID != firstSourceID || businessType != "UNCLASSIFIED" || customerName != "Cliente de teste" || grossAmount != "100.2500" || openBalance != "80.1000" {
		t.Fatalf("normalized receivable is not the expected unclassified open title: stage/status=%s/%s id=%s type=%s customer=%s gross=%s open=%s", stage, status, sourceID, businessType, customerName, grossAmount, openBalance)
	}

	failSecondRecord = true
	failed, err := service.SyncWithCorrelationID(ctx, ReceivablesFilter{}, pgUUID(failedCorrelation))
	if err == nil || failed.Status != "FAILED" {
		t.Fatalf("refresh with mismatched detail = %+v, err=%v; want failure", failed, err)
	}
	var successfulSnapshots, publishedRows int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(snapshot.receivable_id)
		  FROM integration_sync_batches b
		  LEFT JOIN receivable_sync_snapshot_rows snapshot ON snapshot.sync_batch_id = b.id
		 WHERE b.id = $1::uuid AND b.status = 'SUCCESS' AND b.data_stage = 'NORMALIZED'`, first.BatchID).Scan(&successfulSnapshots, &publishedRows); err != nil {
		t.Fatalf("check previous receivable snapshot remains published: %v", err)
	}
	if successfulSnapshots != 1 || publishedRows != 1 {
		t.Fatalf("previous receivable snapshot was changed by failed refresh: successful snapshots=%d rows=%d", successfulSnapshots, publishedRows)
	}
}
