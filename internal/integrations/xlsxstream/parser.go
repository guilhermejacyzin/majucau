// Package xlsxstream reads untrusted XLSX workbooks one worksheet row at a
// time. It exposes source values only; it does not interpret dates, formulas,
// headers, accounting meanings, or financial rules.
package xlsxstream

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"majucau.local/financial-intelligence/internal/securetemp"
)

var (
	ErrInvalidWorkbook = errors.New("xlsx workbook is invalid or unsupported")
	ErrLimitExceeded   = errors.New("xlsx safety limit exceeded")
	ErrRowHandler      = errors.New("xlsx row handler failed")
)

const (
	// MaxFileBytes is the maximum compressed workbook size accepted by Stream.
	MaxFileBytes          int64  = 256 << 20
	maxCentralDirectory   uint32 = 2 << 20
	maxArchiveEntries            = 1024
	maxUncompressedBytes  uint64 = 1 << 30
	maxEntryBytes         int64  = 512 << 20
	maxSheets                    = 64
	maxRows               uint64 = 1_048_576
	maxColumns                   = 256
	maxCellsPerRow               = 256
	maxCellBytes          int    = 256 << 10
	maxXMLTokenBytes      int    = 1 << 20
	maxRowBytes           int    = 4 << 20
	maxSharedStrings             = 1_000_000
	maxSharedStringBytes  int64  = 256 << 20
	maxXMLMetadataBytes   int64  = 4 << 20
	maxZipCommentBytes           = 65_535
	spreadsheetNamespace         = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	relationshipNamespace        = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
)

// Cell contains the source type and raw value. Formula contains the formula
// text when present; it is never evaluated by this package.
type Cell struct {
	Reference string
	Column    int
	Type      string
	Value     string
	Formula   string
}

// Row is delivered once at a time. Cells are sparse and ordered by column.
type Row struct {
	Sheet      string
	SheetIndex int
	Number     int
	Cells      []Cell
}

type Summary struct {
	FileBytes         int64
	UncompressedBytes uint64
	Sheets            int
	Rows              uint64
	Cells             uint64
}

// RowHandler receives one bounded row at a time. The handler must finish
// using the row before returning and should persist its own results atomically.
type RowHandler func(context.Context, Row) error

type sheetReference struct {
	name string
	rel  string
}

type relationship struct {
	target  string
	typeURI string
}

type stringOffset struct {
	start  int64
	length uint32
}

type sharedStringIndex struct {
	file    *os.File
	offsets []stringOffset
	bytes   int64
}

type zipLimits struct {
	entries          int
	uncompressed     uint64
	centralDirectory uint32
}

