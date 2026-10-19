package googledrive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"majucau.local/financial-intelligence/database/gen"
)

const maxSnapshotFilesPerPage = 100

var (
	ErrSnapshotUnavailable = errors.New("Google Drive file snapshot is not current")
	ErrSnapshotInvalid     = errors.New("Google Drive file snapshot is invalid")
)

// FileInventorySnapshotPage contains a bounded page from the last completely
// published inventory for one approved folder. The batch ID pins the page
// sequence to one immutable synchronization result.
type FileInventorySnapshotPage struct {
	Files      []File
	NextFileID string
	HasMore    bool
}

// ReadSuccessfulSnapshotPage reads at most 100 metadata rows and checks that
// batchID is still the current successful cursor. The database connection is
// released before a caller performs any network download.
func (s *FileInventorySyncService) ReadSuccessfulSnapshotPage(
	ctx context.Context,
	scope DriveFolderScope,
	batchID string,
	afterFileID string,
	pageSize int,
) (FileInventorySnapshotPage, error) {
	if s == nil || s.pool == nil || ctx == nil || !scope.valid() || pageSize < 1 || pageSize > maxSnapshotFilesPerPage || len(afterFileID) > 256 {
		return FileInventorySnapshotPage{}, ErrSnapshotInvalid
	}
	parsedBatchID, err := uuid.Parse(strings.TrimSpace(batchID))
	if err != nil {
		return FileInventorySnapshotPage{}, ErrSnapshotInvalid
	}
	var batchUUID pgtype.UUID
	copy(batchUUID.Bytes[:], parsedBatchID[:])
	batchUUID.Valid = true

	connection, err := database.New(s.pool).GetIntegrationConnectionByProvider(ctx, GoogleDriveProvider)
	if err != nil || !connection.ID.Valid {
		return FileInventorySnapshotPage{}, ErrSnapshotUnavailable
	}
	resource := googleDriveCheckpointResource + string(scope)
	var isCurrent bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM sync_cursors cursor
			JOIN integration_sync_batches batch
			  ON batch.connection_id = cursor.connection_id
			 AND batch.resource = cursor.resource
			 AND batch.id::text = cursor.cursor_value
			JOIN integration_connections connection
			  ON connection.id = batch.connection_id
			WHERE cursor.connection_id = $1
			  AND cursor.resource = $2
			  AND cursor.cursor_value = $3
			  AND batch.id = $4
			  AND batch.status = 'SUCCESS'
			  AND connection.provider = $5
		)`, connection.ID, resource, parsedBatchID.String(), batchUUID, GoogleDriveProvider).Scan(&isCurrent); err != nil || !isCurrent {
		return FileInventorySnapshotPage{}, ErrSnapshotUnavailable
	}

	rows, err := s.pool.Query(ctx, `
		SELECT current_file.file_id, file_version.metadata
		FROM google_drive_files current_file
		JOIN google_drive_file_versions file_version
		  ON file_version.id = current_file.current_version_id
		 AND file_version.connection_id = current_file.connection_id
		 AND file_version.folder_scope = current_file.folder_scope
		 AND file_version.file_id = current_file.file_id
		JOIN integration_sync_batches batch
		  ON batch.id = current_file.last_seen_batch_id
		 AND batch.connection_id = current_file.connection_id
		WHERE current_file.connection_id = $1
		  AND current_file.folder_scope = $2
		  AND current_file.last_seen_batch_id = $3
		  AND batch.resource = $4
		  AND batch.status = 'SUCCESS'
		  AND current_file.file_id > $5
		ORDER BY current_file.file_id COLLATE "C"
		LIMIT $6`, connection.ID, string(scope), batchUUID, resource, afterFileID, pageSize+1)
	if err != nil {
		return FileInventorySnapshotPage{}, fmt.Errorf("query Google Drive snapshot page: %w", err)
	}
	defer rows.Close()

	page := FileInventorySnapshotPage{Files: make([]File, 0, pageSize)}
	for rows.Next() {
		var fileID string
		var metadataJSON []byte
		if err := rows.Scan(&fileID, &metadataJSON); err != nil {
			return FileInventorySnapshotPage{}, fmt.Errorf("read Google Drive snapshot metadata: %w", err)
		}
		file, err := decodeSnapshotFile(fileID, metadataJSON)
		if err != nil {
			return FileInventorySnapshotPage{}, err
		}
		if len(page.Files) == pageSize {
			page.HasMore = true
			break
		}
		page.Files = append(page.Files, file)
	}
	if err := rows.Err(); err != nil {
		return FileInventorySnapshotPage{}, fmt.Errorf("iterate Google Drive snapshot metadata: %w", err)
	}
	if len(page.Files) > 0 {
		page.NextFileID = page.Files[len(page.Files)-1].ID
	}
	return page, nil
}

func decodeSnapshotFile(fileID string, metadataJSON []byte) (File, error) {
	if !validDriveID(fileID) || len(metadataJSON) == 0 || len(metadataJSON) > 16<<10 {
		return File{}, ErrSnapshotInvalid
	}
	var metadata driveFileMetadata
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil || metadata.ID != fileID || metadata.Name == "" || len(metadata.Name) > 4096 || metadata.MIMEType == "" || len(metadata.MIMEType) > 256 || metadata.Size < 0 || !metadata.SizeKnown && metadata.Size != 0 || len(metadata.ModifiedTime) == 0 || len(metadata.ModifiedTime) > 128 || len(metadata.MD5Checksum) > 64 || len(metadata.Version) == 0 || len(metadata.Version) > 128 {
		return File{}, ErrSnapshotInvalid
	}
	return File{
		ID: fileID, Name: metadata.Name, MIMEType: metadata.MIMEType, Size: metadata.Size,
		SizeKnown: metadata.SizeKnown, ModifiedTime: metadata.ModifiedTime,
		MD5Checksum: metadata.MD5Checksum, Version: metadata.Version,
	}, nil
}
