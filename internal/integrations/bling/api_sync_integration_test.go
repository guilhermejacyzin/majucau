package bling

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/integrations/httpretry"
)

func TestAPIResourceSyncKeepsFailureAuditAndCommitsCursorWithRAW(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL API sync integration test")
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
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ('BLING', 'CONNECTED', 'api-sync-atomicity-ci')
		ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("prepare Bling connection: %v", err)
	}
	connection, err := database.New(pool).GetIntegrationConnectionByProvider(ctx, "BLING")
	if err != nil || !connection.ID.Valid {
		t.Fatalf("load Bling connection: %v", err)
	}

	const resource = "ci.bling.api_sync.atomicity"
	const sourceEntity = "ci.bling.api_sync.atomicity.page"
	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM bling_sync_page_checkpoints WHERE connection_id = $1 AND resource = $2`, connection.ID, resource)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM raw_records WHERE source_entity = $1`, sourceEntity)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, resource)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_sync_batches WHERE connection_id = $1 AND resource = $2`, connection.ID, resource)
	}
	cleanup()
	defer cleanup()

	var priorBatchID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at)
		VALUES ($1, $2, 'SUCCESS', clock_timestamp())
		RETURNING id::text`, connection.ID, resource).Scan(&priorBatchID); err != nil {
		t.Fatalf("create prior successful batch: %v", err)
	}
	previousPayload := []byte(`{"page":"previous-valid"}`)
	previousDigest := sha256.Sum256(previousPayload)
	previousHash := hex.EncodeToString(previousDigest[:])
	if _, err := pool.Exec(ctx, `
		INSERT INTO raw_records (sync_batch_id, connection_id, source_system, source_entity, source_id, payload_hash, payload, is_current)
		VALUES ($1::uuid, $2, 'BLING', $3, 'pagina:1', $4, $5::jsonb, true)`, priorBatchID, connection.ID, sourceEntity, previousHash, previousPayload); err != nil {
		t.Fatalf("seed latest valid RAW page: %v", err)
	}
	stalePagePayload := []byte(`{"page":"stale-page-four"}`)
	stalePageDigest := sha256.Sum256(stalePagePayload)
	if _, err := pool.Exec(ctx, `
		INSERT INTO raw_records (sync_batch_id, connection_id, source_system, source_entity, source_id, payload_hash, payload, is_current)
		VALUES ($1::uuid, $2, 'BLING', $3, 'pagina:4', $4, $5::jsonb, true)`, priorBatchID, connection.ID, sourceEntity, hex.EncodeToString(stalePageDigest[:]), stalePagePayload); err != nil {
		t.Fatalf("seed stale RAW page: %v", err)
	}
	const previousCursor = "checkpoint:previous-success"
	if _, err := pool.Exec(ctx, `
		INSERT INTO sync_cursors (connection_id, resource, cursor_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (connection_id, resource) DO UPDATE SET cursor_value = EXCLUDED.cursor_value`, connection.ID, resource, previousCursor); err != nil {
		t.Fatalf("seed previous cursor: %v", err)
	}

	pageOne := BlingReceivablesPage{Page: 1, Limit: 1, HasNext: true, HasNextKnown: true, Records: []json.RawMessage{json.RawMessage(`{"id":"page-one-new"}`)}}
	pageTwo := BlingReceivablesPage{Page: 2, Limit: 1, HasNext: false, HasNextKnown: true, Records: []json.RawMessage{json.RawMessage(`{"id":"page-two-new"}`)}}
	unknownEmptyPage := BlingReceivablesPage{Page: 2, Limit: 1, HasNext: false, HasNextKnown: false, Records: []json.RawMessage{}}
	pageTwoFailure := errors.New("injected page read failure")
	transientPageFailure := errors.New("injected transient page request")
	retryRequests := 0
	failSecondPage := true
	unknownPagination := false
	pageOneCalls := 0
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("Brasilia", -3*60*60))
	dailyCorrelationID, err := dailyBlingCorrelationID(scheduledAt)
	if err != nil {
		t.Fatalf("create daily sync correlation ID: %v", err)
	}
	persistedCorrelation := pgUUID(dailyCorrelationID)
	service := &apiResourceSyncService{
		pool:         pool,
		client:       &BlingAPIClient{},
		maxPages:     3,
		resource:     resource,
		sourceEntity: sourceEntity,
		list: func(requestCtx context.Context, filter ReceivablesFilter) (BlingReceivablesPage, error) {
			switch filter.Page {
			case 1:
				pageOneCalls++
				_, retryErr := httpretry.Do(requestCtx, httpretry.Policy{
					Method:    http.MethodGet,
					Retryable: func(err error) bool { return errors.Is(err, transientPageFailure) },
					Wait:      func(context.Context, time.Duration) error { return nil },
				}, func() (struct{}, error) {
					retryRequests++
					if retryRequests == 1 {
						return struct{}{}, transientPageFailure
					}
					return struct{}{}, nil
				})
				if retryErr != nil {
					return BlingReceivablesPage{}, retryErr
				}
				return pageOne, nil
			case 2:
				if failSecondPage {
					return BlingReceivablesPage{}, pageTwoFailure
				}
				if unknownPagination {
					return unknownEmptyPage, nil
				}
				return pageTwo, nil
			default:
				return BlingReceivablesPage{}, errors.New("unexpected page request")
			}
		},
	}

	failed, err := service.sync(ctx, ReceivablesFilter{Page: 1, Limit: 1}, &persistedCorrelation)
	if !errors.Is(err, pageTwoFailure) {
		t.Fatalf("failed page sync error = %v, want injected read failure", err)
	}
	if failed.Status != "FAILED" || failed.BatchID == "" || failed.RetryCount != 1 || retryRequests != 2 {
		t.Fatalf("failed sync result = %+v", failed)
	}
	var failedStatus, failedCode string
	var failedRecordsRead int
	var failedRetryCount int32
	if err := pool.QueryRow(ctx, `
		SELECT status, COALESCE(error_code, ''), records_read, retry_count
		FROM integration_sync_batches WHERE id = $1::uuid`, failed.BatchID).Scan(&failedStatus, &failedCode, &failedRecordsRead, &failedRetryCount); err != nil {
		t.Fatalf("read persisted failed batch: %v", err)
	}
	if failedStatus != "FAILED" || failedCode != "BLING_API_READ_FAILED" || failedRecordsRead != 1 || failedRetryCount != 1 {
		t.Fatalf("failed batch status/code/records/retries = %q/%q/%d/%d", failedStatus, failedCode, failedRecordsRead, failedRetryCount)
	}
	var currentHash, cursor string
	if err := pool.QueryRow(ctx, `
		SELECT payload_hash FROM raw_records
		WHERE connection_id = $1 AND source_entity = $2 AND source_id = 'pagina:1' AND is_current`, connection.ID, sourceEntity).Scan(&currentHash); err != nil {
		t.Fatalf("read preserved current RAW page: %v", err)
	}
	if currentHash != previousHash {
		t.Fatalf("failed sync replaced the latest valid RAW hash: got %s, want %s", currentHash, previousHash)
	}
	if err := pool.QueryRow(ctx, `SELECT cursor_value FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, resource).Scan(&cursor); err != nil {
		t.Fatalf("read cursor after failed sync: %v", err)
	}
	if cursor != previousCursor {
		t.Fatalf("failed sync advanced cursor to %q, want %q", cursor, previousCursor)
	}
	var rawCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM raw_records WHERE connection_id = $1 AND source_entity = $2`, connection.ID, sourceEntity).Scan(&rawCount); err != nil {
		t.Fatalf("count RAW pages after failed sync: %v", err)
	}
	if rawCount != 2 {
		t.Fatalf("failed sync changed the prior RAW snapshot: got %d rows, want the original 2", rawCount)
	}
	var stagedPages int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM bling_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`, connection.ID, resource, persistedCorrelation).Scan(&stagedPages); err != nil {
		t.Fatalf("count checkpoints after interrupted sync: %v", err)
	}
	if stagedPages != 1 || pageOneCalls != 1 {
		t.Fatalf("interrupted sync checkpoint/pages called = %d/%d, want one saved page and one fetched page", stagedPages, pageOneCalls)
	}

	failSecondPage = false
	succeeded, err := service.sync(ctx, ReceivablesFilter{Page: 1, Limit: 1}, &persistedCorrelation)
	if err != nil {
		t.Fatalf("successful paginated sync: %v", err)
	}
	if succeeded.Status != "SUCCESS" || succeeded.PagesRead != 2 || succeeded.RecordsRead != 2 || succeeded.RetryCount != 0 {
		t.Fatalf("successful sync result = %+v", succeeded)
	}
	if pageOneCalls != 1 {
		t.Fatalf("resumed sync re-read page 1 %d times; want only the original read", pageOneCalls)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM bling_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`, connection.ID, resource, persistedCorrelation).Scan(&stagedPages); err != nil {
		t.Fatalf("count checkpoints after successful publish: %v", err)
	}
	if stagedPages != 0 {
		t.Fatalf("successful sync left %d page checkpoints; want none", stagedPages)
	}
	var persistedCorrelationID string
	if err := pool.QueryRow(ctx, `SELECT correlation_id::text FROM integration_sync_batches WHERE id = $1::uuid`, succeeded.BatchID).Scan(&persistedCorrelationID); err != nil {
		t.Fatalf("read persisted daily correlation ID: %v", err)
	}
	if persistedCorrelationID != dailyCorrelationID.String() {
		t.Fatalf("persisted correlation ID = %q, want %q", persistedCorrelationID, dailyCorrelationID)
	}
	if err := pool.QueryRow(ctx, `SELECT cursor_value FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, resource).Scan(&cursor); err != nil {
		t.Fatalf("read cursor after successful sync: %v", err)
	}
	if cursor != "page:2" {
		t.Fatalf("successful sync cursor = %q, want page:2", cursor)
	}
	var currentPages, allPages int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE is_current), count(*)
		FROM raw_records WHERE connection_id = $1 AND source_entity = $2`, connection.ID, sourceEntity).Scan(&currentPages, &allPages); err != nil {
		t.Fatalf("count RAW pages after successful sync: %v", err)
	}
	if currentPages != 2 || allPages != 4 {
		t.Fatalf("successful sync RAW current/total counts = %d/%d, want 2/4", currentPages, allPages)
	}
	var staleCurrentPages int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM raw_records
		WHERE connection_id = $1 AND source_entity = $2 AND source_id = 'pagina:4' AND is_current`, connection.ID, sourceEntity).Scan(&staleCurrentPages); err != nil {
		t.Fatalf("check stale page retirement: %v", err)
	}
	if staleCurrentPages != 0 {
		t.Fatalf("stale page 4 is still current after a complete read: %d rows", staleCurrentPages)
	}

	unknownPagination = true
	unknown, err := service.Sync(ctx, ReceivablesFilter{Page: 1, Limit: 1})
	if !errors.Is(err, ErrBlingAPIPaginationUnknown) {
		t.Fatalf("sync without page-completion metadata error = %v, want pagination unknown", err)
	}
	if unknown.Status != "FAILED" {
		t.Fatalf("sync without page-completion metadata status = %q, want FAILED", unknown.Status)
	}
	var unknownStatus, unknownCode string
	if err := pool.QueryRow(ctx, `
		SELECT status, COALESCE(error_code, '')
		FROM integration_sync_batches WHERE id = $1::uuid`, unknown.BatchID).Scan(&unknownStatus, &unknownCode); err != nil {
		t.Fatalf("read unknown-pagination audit: %v", err)
	}
	if unknownStatus != "FAILED" || unknownCode != "BLING_API_PAGINATION_UNKNOWN" {
		t.Fatalf("unknown-pagination audit = %q/%q", unknownStatus, unknownCode)
	}
	var unknownCorrelation pgtype.UUID
	if err := pool.QueryRow(ctx, `SELECT correlation_id FROM integration_sync_batches WHERE id = $1::uuid`, unknown.BatchID).Scan(&unknownCorrelation); err != nil {
		t.Fatalf("read unknown-pagination correlation ID: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT cursor_value FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, resource).Scan(&cursor); err != nil {
		t.Fatalf("read cursor after unknown pagination: %v", err)
	}
	if cursor != "page:2" {
		t.Fatalf("unknown pagination advanced cursor to %q, want page:2", cursor)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE is_current), count(*)
		FROM raw_records WHERE connection_id = $1 AND source_entity = $2`, connection.ID, sourceEntity).Scan(&currentPages, &allPages); err != nil {
		t.Fatalf("count RAW pages after unknown pagination: %v", err)
	}
	if currentPages != 2 || allPages != 4 {
		t.Fatalf("unknown pagination changed RAW current/total counts to %d/%d, want 2/4", currentPages, allPages)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM bling_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`, connection.ID, resource, unknownCorrelation).Scan(&stagedPages); err != nil {
		t.Fatalf("count staged pages after unknown pagination: %v", err)
	}
	if stagedPages != 1 {
		t.Fatalf("unknown pagination checkpoint count = %d, want only the confirmed first page", stagedPages)
	}

	unknownPagination = false
	replayed, err := service.Sync(ctx, ReceivablesFilter{Page: 1, Limit: 1})
	if err != nil {
		t.Fatalf("idempotent API sync replay: %v", err)
	}
	if replayed.Status != "SUCCESS" || replayed.PagesRead != 2 || replayed.PagesCreated != 0 || replayed.PagesUpdated != 0 {
		t.Fatalf("idempotent replay result = %+v", replayed)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE is_current), count(*)
		FROM raw_records WHERE connection_id = $1 AND source_entity = $2`, connection.ID, sourceEntity).Scan(&currentPages, &allPages); err != nil {
		t.Fatalf("count RAW pages after idempotent replay: %v", err)
	}
	if currentPages != 2 || allPages != 4 {
		t.Fatalf("idempotent replay changed RAW current/total counts to %d/%d, want 2/4", currentPages, allPages)
	}
}