// Stream parses an XLSX file through ReaderAt, allowing the caller to use a
// private staged file without loading the workbook into memory. tempParent
// must be an application-controlled directory whose ACL is private.
func Stream(ctx context.Context, source io.ReaderAt, size int64, tempParent string, accept RowHandler) (summary Summary, returnErr error) {
	if ctx == nil || source == nil || accept == nil || strings.TrimSpace(tempParent) == "" || size <= 0 {
		return Summary{}, ErrInvalidWorkbook
	}
	if size > MaxFileBytes {
		return Summary{}, ErrLimitExceeded
	}
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}
	zipLimits, err := inspectZipDirectory(source, size)
	if err != nil {
		return Summary{}, err
	}
	archive, err := zip.NewReader(source, size)
	if err != nil || len(archive.File) != zipLimits.entries {
		return Summary{}, ErrInvalidWorkbook
	}
	files, uncompressedBytes, err := indexArchive(archive, zipLimits)
	if err != nil {
		return Summary{}, err
	}
	if err := rejectExternalParts(ctx, files); err != nil {
		return Summary{}, err
	}
	if err := validateContentTypes(ctx, files); err != nil {
		return Summary{}, err
	}
	sheets, err := loadWorkbookSheets(ctx, files)
	if err != nil {
		return Summary{}, err
	}
	relations, err := loadWorkbookRelationships(ctx, files)
	if err != nil {
		return Summary{}, err
	}
	for index := range sheets {
		sheet := &sheets[index]
		relation, ok := relations[sheet.rel]
		if !ok || !strings.HasSuffix(relation.typeURI, "/worksheet") {
			return Summary{}, ErrInvalidWorkbook
		}
		target, err := resolveWorksheetTarget(relation.target)
		if err != nil {
			return Summary{}, err
		}
		if _, ok := files[target]; !ok {
			return Summary{}, ErrInvalidWorkbook
		}
		sheet.rel = target
	}

	workspace, err := securetemp.NewWorkspace(tempParent, "majucau-xlsx-index-")
	if err != nil {
		return Summary{}, ErrInvalidWorkbook
	}
	defer func() {
		returnErr = errors.Join(returnErr, workspace.RemoveAll())
	}()

	shared, err := loadSharedStrings(ctx, files["xl/sharedStrings.xml"], workspace.Path())
	if err != nil {
		return Summary{}, err
	}
	if shared != nil {
		defer func() {
			returnErr = errors.Join(returnErr, shared.close())
		}()
	}

	summary = Summary{FileBytes: size, UncompressedBytes: uncompressedBytes, Sheets: len(sheets)}
	for index, sheet := range sheets {
		if err := ctx.Err(); err != nil {
			return Summary{}, err
		}
		rows, cells, err := streamWorksheet(ctx, files[sheet.rel], sheet.name, index, shared, accept)
		if err != nil {
			return Summary{}, err
		}
		if math.MaxUint64-summary.Rows < rows || math.MaxUint64-summary.Cells < cells {
			return Summary{}, ErrLimitExceeded
		}
		summary.Rows += rows
		summary.Cells += cells
	}
	return summary, nil
}

func inspectZipDirectory(source io.ReaderAt, size int64) (zipLimits, error) {
	if size < 22 {
		return zipLimits{}, ErrInvalidWorkbook
	}
	tailLength := int64(22 + maxZipCommentBytes)
	if size < tailLength {
		tailLength = size
	}
	tail := make([]byte, int(tailLength))
	if _, err := source.ReadAt(tail, size-tailLength); err != nil && !errors.Is(err, io.EOF) {
		return zipLimits{}, ErrInvalidWorkbook
	}
	marker := []byte{'P', 'K', 5, 6}
	index := bytes.LastIndex(tail, marker)
	if index < 0 || len(tail)-index < 22 {
		return zipLimits{}, ErrInvalidWorkbook
	}
	eocd := tail[index:]
	commentLength := int(binary.LittleEndian.Uint16(eocd[20:22]))
	if len(eocd) != 22+commentLength {
		return zipLimits{}, ErrInvalidWorkbook
	}
	disk := binary.LittleEndian.Uint16(eocd[4:6])
	centralDisk := binary.LittleEndian.Uint16(eocd[6:8])
	diskEntries := binary.LittleEndian.Uint16(eocd[8:10])
	totalEntries := binary.LittleEndian.Uint16(eocd[10:12])
	centralSize := binary.LittleEndian.Uint32(eocd[12:16])
	centralOffset := binary.LittleEndian.Uint32(eocd[16:20])
	if disk != 0 || centralDisk != 0 || diskEntries != totalEntries ||
		totalEntries == math.MaxUint16 || centralSize == math.MaxUint32 || centralOffset == math.MaxUint32 {
		return zipLimits{}, ErrInvalidWorkbook
	}
	if int(totalEntries) == 0 || int(totalEntries) > maxArchiveEntries || centralSize > maxCentralDirectory {
		return zipLimits{}, ErrLimitExceeded
	}
	eocdOffset := uint64(size-tailLength) + uint64(index)
	if uint64(centralOffset) > eocdOffset || uint64(centralSize) > eocdOffset-uint64(centralOffset) {
		return zipLimits{}, ErrInvalidWorkbook
	}
	return zipLimits{entries: int(totalEntries), centralDirectory: centralSize}, nil
}

