package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/csvlimit"
	"majucau.local/financial-intelligence/internal/integrations/googledrive"
	"majucau.local/financial-intelligence/internal/integrations/jsonstream"
	"majucau.local/financial-intelligence/internal/integrations/nuvempago"
	"majucau.local/financial-intelligence/internal/integrations/xlsstream"
	"majucau.local/financial-intelligence/internal/integrations/xlsxstream"
	"majucau.local/financial-intelligence/internal/scheduler"
	"majucau.local/financial-intelligence/internal/securetemp"
)

const driveImportSnapshotPageSize = 100

const driveDailyImportGraceDays = 1

var (
	errDriveFutureInventoryMissing = errors.New("Nuvem Pago Drive inventory is unavailable")
	errDriveFutureSnapshotRead     = errors.New("Nuvem Pago Drive inventory could not be read")
	errDriveFutureStaging          = errors.New("Nuvem Pago Drive source could not be staged")
	errDriveFutureFileMismatch     = errors.New("staged Drive source does not match its approved inventory metadata")
	errDriveFutureInvalidTSV       = errors.New("Nuvem Pago TSV structure is invalid or exceeds safety limits")
	errDriveFutureInvalidJSON      = errors.New("Nuvem Pago JSON structure is invalid or exceeds safety limits")
	errDriveFutureInvalidXLS       = errors.New("Nuvem Pago XLS structure is invalid or exceeds safety limits")
	errDriveSourceSnapshotRead     = errors.New("Google Drive source inventory could not be read")
	errDriveSourceValidation       = errors.New("Google Drive source could not be structurally validated")
	errDriveSourceInvalidFile      = errors.New("Google Drive source file is invalid or exceeds safety limits")
)

type dailyDriveInventoryService interface {
	SyncDaily(context.Context, time.Time) (googledrive.DailyFileInventorySyncResult, error)
}

type driveSnapshotPageReader interface {
	ReadSuccessfulSnapshotPage(context.Context, googledrive.DriveFolderScope, string, string, int) (googledrive.FileInventorySnapshotPage, error)
	RecordSourceValidation(context.Context, googledrive.DriveSourceValidationRecord) error
}

type drivePrivateFileStager interface {
	WithPrivateStage(context.Context, string, string, int64, func(context.Context, googledrive.File, io.Reader, int64, string) error) (googledrive.StreamResult, error)
}

type stagedFutureImporter interface {
	ImportStaged(context.Context, string, nuvempago.FutureImportPending, func() error) (nuvempago.FutureImportResult, error)
	RecordPending(context.Context, nuvempago.FutureImportPending) (nuvempago.FutureImportResult, error)
}

type dailyGoogleDriveImportSyncer struct {
	inventory dailyDriveInventoryService
	snapshots driveSnapshotPageReader
	stager    drivePrivateFileStager
	importer  stagedFutureImporter
	tempRoot  string
	folderIDs googledrive.DriveFolderIDs
}

func newDailyGoogleDriveImportSyncer(
	inventory dailyDriveInventoryService,
	snapshots driveSnapshotPageReader,
	stager drivePrivateFileStager,
	importer stagedFutureImporter,
	tempRoot string,
	folderIDs googledrive.DriveFolderIDs,
) (*dailyGoogleDriveImportSyncer, error) {
	if inventory == nil || snapshots == nil || stager == nil || importer == nil || strings.TrimSpace(tempRoot) == "" || googledrive.ValidateDriveFolderIDs(folderIDs) != nil {
		return nil, googledrive.ErrSyncConfiguration
	}
	copyFolderIDs := make(googledrive.DriveFolderIDs, len(folderIDs))
	for scope, folderID := range folderIDs {
		copyFolderIDs[scope] = folderID
	}
	return &dailyGoogleDriveImportSyncer{inventory: inventory, snapshots: snapshots, stager: stager, importer: importer, tempRoot: tempRoot, folderIDs: copyFolderIDs}, nil
}