func TestAPIResourceSyncSerializesConcurrentRunsForSameResource(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL API sync integration test")
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
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ('BLING', 'CONNECTED', 'api-sync-concurrency-ci')
		ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("prepare Bling connection: %v", err)
	}
	connection, err := database.New(pool).GetIntegrationConnectionByProvider(ctx, "BLING")
	if err != nil || !connection.ID.Valid {
		t.Fatalf("load Bling connection: %v", err)
	}

	const resource = "ci.bling.api_sync.concurrent"
	const sourceEntity = "ci.bling.api_sync.concurrent.page"
	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM bling_sync_page_checkpoints WHERE connection_id = $1 AND resource = $2`, connection.ID, resource)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM raw_records WHERE source_entity = $1`, sourceEntity)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, resource)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_sync_batches WHERE connection_id = $1 AND resource = $2`, connection.ID, resource)
	}
	cleanup()
	defer cleanup()

	firstReadStarted := make(chan struct{})
	releaseFirstRead := make(chan struct{})
	var releaseOnce sync.Once
	allowFirstReadToFinish := func() { releaseOnce.Do(func() { close(releaseFirstRead) }) }
	defer allowFirstReadToFinish()
	secondReadStarted := make(chan struct{}, 1)
	page := BlingReceivablesPage{Page: 1, Limit: 1, HasNextKnown: true, Records: []json.RawMessage{json.RawMessage(`{"id":"same-page"}`)}}

	firstService := &apiResourceSyncService{
		pool: pool, client: &BlingAPIClient{}, maxPages: 1, resource: resource, sourceEntity: sourceEntity,
		list: func(ctx context.Context, _ ReceivablesFilter) (BlingReceivablesPage, error) {
			close(firstReadStarted)
			select {
			case <-releaseFirstRead:
				return page, nil
			case <-ctx.Done():
				return BlingReceivablesPage{}, ctx.Err()
			}
		},
	}
	secondService := &apiResourceSyncService{
		pool: pool, client: &BlingAPIClient{}, maxPages: 1, resource: resource, sourceEntity: sourceEntity,
		list: func(context.Context, ReceivablesFilter) (BlingReceivablesPage, error) {
			secondReadStarted <- struct{}{}
			return page, nil
		},
	}
	type syncOutcome struct{ err error }
	firstDone := make(chan syncOutcome, 1)
	secondDone := make(chan syncOutcome, 1)
	go func() {
		_, err := firstService.Sync(ctx, ReceivablesFilter{Page: 1, Limit: 1})
		firstDone <- syncOutcome{err: err}
	}()
	select {
	case <-firstReadStarted:
	case <-ctx.Done():
		t.Fatalf("first sync did not reach the Bling read: %v", ctx.Err())
	}
	go func() {
		_, err := secondService.Sync(ctx, ReceivablesFilter{Page: 1, Limit: 1})
		secondDone <- syncOutcome{err: err}
	}()
	select {
	case <-secondReadStarted:
		allowFirstReadToFinish()
		<-firstDone
		<-secondDone
		t.Fatal("second sync reached the Bling read before the first same-resource transaction finished")
	case <-time.After(200 * time.Millisecond):
	case <-ctx.Done():
		allowFirstReadToFinish()
		t.Fatalf("waiting for the resource lock check: %v", ctx.Err())
	}
	allowFirstReadToFinish()
	select {
	case outcome := <-firstDone:
		if outcome.err != nil {
			t.Fatalf("first sync failed: %v", outcome.err)
		}
	case <-ctx.Done():
		t.Fatalf("first sync did not finish: %v", ctx.Err())
	}
	select {
	case <-secondReadStarted:
	case <-ctx.Done():
		t.Fatalf("second sync did not continue after first transaction: %v", ctx.Err())
	}
	select {
	case outcome := <-secondDone:
		if outcome.err != nil {
			t.Fatalf("second sync failed: %v", outcome.err)
		}
	case <-ctx.Done():
		t.Fatalf("second sync did not finish: %v", ctx.Err())
	}

	var successfulBatches int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM integration_sync_batches
		WHERE connection_id = $1 AND resource = $2 AND status = 'SUCCESS'`, connection.ID, resource).Scan(&successfulBatches); err != nil {
		t.Fatalf("count successful concurrent batches: %v", err)
	}
	if successfulBatches != 2 {
		t.Fatalf("successful batches = %d, want 2 serialized runs", successfulBatches)
	}
}