func indexArchive(archive *zip.Reader, limits zipLimits) (map[string]*zip.File, uint64, error) {
	if len(archive.File) > maxArchiveEntries || len(archive.File) != limits.entries {
		return nil, 0, ErrLimitExceeded
	}
	files := make(map[string]*zip.File, len(archive.File))
	var total uint64
	for _, file := range archive.File {
		name := file.Name
		if !utf8.ValidString(name) || name == "" || strings.Contains(name, "\\") || path.IsAbs(name) {
			return nil, 0, ErrInvalidWorkbook
		}
		directory := strings.HasSuffix(name, "/")
		cleanName := strings.TrimSuffix(name, "/")
		if cleanName == "" || path.Clean(cleanName) != cleanName || cleanName == "." || strings.HasPrefix(cleanName, "../") || strings.Contains(cleanName, "/../") {
			return nil, 0, ErrInvalidWorkbook
		}
		if directory != file.FileInfo().IsDir() || (file.Mode()&os.ModeSymlink) != 0 || (file.Flags&1) != 0 {
			return nil, 0, ErrInvalidWorkbook
		}
		if file.Method != zip.Store && file.Method != zip.Deflate {
			return nil, 0, ErrInvalidWorkbook
		}
		if _, exists := files[name]; exists {
			return nil, 0, ErrInvalidWorkbook
		}
		if file.UncompressedSize64 > uint64(maxEntryBytes) || file.CompressedSize64 > uint64(MaxFileBytes) {
			return nil, 0, ErrLimitExceeded
		}
		if math.MaxUint64-total < file.UncompressedSize64 {
			return nil, 0, ErrLimitExceeded
		}
		total += file.UncompressedSize64
		if total > maxUncompressedBytes {
			return nil, 0, ErrLimitExceeded
		}
		files[name] = file
	}
	if total == 0 {
		return nil, 0, ErrInvalidWorkbook
	}
	return files, total, nil
}