func (s *dailyGoogleDriveImportSyncer) SyncDaily(ctx context.Context, scheduledAt time.Time) (out googledrive.DailyFileInventorySyncResult, returnErr error) {
	if s == nil || ctx == nil {
		return googledrive.DailyFileInventorySyncResult{Status: "FAILED"}, googledrive.ErrSyncConfiguration
	}
	out, inventoryErr := s.inventory.SyncDaily(ctx, scheduledAt)
	var failures []error
	if inventoryErr != nil {
		failures = append(failures, inventoryErr)
	}
	if err := ctx.Err(); err != nil {
		out.Status = "FAILED"
		return out, errors.Join(googledrive.ErrDailySyncIncomplete, inventoryErr, err)
	}

	folder, found := successfulDriveFolder(out.FolderResults, googledrive.DriveFolderNuvemPagoReceipts)
	if !found {
		out.FutureStatus = "FAILED"
		failures = append(failures, errDriveFutureInventoryMissing)
	} else {
		importResult, importErr := s.importFutureSnapshot(ctx, scheduledAt, folder.BatchID)
		out.FutureStatus = importResult.Status
		if importErr != nil && out.FutureStatus == "" {
			out.FutureStatus = "FAILED"
		}
		out.FutureRead = int64(importResult.RecordsRead)
		out.FutureCreated = int64(importResult.RecordsCreated)
		out.FutureUpdated = int64(importResult.RecordsUpdated)
		out.FutureFailed = int64(importResult.RecordsFailed)
		out.FuturePending = importResult.PendingFiles
		out.FutureStale = importResult.StaleFiles
		if importErr != nil {
			failures = append(failures, importErr)
		} else if importResult.Status != "SUCCESS" {
			failures = append(failures, googledrive.ErrDailySyncIncomplete)
		}
	}

	for _, scope := range []googledrive.DriveFolderScope{googledrive.DriveFolderOperationalLosses, googledrive.DriveFolderNuvemShipping} {
		folder, found := successfulDriveFolder(out.FolderResults, scope)
		validation := googledrive.DriveSourceValidationResult{}
		validationStatus := "FAILED"
		errorCode := "GOOGLE_DRIVE_SOURCE_INVENTORY_UNAVAILABLE"
		inventoryBatchID := ""
		var validationErr error
		if !found {
			if scope == googledrive.DriveFolderOperationalLosses {
				out.OperationalValidation.Status = "FAILED"
			} else {
				out.ShippingValidation.Status = "FAILED"
			}
			validation.Status = "FAILED"
			validationErr = errDriveSourceSnapshotRead
		} else {
			inventoryBatchID = folder.BatchID
			validation, validationErr = s.validateDriveSourceSnapshot(ctx, scheduledAt, scope, folder.BatchID)
			if validationErr == nil {
				// Structure can be checked safely, but these sources have no approved
				// financial contract yet, so their contents must remain pending.
				validationStatus = "PENDING"
				errorCode = "GOOGLE_DRIVE_SOURCE_CONTRACT_PENDING"
			} else {
				validationStatus = validation.Status
				if validationStatus != "PARTIAL" {
					validationStatus = "FAILED"
				}
				errorCode = "GOOGLE_DRIVE_SOURCE_VALIDATION_FAILED"
			}
		}
		if scope == googledrive.DriveFolderOperationalLosses {
			out.OperationalValidation = validation
		} else {
			out.ShippingValidation = validation
		}
		if persistErr := s.snapshots.RecordSourceValidation(ctx, googledrive.DriveSourceValidationRecord{
			ScheduledAt: scheduledAt, Scope: scope, InventoryBatchID: inventoryBatchID,
			Status: validationStatus, ErrorCode: errorCode, Result: validation,
		}); persistErr != nil {
			if scope == googledrive.DriveFolderOperationalLosses {
				out.OperationalValidation.Status = "FAILED"
			} else {
				out.ShippingValidation.Status = "FAILED"
			}
			failures = append(failures, persistErr)
		}
		if validationErr != nil {
			failures = append(failures, validationErr)
		} else if validationStatus == "PENDING" {
			failures = append(failures, googledrive.ErrDailySyncIncomplete)
		}
	}

	if len(failures) > 0 {
		out.Status = "PARTIAL"
		if out.FoldersRead == 0 {
			out.Status = "FAILED"
		}
		return out, errors.Join(append([]error{googledrive.ErrDailySyncIncomplete}, failures...)...)
	}
	return out, nil
}

