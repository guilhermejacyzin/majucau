package reportexport

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestWriteXLSXStreamsRowsAndKeepsExternalTextInert(t *testing.T) {
	rows := [][]string{
		{"Bling", "arquivo & lote", "PARCIAL"},
		{"=HYPERLINK(\"https://example.invalid\")", "linha <2>", "ABERTO"},
	}
	position := 0
	var nextCalls int
	source := RowSourceFunc(func(ctx context.Context) ([]string, bool, error) {
		nextCalls++
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		if position == len(rows) {
			return nil, false, nil
		}
		row := rows[position]
		position++
		return row, true, nil
	})

	var output bytes.Buffer
	err := WriteXLSX(context.Background(), &output, Report{
		SheetName:   "T15 Integrações",
		DisplayDate: "12/10/2026, 09:30",
		Columns:     []string{"Origem", "Lote/arquivo", "Status"},
		Rows:        source,
	})
	if err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	if nextCalls != len(rows)+1 {
		t.Fatalf("source called %d times, want one call per row plus EOF", nextCalls)
	}

	archive, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil {
		t.Fatalf("open xlsx archive: %v", err)
	}
	entries := make(map[string]string, len(archive.File))
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatalf("open workbook entry %q: %v", file.Name, err)
		}
		content, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if err := errors.Join(readErr, closeErr); err != nil {
			t.Fatalf("read workbook entry %q: %v", file.Name, err)
		}
		entries[file.Name] = string(content)
	}
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/worksheets/sheet1.xml"} {
		if _, ok := entries[name]; !ok {
			t.Errorf("workbook entry %q is missing", name)
		}
	}

	var workbook struct {
		Sheets []struct {
			Name string `xml:"name,attr"`
		} `xml:"sheets>sheet"`
	}
	if err := xml.Unmarshal([]byte(entries["xl/workbook.xml"]), &workbook); err != nil {
		t.Fatalf("parse workbook XML: %v", err)
	}
	if len(workbook.Sheets) != 1 || workbook.Sheets[0].Name != "T15 Integrações" {
		t.Fatalf("workbook sheets = %#v", workbook.Sheets)
	}

	var sheet struct {
		Rows []struct {
			Cells []struct {
				Ref  string `xml:"r,attr"`
				Type string `xml:"t,attr"`
				Text struct {
					Value string `xml:"t"`
				} `xml:"is"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := xml.Unmarshal([]byte(entries["xl/worksheets/sheet1.xml"]), &sheet); err != nil {
		t.Fatalf("parse worksheet XML: %v", err)
	}
	if len(sheet.Rows) != 4 {
		t.Fatalf("worksheet has %d rows, want date, header, and two data rows", len(sheet.Rows))
	}
	if got := sheet.Rows[0].Cells[1].Text.Value; got != "12/10/2026, 09:30" {
		t.Fatalf("displayed date = %q", got)
	}
	if got := sheet.Rows[1].Cells[1].Text.Value; got != "Lote/arquivo" {
		t.Fatalf("table header = %q", got)
	}
	formulaCell := sheet.Rows[3].Cells[0]
	if formulaCell.Ref != "A4" || formulaCell.Type != "inlineStr" || formulaCell.Text.Value != rows[1][0] {
		t.Fatalf("external text was not preserved as an inline text cell: %#v", formulaCell)
	}
	if strings.Contains(entries["xl/worksheets/sheet1.xml"], "<f>") {
		t.Fatal("workbook must not contain executable formula cells")
	}
}

func TestWriteXLSXRejectsInvalidRowsAndCellContent(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		columns []string
		row     []string
		want    string
	}{
		{name: "empty date", date: " ", columns: []string{"Descrição"}, row: []string{"value"}, want: "display date"},
		{name: "wrong column count", date: "12/10/2026", columns: []string{"Descrição"}, row: []string{"one", "two"}, want: "does not match"},
		{name: "invalid XML control", date: "12/10/2026", columns: []string{"Descrição"}, row: []string{"bad\x01text"}, want: "invalid XML character"},
		{name: "long cell", date: "12/10/2026", columns: []string{"Descrição"}, row: []string{strings.Repeat("x", MaxCellUTF16+1)}, want: "Excel text limit"},
		{name: "long row", date: "12/10/2026", columns: []string{"A", "B", "C"}, row: []string{strings.Repeat("x", 22000), strings.Repeat("y", 22000), strings.Repeat("z", 22000)}, want: "row byte limit"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := RowSourceFunc(func(context.Context) ([]string, bool, error) {
				return test.row, true, nil
			})
			var output bytes.Buffer
			err := WriteXLSX(context.Background(), &output, Report{
				SheetName:   "T13 Pendências",
				DisplayDate: test.date,
				Columns:     test.columns,
				Rows:        source,
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestWriteXLSXCancelsWhileReading(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	source := RowSourceFunc(func(context.Context) ([]string, bool, error) {
		cancel()
		return []string{"cancel after this row"}, true, nil
	})
	var output bytes.Buffer
	err := WriteXLSX(ctx, &output, Report{
		SheetName:   "T14 Alertas",
		DisplayDate: "12/10/2026",
		Columns:     []string{"Alerta"},
		Rows:        source,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context cancellation", err)
	}
}

func TestWriteXLSXRejectsUnsafeSheetNames(t *testing.T) {
	for _, name := range []string{"", "bad/name", "bad\"name", "'bad", "bad'", "bad\x01name", strings.Repeat("a", maxSheetNameUTF+1)} {
		err := WriteXLSX(context.Background(), io.Discard, Report{
			SheetName:   name,
			DisplayDate: "12/10/2026",
			Columns:     []string{"Descrição"},
			Rows:        RowSourceFunc(func(context.Context) ([]string, bool, error) { return nil, false, nil }),
		})
		if err == nil || !strings.Contains(err.Error(), "sheet name") {
			t.Errorf("sheet name %q error = %v", name, err)
		}
	}
}