func rejectExternalParts(ctx context.Context, files map[string]*zip.File) error {
	for name := range files {
		lower := strings.ToLower(name)
		if lower == "xl/vbaproject.bin" || strings.HasPrefix(lower, "xl/externallinks/") {
			return ErrInvalidWorkbook
		}
	}
	for name, file := range files {
		if !strings.HasSuffix(name, ".rels") {
			continue
		}
		if err := parseXMLFile(ctx, file, maxXMLMetadataBytes, func(decoder *xml.Decoder) error {
			for {
				token, err := nextToken(decoder)
				if errors.Is(err, io.EOF) {
					return nil
				}
				if err != nil {
					return ErrInvalidWorkbook
				}
				start, ok := token.(xml.StartElement)
				if !ok || start.Name.Local != "Relationship" {
					continue
				}
				for _, attribute := range start.Attr {
					if attribute.Name.Local == "TargetMode" && strings.EqualFold(attribute.Value, "External") {
						return ErrInvalidWorkbook
					}
				}
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

func validateContentTypes(ctx context.Context, files map[string]*zip.File) error {
	file, ok := files["[Content_Types].xml"]
	if !ok {
		return ErrInvalidWorkbook
	}
	var contentTypes strings.Builder
	if err := parseXMLFile(ctx, file, maxXMLMetadataBytes, func(decoder *xml.Decoder) error {
		for {
			token, err := nextToken(decoder)
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return ErrInvalidWorkbook
			}
			if text, ok := token.(xml.CharData); ok {
				if contentTypes.Len()+len(text) > int(maxXMLMetadataBytes) {
					return ErrLimitExceeded
				}
				contentTypes.Write(text)
			}
			if start, ok := token.(xml.StartElement); ok {
				for _, attribute := range start.Attr {
					if strings.Contains(strings.ToLower(attribute.Value), "macroenabled") {
						return ErrInvalidWorkbook
					}
				}
			}
		}
	}); err != nil {
		return err
	}
	if strings.Contains(strings.ToLower(contentTypes.String()), "macroenabled") {
		return ErrInvalidWorkbook
	}
	return nil
}

func loadWorkbookSheets(ctx context.Context, files map[string]*zip.File) ([]sheetReference, error) {
	file, ok := files["xl/workbook.xml"]
	if !ok {
		return nil, ErrInvalidWorkbook
	}
	var result []sheetReference
	err := parseXMLFile(ctx, file, maxXMLMetadataBytes, func(decoder *xml.Decoder) error {
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			token, err := nextToken(decoder)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return ErrInvalidWorkbook
			}
			start, ok := token.(xml.StartElement)
			if !ok || start.Name.Local != "sheet" || start.Name.Space != spreadsheetNamespace {
				continue
			}
			var sheet sheetReference
			for _, attribute := range start.Attr {
				switch attribute.Name.Local {
				case "name":
					sheet.name = attribute.Value
				case "id":
					if attribute.Name.Space == relationshipNamespace {
						sheet.rel = attribute.Value
					}
				}
			}
			if sheet.name == "" || len(sheet.name) > 31 || sheet.rel == "" || len(result) >= maxSheets {
				return ErrLimitExceeded
			}
			if strings.ContainsAny(sheet.name, "[]:*?/\\") {
				return ErrInvalidWorkbook
			}
			for _, existing := range result {
				if strings.EqualFold(existing.name, sheet.name) || existing.rel == sheet.rel {
					return ErrInvalidWorkbook
				}
			}
			result = append(result, sheet)
		}
		if len(result) == 0 {
			return ErrInvalidWorkbook
		}
		return nil
	})
	return result, err
}

func loadWorkbookRelationships(ctx context.Context, files map[string]*zip.File) (map[string]relationship, error) {
	file, ok := files["xl/_rels/workbook.xml.rels"]
	if !ok {
		return nil, ErrInvalidWorkbook
	}
	result := make(map[string]relationship)
	err := parseXMLFile(ctx, file, maxXMLMetadataBytes, func(decoder *xml.Decoder) error {
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			token, err := nextToken(decoder)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return ErrInvalidWorkbook
			}
			start, ok := token.(xml.StartElement)
			if !ok || start.Name.Local != "Relationship" {
				continue
			}
			var id, target, typeURI, targetMode string
			for _, attribute := range start.Attr {
				switch attribute.Name.Local {
				case "Id":
					id = attribute.Value
				case "Target":
					target = attribute.Value
				case "Type":
					typeURI = attribute.Value
				case "TargetMode":
					targetMode = attribute.Value
				}
			}
			if id == "" || target == "" || typeURI == "" || strings.EqualFold(targetMode, "External") {
				return ErrInvalidWorkbook
			}
			if _, exists := result[id]; exists {
				return ErrInvalidWorkbook
			}
			result[id] = relationship{target: target, typeURI: typeURI}
		}
		if len(result) == 0 {
			return ErrInvalidWorkbook
		}
		return nil
	})
	return result, err
}

func resolveWorksheetTarget(target string) (string, error) {
	if strings.Contains(target, "\\") || strings.Contains(target, ":") {
		return "", ErrInvalidWorkbook
	}
	resolved := target
	if strings.HasPrefix(resolved, "/") {
		resolved = strings.TrimPrefix(resolved, "/")
	} else {
		resolved = path.Join("xl", resolved)
	}
	if path.Clean(resolved) != resolved || !strings.HasPrefix(resolved, "xl/worksheets/") || strings.Contains(resolved, "/../") {
		return "", ErrInvalidWorkbook
	}
	return resolved, nil
}

func loadSharedStrings(ctx context.Context, file *zip.File, tempParent string) (*sharedStringIndex, error) {
	if file == nil {
		return nil, nil
	}
	temporaryPath := filepath.Join(tempParent, "shared-strings.bin")
	output, err := os.OpenFile(temporaryPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, ErrInvalidWorkbook
	}
	index := &sharedStringIndex{file: output}
	err = parseXMLFile(ctx, file, maxEntryBytes, func(decoder *xml.Decoder) error {
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			token, err := nextToken(decoder)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return ErrInvalidWorkbook
			}
			start, ok := token.(xml.StartElement)
			if !ok || start.Name.Local != "si" {
				continue
			}
			value, err := readRichText(decoder, "si", maxCellBytes)
			if err != nil {
				return err
			}
			if err := index.add(value); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		err = output.Sync()
	}
	closeErr := output.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = securetemp.RemoveFile(temporaryPath)
		return nil, err
	}
	input, err := os.Open(temporaryPath)
	if err != nil {
		_ = securetemp.RemoveFile(temporaryPath)
		return nil, ErrInvalidWorkbook
	}
	index.file = input
	return index, nil
}