func successfulDriveFolder(results []googledrive.FileInventorySyncResult, scope googledrive.DriveFolderScope) (googledrive.FileInventorySyncResult, bool) {
	for _, result := range results {
		if result.Scope == scope && result.Status == "SUCCESS" && strings.TrimSpace(result.BatchID) != "" {
			return result, true
		}
	}
	return googledrive.FileInventorySyncResult{}, false
}

func (s *dailyGoogleDriveImportSyncer) importFutureSnapshot(ctx context.Context, scheduledAt time.Time, batchID string) (out nuvempago.FutureImportResult, returnErr error) {
	previousSchedule, err := previousGoogleDriveSchedule(scheduledAt)
	if err != nil {
		return nuvempago.FutureImportResult{}, err
	}
	workspace, err := securetemp.NewWorkspace(s.tempRoot, "majucau-drive-future-")
	if err != nil {
		return nuvempago.FutureImportResult{}, errDriveFutureStaging
	}
	workspaceRemoved := false
	removeWorkspace := func() error {
		if workspaceRemoved {
			return nil
		}
		if err := workspace.RemoveAll(); err != nil {
			return err
		}
		workspaceRemoved = true
		return nil
	}
	defer func() {
		if err := removeWorkspace(); err != nil {
			returnErr = errors.Join(returnErr, err)
			out = nuvempago.FutureImportResult{}
		}
	}()

	var afterFileID string
	var stagedFiles int64
	var stagedSequence int64
	var pending nuvempago.FutureImportPending
	for {
		if err := ctx.Err(); err != nil {
			return nuvempago.FutureImportResult{}, err
		}
		page, err := s.snapshots.ReadSuccessfulSnapshotPage(ctx, googledrive.DriveFolderNuvemPagoReceipts, batchID, afterFileID, driveImportSnapshotPageSize)
		if err != nil {
			return nuvempago.FutureImportResult{}, errors.Join(errDriveFutureSnapshotRead, err)
		}
		for _, file := range page.Files {
			if err := ctx.Err(); err != nil {
				return nuvempago.FutureImportResult{}, err
			}
			if file.MIMEType == "application/vnd.google-apps.folder" {
				continue
			}
			modifiedAt, err := time.Parse(time.RFC3339Nano, file.ModifiedTime)
			if err != nil {
				if !addPendingCount(&pending.InvalidMetadataFiles, 1) {
					return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
				}
				continue
			}
			if modifiedAt.After(scheduledAt) {
				if !addPendingCount(&pending.DeferredFiles, 1) {
					return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
				}
				continue
			}
			if !modifiedAt.After(previousSchedule) {
				if !addPendingCount(&pending.StaleFiles, 1) {
					return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
				}
				continue
			}
			extension := filepath.Ext(file.Name)
			isCSV := strings.EqualFold(extension, ".csv")
			isTSV := strings.EqualFold(extension, ".tsv")
			isJSON := strings.EqualFold(extension, ".json")
			isXLS := strings.EqualFold(extension, ".xls")
			isXLSX := strings.EqualFold(extension, ".xlsx")
			if !isCSV && !isTSV && !isJSON && !isXLS && !isXLSX {
				if !addPendingCount(&pending.UnsupportedFiles, 1) {
					return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
				}
				continue
			}
			if !file.SizeKnown || file.Size < 0 || ((isXLS || isXLSX) && file.Size == 0) || stagedSequence == int64(^uint64(0)>>1) {
				if !addPendingCount(&pending.InvalidMetadataFiles, 1) {
					return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
				}
				continue
			}
			if (isXLS && file.Size > xlsstream.MaxFileBytes) || (isXLSX && file.Size > xlsxstream.MaxFileBytes) || (isJSON && file.Size > jsonstream.MaxFileBytes) {
				if !addPendingCount(&pending.UnsupportedFiles, 1) {
					return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
				}
				continue
			}
			limit := file.Size
			if limit == 0 {
				limit = 1
			}
			stagedSequence++
			stagedExtension := ".csv"
			if isTSV {
				stagedExtension = ".tsv"
			} else if isJSON {
				stagedExtension = ".json"
			} else if isXLS {
				stagedExtension = ".xls"
			} else if isXLSX {
				stagedExtension = ".xlsx"
			}
			stagedName := fmt.Sprintf("%020d%s", stagedSequence, stagedExtension)
			stagedPath := filepath.Join(workspace.Path(), stagedName)
			_, stageErr := s.stager.WithPrivateStage(ctx, s.tempRoot, file.ID, limit,
				func(stageCtx context.Context, actual googledrive.File, reader io.Reader, size int64, sha256Hex string) error {
					if err := stageCtx.Err(); err != nil {
						return err
					}
					if err := copyVerifiedDriveFile(stagedPath, file, actual, s.folderIDs[googledrive.DriveFolderNuvemPagoReceipts], reader, size, sha256Hex); err != nil {
						return err
					}
					if isTSV {
						return validateDriveTSV(stageCtx, stagedPath)
					}
					if isJSON {
						return validateDriveJSON(stageCtx, stagedPath, size)
					}
					if isXLS {
						return validateDriveXLS(stageCtx, stagedPath, size)
					}
					if !isXLSX {
						return nil
					}
					return validateDriveXLSX(stageCtx, stagedPath, workspace.Path(), size)
				},
			)
			if isXLS || isXLSX || isTSV || isJSON {
				if cleanupErr := securetemp.RemoveFile(stagedPath); cleanupErr != nil {
					return nuvempago.FutureImportResult{}, cleanupErr
				}
			}
			if stageErr != nil {
				if ctx.Err() != nil {
					return nuvempago.FutureImportResult{}, ctx.Err()
				}
				var cleanupErr *securetemp.CleanupError
				if errors.As(stageErr, &cleanupErr) {
					return nuvempago.FutureImportResult{}, stageErr
				}
				if isXLSX && (errors.Is(stageErr, xlsxstream.ErrInvalidWorkbook) || errors.Is(stageErr, xlsxstream.ErrLimitExceeded)) {
					if !addPendingCount(&pending.UnsupportedFiles, 1) {
						return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
					}
					continue
				}
				if isXLS && (errors.Is(stageErr, errDriveFutureInvalidXLS) || errors.Is(stageErr, xlsstream.ErrInvalidWorkbook) || errors.Is(stageErr, xlsstream.ErrLimitExceeded)) {
					if !addPendingCount(&pending.UnsupportedFiles, 1) {
						return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
					}
					continue
				}
				if isTSV && (errors.Is(stageErr, errDriveFutureInvalidTSV) || errors.Is(stageErr, csvlimit.ErrLimitExceeded)) {
					if !addPendingCount(&pending.UnsupportedFiles, 1) {
						return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
					}
					continue
				}
				if isJSON && (errors.Is(stageErr, errDriveFutureInvalidJSON) || errors.Is(stageErr, jsonstream.ErrLimitExceeded)) {
					if !addPendingCount(&pending.UnsupportedFiles, 1) {
						return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
					}
					continue
				}
				if !addPendingCount(&pending.DownloadFailedFiles, 1) {
					return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
				}
				continue
			}
			if isCSV {
				stagedFiles++
			} else if !addPendingCount(&pending.UnsupportedFiles, 1) {
				return nuvempago.FutureImportResult{}, nuvempago.ErrFuturePendingInvalid
			}
		}
		if page.HasMore {
			if len(page.Files) == 0 || page.NextFileID == "" || page.NextFileID <= afterFileID {
				return nuvempago.FutureImportResult{}, errDriveFutureSnapshotRead
			}
			afterFileID = page.NextFileID
			continue
		}
		break
	}

	if stagedFiles == 0 {
		pending.NoCompatibleCSV = true
		if err := removeWorkspace(); err != nil {
			return nuvempago.FutureImportResult{}, err
		}
		return s.importer.RecordPending(ctx, pending)
	}
	return s.importer.ImportStaged(ctx, workspace.Path(), pending, removeWorkspace)
}

