package googledrive

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/dbtx"
)

const (
	GoogleDriveProvider           = "GOOGLE_DRIVE"
	maxSyncPagesPerCall           = 64
	maxCheckpointPayloadBytes     = 1 << 20
	googleDriveCheckpointResource = "google_drive.files."
)

var (
	ErrSyncDatabaseUnavailable = errors.New("Google Drive sync database is unavailable")
	ErrSyncClientRequired      = errors.New("Google Drive sync client is required")
	ErrSyncConnectionMissing   = errors.New("Google Drive sync connection is not configured")
	ErrSyncConfiguration       = errors.New("Google Drive sync configuration is invalid")
	ErrSyncCheckpointInvalid   = errors.New("Google Drive page checkpoint is invalid")
	ErrSyncCheckpointConflict  = errors.New("Google Drive page checkpoint conflicts with a previous read")
	ErrSyncCheckpointStore     = errors.New("Google Drive page checkpoint could not be stored")
	ErrSyncDuplicateFile       = errors.New("Google Drive listed the same file more than once")
	ErrSyncPageLimit           = errors.New("Google Drive listing did not advance to another page")
)

// DriveFolderScope is a stable internal key. The associated Drive IDs are
// supplied through protected worker configuration and are never logged.
type DriveFolderScope string

const (
	DriveFolderOperationalLosses DriveFolderScope = "01_perdas_operacionais"
	DriveFolderNuvemPagoReceipts DriveFolderScope = "02_nuvem_pago/recebimentos_futuros"
	DriveFolderNuvemShipping     DriveFolderScope = "03_nuvem_envio"
)

var approvedDriveFolderScopes = []DriveFolderScope{
	DriveFolderOperationalLosses,
	DriveFolderNuvemPagoReceipts,
	DriveFolderNuvemShipping,
}

func (s DriveFolderScope) valid() bool {
	switch s {
	case DriveFolderOperationalLosses, DriveFolderNuvemPagoReceipts, DriveFolderNuvemShipping:
		return true
	default:
		return false
	}
}

type FileInventorySyncResult struct {
	BatchID      string
	Scope        DriveFolderScope
	Status       string
	PagesRead    int
	FilesRead    int64
	FilesCreated int64
	FilesUpdated int64
	Skipped      bool
}

type FileInventorySyncService struct {
	pool   *pgxpool.Pool
	client *Client
}

func NewFileInventorySyncService(pool *pgxpool.Pool, client *Client) (*FileInventorySyncService, error) {
	if pool == nil {
		return nil, ErrSyncDatabaseUnavailable
	}
	if client == nil {
		return nil, ErrSyncClientRequired
	}
	return &FileInventorySyncService{pool: pool, client: client}, nil
}