func (index *sharedStringIndex) add(value string) error {
	if index == nil || index.file == nil || len(index.offsets) >= maxSharedStrings || len(value) > maxCellBytes || index.bytes > maxSharedStringBytes-int64(len(value)) {
		return ErrLimitExceeded
	}
	start, err := index.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return ErrInvalidWorkbook
	}
	if _, err := io.WriteString(index.file, value); err != nil {
		return ErrInvalidWorkbook
	}
	index.offsets = append(index.offsets, stringOffset{start: start, length: uint32(len(value))})
	index.bytes += int64(len(value))
	return nil
}

func (index *sharedStringIndex) get(position uint64) (string, error) {
	if index == nil || index.file == nil || position >= uint64(len(index.offsets)) {
		return "", ErrInvalidWorkbook
	}
	entry := index.offsets[position]
	value := make([]byte, int(entry.length))
	if len(value) > 0 {
		if _, err := index.file.ReadAt(value, entry.start); err != nil {
			return "", ErrInvalidWorkbook
		}
	}
	return string(value), nil
}

func (index *sharedStringIndex) close() error {
	if index == nil || index.file == nil {
		return nil
	}
	err := index.file.Close()
	index.file = nil
	if err != nil {
		return ErrInvalidWorkbook
	}
	return nil
}

func streamWorksheet(ctx context.Context, file *zip.File, sheetName string, sheetIndex int, shared *sharedStringIndex, accept RowHandler) (uint64, uint64, error) {
	var rows, cells uint64
	var lastRow uint64
	err := parseXMLFile(ctx, file, maxEntryBytes, func(decoder *xml.Decoder) error {
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			token, err := nextToken(decoder)
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return ErrInvalidWorkbook
			}
			start, ok := token.(xml.StartElement)
			if !ok || start.Name.Local != "row" || start.Name.Space != spreadsheetNamespace {
				continue
			}
			row, err := parseRow(ctx, decoder, start, sheetName, sheetIndex, shared, lastRow)
			if err != nil {
				return err
			}
			lastRow = uint64(row.Number)
			if rows >= maxRows {
				return ErrLimitExceeded
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := accept(ctx, row); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return fmt.Errorf("%w: %w", ErrRowHandler, err)
			}
			rows++
			cells += uint64(len(row.Cells))
		}
	})
	return rows, cells, err
}

func parseRow(ctx context.Context, decoder *xml.Decoder, element xml.StartElement, sheet string, sheetIndex int, shared *sharedStringIndex, previous uint64) (Row, error) {
	var rowNumber uint64
	for _, attribute := range element.Attr {
		if attribute.Name.Local == "r" {
			parsed, err := strconv.ParseUint(attribute.Value, 10, 32)
			if err != nil {
				return Row{}, ErrInvalidWorkbook
			}
			rowNumber = parsed
		}
	}
	if rowNumber == 0 || rowNumber > maxRows || rowNumber <= previous {
		return Row{}, ErrLimitExceeded
	}
	result := Row{Sheet: sheet, SheetIndex: sheetIndex, Number: int(rowNumber)}
	lastColumn := 0
	rowBytes := 0
	for {
		if err := ctx.Err(); err != nil {
			return Row{}, err
		}
		token, err := nextToken(decoder)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return Row{}, ErrInvalidWorkbook
			}
			return Row{}, ErrInvalidWorkbook
		}
		switch token := token.(type) {
		case xml.StartElement:
			if token.Name.Local == "row" {
				return Row{}, ErrInvalidWorkbook
			}
			if token.Name.Local != "c" || token.Name.Space != spreadsheetNamespace {
				if err := skipElement(decoder, token.Name); err != nil {
					return Row{}, err
				}
				continue
			}
			cell, err := parseCell(decoder, token, rowNumber, shared)
			if err != nil {
				return Row{}, err
			}
			if cell.Column <= lastColumn || cell.Column > maxColumns || len(result.Cells) >= maxCellsPerRow {
				return Row{}, ErrLimitExceeded
			}
			lastColumn = cell.Column
			rowBytes += len(cell.Reference) + len(cell.Type) + len(cell.Value) + len(cell.Formula)
			if rowBytes > maxRowBytes {
				return Row{}, ErrLimitExceeded
			}
			result.Cells = append(result.Cells, cell)
		case xml.EndElement:
			if token.Name == element.Name {
				return result, nil
			}
		}
	}
}

