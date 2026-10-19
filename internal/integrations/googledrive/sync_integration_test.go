package googledrive

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestFileInventorySyncResumesAndPublishesWholeSnapshot(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL Google Drive sync integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	var seededStatus string
	if err := pool.QueryRow(ctx, `SELECT status FROM integration_connections WHERE provider = $1`, GoogleDriveProvider).Scan(&seededStatus); err != nil {
		t.Fatalf("load migrated Google Drive connection: %v", err)
	}
	if seededStatus != "NOT_CONFIGURED" {
		t.Fatalf("initial Google Drive status = %q, want NOT_CONFIGURED", seededStatus)
	}

	connectionID, connectionCreated := ensureGoogleDriveTestConnection(t, ctx, pool)
	scope := DriveFolderOperationalLosses
	resource := googleDriveCheckpointResource + string(scope)
	fileID := fmt.Sprintf("file-ci-%x", time.Now().UnixNano())
	missingFileID := fileID + "-old"
	newFileID := fileID + "-new"
	nestedFileID := fileID + "-nested"
	childFolderID := fileID + "-freight"
	correlationOne := randomTestUUID(t)
	correlationTwo := randomTestUUID(t)
	previousCursor, previousWatermark, previousCursorExists := readTestCursor(t, ctx, pool, connectionID, resource)

	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM google_drive_sync_page_checkpoints WHERE connection_id = $1 AND resource = $2 AND correlation_id = ANY($3::uuid[])`, connectionID, resource, []pgtype.UUID{correlationOne, correlationTwo})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM google_drive_sync_folders WHERE connection_id = $1 AND resource = $2 AND correlation_id = ANY($3::uuid[])`, connectionID, resource, []pgtype.UUID{correlationOne, correlationTwo})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM google_drive_files WHERE connection_id = $1 AND folder_scope = $2 AND file_id = ANY($3::text[])`, connectionID, string(scope), []string{fileID, missingFileID, newFileID, nestedFileID, childFolderID})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM google_drive_file_versions WHERE connection_id = $1 AND folder_scope = $2 AND file_id = ANY($3::text[])`, connectionID, string(scope), []string{fileID, missingFileID, newFileID, nestedFileID, childFolderID})
		if previousCursorExists {
			_, _ = pool.Exec(cleanupCtx, `UPDATE sync_cursors SET cursor_value = $3, watermark_at = $4 WHERE connection_id = $1 AND resource = $2`, connectionID, resource, previousCursor, previousWatermark)
		} else {
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connectionID, resource)
		}
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_sync_batches WHERE connection_id = $1 AND resource = $2 AND correlation_id = ANY($3::uuid[])`, connectionID, resource, []pgtype.UUID{correlationOne, correlationTwo})
		if connectionCreated {
			_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_connections WHERE id = $1 AND provider = $2`, connectionID, GoogleDriveProvider)
		}
	}
	defer cleanup()

	var mu sync.Mutex
	phase := 1
	pageCalls := map[string]int{}
	callsFor := func(key string) int {
		mu.Lock()
		defer mu.Unlock()
		return pageCalls[key]
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/drive/v3/files" {
			t.Errorf("request = %s %s, want GET /drive/v3/files", request.Method, request.URL.Path)
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		folderQuery := request.URL.Query().Get("q")
		mu.Lock()
		currentPhase := phase
		token := request.URL.Query().Get("pageToken")
		pageCalls[fmt.Sprintf("%d:%s:%s", currentPhase, folderQuery, token)]++
		mu.Unlock()

		writer.Header().Set("Content-Type", "application/json")
		switch folderQuery {
		case "'folder-a' in parents and trashed = false":
			switch currentPhase {
			case 1:
				if token == "" {
					writeDriveFilePageMany(writer, "next-one", testDriveFile(fileID, "version-one", "1"), testDriveFolder(childFolderID, "folder-a"))
					return
				}
				writeDriveFilePage(writer, "", testDriveFile(missingFileID, "missing-old", "1"))
			case 2:
				if token == "" {
					writeDriveFilePageMany(writer, "next-two", testDriveFile(fileID, "version-two-staged", "2"), testDriveFolder(childFolderID, "folder-a"))
					return
				}
				writer.WriteHeader(http.StatusBadRequest)
			case 3:
				if token != "next-two" {
					t.Errorf("resumed page token = %q, want next-two", token)
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				writeDriveFilePage(writer, "", testDriveFile(newFileID, "new-file", "1"))
			default:
				writer.WriteHeader(http.StatusInternalServerError)
			}
		case "'" + childFolderID + "' in parents and trashed = false":
			if currentPhase == 1 || currentPhase == 3 {
				writeDriveFilePage(writer, "", testDriveFileWithParent(nestedFileID, "nested-shipping", "1", childFolderID))
				return
			}
			writer.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("unexpected folder query = %q", folderQuery)
			writer.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	client, err := newClient(server.URL+"/drive/v3", AccessTokenFunc(func(context.Context) (string, error) { return "test-token", nil }), testFolderIDs, server.Client())
	if err != nil {
		t.Fatalf("create test Drive client: %v", err)
	}
	service, err := NewFileInventorySyncService(pool, client)
	if err != nil {
		t.Fatalf("create inventory sync service: %v", err)
	}

	first, err := service.SyncFolder(ctx, scope, testFolderIDs[0], correlationOne)
	if err != nil || first.Status != "SUCCESS" || first.PagesRead != 3 || first.FilesRead != 4 || first.FilesCreated != 4 || first.FilesUpdated != 0 {
		t.Fatalf("initial snapshot = %#v, error=%v", first, err)
	}
	firstBatch := loadTestBatchID(t, ctx, pool, first.BatchID)
	initialCursor, _, cursorExists := readTestCursor(t, ctx, pool, connectionID, resource)
	if !cursorExists || initialCursor != first.BatchID {
		t.Fatalf("initial cursor = %q, exists=%t; want batch %q", initialCursor, cursorExists, first.BatchID)
	}

	replayed, err := service.SyncFolder(ctx, scope, testFolderIDs[0], correlationOne)
	if err != nil || !replayed.Skipped || replayed.BatchID != first.BatchID || callsFor("1:'folder-a' in parents and trashed = false:") != 1 || callsFor("1:'folder-a' in parents and trashed = false:next-one") != 1 {
		t.Fatalf("completed retry = %#v, error=%v, calls=%v", replayed, err, pageCalls)
	}

	mu.Lock()
	phase = 2
	mu.Unlock()
	failed, err := service.SyncFolder(ctx, scope, testFolderIDs[0], correlationTwo)
	if err == nil || failed.Status != "FAILED" {
		t.Fatalf("interrupted snapshot = %#v, error=%v; want audited failure", failed, err)
	}
	failedCursor, _, _ := readTestCursor(t, ctx, pool, connectionID, resource)
	if failedCursor != first.BatchID {
		t.Fatalf("failed snapshot advanced cursor to %q; want previous batch %q", failedCursor, first.BatchID)
	}
	var stagedPages, visibleFileVersion int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM google_drive_sync_page_checkpoints WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`, connectionID, resource, correlationTwo).Scan(&stagedPages); err != nil || stagedPages != 1 {
		t.Fatalf("failed run checkpoint count = %d, error=%v; want one staged page", stagedPages, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM google_drive_files WHERE connection_id = $1 AND folder_scope = $2 AND file_id = $3 AND last_seen_batch_id = $4`, connectionID, string(scope), fileID, firstBatch).Scan(&visibleFileVersion); err != nil || visibleFileVersion != 1 {
		t.Fatalf("old file snapshot changed during failed run: count=%d error=%v", visibleFileVersion, err)
	}

	mu.Lock()
	phase = 3
	mu.Unlock()
	resumed, err := service.SyncFolder(ctx, scope, testFolderIDs[0], correlationTwo)
	if err != nil || resumed.Status != "SUCCESS" || resumed.PagesRead != 3 || resumed.FilesRead != 4 || resumed.FilesCreated != 1 || resumed.FilesUpdated != 1 {
		t.Fatalf("resumed snapshot = %#v, error=%v", resumed, err)
	}
	if callsFor("3:'folder-a' in parents and trashed = false:") != 0 || callsFor("3:'folder-a' in parents and trashed = false:next-two") != 1 || callsFor("3:'"+childFolderID+"' in parents and trashed = false:") != 1 {
		t.Fatalf("resumed run did not continue from its checkpoint: calls=%v", pageCalls)
	}
	finalCursor, _, cursorExists := readTestCursor(t, ctx, pool, connectionID, resource)
	if !cursorExists || finalCursor != resumed.BatchID {
		t.Fatalf("final cursor = %q, exists=%t; want batch %q", finalCursor, cursorExists, resumed.BatchID)
	}
	var finalSnapshotFiles, staleFileVisible, checkpointCount, fileVersionCount int
	resumedBatch := loadTestBatchID(t, ctx, pool, resumed.BatchID)
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM google_drive_files WHERE connection_id = $1 AND folder_scope = $2 AND last_seen_batch_id = $3`, connectionID, string(scope), resumedBatch).Scan(&finalSnapshotFiles); err != nil || finalSnapshotFiles != 4 {
		t.Fatalf("final snapshot file count=%d error=%v; want 4", finalSnapshotFiles, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM google_drive_files WHERE connection_id = $1 AND folder_scope = $2 AND file_id = $3 AND last_seen_batch_id = $4`, connectionID, string(scope), missingFileID, resumedBatch).Scan(&staleFileVisible); err != nil || staleFileVisible != 0 {
		t.Fatalf("file absent from latest Drive listing is visible: count=%d error=%v", staleFileVisible, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM google_drive_sync_page_checkpoints WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`, connectionID, resource, correlationTwo).Scan(&checkpointCount); err != nil || checkpointCount != 0 {
		t.Fatalf("published run checkpoint count=%d error=%v; want cleanup", checkpointCount, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM google_drive_file_versions WHERE connection_id = $1 AND folder_scope = $2 AND file_id = $3`, connectionID, string(scope), fileID).Scan(&fileVersionCount); err != nil || fileVersionCount != 2 {
		t.Fatalf("file version count=%d error=%v; want both immutable versions", fileVersionCount, err)
	}
}

func ensureGoogleDriveTestConnection(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (pgtype.UUID, bool) {
	t.Helper()
	var id pgtype.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ($1, 'CONNECTED', 'google-drive-sync-test')
		ON CONFLICT (provider) DO NOTHING
		RETURNING id`, GoogleDriveProvider).Scan(&id)
	created := err == nil
	if errors.Is(err, pgx.ErrNoRows) {
		err = pool.QueryRow(ctx, `SELECT id FROM integration_connections WHERE provider = $1`, GoogleDriveProvider).Scan(&id)
	}
	if err != nil || !id.Valid {
		t.Fatalf("prepare Google Drive integration connection: %v", err)
	}
	return id, created
}

func randomTestUUID(t *testing.T) pgtype.UUID {
	t.Helper()
	var value pgtype.UUID
	if _, err := rand.Read(value.Bytes[:]); err != nil {
		t.Fatalf("generate test correlation ID: %v", err)
	}
	value.Valid = true
	return value
}

func readTestCursor(t *testing.T, ctx context.Context, pool *pgxpool.Pool, connectionID pgtype.UUID, resource string) (string, pgtype.Timestamptz, bool) {
	t.Helper()
	var value string
	var watermark pgtype.Timestamptz
	err := pool.QueryRow(ctx, `SELECT cursor_value, watermark_at FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connectionID, resource).Scan(&value, &watermark)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", pgtype.Timestamptz{}, false
	}
	if err != nil {
		t.Fatalf("read sync cursor: %v", err)
	}
	return value, watermark, true
}

