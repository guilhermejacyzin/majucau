package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/googledrive"
	"majucau.local/financial-intelligence/internal/integrations/nuvempago"
)

const testNuvemPagoFolderID = "folder-nuvem"

type fakeDailyDriveInventory struct {
	result googledrive.DailyFileInventorySyncResult
	err    error
}

func (f fakeDailyDriveInventory) SyncDaily(context.Context, time.Time) (googledrive.DailyFileInventorySyncResult, error) {
	return f.result, f.err
}

type fakeDriveSnapshotReader struct {
	pages map[string]googledrive.FileInventorySnapshotPage
	err   error
	calls []string
}

func (f *fakeDriveSnapshotReader) ReadSuccessfulSnapshotPage(_ context.Context, scope googledrive.DriveFolderScope, batchID, after string, pageSize int) (googledrive.FileInventorySnapshotPage, error) {
	f.calls = append(f.calls, after)
	if scope != googledrive.DriveFolderNuvemPagoReceipts || batchID != "batch-current" || pageSize != driveImportSnapshotPageSize {
		return googledrive.FileInventorySnapshotPage{}, errDriveFutureSnapshotRead
	}
	if f.err != nil {
		return googledrive.FileInventorySnapshotPage{}, f.err
	}
	return f.pages[after], nil
}

type fakeDrivePrivateStager struct {
	files    map[string]driveStageFixture
	failures map[string]error
	calls    []string
}

type driveStageFixture struct {
	file    googledrive.File
	content []byte
}

func (f *fakeDrivePrivateStager) WithPrivateStage(ctx context.Context, _ string, fileID string, maxBytes int64, consume func(context.Context, googledrive.File, io.Reader, int64, string) error) (googledrive.StreamResult, error) {
	f.calls = append(f.calls, fileID)
	if err := ctx.Err(); err != nil {
		return googledrive.StreamResult{}, err
	}
	if err := f.failures[fileID]; err != nil {
		return googledrive.StreamResult{}, err
	}
	fixture, exists := f.files[fileID]
	if !exists {
		return googledrive.StreamResult{}, errDriveFutureStaging
	}
	if int64(len(fixture.content)) > maxBytes {
		return googledrive.StreamResult{}, googledrive.ErrFileTooLarge
	}
	digest := sha256.Sum256(fixture.content)
	sha := hex.EncodeToString(digest[:])
	err := consume(ctx, fixture.file, bytes.NewReader(fixture.content), int64(len(fixture.content)), sha)
	if err != nil {
		return googledrive.StreamResult{}, err
	}
	return googledrive.StreamResult{File: fixture.file, Bytes: int64(len(fixture.content)), SHA256: sha}, nil
}

type fakeStagedFutureImporter struct {
	importCalls   int
	pendingCalls  int
	pending       nuvempago.FutureImportPending
	pendingResult nuvempago.FutureImportResult
	importResult  nuvempago.FutureImportResult
	importErr     error
	recorded      nuvempago.FutureImportPending
	stagedFiles   []string
}

func (f *fakeStagedFutureImporter) ImportStaged(_ context.Context, folder string, pending nuvempago.FutureImportPending, cleanup func() error) (nuvempago.FutureImportResult, error) {
	f.importCalls++
	f.pending = pending
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nuvempago.FutureImportResult{}, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			f.stagedFiles = append(f.stagedFiles, entry.Name())
		}
	}
	if cleanup != nil {
		if err := cleanup(); err != nil {
			return nuvempago.FutureImportResult{}, err
		}
	}
	if _, err := os.Stat(folder); !errors.Is(err, os.ErrNotExist) {
		return nuvempago.FutureImportResult{}, errors.New("staging folder was not removed before commit")
	}
	return f.importResult, f.importErr
}

func (f *fakeStagedFutureImporter) RecordPending(_ context.Context, pending nuvempago.FutureImportPending) (nuvempago.FutureImportResult, error) {
	f.pendingCalls++
	f.recorded = pending
	return f.pendingResult, nil
}

