package bling

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	database "majucau.local/financial-intelligence/database/gen"
)

func TestPaymentMethodTermsAreNormalizedAtomically(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL payment-method normalization test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		t.Skip("payment-method normalization integration test only runs against the dedicated majucau_test database")
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ('BLING', 'CONNECTED', 'payment-method-normalizer-ci')
		ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("prepare Bling connection: %v", err)
	}
	connection, err := database.New(pool).GetIntegrationConnectionByProvider(ctx, "BLING")
	if err != nil || !connection.ID.Valid {
		t.Fatalf("load Bling connection: %v", err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin normalization transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var batch pgtype.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO integration_sync_batches (connection_id, resource)
		VALUES ($1, $2) RETURNING id`, connection.ID, PaymentMethodsAPIResource).Scan(&batch); err != nil {
		t.Fatalf("create payment-method batch: %v", err)
	}
	_, normalize := normalizePaymentMethodsAPISnapshotPage(nil)
	if err := normalize(ctx, database.New(tx), connection.ID, batch, []json.RawMessage{
		json.RawMessage(`{"list":{"id":910001,"descricao":"cartao"},"detail":{"id":910001,"destino":1,"taxas":{"aliquota":2.5,"valor":1.25,"prazo":30},"utilizaDiasUteis":false}}`),
		json.RawMessage(`{"list":{"id":910002,"descricao":"sem configuracao"},"detail":{"id":910002,"destino":1}}`),
	}); err != nil {
		t.Fatalf("normalize payment-method settings: %v", err)
	}
	rows, err := tx.Query(ctx, `
		SELECT source_id, destination, fee_rate::text, fee_rate_known,
		       fixed_fee::text, fixed_fee_known, settlement_days, settlement_known,
		       use_business_days, business_days_known,
		       raw_record_id IS NOT NULL
		  FROM bling_payment_method_terms
		 WHERE sync_batch_id = $1
		 ORDER BY source_id`, batch)
	if err != nil {
		t.Fatalf("read normalized payment-method terms: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("normalized payment-method terms are missing")
	}
	var (
		sourceID, feeRate, fixedFee                                         string
		destination, settlementDays                                         *int32
		feeKnown, fixedKnown, settlementKnown, businessDaysKnown, rawLinked bool
		businessDays                                                        *bool
	)
	if err := rows.Scan(&sourceID, &destination, &feeRate, &feeKnown, &fixedFee, &fixedKnown, &settlementDays, &settlementKnown, &businessDays, &businessDaysKnown, &rawLinked); err != nil {
		t.Fatalf("scan configured payment-method terms: %v", err)
	}
	if sourceID != "910001" || destination == nil || *destination != 1 || feeRate != "2.50000000" || !feeKnown || fixedFee != "1.2500" || !fixedKnown || settlementDays == nil || *settlementDays != 30 || !settlementKnown || businessDays == nil || *businessDays || !businessDaysKnown || !rawLinked {
		t.Fatalf("configured terms were not preserved: source=%s destination=%v rate=%s/%v fixed=%s/%v days=%v/%v business=%v/%v raw=%v", sourceID, destination, feeRate, feeKnown, fixedFee, fixedKnown, settlementDays, settlementKnown, businessDays, businessDaysKnown, rawLinked)
	}
	if !rows.Next() {
		t.Fatal("second normalized payment-method row is missing")
	}
	if err := rows.Scan(&sourceID, &destination, &feeRate, &feeKnown, &fixedFee, &fixedKnown, &settlementDays, &settlementKnown, &businessDays, &businessDaysKnown, &rawLinked); err != nil {
		t.Fatalf("scan incomplete payment-method terms: %v", err)
	}
	if sourceID != "910002" || destination == nil || *destination != 1 || feeRate != "0.00000000" || feeKnown || fixedFee != "0.0000" || fixedKnown || settlementDays != nil || settlementKnown || businessDays != nil || businessDaysKnown || !rawLinked {
		t.Fatalf("unknown settings were confused with configured zero values: source=%s destination=%v rate=%s/%v fixed=%s/%v days=%v/%v business=%v/%v raw=%v", sourceID, destination, feeRate, feeKnown, fixedFee, fixedKnown, settlementDays, settlementKnown, businessDays, businessDaysKnown, rawLinked)
	}
	if rows.Next() {
		t.Fatal("unexpected extra payment-method row")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read normalized payment-method rows: %v", err)
	}
}