func parseCell(decoder *xml.Decoder, element xml.StartElement, rowNumber uint64, shared *sharedStringIndex) (Cell, error) {
	var reference, kind string
	for _, attribute := range element.Attr {
		switch attribute.Name.Local {
		case "r":
			reference = attribute.Value
		case "t":
			kind = attribute.Value
		}
	}
	column, row, err := parseCellReference(reference)
	if err != nil || row != rowNumber || column > maxColumns {
		return Cell{}, ErrLimitExceeded
	}
	var value, formula string
	seenValue, seenFormula, seenInline := false, false, false
	for {
		token, err := nextToken(decoder)
		if err != nil {
			return Cell{}, ErrInvalidWorkbook
		}
		switch token := token.(type) {
		case xml.StartElement:
			if token.Name.Local == "v" {
				if seenValue {
					return Cell{}, ErrInvalidWorkbook
				}
				seenValue = true
				value, err = readSimpleText(decoder, "v", maxCellBytes)
				if err != nil {
					return Cell{}, err
				}
			} else if token.Name.Local == "f" {
				if seenFormula {
					return Cell{}, ErrInvalidWorkbook
				}
				seenFormula = true
				formula, err = readSimpleText(decoder, "f", maxCellBytes)
				if err != nil {
					return Cell{}, err
				}
			} else if token.Name.Local == "is" {
				if seenInline {
					return Cell{}, ErrInvalidWorkbook
				}
				seenInline = true
				value, err = readRichText(decoder, "is", maxCellBytes)
				if err != nil {
					return Cell{}, err
				}
			} else if err := skipElement(decoder, token.Name); err != nil {
				return Cell{}, err
			}
		case xml.EndElement:
			if token.Name == element.Name {
				if kind == "s" {
					if !seenValue {
						return Cell{}, ErrInvalidWorkbook
					}
					position, parseErr := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
					if parseErr != nil {
						return Cell{}, ErrInvalidWorkbook
					}
					value, err = shared.get(position)
					if err != nil {
						return Cell{}, err
					}
				} else if kind == "inlineStr" && !seenInline {
					return Cell{}, ErrInvalidWorkbook
				} else if kind != "" && kind != "n" && kind != "b" && kind != "e" && kind != "str" && kind != "inlineStr" {
					return Cell{}, ErrInvalidWorkbook
				}
				return Cell{Reference: reference, Column: column, Type: kind, Value: value, Formula: formula}, nil
			}
		}
	}
}

func parseCellReference(reference string) (int, uint64, error) {
	if reference == "" || len(reference) > 16 {
		return 0, 0, ErrInvalidWorkbook
	}
	column, index := 0, 0
	for index < len(reference) {
		character := reference[index]
		if character >= 'a' && character <= 'z' {
			character -= 'a' - 'A'
		}
		if character < 'A' || character > 'Z' || index >= 3 {
			break
		}
		column = column*26 + int(character-'A'+1)
		index++
	}
	if index == 0 || index == len(reference) {
		return 0, 0, ErrInvalidWorkbook
	}
	for _, character := range reference[index:] {
		if character < '0' || character > '9' {
			return 0, 0, ErrInvalidWorkbook
		}
	}
	row, err := strconv.ParseUint(reference[index:], 10, 32)
	if err != nil || row == 0 {
		return 0, 0, ErrInvalidWorkbook
	}
	return column, row, nil
}