func loadTestBatchID(t *testing.T, ctx context.Context, pool *pgxpool.Pool, value string) pgtype.UUID {
	t.Helper()
	var parsed pgtype.UUID
	if err := pool.QueryRow(ctx, `SELECT $1::uuid`, value).Scan(&parsed); err != nil {
		t.Fatalf("load sync batch ID: %v", err)
	}
	return parsed
}

func testDriveFile(id, name, version string) map[string]any {
	return testDriveFileWithParent(id, name, version, "folder-a")
}

func testDriveFileWithParent(id, name, version, parentID string) map[string]any {
	return map[string]any{
		"id": id, "name": name + ".csv", "mimeType": "text/csv", "size": "12",
		"modifiedTime": "2026-10-18T09:00:00Z", "md5Checksum": "0123456789abcdef0123456789abcdef",
		"version": version, "parents": []string{parentID},
	}
}

func testDriveFolder(id, parentID string) map[string]any {
	return map[string]any{
		"id": id, "name": "Frete", "mimeType": googleDriveFolderMIMEType,
		"modifiedTime": "2026-10-18T09:00:00Z", "version": 1, "parents": []string{parentID},
	}
}

func writeDriveFilePage(writer http.ResponseWriter, nextToken string, file map[string]any) {
	writeDriveFilePageMany(writer, nextToken, file)
}

func writeDriveFilePageMany(writer http.ResponseWriter, nextToken string, files ...map[string]any) {
	writer.Header().Set("Content-Type", "application/json")
	items := make([]any, 0, len(files))
	for _, file := range files {
		items = append(items, file)
	}
	payload, _ := json.Marshal(map[string]any{"files": items, "nextPageToken": nextToken})
	_, _ = io.WriteString(writer, string(payload))
}