func successfulDriveInventory(status string, failed int) googledrive.DailyFileInventorySyncResult {
	operationalStatus := "SUCCESS"
	if failed > 0 {
		operationalStatus = "FAILED"
	}
	return googledrive.DailyFileInventorySyncResult{
		Status: status, FoldersRead: 3 - failed, FoldersFailed: failed,
		FolderResults: []googledrive.FileInventorySyncResult{
			{Scope: googledrive.DriveFolderOperationalLosses, Status: operationalStatus, BatchID: "batch-losses"},
			{Scope: googledrive.DriveFolderNuvemPagoReceipts, Status: "SUCCESS", BatchID: "batch-current"},
			{Scope: googledrive.DriveFolderNuvemShipping, Status: "SUCCESS", BatchID: "batch-shipping"},
		},
	}
}

func driveSnapshotFile(id, name, modifiedAt string, content []byte) (googledrive.File, driveStageFixture) {
	md5sum := md5.Sum(content)
	file := googledrive.File{
		ID: id, Name: name, MIMEType: "text/csv", Size: int64(len(content)), SizeKnown: true,
		ModifiedTime: modifiedAt, MD5Checksum: hex.EncodeToString(md5sum[:]), Version: "7",
		Parents: []string{testNuvemPagoFolderID}, ScopeRootID: testNuvemPagoFolderID,
	}
	actual := file
	return file, driveStageFixture{file: actual, content: content}
}

func TestDailyDrivePipelineImportsAvailableCSVAndKeepsOtherScopesIndependent(t *testing.T) {
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	csvBytes := []byte("sanitized streaming fixture")
	onTime, onTimeData := driveSnapshotFile("file-a", "report.csv", "2026-10-18T09:20:00-03:00", csvBytes)
	late, _ := driveSnapshotFile("file-b", "late.csv", "2026-10-18T09:31:00-03:00", []byte("late"))
	unsupported, _ := driveSnapshotFile("file-c", "sheet.xlsx", "2026-10-18T09:10:00-03:00", []byte("unsupported"))
	inventoryErr := errors.New("one independent folder failed")
	inventory := fakeDailyDriveInventory{result: successfulDriveInventory("PARTIAL", 1), err: inventoryErr}
	reader := &fakeDriveSnapshotReader{pages: map[string]googledrive.FileInventorySnapshotPage{
		"":       {Files: []googledrive.File{onTime, late}, NextFileID: "file-b", HasMore: true},
		"file-b": {Files: []googledrive.File{unsupported}, HasMore: false},
	}}
	stager := &fakeDrivePrivateStager{files: map[string]driveStageFixture{"file-a": onTimeData}}
	importer := &fakeStagedFutureImporter{importResult: nuvempago.FutureImportResult{Status: "PARTIAL", RecordsRead: 2, RecordsCreated: 2, PendingFiles: 2}}
	syncer, err := newDailyGoogleDriveImportSyncer(inventory, reader, stager, importer, t.TempDir(), testNuvemPagoFolderID)
	if err != nil {
		t.Fatal(err)
	}

	result, err := syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, googledrive.ErrDailySyncIncomplete) || !errors.Is(err, inventoryErr) {
		t.Fatalf("daily import error = %v, want partial independent source failure", err)
	}
	if result.Status != "PARTIAL" || result.FutureStatus != "PARTIAL" || result.FuturePending != 2 || result.FutureRead != 2 || importer.importCalls != 1 || importer.pendingCalls != 0 {
		t.Fatalf("daily import result=%+v importer=%+v", result, importer)
	}
	if len(stager.calls) != 1 || stager.calls[0] != "file-a" {
		t.Fatalf("staged IDs = %#v, want only the on-time CSV", stager.calls)
	}
	if len(reader.calls) != 2 || reader.calls[0] != "" || reader.calls[1] != "file-b" {
		t.Fatalf("snapshot page cursors = %#v", reader.calls)
	}
	if importer.pending.DeferredFiles != 1 || importer.pending.UnsupportedFiles != 1 || importer.pending.NoCompatibleCSV {
		t.Fatalf("pending source summary = %+v", importer.pending)
	}
	if len(importer.stagedFiles) != 1 || importer.stagedFiles[0] != "00000000000000000001.csv" {
		t.Fatalf("private staged names = %#v", importer.stagedFiles)
	}
}

