package xlsstream

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func workbookBytes(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestStreamSimpleWorkbook(t *testing.T) {
	data := workbookBytes(t, "simple.xls")
	var rows []Row
	summary, err := Stream(context.Background(), bytes.NewReader(data), int64(len(data)), func(_ context.Context, row Row) error {
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.FileBytes != int64(len(data)) || summary.Sheets != 1 || summary.Rows != 4 || summary.Cells != 12 {
		t.Fatalf("summary = %+v", summary)
	}
	if len(rows) != 4 || rows[0].Sheet != "Sheet1" || rows[0].Number != 0 || len(rows[0].Cells) != 3 || rows[0].Cells[0].Raw != "Name" {
		t.Fatalf("first row = %+v", rows[:min(len(rows), 1)])
	}
	if rows[1].Cells[0].Raw != "Widget" || rows[1].Cells[1].Raw != "10" {
		t.Fatalf("source values were not preserved: %+v", rows[1])
	}
}

func TestStreamMultipleSheets(t *testing.T) {
	data := workbookBytes(t, "multisheet.xls")
	var names []string
	summary, err := Stream(context.Background(), bytes.NewReader(data), int64(len(data)), func(_ context.Context, row Row) error {
		if len(names) <= row.SheetIndex {
			names = append(names, row.Sheet)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.Sheets != 3 || summary.Rows != 5 || summary.Cells != 9 {
		t.Fatalf("summary = %+v", summary)
	}
	if len(names) != 3 || names[0] != "Alpha" || names[1] != "Beta" || names[2] != "Gamma With Spaces" {
		t.Fatalf("sheet names = %#v", names)
	}
}

func TestStreamRejectsMalformedFiles(t *testing.T) {
	tests := map[string][]byte{
		"empty":      nil,
		"not a file": []byte("not an XLS workbook"),
		"truncated":  {0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1},
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Stream(context.Background(), bytes.NewReader(data), int64(len(data)), func(context.Context, Row) error { return nil })
			if !errors.Is(err, ErrInvalidWorkbook) {
				t.Fatalf("error = %v, want ErrInvalidWorkbook", err)
			}
		})
	}
}

func TestStreamRejectsOversizedFileBeforeReading(t *testing.T) {
	reader := &countingReaderAt{}
	_, err := Stream(context.Background(), reader, MaxFileBytes+1, func(context.Context, Row) error { return nil })
	if !errors.Is(err, ErrLimitExceeded) || reader.calls != 0 {
		t.Fatalf("error=%v read calls=%d", err, reader.calls)
	}
}

func TestStreamPreservesCancellation(t *testing.T) {
	data := workbookBytes(t, "simple.xls")
	ctx, cancel := context.WithCancel(context.Background())
	reader := &cancelingReaderAt{source: bytes.NewReader(data), cancel: cancel}
	_, err := Stream(ctx, reader, int64(len(data)), func(context.Context, Row) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestStreamPreservesRowHandlerError(t *testing.T) {
	data := workbookBytes(t, "simple.xls")
	want := errors.New("stop after first row")
	_, err := Stream(context.Background(), bytes.NewReader(data), int64(len(data)), func(context.Context, Row) error { return want })
	if !errors.Is(err, ErrRowHandler) || !errors.Is(err, want) {
		t.Fatalf("error = %v, want handler and source errors", err)
	}
}

func TestStreamRejectsForgedFATCountWithoutLargeAllocation(t *testing.T) {
	data := workbookBytes(t, "simple.xls")
	if len(data) < 48 {
		t.Fatal("fixture is too short for CFB header")
	}
	binary.LittleEndian.PutUint32(data[44:48], math.MaxUint32)
	_, err := Stream(context.Background(), bytes.NewReader(data), int64(len(data)), func(context.Context, Row) error { return nil })
	if !errors.Is(err, ErrInvalidWorkbook) {
		t.Fatalf("error = %v, want ErrInvalidWorkbook", err)
	}
}

func FuzzStreamRejectsMalformedInputSafely(f *testing.F) {
	for _, name := range []string{"simple.xls", "types.xls", "multisheet.xls"} {
		f.Add(workbookBytes(f, name))
	}
	f.Add([]byte("not a workbook"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if int64(len(data)) > MaxFileBytes {
			return
		}
		_, _ = Stream(context.Background(), bytes.NewReader(data), int64(len(data)), func(context.Context, Row) error { return nil })
	})
}

type countingReaderAt struct{ calls int }

func (reader *countingReaderAt) ReadAt([]byte, int64) (int, error) {
	reader.calls++
	return 0, errors.New("unexpected read")
}

type cancelingReaderAt struct {
	source io.ReaderAt
	cancel context.CancelFunc
	calls  int
}

func (reader *cancelingReaderAt) ReadAt(buffer []byte, offset int64) (int, error) {
	reader.calls++
	if reader.calls == 5 {
		reader.cancel()
		return 0, context.Canceled
	}
	return reader.source.ReadAt(buffer, offset)
}
