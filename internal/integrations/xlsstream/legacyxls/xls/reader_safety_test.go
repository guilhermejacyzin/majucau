package xls

import (
	"encoding/binary"
	"errors"
	"testing"
)

func TestDecodeRejectsWideCellBeforeAllocatingRow(t *testing.T) {
	data := make([]byte, 14)
	binary.LittleEndian.PutUint16(data[2:4], 0xFFFF)
	r := &Reader{rr: &recordReader{typ: recNumber, data: data}}

	_, cells, ok := r.decodeCellRecord()
	if ok || cells != nil || !errors.Is(r.Err(), ErrColumnLimit) {
		t.Fatalf("decode = (ok=%v, cells=%d, err=%v), want rejected before allocation", ok, len(cells), r.Err())
	}
}

func TestDecodeRejectsWideMultiCellRangeBeforeAllocating(t *testing.T) {
	data := make([]byte, 2+2+257*2+2)
	binary.LittleEndian.PutUint16(data[2:4], 0)
	r := &Reader{rr: &recordReader{typ: recMulBlank, data: data}}

	_, cells, ok := r.decodeCellRecord()
	if ok || cells != nil || !errors.Is(r.Err(), ErrColumnLimit) {
		t.Fatalf("decode = (ok=%v, cells=%d, err=%v), want rejected before allocation", ok, len(cells), r.Err())
	}
}

func TestWorkbookRejectsWorksheetCountBeforeGrowingTable(t *testing.T) {
	r := &Reader{sheets: make([]boundSheet, maxSheets)}
	if err := r.addSheet(boundSheet{kind: sheetKindWorksheet}); !errors.Is(err, ErrSheetLimit) {
		t.Fatalf("addSheet error=%v, want ErrSheetLimit", err)
	}
	if len(r.sheets) != maxSheets {
		t.Fatalf("worksheet table grew beyond the limit: %d", len(r.sheets))
	}
}

func TestWorkbookRejectsOversizedStyleTables(t *testing.T) {
	r := &Reader{xfNumFmt: make([]int, maxStyleEntries), formats: make(map[int]string, maxStyleEntries)}
	for id := range maxStyleEntries {
		r.formats[id] = "format"
	}
	if err := r.addXF(0); !errors.Is(err, ErrStyleLimit) || len(r.xfNumFmt) != maxStyleEntries {
		t.Fatalf("addXF error=%v, entries=%d", err, len(r.xfNumFmt))
	}
	if err := r.addFormat(maxStyleEntries, "new"); !errors.Is(err, ErrStyleLimit) || len(r.formats) != maxStyleEntries {
		t.Fatalf("addFormat error=%v, entries=%d", err, len(r.formats))
	}
	if err := r.addFormat(0, "updated"); err != nil || r.formats[0] != "updated" {
		t.Fatalf("existing style update error=%v, value=%q", err, r.formats[0])
	}
}