func TestDailyDrivePipelineReportsInvalidMetadataAlongsideImportedCSV(t *testing.T) {
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	valid, fixture := driveSnapshotFile("file-valid", "receipts.csv", "2026-10-18T09:00:00-03:00", []byte("csv"))
	invalid, _ := driveSnapshotFile("file-invalid", "unknown.csv", "not-a-date", []byte("not downloaded"))
	reader := &fakeDriveSnapshotReader{pages: map[string]googledrive.FileInventorySnapshotPage{"": {Files: []googledrive.File{valid, invalid}}}}
	stager := &fakeDrivePrivateStager{files: map[string]driveStageFixture{"file-valid": fixture}}
	importer := &fakeStagedFutureImporter{importResult: nuvempago.FutureImportResult{Status: "PARTIAL", RecordsRead: 1, RecordsCreated: 1, PendingFiles: 1}}
	syncer, err := newDailyGoogleDriveImportSyncer(fakeDailyDriveInventory{result: successfulDriveInventory("SUCCESS", 0)}, reader, stager, importer, t.TempDir(), testNuvemPagoFolderID)
	if err != nil {
		t.Fatal(err)
	}

	result, err := syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, googledrive.ErrDailySyncIncomplete) || result.Status != "PARTIAL" || result.FuturePending != 1 {
		t.Fatalf("daily import result=%+v error=%v", result, err)
	}
	if importer.importCalls != 1 || importer.pending.InvalidMetadataFiles != 1 || len(stager.calls) != 1 || stager.calls[0] != "file-valid" {
		t.Fatalf("importCalls=%d pending=%+v staged=%#v", importer.importCalls, importer.pending, stager.calls)
	}
}

func TestDailyDrivePipelineDefersLateCSVUntilFollowingSchedule(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	firstSchedule := time.Date(2026, 10, 18, 9, 30, 0, 0, location)
	file, fixture := driveSnapshotFile("file-late", "daily.csv", "2026-10-18T09:31:00-03:00", []byte("late content"))
	inventory := fakeDailyDriveInventory{result: successfulDriveInventory("SUCCESS", 0)}
	reader := &fakeDriveSnapshotReader{pages: map[string]googledrive.FileInventorySnapshotPage{"": {Files: []googledrive.File{file}}}}
	stager := &fakeDrivePrivateStager{files: map[string]driveStageFixture{"file-late": fixture}}
	importer := &fakeStagedFutureImporter{pendingResult: nuvempago.FutureImportResult{Status: "PARTIAL", SourceMissing: true, PendingFiles: 1}}
	syncer, err := newDailyGoogleDriveImportSyncer(inventory, reader, stager, importer, t.TempDir(), testNuvemPagoFolderID)
	if err != nil {
		t.Fatal(err)
	}
	firstResult, firstErr := syncer.SyncDaily(context.Background(), firstSchedule)
	if !errors.Is(firstErr, googledrive.ErrDailySyncIncomplete) || firstResult.Status != "PARTIAL" || importer.pendingCalls != 1 || importer.importCalls != 0 || len(stager.calls) != 0 {
		t.Fatalf("first schedule result=%+v error=%v importer=%+v staged=%#v", firstResult, firstErr, importer, stager.calls)
	}
	if !importer.recorded.NoCompatibleCSV || importer.recorded.DeferredFiles != 1 {
		t.Fatalf("first schedule pending = %+v", importer.recorded)
	}

	importer.pendingCalls = 0
	importer.importResult = nuvempago.FutureImportResult{Status: "SUCCESS", RecordsRead: 1, RecordsCreated: 1}
	nextSchedule := firstSchedule.Add(24 * time.Hour)
	secondResult, secondErr := syncer.SyncDaily(context.Background(), nextSchedule)
	if secondErr != nil || secondResult.Status != "SUCCESS" || importer.importCalls != 1 || len(stager.calls) != 1 {
		t.Fatalf("following schedule result=%+v error=%v importer=%+v staged=%#v", secondResult, secondErr, importer, stager.calls)
	}
}