// SyncFolder scans one approved folder. The caller must pass the stable
// correlation ID for the scheduled operation so an interrupted run resumes
// from its last committed page. The previous complete inventory stays
// available until the new scan, file versions, batch status and cursor commit
// together.
func (s *FileInventorySyncService) SyncFolder(ctx context.Context, scope DriveFolderScope, folderID string, correlationID pgtype.UUID) (out FileInventorySyncResult, returnErr error) {
	out.Scope = scope
	if s == nil || s.pool == nil {
		return out, ErrSyncDatabaseUnavailable
	}
	if s.client == nil {
		return out, ErrSyncClientRequired
	}
	if ctx == nil || !scope.valid() || !correlationID.Valid || correlationID.Bytes == ([16]byte{}) {
		return out, ErrSyncConfiguration
	}
	if err := s.client.validateFolder(folderID); err != nil {
		return out, ErrSyncConfiguration
	}
	resource := googleDriveCheckpointResource + string(scope)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return out, fmt.Errorf("begin Google Drive sync: %w", err)
	}
	defer func() {
		if rollbackErr := dbtx.Rollback(ctx, tx); rollbackErr != nil {
			out = FileInventorySyncResult{Scope: scope}
			returnErr = errors.Join(returnErr, rollbackErr)
		}
	}()
	queries := database.New(tx)
	connection, err := queries.GetIntegrationConnectionByProvider(ctx, GoogleDriveProvider)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, ErrSyncConnectionMissing
		}
		return out, fmt.Errorf("load Google Drive sync connection: %w", err)
	}
	if !connection.ID.Valid {
		return out, ErrSyncConnectionMissing
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, driveSyncLockKey(connection.ID, resource)); err != nil {
		return out, fmt.Errorf("lock Google Drive folder sync: %w", err)
	}
	var completedBatchID pgtype.UUID
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM integration_sync_batches
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND status = 'SUCCESS'
		ORDER BY finished_at DESC, id DESC
		LIMIT 1`, connection.ID, resource, correlationID).Scan(&completedBatchID)
	if err == nil {
		out.BatchID = completedBatchID.String()
		out.Status = "SUCCESS"
		out.Skipped = true
		return out, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return out, fmt.Errorf("check completed Google Drive sync: %w", err)
	}
	batch, err := queries.InsertIntegrationSyncBatch(ctx, database.InsertIntegrationSyncBatchParams{ConnectionID: connection.ID, Resource: resource})
	if err != nil {
		return out, fmt.Errorf("create Google Drive sync batch: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE integration_sync_batches SET correlation_id = $1 WHERE id = $2`, correlationID, batch.ID); err != nil {
		return out, fmt.Errorf("set Google Drive sync correlation: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM google_drive_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id <> $3`, connection.ID, resource, correlationID); err != nil {
		return out, fmt.Errorf("remove superseded Google Drive checkpoints: %w", err)
	}
	out.BatchID = batch.ID.String()
	out.Status = "RUNNING"
	if _, err := tx.Exec(ctx, "SAVEPOINT google_drive_sync_data"); err != nil {
		return out, fmt.Errorf("create Google Drive sync savepoint: %w", err)
	}
	fail := func(code string, cause error) (FileInventorySyncResult, error) {
		out.Status = "FAILED"
		out.FilesCreated = 0
		out.FilesUpdated = 0
		return out, s.finishFailed(ctx, tx, queries, batch.ID, out, code, cause)
	}

	checkpoint, err := loadDriveCheckpointSummary(ctx, tx, connection.ID, resource, correlationID)
	if err != nil {
		return fail("GOOGLE_DRIVE_CHECKPOINT_INVALID", err)
	}
	out.PagesRead = checkpoint.PageCount
	out.FilesRead = checkpoint.FileCount
	nextToken := checkpoint.NextPageToken
	complete := checkpoint.PageCount > 0 && !checkpoint.LastHasNext
	for !complete {
		if err := ctx.Err(); err != nil {
			return fail("SYNC_CANCELLED", err)
		}
		pageToken := nextToken
		acceptedPages := 0
		walk, walkErr := s.client.WalkFilePages(ctx, folderID, pageToken, maxSyncPagesPerCall, func(page FilePage, nextPageToken string) error {
			if len(page.Files) > maxPageSize || page.NextPageToken != nextPageToken {
				return ErrSyncCheckpointInvalid
			}
			payload, marshalErr := marshalDriveCheckpointFiles(page.Files)
			if marshalErr != nil {
				return ErrSyncCheckpointInvalid
			}
			pageStartToken := pageToken
			pageNumber := checkpoint.PageCount + 1
			if err := persistDrivePageCheckpoint(ctx, s.pool, connection.ID, resource, correlationID, pageNumber, pageStartToken, nextPageToken, len(page.Files), payload); err != nil {
				return err
			}
			acceptedPages++
			checkpoint.PageCount++
			checkpoint.FileCount += int64(len(page.Files))
			checkpoint.LastHasNext = nextPageToken != ""
			checkpoint.NextPageToken = nextPageToken
			pageToken = nextPageToken
			return nil
		})
		out.PagesRead = checkpoint.PageCount
		out.FilesRead = checkpoint.FileCount
		nextToken = walk.NextPageToken
		if walk.Complete {
			complete = true
			break
		}
		if walkErr != nil && !errors.Is(walkErr, ErrFileWalkLimit) {
			if errors.Is(walkErr, ErrFileVisitor) {
				if errors.Is(walkErr, ErrSyncCheckpointConflict) {
					return fail("GOOGLE_DRIVE_CHECKPOINT_CONFLICT", ErrSyncCheckpointConflict)
				}
				if errors.Is(walkErr, ErrSyncCheckpointInvalid) {
					return fail("GOOGLE_DRIVE_CHECKPOINT_INVALID", ErrSyncCheckpointInvalid)
				}
				return fail("GOOGLE_DRIVE_CHECKPOINT_FAILED", ErrSyncCheckpointStore)
			}
			return fail(driveReadErrorCode(walkErr), walkErr)
		}
		if acceptedPages == 0 || nextToken == "" {
			return fail("GOOGLE_DRIVE_PAGINATION_INVALID", ErrSyncPageLimit)
		}
	}

	if checkpoint.PageCount < 1 || checkpoint.LastHasNext || checkpoint.NextPageToken != "" {
		return fail("GOOGLE_DRIVE_CHECKPOINT_INVALID", ErrSyncCheckpointInvalid)
	}
	created, updated, err := s.publishSnapshot(ctx, tx, connection.ID, batch.ID, resource, scope, folderID, correlationID, checkpoint.PageCount)
	if err != nil {
		return fail("GOOGLE_DRIVE_SNAPSHOT_FAILED", err)
	}
	out.FilesCreated = created
	out.FilesUpdated = updated
	cursorValue := batch.ID.String()
	watermark := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	if err := queries.UpsertSyncCursor(ctx, database.UpsertSyncCursorParams{ConnectionID: connection.ID, Resource: resource, CursorValue: &cursorValue, WatermarkAt: watermark}); err != nil {
		return fail("GOOGLE_DRIVE_CURSOR_FAILED", err)
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM google_drive_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`, connection.ID, resource, correlationID); err != nil {
		return fail("GOOGLE_DRIVE_CHECKPOINT_CLEANUP_FAILED", err)
	}
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batch.ID, Status: "SUCCESS", RecordsRead: out.FilesRead, RecordsCreated: out.FilesCreated, RecordsUpdated: out.FilesUpdated,
	}); err != nil {
		return fail("GOOGLE_DRIVE_BATCH_FINISH_FAILED", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return out, fmt.Errorf("commit Google Drive file snapshot: %w", err)
	}
	out.Status = "SUCCESS"
	return out, nil
}

type driveCheckpointSummary struct {
	PageCount     int
	FileCount     int64
	LastHasNext   bool
	NextPageToken string
}

type driveCheckpointFile struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	MIMEType     string   `json:"mime_type"`
	Size         int64    `json:"size"`
	SizeKnown    bool     `json:"size_known"`
	ModifiedTime string   `json:"modified_time"`
	MD5Checksum  string   `json:"md5_checksum"`
	Version      string   `json:"version"`
	Trashed      bool     `json:"trashed"`
	Parents      []string `json:"parents"`
}

func marshalDriveCheckpointFiles(files []File) ([]byte, error) {
	if len(files) > maxPageSize {
		return nil, ErrSyncCheckpointInvalid
	}
	items := make([]driveCheckpointFile, 0, len(files))
	for _, file := range files {
		if !validDriveID(file.ID) || file.Trashed || file.MIMEType == "" || len(file.Name) > 4096 || len(file.ModifiedTime) > 128 || len(file.MD5Checksum) > 64 || len(file.Version) > 128 || len(file.Parents) > approvedFolderCount {
			return nil, ErrSyncCheckpointInvalid
		}
		items = append(items, driveCheckpointFile{ID: file.ID, Name: file.Name, MIMEType: file.MIMEType, Size: file.Size, SizeKnown: file.SizeKnown, ModifiedTime: file.ModifiedTime, MD5Checksum: file.MD5Checksum, Version: file.Version, Trashed: file.Trashed, Parents: append([]string(nil), file.Parents...)})
	}
	payload, err := json.Marshal(items)
	if err != nil || len(payload) == 0 || len(payload) > maxCheckpointPayloadBytes {
		return nil, ErrSyncCheckpointInvalid
	}
	return payload, nil
}

func canonicalDriveCheckpointPayload(payload []byte) ([]byte, []driveCheckpointFile, error) {
	var files []driveCheckpointFile
	if len(payload) == 0 || len(payload) > maxCheckpointPayloadBytes {
		return nil, nil, ErrSyncCheckpointInvalid
	}
	if err := json.Unmarshal(payload, &files); err != nil || files == nil || len(files) > maxPageSize {
		return nil, nil, ErrSyncCheckpointInvalid
	}
	canonical, err := json.Marshal(files)
	if err != nil || len(canonical) > maxCheckpointPayloadBytes {
		return nil, nil, ErrSyncCheckpointInvalid
	}
	return canonical, files, nil
}

func persistDrivePageCheckpoint(ctx context.Context, pool *pgxpool.Pool, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID, pageNumber int, startToken, nextToken string, fileCount int, payload []byte) error {
	if pool == nil || !connectionID.Valid || !correlationID.Valid || !driveResourceValid(resource) || pageNumber < 1 || len(startToken) > maxPageTokenBytes || len(nextToken) > maxPageTokenBytes || fileCount < 0 || fileCount > maxPageSize {
		return ErrSyncCheckpointInvalid
	}
	canonical, files, err := canonicalDriveCheckpointPayload(payload)
	if err != nil || len(files) != fileCount {
		return ErrSyncCheckpointInvalid
	}
	digest := sha256.Sum256(canonical)
	payloadHash := hex.EncodeToString(digest[:])
	startDigest := sha256.Sum256([]byte(startToken))
	startTokenHash := hex.EncodeToString(startDigest[:])
	hasNext := nextToken != ""
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ErrSyncCheckpointStore
	}
	defer func() { _ = dbtx.Rollback(ctx, tx) }()
	var inserted int
	err = tx.QueryRow(ctx, `
		INSERT INTO google_drive_sync_page_checkpoints (
			connection_id, resource, correlation_id, page_number, start_page_token,
			start_page_token_hash, next_page_token, has_next, file_count, payload_hash, payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
		ON CONFLICT (connection_id, resource, correlation_id, page_number) DO NOTHING
		RETURNING page_number`, connectionID, resource, correlationID, pageNumber, startToken, startTokenHash, nextToken, hasNext, fileCount, payloadHash, canonical).Scan(&inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		var existingStartHash, existingNextToken, existingPayloadHash string
		var existingHasNext bool
		var existingFileCount int
		var existingPayload []byte
		if err := tx.QueryRow(ctx, `
			SELECT start_page_token_hash, next_page_token, has_next, file_count, payload_hash, payload
			FROM google_drive_sync_page_checkpoints
			WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND page_number = $4`, connectionID, resource, correlationID, pageNumber).Scan(&existingStartHash, &existingNextToken, &existingHasNext, &existingFileCount, &existingPayloadHash, &existingPayload); err != nil {
			return ErrSyncCheckpointStore
		}
		existingCanonical, _, canonicalErr := canonicalDriveCheckpointPayload(existingPayload)
		if canonicalErr != nil || existingStartHash != startTokenHash || existingNextToken != nextToken || existingHasNext != hasNext || existingFileCount != fileCount || existingPayloadHash != payloadHash || !jsonBytesEqual(existingCanonical, canonical) {
			return ErrSyncCheckpointConflict
		}
	} else if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrSyncCheckpointConflict
		}
		return ErrSyncCheckpointStore
	}
	if err := tx.Commit(ctx); err != nil {
		return ErrSyncCheckpointStore
	}
	return nil
}

func loadDriveCheckpointSummary(ctx context.Context, tx pgx.Tx, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID) (driveCheckpointSummary, error) {
	if !connectionID.Valid || !correlationID.Valid || !driveResourceValid(resource) {
		return driveCheckpointSummary{}, ErrSyncCheckpointInvalid
	}
	var summary driveCheckpointSummary
	var firstPage, lastPage int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::integer, COALESCE(MIN(page_number), 0), COALESCE(MAX(page_number), 0), COALESCE(SUM(file_count), 0)::bigint
		FROM google_drive_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`, connectionID, resource, correlationID).Scan(&summary.PageCount, &firstPage, &lastPage, &summary.FileCount)
	if err != nil {
		return driveCheckpointSummary{}, ErrSyncCheckpointStore
	}
	if summary.PageCount == 0 {
		return summary, nil
	}
	if firstPage != 1 || lastPage != summary.PageCount || summary.FileCount < 0 {
		return driveCheckpointSummary{}, ErrSyncCheckpointInvalid
	}
	var chainInvalid bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM (
				SELECT page_number, start_page_token,
				       lag(next_page_token) OVER (ORDER BY page_number) AS previous_token
				FROM google_drive_sync_page_checkpoints
				WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3
			) pages
			WHERE (page_number = 1 AND start_page_token <> '')
			   OR (page_number > 1 AND start_page_token IS DISTINCT FROM previous_token)
		)`, connectionID, resource, correlationID).Scan(&chainInvalid); err != nil || chainInvalid {
		return driveCheckpointSummary{}, ErrSyncCheckpointInvalid
	}
	var startToken, startHash string
	if err := tx.QueryRow(ctx, `
		SELECT start_page_token, start_page_token_hash, next_page_token, has_next
		FROM google_drive_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND page_number = $4`, connectionID, resource, correlationID, lastPage).Scan(&startToken, &startHash, &summary.NextPageToken, &summary.LastHasNext); err != nil {
		return driveCheckpointSummary{}, ErrSyncCheckpointStore
	}
	startDigest := sha256.Sum256([]byte(startToken))
	if startHash != hex.EncodeToString(startDigest[:]) || summary.LastHasNext != (summary.NextPageToken != "") {
		return driveCheckpointSummary{}, ErrSyncCheckpointInvalid
	}
	return summary, nil
}

func (s *FileInventorySyncService) publishSnapshot(ctx context.Context, tx pgx.Tx, connectionID, batchID pgtype.UUID, resource string, scope DriveFolderScope, folderID string, correlationID pgtype.UUID, pageCount int) (created, updated int64, returnErr error) {
	if !connectionID.Valid || !batchID.Valid || !correlationID.Valid || !scope.valid() || !driveResourceValid(resource) || pageCount < 1 {
		return 0, 0, ErrSyncCheckpointInvalid
	}
	for pageNumber := 1; pageNumber <= pageCount; pageNumber++ {
		if err := ctx.Err(); err != nil {
			return 0, 0, err
		}
		var storedStartToken, startTokenHash, nextToken, storedHash string
		var storedHasNext bool
		var storedFileCount int
		var payload []byte
		if err := tx.QueryRow(ctx, `
			SELECT start_page_token, start_page_token_hash, next_page_token, has_next, file_count, payload_hash, payload
			FROM google_drive_sync_page_checkpoints
			WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND page_number = $4`, connectionID, resource, correlationID, pageNumber).Scan(&storedStartToken, &startTokenHash, &nextToken, &storedHasNext, &storedFileCount, &storedHash, &payload); err != nil {
			return 0, 0, ErrSyncCheckpointInvalid
		}
		canonical, files, err := canonicalDriveCheckpointPayload(payload)
		if err != nil || len(files) != storedFileCount {
			return 0, 0, ErrSyncCheckpointInvalid
		}
		digest := sha256.Sum256(canonical)
		startDigest := sha256.Sum256([]byte(storedStartToken))
		if storedHash != hex.EncodeToString(digest[:]) || startTokenHash != hex.EncodeToString(startDigest[:]) || !storedHasNext && nextToken != "" || storedHasNext != (nextToken != "") || pageNumber == 1 && storedStartToken != "" || pageNumber > 1 && len(storedStartToken) == 0 {
			return 0, 0, ErrSyncCheckpointInvalid
		}
		if err := s.validateCheckpointFiles(files, folderID); err != nil {
			return 0, 0, err
		}
		for _, file := range files {
			if err := ctx.Err(); err != nil {
				return 0, 0, err
			}
			metadata := driveFileMetadata{ID: file.ID, Name: file.Name, MIMEType: file.MIMEType, Size: file.Size, SizeKnown: file.SizeKnown, ModifiedTime: file.ModifiedTime, MD5Checksum: file.MD5Checksum, Version: file.Version}
			metadataBytes, err := json.Marshal(metadata)
			if err != nil {
				return 0, 0, ErrSyncCheckpointInvalid
			}
			metadataDigest := sha256.Sum256(metadataBytes)
			metadataHash := hex.EncodeToString(metadataDigest[:])
			var versionID pgtype.UUID
			err = tx.QueryRow(ctx, `
				INSERT INTO google_drive_file_versions (connection_id, folder_scope, file_id, payload_hash, metadata)
				VALUES ($1, $2, $3, $4, $5::jsonb)
				ON CONFLICT (connection_id, folder_scope, file_id, payload_hash) DO NOTHING
				RETURNING id`, connectionID, string(scope), file.ID, metadataHash, metadataBytes).Scan(&versionID)
			if errors.Is(err, pgx.ErrNoRows) {
				err = tx.QueryRow(ctx, `
					SELECT id FROM google_drive_file_versions
					WHERE connection_id = $1 AND folder_scope = $2 AND file_id = $3 AND payload_hash = $4`, connectionID, string(scope), file.ID, metadataHash).Scan(&versionID)
			}
			if err != nil || !versionID.Valid {
				return 0, 0, fmt.Errorf("store Google Drive file metadata version: %w", errors.Join(ErrSyncCheckpointStore, err))
			}
			var previousVersionID, lastSeenBatchID pgtype.UUID
			err = tx.QueryRow(ctx, `
				SELECT current_version_id, last_seen_batch_id FROM google_drive_files
				WHERE connection_id = $1 AND folder_scope = $2 AND file_id = $3 FOR UPDATE`, connectionID, string(scope), file.ID).Scan(&previousVersionID, &lastSeenBatchID)
			if err == nil && lastSeenBatchID == batchID {
				return 0, 0, ErrSyncDuplicateFile
			}
			if errors.Is(err, pgx.ErrNoRows) {
				created++
			} else if err != nil {
				return 0, 0, fmt.Errorf("read current Google Drive file pointer: %w", err)
			} else if previousVersionID != versionID {
				updated++
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO google_drive_files (connection_id, folder_scope, file_id, current_version_id, last_seen_batch_id)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (connection_id, folder_scope, file_id) DO UPDATE
				SET current_version_id = EXCLUDED.current_version_id,
				    last_seen_batch_id = EXCLUDED.last_seen_batch_id,
				    last_seen_at = clock_timestamp()`, connectionID, string(scope), file.ID, versionID, batchID); err != nil {
				return 0, 0, fmt.Errorf("publish Google Drive file metadata pointer: %w", err)
			}
		}
	}
	return created, updated, nil
}

