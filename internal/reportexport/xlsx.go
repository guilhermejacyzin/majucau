// Package reportexport writes spreadsheet files without buffering complete
// reports in memory. It does not load data or decide who may export it.
package reportexport

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	// MaxColumns caps the number of table columns supported by one report.
	MaxColumns = 64
	// MaxRowBytes caps the combined UTF-8 bytes emitted for one row.
	MaxRowBytes = 64 << 10
	// MaxCellUTF16 is Excel's maximum text length for one cell.
	MaxCellUTF16    = 32767
	maxSheetNameUTF = 31
	maxSheetRows    = 1_048_576
	firstDataRow    = 3 // Row 1 carries the displayed date; row 2 is the table header.
)

// RowSource supplies one table row at a time. Implementations must enforce
// MaxColumns, MaxRowBytes, and MaxCellUTF16 before materializing a row. The
// writer checks these limits after Next returns, which bounds output but cannot
// prevent a faulty source from allocating an oversized database value first.
// Database-backed sources should use bounded projections and database-side
// length checks. Implementations should keep their source open only for the
// export and release it on error or when no more rows remain.
type RowSource interface {
	Next(context.Context) ([]string, bool, error)
}

// RowSourceFunc adapts a function into a RowSource.
type RowSourceFunc func(context.Context) ([]string, bool, error)

func (f RowSourceFunc) Next(ctx context.Context) ([]string, bool, error) {
	return f(ctx)
}

// Report contains only the date and the exact visible table headings and rows.
// DisplayDate must be the same date text shown on screen; values are written
// as text so Excel cannot execute content supplied by an external record.
type Report struct {
	SheetName   string
	DisplayDate string
	Columns     []string
	Rows        RowSource
}

