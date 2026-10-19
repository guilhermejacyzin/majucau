package googledrive

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrDailySyncScheduleInvalid = errors.New("daily Google Drive sync schedule time is invalid")
	ErrDailySyncIncomplete      = errors.New("daily Google Drive inventory did not complete")
)

type FolderInventorySyncer interface {
	SyncFolder(context.Context, DriveFolderScope, string, pgtype.UUID) (FileInventorySyncResult, error)
}

type DailyFileInventorySyncResult struct {
	Status                string
	FoldersRead           int
	FoldersFailed         int
	FoldersSkipped        int
	FilesRead             int64
	FolderResults         []FileInventorySyncResult
	FutureStatus          string
	FutureRead            int64
	FutureCreated         int64
	FutureUpdated         int64
	FutureFailed          int64
	FuturePending         int64
	FutureStale           int64
	OperationalValidation DriveSourceValidationResult
	ShippingValidation    DriveSourceValidationResult
}

// DriveSourceValidationResult contains only bounded, non-sensitive counts.
// It describes structural inspection, not an approved financial import.
type DriveSourceValidationResult struct {
	Status           string
	FilesRead        int64
	FilesValidated   int64
	FilesPending     int64
	FilesInvalid     int64
	FilesFailed      int64
	FilesStale       int64
	FilesDeferred    int64
	FilesUnsupported int64
	SourceMissing    bool
}

// DriveSourceValidationRecord stores only status and bounded counters for a
// source whose financial interpretation is not approved yet.
type DriveSourceValidationRecord struct {
	ScheduledAt      time.Time
	Scope            DriveFolderScope
	InventoryBatchID string
	Status           string
	ErrorCode        string
	Result           DriveSourceValidationResult
}

type DailyFileInventorySyncService struct {
	syncer    FolderInventorySyncer
	folderIDs map[DriveFolderScope]string
}

func NewDailyFileInventorySyncService(syncer FolderInventorySyncer, folderIDs map[DriveFolderScope]string) (*DailyFileInventorySyncService, error) {
	if syncer == nil || ValidateDriveFolderIDs(folderIDs) != nil {
		return nil, ErrSyncConfiguration
	}
	copyIDs := make(map[DriveFolderScope]string, approvedFolderCount)
	for _, scope := range approvedDriveFolderScopes {
		copyIDs[scope] = folderIDs[scope]
	}
	return &DailyFileInventorySyncService{syncer: syncer, folderIDs: copyIDs}, nil
}

// SyncDaily refreshes each approved folder independently. A failure in one
// folder does not prevent the other folders from getting a fresh inventory.
// This step records metadata only; it does not import financial facts.
func (s *DailyFileInventorySyncService) SyncDaily(ctx context.Context, scheduledAt time.Time) (out DailyFileInventorySyncResult, returnErr error) {
	if s == nil || s.syncer == nil || len(s.folderIDs) != approvedFolderCount || ctx == nil {
		return DailyFileInventorySyncResult{Status: "FAILED"}, ErrSyncConfiguration
	}
	if err := ctx.Err(); err != nil {
		return DailyFileInventorySyncResult{Status: "FAILED"}, err
	}
	out.Status = "SUCCESS"
	var failures []error
	for _, scope := range approvedDriveFolderScopes {
		if err := ctx.Err(); err != nil {
			out.FoldersFailed++
			failures = append(failures, err)
			break
		}
		correlationID, err := dailyDriveCorrelationID(scheduledAt, scope)
		if err != nil {
			return DailyFileInventorySyncResult{Status: "FAILED"}, err
		}
		result, err := s.syncer.SyncFolder(ctx, scope, s.folderIDs[scope], pgUUID(correlationID))
		if result.Scope == "" {
			result.Scope = scope
		}
		if result.Status == "" && err != nil {
			result.Status = "FAILED"
		}
		out.FolderResults = append(out.FolderResults, result)
		if err != nil || result.Status != "SUCCESS" {
			out.FoldersFailed++
			if err == nil {
				err = ErrDailySyncIncomplete
			}
			failures = append(failures, err)
			continue
		}
		out.FoldersRead++
		out.FilesRead += result.FilesRead
		if result.Skipped {
			out.FoldersSkipped++
		}
	}
	if len(failures) == 0 {
		return out, nil
	}
	if out.FoldersRead == 0 {
		out.Status = "FAILED"
	} else {
		out.Status = "PARTIAL"
	}
	return out, errors.Join(append([]error{ErrDailySyncIncomplete}, failures...)...)
}

func dailyDriveCorrelationID(scheduledAt time.Time, scope DriveFolderScope) (uuid.UUID, error) {
	if scheduledAt.IsZero() || !scope.valid() {
		return uuid.Nil, ErrDailySyncScheduleInvalid
	}
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return uuid.Nil, errors.Join(ErrDailySyncScheduleInvalid, err)
	}
	local := scheduledAt.In(location)
	if local.Hour() != 9 || local.Minute() != 30 || local.Second() != 0 || local.Nanosecond() != 0 {
		return uuid.Nil, ErrDailySyncScheduleInvalid
	}
	name := fmt.Sprintf("majucau/google-drive/daily-inventory/v1:%s:%s", local.Format("2006-01-02"), scope)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)), nil
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	var bytes [16]byte
	copy(bytes[:], value[:])
	return pgtype.UUID{Bytes: bytes, Valid: true}
}

var _ FolderInventorySyncer = (*FileInventorySyncService)(nil)