func parseXMLFile(ctx context.Context, file *zip.File, maxBytes int64, parse func(*xml.Decoder) error) error {
	if file == nil || file.UncompressedSize64 > uint64(maxBytes) {
		return ErrLimitExceeded
	}
	reader, err := file.Open()
	if err != nil {
		return ErrInvalidWorkbook
	}
	bounded := &countingLimitReader{reader: reader, limit: maxBytes}
	limited := newXMLTokenLimitReader(ctx, bounded, maxXMLTokenBytes)
	decoder := xml.NewDecoder(limited)
	decoder.Strict = true
	err = parse(decoder)
	if err == nil {
		_, err = io.Copy(io.Discard, limited)
	}
	closeErr := reader.Close()
	if err == nil {
		err = closeErr
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if limited.exceeded || errors.Is(err, errXMLTokenLimit) {
		return ErrLimitExceeded
	}
	if errors.Is(err, errReadLimit) {
		return ErrLimitExceeded
	}
	if err != nil {
		return err
	}
	if bounded.read > file.UncompressedSize64 {
		return ErrInvalidWorkbook
	}
	return nil
}

type countingLimitReader struct {
	reader io.Reader
	limit  int64
	read   uint64
}

var errReadLimit = errors.New("xlsx entry exceeds configured limit")

func (reader *countingLimitReader) Read(buffer []byte) (int, error) {
	if reader.read > uint64(reader.limit) {
		return 0, errReadLimit
	}
	remaining := uint64(reader.limit) - reader.read
	if remaining == 0 {
		var extra [1]byte
		n, err := reader.reader.Read(extra[:])
		if n > 0 {
			reader.read++
			return 0, errReadLimit
		}
		return 0, err
	}
	if uint64(len(buffer)) > remaining+1 {
		buffer = buffer[:int(remaining+1)]
	}
	n, err := reader.reader.Read(buffer)
	reader.read += uint64(n)
	if reader.read > uint64(reader.limit) {
		return n, errReadLimit
	}
	return n, err
}

func nextToken(decoder *xml.Decoder) (xml.Token, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch token := token.(type) {
	case xml.Directive:
		return nil, ErrInvalidWorkbook
	case xml.ProcInst:
		if !strings.EqualFold(token.Target, "xml") {
			return nil, ErrInvalidWorkbook
		}
	}
	return token, nil
}

func readSimpleText(decoder *xml.Decoder, endName string, maxBytes int) (string, error) {
	var value strings.Builder
	for {
		token, err := nextToken(decoder)
		if err != nil {
			return "", ErrInvalidWorkbook
		}
		switch token := token.(type) {
		case xml.CharData:
			if value.Len()+len(token) > maxBytes {
				return "", ErrLimitExceeded
			}
			value.Write(token)
		case xml.StartElement:
			return "", ErrInvalidWorkbook
		case xml.EndElement:
			if token.Name.Local == endName {
				return value.String(), nil
			}
			return "", ErrInvalidWorkbook
		}
	}
}

func readRichText(decoder *xml.Decoder, endName string, maxBytes int) (string, error) {
	var value strings.Builder
	depth, textDepth, phoneticDepth := 1, 0, 0
	for depth > 0 {
		token, err := nextToken(decoder)
		if err != nil {
			return "", ErrInvalidWorkbook
		}
		switch token := token.(type) {
		case xml.StartElement:
			depth++
			if phoneticDepth > 0 {
				phoneticDepth++
			} else if token.Name.Local == "rPh" {
				phoneticDepth = 1
			} else if token.Name.Local == "t" {
				textDepth++
			}
		case xml.CharData:
			if textDepth > 0 && phoneticDepth == 0 {
				if value.Len()+len(token) > maxBytes {
					return "", ErrLimitExceeded
				}
				value.Write(token)
			}
		case xml.EndElement:
			if phoneticDepth > 0 {
				phoneticDepth--
			} else if token.Name.Local == "t" && textDepth > 0 {
				textDepth--
			}
			depth--
			if depth == 0 && token.Name.Local != endName {
				return "", ErrInvalidWorkbook
			}
		}
	}
	return value.String(), nil
}

func skipElement(decoder *xml.Decoder, start xml.Name) error {
	depth := 1
	for depth > 0 {
		token, err := nextToken(decoder)
		if err != nil {
			return ErrInvalidWorkbook
		}
		switch token.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}
