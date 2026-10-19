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
	maxSyncFoldersPerTree         = 10000
	maxSyncFolderDepth            = 32
	maxSyncPagesPerTree           = 10000
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
	ErrSyncTreeLimit           = errors.New("Google Drive tree listing reached the page limit")
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
	if err := discardLegacyDriveCheckpoints(ctx, s.pool, connection.ID, resource, correlationID); err != nil {
		return out, err
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
	if _, err := tx.Exec(ctx, `
		DELETE FROM google_drive_sync_folders
		WHERE connection_id = $1 AND resource = $2 AND correlation_id <> $3`, connection.ID, resource, correlationID); err != nil {
		return out, fmt.Errorf("remove superseded Google Drive folder state: %w", err)
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
	if err := s.client.resetDescendants(folderID); err != nil {
		return fail("GOOGLE_DRIVE_FOLDER_AUTHORIZATION_FAILED", err)
	}
	if err := seedDriveRootFolder(ctx, s.pool, connection.ID, resource, correlationID, folderID); err != nil {
		return fail("GOOGLE_DRIVE_FOLDER_STATE_FAILED", err)
	}
	if err := s.restoreDriveFolderAuthorization(ctx, tx, connection.ID, resource, correlationID, folderID); err != nil {
		return fail("GOOGLE_DRIVE_FOLDER_AUTHORIZATION_FAILED", err)
	}

	checkpoint, err := loadDriveTreeCheckpointSummary(ctx, tx, connection.ID, resource, correlationID)
	if err != nil {
		return fail("GOOGLE_DRIVE_CHECKPOINT_INVALID", err)
	}
	out.PagesRead = checkpoint.PageCount
	out.FilesRead = checkpoint.FileCount
	for !checkpoint.Complete {
		if err := ctx.Err(); err != nil {
			return fail("SYNC_CANCELLED", err)
		}
		if checkpoint.PageCount >= maxSyncPagesPerTree {
			return fail("GOOGLE_DRIVE_PAGE_LIMIT", ErrSyncTreeLimit)
		}
		listingFolder, found, err := nextDriveFolder(ctx, tx, connection.ID, resource, correlationID)
		if err != nil {
			return fail("GOOGLE_DRIVE_FOLDER_STATE_FAILED", err)
		}
		if !found {
			return fail("GOOGLE_DRIVE_CHECKPOINT_INVALID", ErrSyncCheckpointInvalid)
		}
		if listingFolder.RootFolderID != folderID || !s.client.folderBelongsToRoot(listingFolder.FolderID, folderID) {
			return fail("GOOGLE_DRIVE_FOLDER_AUTHORIZATION_FAILED", ErrFileNotAllowed)
		}
		page, readErr := s.client.ListFiles(ctx, listingFolder.FolderID, listingFolder.NextPageToken)
		if readErr != nil {
			return fail(driveReadErrorCode(readErr), readErr)
		}
		payload, marshalErr := marshalDriveCheckpointFiles(page.Files)
		if marshalErr != nil {
			return fail("GOOGLE_DRIVE_CHECKPOINT_INVALID", marshalErr)
		}
		pageNumber := checkpoint.PageCount + 1
		if err := persistDriveTreePage(ctx, s.pool, connection.ID, resource, correlationID, pageNumber, listingFolder, page, payload); err != nil {
			switch {
			case errors.Is(err, ErrSyncCheckpointConflict):
				return fail("GOOGLE_DRIVE_CHECKPOINT_CONFLICT", err)
			case errors.Is(err, ErrSyncCheckpointInvalid):
				return fail("GOOGLE_DRIVE_CHECKPOINT_INVALID", err)
			default:
				return fail("GOOGLE_DRIVE_CHECKPOINT_FAILED", err)
			}
		}
		for _, listed := range page.Files {
			if listed.MIMEType == googleDriveFolderMIMEType {
				if err := s.client.authorizeDescendant(listingFolder.FolderID, folderID, listed.ID); err != nil {
					return fail("GOOGLE_DRIVE_FOLDER_AUTHORIZATION_FAILED", err)
				}
			}
		}
		checkpoint.PageCount++
		checkpoint.FileCount += int64(len(page.Files))
		checkpoint.Complete, err = driveTreeComplete(ctx, tx, connection.ID, resource, correlationID)
		if err != nil {
			return fail("GOOGLE_DRIVE_FOLDER_STATE_FAILED", err)
		}
		out.PagesRead = checkpoint.PageCount
		out.FilesRead = checkpoint.FileCount
	}

	if checkpoint.PageCount < 1 || !checkpoint.Complete {
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
	if _, err := tx.Exec(ctx, `
		DELETE FROM google_drive_sync_folders
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

type driveFolderCheckpoint struct {
	FolderID       string
	RootFolderID   string
	ParentFolderID string
	Depth          int
	Status         string
	NextPageToken  string
}

type driveTreeCheckpointSummary struct {
	PageCount int
	FileCount int64
	Complete  bool
}

func discardLegacyDriveCheckpoints(ctx context.Context, pool *pgxpool.Pool, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID) error {
	if ctx == nil || pool == nil || !connectionID.Valid || !correlationID.Valid || !driveResourceValid(resource) {
		return ErrSyncCheckpointInvalid
	}
	if _, err := pool.Exec(ctx, `
		DELETE FROM google_drive_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3
		  AND listing_folder_id IS NULL`, connectionID, resource, correlationID); err != nil {
		return fmt.Errorf("remove legacy Google Drive checkpoints: %w", ErrSyncCheckpointStore)
	}
	return nil
}

func seedDriveRootFolder(ctx context.Context, pool *pgxpool.Pool, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID, rootFolderID string) error {
	if ctx == nil || pool == nil || !connectionID.Valid || !correlationID.Valid || !driveResourceValid(resource) || !validDriveID(rootFolderID) {
		return ErrSyncCheckpointInvalid
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return driveCheckpointStoreError("begin folder-state transaction", err)
	}
	defer func() { _ = dbtx.Rollback(ctx, tx) }()
	if _, err := tx.Exec(ctx, `
		INSERT INTO google_drive_sync_folders (
			connection_id, resource, correlation_id, root_folder_id, folder_id,
			parent_folder_id, depth, status, next_page_token
		) VALUES ($1, $2, $3, $4, $4, NULL, 0, 'PENDING', '')
		ON CONFLICT (connection_id, resource, correlation_id, folder_id) DO NOTHING`, connectionID, resource, correlationID, rootFolderID); err != nil {
		return ErrSyncCheckpointStore
	}
	var storedRoot string
	var storedParent pgtype.Text
	var depth int
	if err := tx.QueryRow(ctx, `
		SELECT root_folder_id, parent_folder_id, depth
		FROM google_drive_sync_folders
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND folder_id = $4`, connectionID, resource, correlationID, rootFolderID).Scan(&storedRoot, &storedParent, &depth); err != nil || storedRoot != rootFolderID || storedParent.Valid || depth != 0 {
		return ErrSyncCheckpointInvalid
	}
	if err := tx.Commit(ctx); err != nil {
		return ErrSyncCheckpointStore
	}
	return nil
}

func (s *FileInventorySyncService) restoreDriveFolderAuthorization(ctx context.Context, tx pgx.Tx, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID, rootFolderID string) error {
	rows, err := tx.Query(ctx, `
		SELECT folder_id, root_folder_id, parent_folder_id, depth
		FROM google_drive_sync_folders
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3
		ORDER BY depth, folder_id COLLATE "C"`, connectionID, resource, correlationID)
	if err != nil {
		return ErrSyncCheckpointStore
	}
	defer rows.Close()
	seenRoot := false
	count := 0
	for rows.Next() {
		var folderID, storedRoot string
		var parentID pgtype.Text
		var depth int
		if err := rows.Scan(&folderID, &storedRoot, &parentID, &depth); err != nil {
			return ErrSyncCheckpointInvalid
		}
		count++
		if count > maxSyncFoldersPerTree || storedRoot != rootFolderID || !validDriveID(folderID) || depth < 0 || depth > maxSyncFolderDepth {
			return ErrSyncCheckpointInvalid
		}
		if depth == 0 {
			if folderID != rootFolderID || parentID.Valid || seenRoot {
				return ErrSyncCheckpointInvalid
			}
			seenRoot = true
			continue
		}
		if !parentID.Valid || s.client.authorizeDescendant(parentID.String, rootFolderID, folderID) != nil {
			return ErrFileNotAllowed
		}
	}
	if err := rows.Err(); err != nil || !seenRoot {
		return ErrSyncCheckpointInvalid
	}
	return nil
}

func nextDriveFolder(ctx context.Context, tx pgx.Tx, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID) (driveFolderCheckpoint, bool, error) {
	var folder driveFolderCheckpoint
	var parentID pgtype.Text
	err := tx.QueryRow(ctx, `
		SELECT folder_id, root_folder_id, parent_folder_id, depth, status, next_page_token
		FROM google_drive_sync_folders
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND status <> 'DONE'
		ORDER BY CASE status WHEN 'ACTIVE' THEN 0 ELSE 1 END, depth, folder_id COLLATE "C"
		LIMIT 1`, connectionID, resource, correlationID).Scan(&folder.FolderID, &folder.RootFolderID, &parentID, &folder.Depth, &folder.Status, &folder.NextPageToken)
	if errors.Is(err, pgx.ErrNoRows) {
		return driveFolderCheckpoint{}, false, nil
	}
	if err != nil {
		return driveFolderCheckpoint{}, false, ErrSyncCheckpointStore
	}
	if parentID.Valid {
		folder.ParentFolderID = parentID.String
	}
	return folder, true, nil
}

func driveTreeComplete(ctx context.Context, tx pgx.Tx, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID) (bool, error) {
	var incomplete bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM google_drive_sync_folders
			WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND status <> 'DONE'
		)`, connectionID, resource, correlationID).Scan(&incomplete); err != nil {
		return false, ErrSyncCheckpointStore
	}
	return !incomplete, nil
}

func loadDriveTreeCheckpointSummary(ctx context.Context, tx pgx.Tx, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID) (driveTreeCheckpointSummary, error) {
	if !connectionID.Valid || !correlationID.Valid || !driveResourceValid(resource) {
		return driveTreeCheckpointSummary{}, ErrSyncCheckpointInvalid
	}
	var summary driveTreeCheckpointSummary
	var firstPage, lastPage int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::integer, COALESCE(MIN(page_number), 0), COALESCE(MAX(page_number), 0), COALESCE(SUM(file_count), 0)::bigint
		FROM google_drive_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND listing_folder_id IS NOT NULL`, connectionID, resource, correlationID).Scan(&summary.PageCount, &firstPage, &lastPage, &summary.FileCount); err != nil {
		return driveTreeCheckpointSummary{}, ErrSyncCheckpointStore
	}
	if summary.PageCount > 0 && (firstPage != 1 || lastPage != summary.PageCount || summary.FileCount < 0) {
		return driveTreeCheckpointSummary{}, ErrSyncCheckpointInvalid
	}
	var invalid bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM (
				SELECT page.page_number, page.listing_folder_id, page.root_folder_id, page.start_page_token, page.next_page_token,
				       row_number() OVER (PARTITION BY page.listing_folder_id ORDER BY page.page_number) AS folder_page,
				       lag(page.next_page_token) OVER (PARTITION BY page.listing_folder_id ORDER BY page.page_number) AS previous_token,
			       folder.folder_id AS known_folder, folder.root_folder_id AS known_root
				FROM google_drive_sync_page_checkpoints page
				LEFT JOIN google_drive_sync_folders folder
				  ON folder.connection_id = page.connection_id AND folder.resource = page.resource
				 AND folder.correlation_id = page.correlation_id AND folder.folder_id = page.listing_folder_id
				WHERE page.connection_id = $1 AND page.resource = $2 AND page.correlation_id = $3
			) pages
			WHERE listing_folder_id IS NULL OR root_folder_id IS NULL OR known_folder IS NULL
			   OR known_root IS DISTINCT FROM root_folder_id
			   OR (folder_page = 1 AND start_page_token <> '')
			   OR (folder_page > 1 AND start_page_token IS DISTINCT FROM previous_token)
		)`, connectionID, resource, correlationID).Scan(&invalid); err != nil || invalid {
		return driveTreeCheckpointSummary{}, ErrSyncCheckpointInvalid
	}
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM google_drive_sync_folders folder
			LEFT JOIN LATERAL (
				SELECT page_number, has_next, next_page_token
				FROM google_drive_sync_page_checkpoints page
				WHERE page.connection_id = folder.connection_id AND page.resource = folder.resource
				  AND page.correlation_id = folder.correlation_id AND page.listing_folder_id = folder.folder_id
				ORDER BY page_number DESC LIMIT 1
			) latest ON true
			WHERE folder.connection_id = $1 AND folder.resource = $2 AND folder.correlation_id = $3
			  AND (
				(folder.status = 'PENDING' AND latest.page_number IS NOT NULL) OR
				(folder.status = 'ACTIVE' AND (latest.page_number IS NULL OR NOT latest.has_next OR folder.next_page_token IS DISTINCT FROM latest.next_page_token)) OR
				(folder.status = 'DONE' AND (latest.page_number IS NULL OR latest.has_next OR folder.next_page_token <> ''))
			  )
		)`, connectionID, resource, correlationID).Scan(&invalid); err != nil || invalid {
		return driveTreeCheckpointSummary{}, ErrSyncCheckpointInvalid
	}
	complete, err := driveTreeComplete(ctx, tx, connectionID, resource, correlationID)
	if err != nil {
		return driveTreeCheckpointSummary{}, err
	}
	summary.Complete = complete
	return summary, nil
}

func persistDriveTreePage(ctx context.Context, pool *pgxpool.Pool, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID, pageNumber int, folder driveFolderCheckpoint, page FilePage, payload []byte) error {
	if ctx == nil || pool == nil || !connectionID.Valid || !correlationID.Valid || !driveResourceValid(resource) || pageNumber < 1 || !validDriveID(folder.FolderID) || !validDriveID(folder.RootFolderID) || folder.Depth < 0 || folder.Depth > maxSyncFolderDepth || len(folder.NextPageToken) > maxPageTokenBytes || len(page.NextPageToken) > maxPageTokenBytes || len(page.Files) > maxPageSize {
		return ErrSyncCheckpointInvalid
	}
	canonical, files, err := canonicalDriveCheckpointPayload(payload)
	if err != nil || len(files) != len(page.Files) {
		return ErrSyncCheckpointInvalid
	}
	for _, file := range page.Files {
		if file.Trashed || !hasParent(file.Parents, folder.FolderID) {
			return ErrSyncCheckpointInvalid
		}
	}
	digest := sha256.Sum256(canonical)
	payloadHash := hex.EncodeToString(digest[:])
	startHash := sha256.Sum256([]byte(folder.NextPageToken))
	startTokenHash := hex.EncodeToString(startHash[:])
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return driveCheckpointStoreError("begin page transaction", err)
	}
	defer func() { _ = dbtx.Rollback(ctx, tx) }()
	var inserted int
	err = tx.QueryRow(ctx, `
		INSERT INTO google_drive_sync_page_checkpoints (
			connection_id, resource, correlation_id, page_number, start_page_token,
			start_page_token_hash, next_page_token, has_next, file_count, payload_hash, payload,
			listing_folder_id, root_folder_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12, $13)
		ON CONFLICT (connection_id, resource, correlation_id, page_number) DO NOTHING
		RETURNING page_number`, connectionID, resource, correlationID, pageNumber, folder.NextPageToken, startTokenHash, page.NextPageToken, page.NextPageToken != "", len(page.Files), payloadHash, canonical, folder.FolderID, folder.RootFolderID).Scan(&inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		var oldStartHash, oldNext, oldHash, oldFolder, oldRoot string
		var oldHasNext bool
		var oldCount int
		var oldPayload []byte
		if err := tx.QueryRow(ctx, `
			SELECT start_page_token_hash, next_page_token, has_next, file_count, payload_hash, payload, listing_folder_id, root_folder_id
			FROM google_drive_sync_page_checkpoints
			WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND page_number = $4`, connectionID, resource, correlationID, pageNumber).Scan(&oldStartHash, &oldNext, &oldHasNext, &oldCount, &oldHash, &oldPayload, &oldFolder, &oldRoot); err != nil {
			return driveCheckpointStoreError("read existing page", err)
		}
		oldCanonical, _, canonicalErr := canonicalDriveCheckpointPayload(oldPayload)
		if canonicalErr != nil || oldStartHash != startTokenHash || oldNext != page.NextPageToken || oldHasNext != (page.NextPageToken != "") || oldCount != len(page.Files) || oldHash != payloadHash || oldFolder != folder.FolderID || oldRoot != folder.RootFolderID || !jsonBytesEqual(oldCanonical, canonical) {
			return ErrSyncCheckpointConflict
		}
	} else if err != nil {
		return driveCheckpointStoreError("insert page", err)
	}
	var storedRoot string
	var storedParent pgtype.Text
	var storedDepth int
	if err := tx.QueryRow(ctx, `
		SELECT root_folder_id, parent_folder_id, depth FROM google_drive_sync_folders
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND folder_id = $4 FOR UPDATE`, connectionID, resource, correlationID, folder.FolderID).Scan(&storedRoot, &storedParent, &storedDepth); err != nil {
		return driveCheckpointStoreError("lock listing folder", err)
	}
	if storedRoot != folder.RootFolderID || storedDepth != folder.Depth || storedDepth == 0 && storedParent.Valid || storedDepth > 0 && (!storedParent.Valid || storedParent.String != folder.ParentFolderID) {
		return ErrSyncCheckpointInvalid
	}
	newStatus := "DONE"
	if page.NextPageToken != "" {
		newStatus = "ACTIVE"
	}
	command, err := tx.Exec(ctx, `
		UPDATE google_drive_sync_folders
		SET status = $7, next_page_token = $8
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND folder_id = $4
		  AND root_folder_id = $5 AND next_page_token = $6 AND status IN ('PENDING', 'ACTIVE')`, connectionID, resource, correlationID, folder.FolderID, folder.RootFolderID, folder.NextPageToken, newStatus, page.NextPageToken)
	if err != nil {
		return driveCheckpointStoreError("advance listing folder", err)
	}
	if command.RowsAffected() != 1 {
		return ErrSyncCheckpointConflict
	}
	for _, item := range files {
		if item.MIMEType != googleDriveFolderMIMEType {
			continue
		}
		if folder.Depth >= maxSyncFolderDepth {
			return ErrSyncCheckpointInvalid
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO google_drive_sync_folders (
				connection_id, resource, correlation_id, root_folder_id, folder_id,
				parent_folder_id, depth, status, next_page_token
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'PENDING', '')
			ON CONFLICT (connection_id, resource, correlation_id, folder_id) DO NOTHING`, connectionID, resource, correlationID, folder.RootFolderID, item.ID, folder.FolderID, folder.Depth+1); err != nil {
			return driveCheckpointStoreError("add discovered folder", err)
		}
		var childRoot, childParent string
		var childDepth int
		if err := tx.QueryRow(ctx, `
			SELECT root_folder_id, parent_folder_id, depth FROM google_drive_sync_folders
			WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND folder_id = $4`, connectionID, resource, correlationID, item.ID).Scan(&childRoot, &childParent, &childDepth); err != nil {
			return driveCheckpointStoreError("verify discovered folder", err)
		}
		if childRoot != folder.RootFolderID || childParent != folder.FolderID || childDepth != folder.Depth+1 {
			return ErrSyncCheckpointInvalid
		}
	}
	var folderCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM google_drive_sync_folders WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`, connectionID, resource, correlationID).Scan(&folderCount); err != nil {
		return driveCheckpointStoreError("count discovered folders", err)
	}
	if folderCount > maxSyncFoldersPerTree {
		return ErrSyncCheckpointInvalid
	}
	if err := tx.Commit(ctx); err != nil {
		return driveCheckpointStoreError("commit page transaction", err)
	}
	return nil
}

// driveCheckpointStoreError keeps database diagnostics useful without exposing
// constraint values, tokens, file metadata, or other potentially sensitive SQL details.
func driveCheckpointStoreError(stage string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fmt.Errorf("%w: %s (database code %s)", ErrSyncCheckpointStore, stage, pgErr.Code)
	}
	return fmt.Errorf("%w: %s", ErrSyncCheckpointStore, stage)
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
	lastTokenByFolder := make(map[string]string)
	for pageNumber := 1; pageNumber <= pageCount; pageNumber++ {
		if err := ctx.Err(); err != nil {
			return 0, 0, err
		}
		var storedStartToken, startTokenHash, nextToken, storedHash, listingFolderID, rootFolderID string
		var storedHasNext bool
		var storedFileCount int
		var payload []byte
		if err := tx.QueryRow(ctx, `
			SELECT start_page_token, start_page_token_hash, next_page_token, has_next, file_count, payload_hash, payload, listing_folder_id, root_folder_id
			FROM google_drive_sync_page_checkpoints
			WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND page_number = $4`, connectionID, resource, correlationID, pageNumber).Scan(&storedStartToken, &startTokenHash, &nextToken, &storedHasNext, &storedFileCount, &storedHash, &payload, &listingFolderID, &rootFolderID); err != nil {
			return 0, 0, ErrSyncCheckpointInvalid
		}
		canonical, files, err := canonicalDriveCheckpointPayload(payload)
		if err != nil || len(files) != storedFileCount {
			return 0, 0, ErrSyncCheckpointInvalid
		}
		digest := sha256.Sum256(canonical)
		startDigest := sha256.Sum256([]byte(storedStartToken))
		if storedHash != hex.EncodeToString(digest[:]) || startTokenHash != hex.EncodeToString(startDigest[:]) || !storedHasNext && nextToken != "" || storedHasNext != (nextToken != "") {
			return 0, 0, ErrSyncCheckpointInvalid
		}
		if rootFolderID != folderID || !s.client.folderBelongsToRoot(listingFolderID, rootFolderID) {
			return 0, 0, ErrSyncCheckpointInvalid
		}
		previousToken, hasPreviousPage := lastTokenByFolder[listingFolderID]
		if !hasPreviousPage && storedStartToken != "" || hasPreviousPage && storedStartToken != previousToken {
			return 0, 0, ErrSyncCheckpointInvalid
		}
		lastTokenByFolder[listingFolderID] = nextToken
		if err := s.validateCheckpointFiles(files, listingFolderID, rootFolderID); err != nil {
			return 0, 0, err
		}
		for _, file := range files {
			if err := ctx.Err(); err != nil {
				return 0, 0, err
			}
			metadata := driveFileMetadata{ID: file.ID, Name: file.Name, MIMEType: file.MIMEType, Size: file.Size, SizeKnown: file.SizeKnown, ModifiedTime: file.ModifiedTime, MD5Checksum: file.MD5Checksum, Version: file.Version, Parents: append([]string(nil), file.Parents...), ScopeRootID: rootFolderID}
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
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	MIMEType     string   `json:"mime_type"`
	Size         int64    `json:"size"`
	SizeKnown    bool     `json:"size_known"`
	ModifiedTime string   `json:"modified_time"`
	MD5Checksum  string   `json:"md5_checksum"`
	Version      string   `json:"version"`
	Parents      []string `json:"parents"`
	ScopeRootID  string   `json:"scope_root_id"`
}

func (s *FileInventorySyncService) validateCheckpointFiles(files []driveCheckpointFile, listingFolderID, rootFolderID string) error {
	if !validDriveID(listingFolderID) || !validDriveID(rootFolderID) || !s.client.folderBelongsToRoot(listingFolderID, rootFolderID) {
		return ErrSyncCheckpointInvalid
	}
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
			if parent == listingFolderID {
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
	case errors.Is(err, ErrFileWalkLimit), errors.Is(err, ErrSyncTreeLimit):
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