// validateDriveSourceSnapshot streams files from an approved non-receipts root
// for safety checks only. Since no financial contract is approved for these
// sources, every discovered file remains pending and no row becomes a fact.
func (s *dailyGoogleDriveImportSyncer) validateDriveSourceSnapshot(
	ctx context.Context,
	scheduledAt time.Time,
	scope googledrive.DriveFolderScope,
	batchID string,
) (out googledrive.DriveSourceValidationResult, returnErr error) {
	out.Status = "PENDING"
	if ctx == nil || ctx.Err() != nil || (scope != googledrive.DriveFolderOperationalLosses && scope != googledrive.DriveFolderNuvemShipping) || strings.TrimSpace(batchID) == "" {
		out.Status = "FAILED"
		return out, errors.Join(errDriveSourceSnapshotRead, ctxErr(ctx))
	}
	previousSchedule, err := previousGoogleDriveSchedule(scheduledAt)
	if err != nil {
		out.Status = "FAILED"
		return out, err
	}
	workspace, err := securetemp.NewWorkspace(s.tempRoot, "majucau-drive-validate-")
	if err != nil {
		out.Status = "FAILED"
		return out, errDriveSourceValidation
	}
	workspaceRemoved := false
	removeWorkspace := func() error {
		if workspaceRemoved {
			return nil
		}
		if err := workspace.RemoveAll(); err != nil {
			return err
		}
		workspaceRemoved = true
		return nil
	}
	defer func() {
		if cleanupErr := removeWorkspace(); cleanupErr != nil {
			out = googledrive.DriveSourceValidationResult{Status: "FAILED"}
			returnErr = errors.Join(returnErr, cleanupErr)
		}
	}()

	var afterFileID string
	var sequence int64
	var currentScheduleFileFound bool
	for {
		if err := ctx.Err(); err != nil {
			out.Status = "FAILED"
			return out, err
		}
		page, err := s.snapshots.ReadSuccessfulSnapshotPage(ctx, scope, batchID, afterFileID, driveImportSnapshotPageSize)
		if err != nil {
			out.Status = "FAILED"
			return out, errors.Join(errDriveSourceSnapshotRead, err)
		}
		for _, file := range page.Files {
			if err := ctx.Err(); err != nil {
				out.Status = "FAILED"
				return out, err
			}
			if file.MIMEType == "application/vnd.google-apps.folder" {
				continue
			}
			if !addPendingCount(&out.FilesRead, 1) || !addPendingCount(&out.FilesPending, 1) {
				out.Status = "FAILED"
				return out, errDriveSourceValidation
			}
			modifiedAt, err := time.Parse(time.RFC3339Nano, file.ModifiedTime)
			if err != nil {
				if !addPendingCount(&out.FilesInvalid, 1) {
					out.Status = "FAILED"
					return out, errDriveSourceValidation
				}
				continue
			}
			if modifiedAt.After(scheduledAt) {
				if !addPendingCount(&out.FilesDeferred, 1) {
					out.Status = "FAILED"
					return out, errDriveSourceValidation
				}
				continue
			}
			if !modifiedAt.After(previousSchedule) {
				if !addPendingCount(&out.FilesStale, 1) {
					out.Status = "FAILED"
					return out, errDriveSourceValidation
				}
				continue
			}
			currentScheduleFileFound = true

			extension := strings.ToLower(filepath.Ext(file.Name))
			if !supportedDriveStructureExtension(extension) {
				if !addPendingCount(&out.FilesUnsupported, 1) {
					out.Status = "FAILED"
					return out, errDriveSourceValidation
				}
				continue
			}
			if !file.SizeKnown || file.Size <= 0 || (extension == ".json" && file.Size > jsonstream.MaxFileBytes) || (extension == ".xls" && file.Size > xlsstream.MaxFileBytes) || (extension == ".xlsx" && file.Size > xlsxstream.MaxFileBytes) {
				if !addPendingCount(&out.FilesUnsupported, 1) {
					out.Status = "FAILED"
					return out, errDriveSourceValidation
				}
				continue
			}
			if sequence == int64(^uint64(0)>>1) {
				out.Status = "FAILED"
				return out, errDriveSourceValidation
			}
			sequence++
			stagedPath := filepath.Join(workspace.Path(), fmt.Sprintf("%020d%s", sequence, extension))
			_, stageErr := s.stager.WithPrivateStage(ctx, s.tempRoot, file.ID, file.Size,
				func(stageCtx context.Context, actual googledrive.File, reader io.Reader, size int64, sha256Hex string) error {
					if err := stageCtx.Err(); err != nil {
						return err
					}
					if err := copyVerifiedDriveFile(stagedPath, file, actual, s.folderIDs[scope], reader, size, sha256Hex); err != nil {
						return err
					}
					return validateDriveStructure(stageCtx, extension, stagedPath, workspace.Path(), size)
				},
			)
			if cleanupErr := securetemp.RemoveFile(stagedPath); cleanupErr != nil {
				out.Status = "FAILED"
				return out, cleanupErr
			}
			if stageErr != nil {
				if ctx.Err() != nil {
					out.Status = "FAILED"
					return out, ctx.Err()
				}
				var cleanupErr *securetemp.CleanupError
				if errors.As(stageErr, &cleanupErr) {
					out.Status = "FAILED"
					return out, stageErr
				}
				if errors.Is(stageErr, errDriveSourceInvalidFile) || errors.Is(stageErr, csvlimit.ErrLimitExceeded) || errors.Is(stageErr, xlsxstream.ErrInvalidWorkbook) || errors.Is(stageErr, xlsxstream.ErrLimitExceeded) || errors.Is(stageErr, xlsstream.ErrInvalidWorkbook) || errors.Is(stageErr, xlsstream.ErrLimitExceeded) || errors.Is(stageErr, jsonstream.ErrInvalidJSON) || errors.Is(stageErr, jsonstream.ErrLimitExceeded) {
					if !addPendingCount(&out.FilesInvalid, 1) {
						out.Status = "FAILED"
						return out, errDriveSourceValidation
					}
					continue
				}
				if !addPendingCount(&out.FilesFailed, 1) {
					out.Status = "FAILED"
					return out, errDriveSourceValidation
				}
				out.Status = "PARTIAL"
				return out, errDriveSourceValidation
			}
			if !addPendingCount(&out.FilesValidated, 1) {
				out.Status = "FAILED"
				return out, errDriveSourceValidation
			}
		}
		if page.HasMore {
			if len(page.Files) == 0 || page.NextFileID == "" || page.NextFileID <= afterFileID {
				out.Status = "FAILED"
				return out, errDriveSourceSnapshotRead
			}
			afterFileID = page.NextFileID
			continue
		}
		break
	}
	out.SourceMissing = !currentScheduleFileFound
	if out.SourceMissing || out.FilesPending > 0 || out.FilesInvalid > 0 || out.FilesUnsupported > 0 {
		out.Status = "PENDING"
	}
	if err := removeWorkspace(); err != nil {
		out.Status = "FAILED"
		return out, err
	}
	return out, nil
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func supportedDriveStructureExtension(extension string) bool {
	switch extension {
	case ".csv", ".tsv", ".json", ".xls", ".xlsx":
		return true
	default:
		return false
	}
}

func validateDriveStructure(ctx context.Context, extension, path, tempParent string, size int64) error {
	switch extension {
	case ".csv":
		return validateDriveCSV(ctx, path)
	case ".tsv":
		return validateDriveTSV(ctx, path)
	case ".json":
		return validateDriveJSON(ctx, path, size)
	case ".xls":
		return validateDriveXLS(ctx, path, size)
	case ".xlsx":
		return validateDriveXLSX(ctx, path, tempParent, size)
	default:
		return errDriveSourceInvalidFile
	}
}

// validateDriveCSV verifies bounded CSV record structure only. It does not
// interpret columns or retain cell values for these unapproved source types.
func validateDriveCSV(ctx context.Context, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errDriveSourceValidation
	}
	buffered := bufio.NewReaderSize(driveContextReader{ctx: ctx, source: file}, 32<<10)
	firstRecord, delimiter, sniffErr := readDriveCSVFirstRecord(ctx, buffered)
	if sniffErr != nil {
		return errors.Join(sniffErr, file.Close())
	}
	bounded, err := csvlimit.NewDelimitedReader(io.MultiReader(bytes.NewReader(firstRecord), buffered), delimiter)
	if err != nil {
		return errors.Join(errDriveSourceInvalidFile, err, file.Close())
	}
	reader := csv.NewReader(bounded)
	reader.Comma = rune(delimiter)
	reader.FieldsPerRecord = -1
	seenRecord := false
	var parseErr error
	for {
		if err := ctx.Err(); err != nil {
			parseErr = err
			break
		}
		_, err := reader.Read()
		if err := ctx.Err(); err != nil {
			parseErr = err
			break
		}
		if errors.Is(err, io.EOF) {
			if !seenRecord {
				parseErr = errDriveSourceInvalidFile
			}
			break
		}
		if errors.Is(err, csvlimit.ErrLimitExceeded) {
			parseErr = err
			break
		}
		if err != nil {
			parseErr = errors.Join(errDriveSourceInvalidFile, err)
			break
		}
		seenRecord = true
	}
	return errors.Join(parseErr, file.Close())
}