type driveFileMetadata struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MIMEType     string `json:"mime_type"`
	Size         int64  `json:"size"`
	SizeKnown    bool   `json:"size_known"`
	ModifiedTime string `json:"modified_time"`
	MD5Checksum  string `json:"md5_checksum"`
	Version      string `json:"version"`
}

func (s *FileInventorySyncService) validateCheckpointFiles(files []driveCheckpointFile, folderID string) error {
	seen := make(map[string]struct{}, len(files))
	for _, item := range files {
		file := File{ID: item.ID, Name: item.Name, MIMEType: item.MIMEType, Size: item.Size, SizeKnown: item.SizeKnown, ModifiedTime: item.ModifiedTime, MD5Checksum: item.MD5Checksum, Version: item.Version, Trashed: item.Trashed, Parents: item.Parents}
		if !validDriveID(file.ID) || file.Trashed || file.MIMEType == "" || file.Size < 0 || !file.SizeKnown && file.Size != 0 || len(file.Name) > 4096 || len(file.ModifiedTime) > 128 || len(file.MD5Checksum) > 64 || len(file.Version) > 128 || len(file.Parents) > approvedFolderCount || !s.client.fileAllowed(file) {
			return ErrSyncCheckpointInvalid
		}
		if _, exists := seen[file.ID]; exists {
			return ErrSyncDuplicateFile
		}
		seen[file.ID] = struct{}{}
		foundFolder := false
		for _, parent := range file.Parents {
			if parent == folderID {
				foundFolder = true
				break
			}
		}
		if !foundFolder {
			return ErrSyncCheckpointInvalid
		}
	}
	return nil
}

