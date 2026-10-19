package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/googledrive"
	"majucau.local/financial-intelligence/internal/integrations/nuvempago"
	"majucau.local/financial-intelligence/internal/integrations/xlsxstream"
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
	mimeType := "text/csv"
	if strings.EqualFold(filepath.Ext(name), ".xlsx") {
		mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	file := googledrive.File{
		ID: id, Name: name, MIMEType: mimeType, Size: int64(len(content)), SizeKnown: true,
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
	validXLSX, validXLSXData := driveSnapshotFile("file-c", "sheet.xlsx", "2026-10-18T09:10:00-03:00", validDriveXLSX(t))
	invalidXLSX, invalidXLSXData := driveSnapshotFile("file-d", "broken.xlsx", "2026-10-18T09:11:00-03:00", []byte("not a workbook"))
	inventoryErr := errors.New("one independent folder failed")
	inventory := fakeDailyDriveInventory{result: successfulDriveInventory("PARTIAL", 1), err: inventoryErr}
	reader := &fakeDriveSnapshotReader{pages: map[string]googledrive.FileInventorySnapshotPage{
		"":       {Files: []googledrive.File{onTime, late}, NextFileID: "file-b", HasMore: true},
		"file-b": {Files: []googledrive.File{validXLSX, invalidXLSX}, HasMore: false},
	}}
	stager := &fakeDrivePrivateStager{files: map[string]driveStageFixture{
		"file-a": onTimeData, "file-c": validXLSXData, "file-d": invalidXLSXData,
	}}
	importer := &fakeStagedFutureImporter{importResult: nuvempago.FutureImportResult{Status: "PARTIAL", RecordsRead: 2, RecordsCreated: 2, PendingFiles: 3}}
	syncer, err := newDailyGoogleDriveImportSyncer(inventory, reader, stager, importer, t.TempDir(), testNuvemPagoFolderID)
	if err != nil {
		t.Fatal(err)
	}

	result, err := syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, googledrive.ErrDailySyncIncomplete) || !errors.Is(err, inventoryErr) {
		t.Fatalf("daily import error = %v, want partial independent source failure", err)
	}
	if result.Status != "PARTIAL" || result.FutureStatus != "PARTIAL" || result.FuturePending != 3 || result.FutureRead != 2 || importer.importCalls != 1 || importer.pendingCalls != 0 {
		t.Fatalf("daily import result=%+v importer=%+v", result, importer)
	}
	if len(stager.calls) != 3 || stager.calls[0] != "file-a" || stager.calls[1] != "file-c" || stager.calls[2] != "file-d" {
		t.Fatalf("staged IDs = %#v, want CSV and both on-time XLSX workbooks", stager.calls)
	}
	if len(reader.calls) != 2 || reader.calls[0] != "" || reader.calls[1] != "file-b" {
		t.Fatalf("snapshot page cursors = %#v", reader.calls)
	}
	if importer.pending.DeferredFiles != 1 || importer.pending.UnsupportedFiles != 2 || importer.pending.NoCompatibleCSV {
		t.Fatalf("pending source summary = %+v", importer.pending)
	}
	if len(importer.stagedFiles) != 1 || importer.stagedFiles[0] != "00000000000000000001.csv" {
		t.Fatalf("private staged names = %#v", importer.stagedFiles)
	}
}

func validDriveXLSX(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	files := map[string]string{
		"[Content_Types].xml":        `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
		"_rels/.rels":                `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":            `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Dados" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":   `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>coluna</t></is></c></row></sheetData></worksheet>`,
	}
	for name, content := range files {
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestDailyDrivePipelineDoesNotDownloadXLSXAboveParserLimit(t *testing.T) {
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	oversized, _ := driveSnapshotFile("file-large", "too-large.xlsx", "2026-10-18T09:00:00-03:00", nil)
	oversized.Size = xlsxstream.MaxFileBytes + 1
	reader := &fakeDriveSnapshotReader{pages: map[string]googledrive.FileInventorySnapshotPage{"": {Files: []googledrive.File{oversized}}}}
	stager := &fakeDrivePrivateStager{files: map[string]driveStageFixture{}}
	importer := &fakeStagedFutureImporter{pendingResult: nuvempago.FutureImportResult{Status: "PARTIAL", SourceMissing: true, PendingFiles: 1}}
	syncer, err := newDailyGoogleDriveImportSyncer(fakeDailyDriveInventory{result: successfulDriveInventory("SUCCESS", 0)}, reader, stager, importer, t.TempDir(), testNuvemPagoFolderID)
	if err != nil {
		t.Fatal(err)
	}

	result, err := syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, googledrive.ErrDailySyncIncomplete) || result.Status != "PARTIAL" || result.FuturePending != 1 {
		t.Fatalf("daily import result=%+v error=%v", result, err)
	}
	if importer.pendingCalls != 1 || importer.recorded.UnsupportedFiles != 1 || len(stager.calls) != 0 {
		t.Fatalf("pending=%+v stage calls=%#v; oversized workbook must stay pending without download", importer.recorded, stager.calls)
	}
}

func TestDailyDrivePipelineStreamsAndKeepsTSVFilesPending(t *testing.T) {
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	valid, validFixture := driveSnapshotFile("file-tsv-valid", "receipts.tsv", "2026-10-18T09:00:00-03:00", []byte("id\tdata\n1\t2026-10-18\n"))
	invalid, invalidFixture := driveSnapshotFile("file-tsv-invalid", "broken.tsv", "2026-10-18T09:01:00-03:00", []byte("id\tdata\n\"unterminated\tvalue"))
	valid.MIMEType, validFixture.file.MIMEType = "text/tab-separated-values", "text/tab-separated-values"
	invalid.MIMEType, invalidFixture.file.MIMEType = "text/tab-separated-values", "text/tab-separated-values"
	reader := &fakeDriveSnapshotReader{pages: map[string]googledrive.FileInventorySnapshotPage{
		"": {Files: []googledrive.File{valid, invalid}},
	}}
	stager := &fakeDrivePrivateStager{files: map[string]driveStageFixture{
		valid.ID: validFixture, invalid.ID: invalidFixture,
	}}
	tempRoot := t.TempDir()
	importer := &fakeStagedFutureImporter{pendingResult: nuvempago.FutureImportResult{Status: "PARTIAL", SourceMissing: true, PendingFiles: 2}}
	syncer, err := newDailyGoogleDriveImportSyncer(fakeDailyDriveInventory{result: successfulDriveInventory("SUCCESS", 0)}, reader, stager, importer, tempRoot, testNuvemPagoFolderID)
	if err != nil {
		t.Fatal(err)
	}

	result, err := syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, googledrive.ErrDailySyncIncomplete) || result.Status != "PARTIAL" || result.FuturePending != 2 {
		t.Fatalf("daily import result=%+v error=%v", result, err)
	}
	if importer.pendingCalls != 1 || importer.importCalls != 0 || importer.recorded.UnsupportedFiles != 2 || !importer.recorded.NoCompatibleCSV {
		t.Fatalf("pending=%+v importer calls=(%d,%d)", importer.recorded, importer.pendingCalls, importer.importCalls)
	}
	if len(stager.calls) != 2 {
		t.Fatalf("TSV stage calls=%#v, want both files to be safely streamed", stager.calls)
	}
	entries, err := os.ReadDir(tempRoot)
	if err != nil || len(entries) != 0 {
		t.Fatalf("private TSV staging remained after processing: entries=%d error=%v", len(entries), err)
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