func readDriveCSVFirstRecord(ctx context.Context, reader *bufio.Reader) ([]byte, byte, error) {
	if ctx == nil || reader == nil {
		return nil, 0, errDriveSourceInvalidFile
	}
	record := make([]byte, 0, 4096)
	var commas, semicolons int
	inQuotes := false
	quotePending := false
	for {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		chunk, readErr := reader.ReadSlice('\n')
		if len(chunk) > csvlimit.MaxRecordBytes-len(record) {
			return nil, 0, csvlimit.ErrLimitExceeded
		}
		record = append(record, chunk...)
		complete := false
		for _, value := range chunk {
			if inQuotes {
				if quotePending {
					if value == '"' {
						quotePending = false
						continue
					}
					inQuotes = false
					quotePending = false
				} else if value == '"' {
					quotePending = true
					continue
				} else {
					continue
				}
			}
			switch value {
			case '"':
				inQuotes = true
			case '\n':
				complete = true
			case ',':
				commas++
			case ';':
				semicolons++
			}
			if complete {
				break
			}
		}
		if complete || errors.Is(readErr, io.EOF) {
			if len(record) == 0 {
				return nil, 0, errDriveSourceInvalidFile
			}
			delimiter := byte(';')
			if commas > semicolons {
				delimiter = ','
			}
			return record, delimiter, nil
		}
		if !errors.Is(readErr, bufio.ErrBufferFull) {
			if readErr != nil {
				return nil, 0, readErr
			}
		}
	}
}