func (s *FileInventorySyncService) finishFailed(ctx context.Context, tx pgx.Tx, queries *database.Queries, batchID pgtype.UUID, result FileInventorySyncResult, code string, cause error) error {
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if _, err := tx.Exec(auditCtx, "ROLLBACK TO SAVEPOINT google_drive_sync_data"); err != nil {
		return fmt.Errorf("%w: rollback Google Drive sync data: %v", cause, err)
	}
	if _, err := tx.Exec(auditCtx, "RELEASE SAVEPOINT google_drive_sync_data"); err != nil {
		return fmt.Errorf("%w: release Google Drive sync savepoint: %v", cause, err)
	}
	message := code
	if err := queries.FinishIntegrationSyncBatch(auditCtx, database.FinishIntegrationSyncBatchParams{ID: batchID, Status: "FAILED", RecordsRead: result.FilesRead, ErrorCode: &code, ErrorMessageSanitized: &message}); err != nil {
		return fmt.Errorf("%w: finish failed Google Drive sync batch: %v", cause, err)
	}
	if err := tx.Commit(auditCtx); err != nil {
		return fmt.Errorf("%w: commit failed Google Drive sync audit: %v", cause, err)
	}
	return cause
}

func driveReadErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrAuthorization):
		return "GOOGLE_DRIVE_AUTHORIZATION_FAILED"
	case errors.Is(err, ErrRateLimited):
		return "GOOGLE_DRIVE_RATE_LIMITED"
	case errors.Is(err, ErrUnavailable):
		return "GOOGLE_DRIVE_UNAVAILABLE"
	case errors.Is(err, ErrInvalidResponse):
		return "GOOGLE_DRIVE_RESPONSE_INVALID"
	case errors.Is(err, ErrFileWalkLimit):
		return "GOOGLE_DRIVE_PAGE_LIMIT"
	default:
		return "GOOGLE_DRIVE_READ_FAILED"
	}
}

func driveResourceValid(resource string) bool {
	if !strings.HasPrefix(resource, googleDriveCheckpointResource) {
		return false
	}
	scope := DriveFolderScope(strings.TrimPrefix(resource, googleDriveCheckpointResource))
	return scope.valid()
}

func driveSyncLockKey(connectionID pgtype.UUID, resource string) int64 {
	hash := sha256.New()
	_, _ = hash.Write([]byte("majucau:google-drive-folder-sync\x00"))
	_, _ = hash.Write(connectionID.Bytes[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(resource))
	return int64(binary.BigEndian.Uint64(hash.Sum(nil)[:8]))
}

func jsonBytesEqual(left, right []byte) bool {
	return string(left) == string(right)
}
