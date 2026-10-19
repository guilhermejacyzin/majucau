package xlsxstream

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

const (
	contentTypesXML = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
<Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`
	workbookXML = `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets><sheet name="Dados" sheetId="1" r:id="rId1"/><sheet name="Apoio" sheetId="2" r:id="rId2"/></sheets>
</workbook>`
	workbookRelsXML = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="/xl/worksheets/sheet2.xml"/>
</Relationships>`
	rootRelsXML = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`
)

func TestStreamDeliversSparseRowsAndSourceValuesWithoutEvaluatingFormulas(t *testing.T) {
	workbook := validWorkbook(t)
	tempParent := t.TempDir()
	var rows []Row
	summary, err := Stream(context.Background(), bytes.NewReader(workbook), int64(len(workbook)), tempParent, func(_ context.Context, row Row) error {
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	if summary.FileBytes != int64(len(workbook)) || summary.Sheets != 2 || summary.Rows != 4 || summary.Cells != 6 || summary.UncompressedBytes == 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if len(rows) != 4 {
		t.Fatalf("got %d rows, want 4", len(rows))
	}
	if rows[0].Sheet != "Dados" || rows[0].Number != 1 || rows[0].Cells[0].Value != "Descrição" || rows[0].Cells[1].Column != 3 || rows[0].Cells[1].Value != "Valor" {
		t.Fatalf("first row was not decoded as expected: %+v", rows[0])
	}
	formulaRow := rows[1]
	if formulaRow.Number != 2 || len(formulaRow.Cells) != 3 || formulaRow.Cells[1].Value != "100.25" || formulaRow.Cells[2].Formula != "SUM(B2:B2)" || formulaRow.Cells[2].Value != "100.25" {
		t.Fatalf("formula row lost source values or formula metadata: %+v", formulaRow)
	}
	if rows[2].Number != 3 || len(rows[2].Cells) != 0 {
		t.Fatalf("blank row was not preserved: %+v", rows[2])
	}
	if rows[3].Sheet != "Apoio" || rows[3].Cells[0].Type != "b" || rows[3].Cells[0].Value != "1" {
		t.Fatalf("second sheet was not decoded: %+v", rows[3])
	}
	assertDirectoryEmpty(t, tempParent)
}

func TestStreamCleansPrivateIndexWhenHandlerFails(t *testing.T) {
	tempParent := t.TempDir()
	want := errors.New("consumer failure")
	_, err := Stream(context.Background(), bytes.NewReader(validWorkbook(t)), int64(len(validWorkbook(t))), tempParent, func(context.Context, Row) error {
		return want
	})
	if !errors.Is(err, ErrRowHandler) || !errors.Is(err, want) {
		t.Fatalf("Stream() error = %v, want wrapped handler error", err)
	}
	assertDirectoryEmpty(t, tempParent)
}

func TestStreamRejectsMalformedAndExternalWorkbooksWithoutPublishingRows(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]string)
		wantErr error
	}{
		{
			name: "external relationship",
			mutate: func(files map[string]string) {
				files["xl/_rels/workbook.xml.rels"] = strings.Replace(files["xl/_rels/workbook.xml.rels"], `Target="worksheets/sheet1.xml"`, `Target="https://example.invalid/sheet1.xml" TargetMode="External"`, 1)
			},
			wantErr: ErrInvalidWorkbook,
		},
		{
			name: "macro part",
			mutate: func(files map[string]string) {
				files["xl/vbaProject.bin"] = "macro-bytes"
			},
			wantErr: ErrInvalidWorkbook,
		},
		{
			name: "missing referenced shared string",
			mutate: func(files map[string]string) {
				files["xl/worksheets/sheet1.xml"] = strings.Replace(files["xl/worksheets/sheet1.xml"], `<v>0</v>`, `<v>99</v>`, 1)
			},
			wantErr: ErrInvalidWorkbook,
		},
		{
			name: "worksheet path traversal",
			mutate: func(files map[string]string) {
				files["xl/_rels/workbook.xml.rels"] = strings.Replace(files["xl/_rels/workbook.xml.rels"], `Target="worksheets/sheet1.xml"`, `Target="../../outside.xml"`, 1)
			},
			wantErr: ErrInvalidWorkbook,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files := baseWorkbookFiles()
			test.mutate(files)
			data := buildWorkbook(t, files)
			tempParent := t.TempDir()
			rows := 0
			_, err := Stream(context.Background(), bytes.NewReader(data), int64(len(data)), tempParent, func(context.Context, Row) error {
				rows++
				return nil
			})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Stream() error = %v, want %v", err, test.wantErr)
			}
			if rows != 0 {
				t.Fatalf("Stream() published %d rows from rejected workbook", rows)
			}
			assertDirectoryEmpty(t, tempParent)
		})
	}
}

func TestStreamEnforcesCellAndColumnLimits(t *testing.T) {
	tests := []struct {
		name      string
		worksheet string
	}{
		{name: "oversized cell", worksheet: worksheetWithInlineCell("A1", strings.Repeat("x", maxCellBytes+1))},
		{name: "wide column", worksheet: worksheetWithInlineCell("JZ1", "value")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files := baseWorkbookFiles()
			files["xl/worksheets/sheet1.xml"] = test.worksheet
			data := buildWorkbook(t, files)
			tempParent := t.TempDir()
			_, err := Stream(context.Background(), bytes.NewReader(data), int64(len(data)), tempParent, func(context.Context, Row) error { return nil })
			if !errors.Is(err, ErrLimitExceeded) {
				t.Fatalf("Stream() error = %v, want %v", err, ErrLimitExceeded)
			}
			assertDirectoryEmpty(t, tempParent)
		})
	}
}

func TestStreamHonorsCancellationAndRejectsNonZipInput(t *testing.T) {
	tempParent := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	data := validWorkbook(t)
	_, err := Stream(ctx, bytes.NewReader(data), int64(len(data)), tempParent, func(context.Context, Row) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Stream() cancellation error = %v", err)
	}
	_, err = Stream(context.Background(), bytes.NewReader([]byte("not a workbook")), int64(len("not a workbook")), tempParent, func(context.Context, Row) error { return nil })
	if !errors.Is(err, ErrInvalidWorkbook) {
		t.Fatalf("Stream() malformed workbook error = %v", err)
	}
	assertDirectoryEmpty(t, tempParent)
}

func validWorkbook(t *testing.T) []byte {
	t.Helper()
	files := baseWorkbookFiles()
	files["xl/sharedStrings.xml"] = `<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="3" uniqueCount="3">
<si><t>Descrição</t></si><si><t>Valor</t></si><si><r><t>Da</t></r><r><t>dos</t></r><rPh><t>ignored</t></rPh></si>
</sst>`
	files["xl/worksheets/sheet1.xml"] = `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
<row r="1"><c r="A1" t="s"><v>0</v></c><c r="C1" t="inlineStr"><is><t>Valor</t></is></c></row>
<row r="2"><c r="A2" t="s"><v>2</v></c><c r="B2"><v>100.25</v></c><c r="C2"><f>SUM(B2:B2)</f><v>100.25</v></c></row>
<row r="3"></row>
</sheetData></worksheet>`
	files["xl/worksheets/sheet2.xml"] = `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="b"><v>1</v></c></row></sheetData></worksheet>`
	return buildWorkbook(t, files)
}

func baseWorkbookFiles() map[string]string {
	return map[string]string{
		"[Content_Types].xml":        contentTypesXML,
		"_rels/.rels":                rootRelsXML,
		"xl/workbook.xml":            workbookXML,
		"xl/_rels/workbook.xml.rels": workbookRelsXML,
		"xl/worksheets/sheet1.xml":   `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="s"><v>0</v></c></row></sheetData></worksheet>`,
		"xl/worksheets/sheet2.xml":   `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="b"><v>1</v></c></row></sheetData></worksheet>`,
		"xl/sharedStrings.xml":       `<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><si><t>Base</t></si></sst>`,
	}
}

func worksheetWithInlineCell(reference, value string) string {
	return fmt.Sprintf(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="%s" t="inlineStr"><is><t>%s</t></is></c></row></sheetData></worksheet>`, reference, value)
}

func buildWorkbook(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q): %v", name, err)
		}
		if _, err := io.WriteString(entry, content); err != nil {
			t.Fatalf("write %q: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close workbook: %v", err)
	}
	return output.Bytes()
}

func assertDirectoryEmpty(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("ReadDir(%q): %v", directory, err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary data was retained in %q: %v", directory, entries)
	}
}
