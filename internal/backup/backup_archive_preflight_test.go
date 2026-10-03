package backup

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestPreflightArchiveDirectoryAcceptsClassicAndZIP64(t *testing.T) {
	for _, count := range []int{3, 4} {
		classic := writePreflightTestArchive(t, count)
		assertPreflightAndZIPReaderAccept(t, classic)

		zip64 := convertPreflightTestArchiveToZIP64(t, classic, uint64(count))
		assertPreflightAndZIPReaderAccept(t, zip64)
	}
}

func TestPreflightArchiveDirectoryRejectsExcessAndMismatchedEntries(t *testing.T) {
	tooMany := writePreflightTestArchive(t, 5)
	assertPreflightRejects(t, tooMany)

	underreportedWithinLimit := rewritePreflightTestEntryCounts(t, writePreflightTestArchive(t, 4), 3)
	assertPreflightRejects(t, underreportedWithinLimit)

	underreported := rewritePreflightTestEntryCounts(t, tooMany, 4)
	assertPreflightRejects(t, underreported)

	zip64Underreported := convertPreflightTestArchiveToZIP64(t, tooMany, 4)
	assertPreflightRejects(t, zip64Underreported)
}

func TestPreflightArchiveDirectoryRejectsOversizedAndMalformedMetadata(t *testing.T) {
	archive := writePreflightTestArchive(t, 3)
	oversized := expandPreflightTestDirectory(t, archive, maxArchiveDirectoryBytes+1)
	assertPreflightRejects(t, oversized)

	truncatedRecord := append([]byte(nil), archive...)
	eocd := preflightTestEOCDOffset(t, truncatedRecord)
	directoryOffset := int(binary.LittleEndian.Uint32(truncatedRecord[eocd+16 : eocd+20]))
	binary.LittleEndian.PutUint16(truncatedRecord[directoryOffset+28:directoryOffset+30], 0xffff)
	assertPreflightRejects(t, truncatedRecord)

	brokenZIP64 := convertPreflightTestArchiveToZIP64(t, archive, 3)
	eocd = preflightTestEOCDOffset(t, brokenZIP64)
	locatorOffset := eocd - 20
	binary.LittleEndian.PutUint64(brokenZIP64[locatorOffset+8:locatorOffset+16], ^uint64(0))
	assertPreflightRejects(t, brokenZIP64)
}

func assertPreflightAndZIPReaderAccept(t *testing.T, archive []byte) {
	t.Helper()
	if err := preflightArchiveDirectory(bytes.NewReader(archive), int64(len(archive))); err != nil {
		t.Fatalf("preflightArchiveDirectory() rejected valid archive: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("zip.NewReader() rejected valid archive after preflight: %v", err)
	}
	if len(reader.File) != 3 && len(reader.File) != 4 {
		t.Fatalf("zip.NewReader() found %d entries, want 3 or 4", len(reader.File))
	}
}

func assertPreflightRejects(t *testing.T, archive []byte) {
	t.Helper()
	if err := preflightArchiveDirectory(bytes.NewReader(archive), int64(len(archive))); err != ErrInvalidPackage {
		t.Fatalf("preflightArchiveDirectory() error = %v, want ErrInvalidPackage", err)
	}
}

func writePreflightTestArchive(t *testing.T, entryCount int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for index := 0; index < entryCount; index++ {
		entry, err := writer.Create(fmt.Sprintf("entry-%d", index))
		if err != nil {
			t.Fatalf("create ZIP test entry: %v", err)
		}
		if _, err := entry.Write([]byte("payload")); err != nil {
			t.Fatalf("write ZIP test entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close ZIP test writer: %v", err)
	}
	return append([]byte(nil), buffer.Bytes()...)
}

func convertPreflightTestArchiveToZIP64(t *testing.T, archive []byte, declaredCount uint64) []byte {
	t.Helper()
	eocd := preflightTestEOCDOffset(t, archive)
	directorySize := binary.LittleEndian.Uint32(archive[eocd+12 : eocd+16])
	directoryOffset := binary.LittleEndian.Uint32(archive[eocd+16 : eocd+20])
	zip64Offset := uint64(eocd)

	converted := append([]byte(nil), archive[:eocd]...)
	var zip64End [56]byte
	binary.LittleEndian.PutUint32(zip64End[0:4], 0x06064b50)
	binary.LittleEndian.PutUint64(zip64End[4:12], 44)
	binary.LittleEndian.PutUint16(zip64End[12:14], 45)
	binary.LittleEndian.PutUint16(zip64End[14:16], 45)
	binary.LittleEndian.PutUint64(zip64End[24:32], declaredCount)
	binary.LittleEndian.PutUint64(zip64End[32:40], declaredCount)
	binary.LittleEndian.PutUint64(zip64End[40:48], uint64(directorySize))
	binary.LittleEndian.PutUint64(zip64End[48:56], uint64(directoryOffset))
	converted = append(converted, zip64End[:]...)

	var locator [20]byte
	binary.LittleEndian.PutUint32(locator[0:4], 0x07064b50)
	binary.LittleEndian.PutUint64(locator[8:16], zip64Offset)
	binary.LittleEndian.PutUint32(locator[16:20], 1)
	converted = append(converted, locator[:]...)

	var end [22]byte
	binary.LittleEndian.PutUint32(end[0:4], 0x06054b50)
	binary.LittleEndian.PutUint16(end[8:10], 0xffff)
	binary.LittleEndian.PutUint16(end[10:12], 0xffff)
	binary.LittleEndian.PutUint32(end[12:16], 0xffffffff)
	binary.LittleEndian.PutUint32(end[16:20], 0xffffffff)
	return append(converted, end[:]...)
}

func rewritePreflightTestEntryCounts(t *testing.T, archive []byte, count uint16) []byte {
	t.Helper()
	rewritten := append([]byte(nil), archive...)
	eocd := preflightTestEOCDOffset(t, rewritten)
	binary.LittleEndian.PutUint16(rewritten[eocd+8:eocd+10], count)
	binary.LittleEndian.PutUint16(rewritten[eocd+10:eocd+12], count)
	return rewritten
}

func expandPreflightTestDirectory(t *testing.T, archive []byte, targetSize int) []byte {
	t.Helper()
	eocd := preflightTestEOCDOffset(t, archive)
	currentSize := int(binary.LittleEndian.Uint32(archive[eocd+12 : eocd+16]))
	if targetSize <= currentSize {
		t.Fatalf("target directory size %d must exceed current size %d", targetSize, currentSize)
	}
	padding := targetSize - currentSize
	expanded := make([]byte, 0, len(archive)+padding)
	expanded = append(expanded, archive[:eocd]...)
	expanded = append(expanded, make([]byte, padding)...)
	expanded = append(expanded, archive[eocd:]...)
	eocd += padding
	binary.LittleEndian.PutUint32(expanded[eocd+12:eocd+16], uint32(targetSize))
	return expanded
}

func preflightTestEOCDOffset(t *testing.T, archive []byte) int {
	t.Helper()
	for offset := len(archive) - 22; offset >= 0; offset-- {
		if binary.LittleEndian.Uint32(archive[offset:offset+4]) == 0x06054b50 {
			return offset
		}
	}
	t.Fatal("ZIP test archive has no EOCD record")
	return 0
}