type driveContextReader struct {
	ctx    context.Context
	source io.Reader
}

func (reader driveContextReader) Read(buffer []byte) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	return reader.source.Read(buffer)
}

// validateDriveXLSX confirms that a downloaded workbook is structurally safe
// to read. Rows are discarded one at a time; without an approved financial
// contract, the workbook remains pending and contributes no financial facts.
func validateDriveXLSX(ctx context.Context, path, tempParent string, size int64) error {
	file, err := os.Open(path)
	if err != nil {
		return errDriveFutureStaging
	}
	_, parseErr := xlsxstream.Stream(ctx, file, size, tempParent, func(context.Context, xlsxstream.Row) error { return nil })
	closeErr := file.Close()
	return errors.Join(parseErr, closeErr)
}

// validateDriveXLS reads BIFF8 worksheets one bounded row at a time. Rows are
// discarded because no approved financial contract maps this format yet.
func validateDriveXLS(ctx context.Context, path string, size int64) error {
	file, err := os.Open(path)
	if err != nil {
		return errDriveFutureStaging
	}
	_, parseErr := xlsstream.Stream(ctx, file, size, func(context.Context, xlsstream.Row) error { return nil })
	closeErr := file.Close()
	if errors.Is(parseErr, xlsstream.ErrInvalidWorkbook) || errors.Is(parseErr, xlsstream.ErrLimitExceeded) {
		parseErr = errors.Join(errDriveFutureInvalidXLS, parseErr)
	}
	return errors.Join(parseErr, closeErr)
}