// WriteXLSX writes a single-sheet XLSX workbook to dst while requesting one
// data row at a time. Callers must write to a staged output and publish it only
// after this function succeeds; a source or write failure leaves an unusable
// partial package that must be discarded.
func WriteXLSX(ctx context.Context, dst io.Writer, report Report) (returnErr error) {
	if ctx == nil || dst == nil || report.Rows == nil {
		return errors.New("report export is not initialized")
	}
	if err := validateReport(report); err != nil {
		return err
	}

	archive := zip.NewWriter(dst)
	defer func() {
		if err := archive.Close(); err != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("finish spreadsheet package: %w", err))
		}
	}()

	if err := writeStaticEntry(archive, "[Content_Types].xml", contentTypesXML); err != nil {
		return err
	}
	if err := writeStaticEntry(archive, "_rels/.rels", rootRelationshipsXML); err != nil {
		return err
	}
	if err := writeWorkbook(archive, report.SheetName); err != nil {
		return err
	}
	if err := writeStaticEntry(archive, "xl/_rels/workbook.xml.rels", workbookRelationshipsXML); err != nil {
		return err
	}

	sheet, err := archive.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		return fmt.Errorf("create spreadsheet table: %w", err)
	}
	if _, err := io.WriteString(sheet, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`); err != nil {
		return fmt.Errorf("write spreadsheet table header: %w", err)
	}
	if err := writeRow(ctx, sheet, 1, []string{"Data exibida na tela", report.DisplayDate}); err != nil {
		return err
	}
	if err := writeRow(ctx, sheet, 2, report.Columns); err != nil {
		return err
	}

	rowNumber := firstDataRow
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("spreadsheet export canceled: %w", err)
		}
		row, more, err := report.Rows.Next(ctx)
		if err != nil {
			return fmt.Errorf("read next spreadsheet row: %w", err)
		}
		if !more {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("spreadsheet export canceled: %w", err)
			}
			break
		}
		if rowNumber > maxSheetRows {
			return errors.New("spreadsheet row limit exceeded")
		}
		if len(row) != len(report.Columns) {
			return errors.New("spreadsheet row does not match the table columns")
		}
		if err := writeRow(ctx, sheet, rowNumber, row); err != nil {
			return err
		}
		rowNumber++
	}
	if _, err := io.WriteString(sheet, `</sheetData></worksheet>`); err != nil {
		return fmt.Errorf("finish spreadsheet table: %w", err)
	}
	return nil
}

func validateReport(report Report) error {
	if !validSheetName(report.SheetName) {
		return errors.New("spreadsheet sheet name is invalid")
	}
	if len(report.Columns) == 0 || len(report.Columns) > MaxColumns {
		return errors.New("spreadsheet column count is invalid")
	}
	if strings.TrimSpace(report.DisplayDate) == "" {
		return errors.New("spreadsheet display date is required")
	}
	if err := validateCell(report.DisplayDate); err != nil {
		return fmt.Errorf("spreadsheet display date is invalid: %w", err)
	}
	for _, column := range report.Columns {
		if strings.TrimSpace(column) == "" {
			return errors.New("spreadsheet column name is empty")
		}
		if err := validateCell(column); err != nil {
			return fmt.Errorf("spreadsheet column name is invalid: %w", err)
		}
	}
	return nil
}

func validSheetName(name string) bool {
	if name == "" || strings.ContainsAny(name, "[]:*?/\\\"") || !utf8.ValidString(name) {
		return false
	}
	if name[0] == '\'' || name[len(name)-1] == '\'' || validateCell(name) != nil {
		return false
	}
	units := 0
	for _, character := range name {
		units++
		if character > 0xffff {
			units++
		}
	}
	return units <= maxSheetNameUTF
}

func writeRow(ctx context.Context, writer io.Writer, rowNumber int, values []string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("spreadsheet export canceled: %w", err)
	}
	rowBytes := 0
	for _, value := range values {
		if err := validateCell(value); err != nil {
			return err
		}
		if len(value) > MaxRowBytes-rowBytes {
			return errors.New("spreadsheet row byte limit exceeded")
		}
		rowBytes += len(value)
	}
	if _, err := io.WriteString(writer, `<row r="`+fmt.Sprint(rowNumber)+`">`); err != nil {
		return fmt.Errorf("write spreadsheet row: %w", err)
	}
	for index, value := range values {
		if _, err := io.WriteString(writer, `<c r="`+columnName(index+1)+fmt.Sprint(rowNumber)+`" t="inlineStr"><is><t xml:space="preserve">`); err != nil {
			return fmt.Errorf("write spreadsheet cell: %w", err)
		}
		if err := xml.EscapeText(writer, []byte(value)); err != nil {
			return fmt.Errorf("write spreadsheet cell text: %w", err)
		}
		if _, err := io.WriteString(writer, `</t></is></c>`); err != nil {
			return fmt.Errorf("finish spreadsheet cell: %w", err)
		}
	}
	if _, err := io.WriteString(writer, `</row>`); err != nil {
		return fmt.Errorf("finish spreadsheet row: %w", err)
	}
	return nil
}

func validateCell(value string) error {
	if !utf8.ValidString(value) {
		return errors.New("spreadsheet cell is not valid UTF-8")
	}
	units := 0
	for _, character := range value {
		if !validXMLCharacter(character) {
			return errors.New("spreadsheet cell contains an invalid XML character")
		}
		units++
		if character > 0xffff {
			units++
		}
		if units > MaxCellUTF16 {
			return errors.New("spreadsheet cell exceeds the Excel text limit")
		}
	}
	return nil
}

func validXMLCharacter(character rune) bool {
	return character == 0x9 || character == 0xa || character == 0xd ||
		(character >= 0x20 && character <= 0xd7ff) ||
		(character >= 0xe000 && character <= 0xfffd) ||
		(character >= 0x10000 && character <= 0x10ffff)
}

func columnName(column int) string {
	var name [3]byte
	position := len(name)
	for column > 0 {
		column--
		position--
		name[position] = byte('A' + column%26)
		column /= 26
	}
	return string(name[position:])
}

func writeStaticEntry(archive *zip.Writer, name, content string) error {
	entry, err := archive.Create(name)
	if err != nil {
		return fmt.Errorf("create spreadsheet package entry: %w", err)
	}
	if _, err := io.WriteString(entry, content); err != nil {
		return fmt.Errorf("write spreadsheet package entry: %w", err)
	}
	return nil
}

func writeWorkbook(archive *zip.Writer, sheetName string) error {
	entry, err := archive.Create("xl/workbook.xml")
	if err != nil {
		return fmt.Errorf("create spreadsheet workbook: %w", err)
	}
	if _, err := io.WriteString(entry, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="`); err != nil {
		return fmt.Errorf("write spreadsheet workbook: %w", err)
	}
	if err := xml.EscapeText(entry, []byte(sheetName)); err != nil {
		return fmt.Errorf("write spreadsheet sheet name: %w", err)
	}
	if _, err := io.WriteString(entry, `" sheetId="1" r:id="rId1"/></sheets></workbook>`); err != nil {
		return fmt.Errorf("finish spreadsheet workbook: %w", err)
	}
	return nil
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`

const rootRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`

const workbookRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`