func TestDailyDrivePipelineDoesNotReuseFileOlderThanPreviousSchedule(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, location)
	stale, _ := driveSnapshotFile("file-stale", "old-report.csv", "2026-10-17T09:30:00-03:00", []byte("old report"))
	reader := &fakeDriveSnapshotReader{pages: map[string]googledrive.FileInventorySnapshotPage{"": {Files: []googledrive.File{stale}}}}
	stager := &fakeDrivePrivateStager{files: map[string]driveStageFixture{}}
	importer := &fakeStagedFutureImporter{pendingResult: nuvempago.FutureImportResult{Status: "PARTIAL", SourceMissing: true, PendingFiles: 1, StaleFiles: 1}}
	syncer, err := newDailyGoogleDriveImportSyncer(fakeDailyDriveInventory{result: successfulDriveInventory("SUCCESS", 0)}, reader, stager, importer, t.TempDir(), testNuvemPagoFolderID)
	if err != nil {
		t.Fatal(err)
	}

	result, err := syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, googledrive.ErrDailySyncIncomplete) || result.Status != "PARTIAL" || result.FutureStale != 1 || result.FuturePending != 1 {
		t.Fatalf("daily import result=%+v error=%v", result, err)
	}
	if importer.pendingCalls != 1 || importer.importCalls != 0 || importer.recorded.StaleFiles != 1 || !importer.recorded.NoCompatibleCSV || len(stager.calls) != 0 {
		t.Fatalf("importer=%+v staged=%#v", importer, stager.calls)
	}
}

func TestCopyVerifiedDriveFileRejectsMovedOrChangedFileAndRemovesPartialCopy(t *testing.T) {
	content := []byte("private csv bytes")
	expected, actualFixture := driveSnapshotFile("file-a", "input.csv", "2026-10-18T09:00:00Z", content)
	destination := filepath.Join(t.TempDir(), "staged.csv")
	actualFixture.file.Parents = []string{"different-authorized-folder"}
	digest := sha256.Sum256(content)
	err := copyVerifiedDriveFile(destination, expected, actualFixture.file, testNuvemPagoFolderID, bytes.NewReader(content), int64(len(content)), hex.EncodeToString(digest[:]))
	if !errors.Is(err, errDriveFutureFileMismatch) {
		t.Fatalf("moved file error = %v, want metadata mismatch", err)
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("mismatched staging file was retained: %v", err)
	}
}

func TestCopyVerifiedDriveFileCopiesAndVerifiesInChunks(t *testing.T) {
	content := bytes.Repeat([]byte("chunk"), 8192)
	expected, actual := driveSnapshotFile("file-a", "input.csv", "2026-10-18T09:00:00Z", content)
	destination := filepath.Join(t.TempDir(), "staged.csv")
	digest := sha256.Sum256(content)
	if err := copyVerifiedDriveFile(destination, expected, actual.file, testNuvemPagoFolderID, bytes.NewReader(content), int64(len(content)), hex.EncodeToString(digest[:])); err != nil {
		t.Fatalf("copyVerifiedDriveFile() error = %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil || !bytes.Equal(got, content) {
		t.Fatalf("staged bytes match=%v error=%v", bytes.Equal(got, content), err)
	}
}