// validateDriveJSON checks JSON syntax and bounded dimensions without keeping
// fields in memory or mapping values to financial facts. JSON remains pending
// until its structure-specific data contract is approved.
func validateDriveJSON(ctx context.Context, path string, size int64) error {
	file, err := os.Open(path)
	if err != nil {
		return errDriveFutureStaging
	}
	_, parseErr := jsonstream.Validate(ctx, file, size)
	closeErr := file.Close()
	if errors.Is(parseErr, jsonstream.ErrInvalidJSON) {
		parseErr = errors.Join(errDriveFutureInvalidJSON, parseErr)
	}
	return errors.Join(parseErr, closeErr)
}

// validateDriveTSV checks quoting, record boundaries, and bounded dimensions
// one row at a time. It deliberately does not map values to financial facts;
// TSV remains pending until its source contract is approved.
func validateDriveTSV(ctx context.Context, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errDriveFutureStaging
	}
	reader := csv.NewReader(csvlimit.NewTSVReader(file))
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	seenRecord := false
	var parseErr error
	for {
		if err := ctx.Err(); err != nil {
			parseErr = err
			break
		}
		_, err := reader.Read()
		if ctx.Err() != nil {
			parseErr = ctx.Err()
			break
		}
		if errors.Is(err, io.EOF) {
			if !seenRecord {
				parseErr = errDriveFutureInvalidTSV
			}
			break
		}
		if errors.Is(err, csvlimit.ErrLimitExceeded) {
			parseErr = err
			break
		}
		if err != nil {
			parseErr = errors.Join(errDriveFutureInvalidTSV, err)
			break
		}
		seenRecord = true
	}
	return errors.Join(parseErr, file.Close())
}

func previousGoogleDriveSchedule(scheduledAt time.Time) (time.Time, error) {
	if scheduledAt.IsZero() {
		return time.Time{}, googledrive.ErrDailySyncScheduleInvalid
	}
	location, err := time.LoadLocation(scheduler.SaoPauloTimeZone)
	if err != nil {
		return time.Time{}, errors.Join(googledrive.ErrDailySyncScheduleInvalid, err)
	}
	local := scheduledAt.In(location)
	if local.Hour() != scheduler.DailySyncHour || local.Minute() != scheduler.DailySyncMinute || local.Second() != 0 || local.Nanosecond() != 0 {
		return time.Time{}, googledrive.ErrDailySyncScheduleInvalid
	}
	return local.AddDate(0, 0, -driveDailyImportGraceDays), nil
}

func addPendingCount(target *int64, amount int64) bool {
	if target == nil || amount < 0 || *target > int64(^uint64(0)>>1)-amount {
		return false
	}
	*target += amount
	return true
}

func copyVerifiedDriveFile(destination string, expected, actual googledrive.File, expectedFolderID string, reader io.Reader, size int64, expectedSHA256 string) error {
	if reader == nil || expectedFolderID == "" || expected.ScopeRootID != expectedFolderID || actual.Trashed || !sameDriveParents(actual.Parents, expected.Parents) || actual.ID != expected.ID || actual.Name != expected.Name || actual.MIMEType != expected.MIMEType || actual.Size != expected.Size || actual.SizeKnown != expected.SizeKnown || actual.ModifiedTime != expected.ModifiedTime || actual.MD5Checksum != expected.MD5Checksum || actual.Version != expected.Version || size != expected.Size || expectedSHA256 == "" {
		return errDriveFutureFileMismatch
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return errDriveFutureStaging
	}
	hasher := sha256.New()
	count, copyErr := io.Copy(io.MultiWriter(file, hasher), reader)
	syncErr := file.Sync()
	closeErr := file.Close()
	if copyErr != nil || syncErr != nil || closeErr != nil || count != size || !equalHash(expectedSHA256, hasher) {
		cleanupErr := securetemp.RemoveFile(destination)
		return errors.Join(errDriveFutureStaging, cleanupErr)
	}
	return nil
}

func equalHash(expected string, actual hash.Hash) bool {
	if actual == nil || len(expected) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	return bytes.Equal(decoded, actual.Sum(nil))
}

func sameDriveParents(left, right []string) bool {
	if len(left) == 0 || len(left) != len(right) {
		return false
	}
	seen := make(map[string]struct{}, len(left))
	for _, parent := range left {
		if parent == "" {
			return false
		}
		seen[parent] = struct{}{}
	}
	for _, parent := range right {
		if _, ok := seen[parent]; !ok {
			return false
		}
	}
	return true
}

var _ dailyGoogleDriveSyncer = (*dailyGoogleDriveImportSyncer)(nil)
var _ driveSnapshotPageReader = (*googledrive.FileInventorySyncService)(nil)
var _ stagedFutureImporter = (*nuvempago.FutureImportService)(nil)
