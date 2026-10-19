// Package xlsstream validates and reads legacy Excel BIFF8 workbooks row by
// row. It exposes source text only; it does not interpret business meaning.
package xlsstream

import (
	"context"
	"errors"
	"io"

	legacyxls "majucau.local/financial-intelligence/internal/integrations/xlsstream/legacyxls/xls"
)

var (
	ErrInvalidWorkbook = errors.New("xls workbook is invalid or unsupported")
	ErrLimitExceeded   = errors.New("xls safety limit exceeded")
	ErrRowHandler      = errors.New("xls row handler failed")
)

const (
	// MaxFileBytes bounds both the input and the legacy reader's in-memory
	// workbook metadata. The BIFF8 worksheet rows themselves are read by cursor.
	MaxFileBytes        int64  = 64 << 20
	maxSheets                  = 256
	maxRowsPerSheet     uint64 = 65_536
	maxRowsPerWorkbook  uint64 = 1_048_576
	maxCellsPerRow             = 256
	maxCellsPerWorkbook uint64 = 8_388_608
	maxCellBytes               = 256 << 10
	maxRowBytes         uint64 = 4 << 20
	maxSheetNameBytes          = 256
)

type Cell struct {
	Column int
	Raw    string
}

type Row struct {
	Sheet      string
	SheetIndex int
	Number     int
	Cells      []Cell
}

type Summary struct {
	FileBytes int64
	Sheets    int
	Rows      uint64
	Cells     uint64
}

type RowHandler func(context.Context, Row) error

// Stream validates a staged BIFF8 workbook through ReaderAt and sends bounded
// rows to accept. The shared string table is format metadata and is limited by
// MaxFileBytes; cell rows are never accumulated. source must remain open until
// this function returns.
func Stream(ctx context.Context, source io.ReaderAt, size int64, accept RowHandler) (summary Summary, returnErr error) {
	if ctx == nil || source == nil || accept == nil || size <= 0 {
		return Summary{}, ErrInvalidWorkbook
	}
	if size > MaxFileBytes {
		return Summary{}, ErrLimitExceeded
	}
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}

	reader, err := legacyxls.Open(contextReaderAt{ctx: ctx, source: source}, size)
	if err != nil {
		if isLegacyLimit(err) {
			return Summary{}, ErrLimitExceeded
		}
		return Summary{}, classifyError(ctx, err)
	}
	defer func() {
		returnErr = errors.Join(returnErr, reader.Close())
	}()

	sheetCount := reader.SheetCount()
	if sheetCount == 0 {
		return Summary{}, ErrInvalidWorkbook
	}
	if sheetCount > maxSheets {
		return Summary{}, ErrLimitExceeded
	}
	summary = Summary{FileBytes: size, Sheets: sheetCount}
	for sheetIndex := 0; reader.NextSheet(); sheetIndex++ {
		if err := ctx.Err(); err != nil {
			return Summary{}, err
		}
		if err := reader.Err(); err != nil {
			if isLegacyLimit(err) {
				return Summary{}, ErrLimitExceeded
			}
			return Summary{}, classifyError(ctx, err)
		}
		if sheetIndex >= sheetCount {
			return Summary{}, ErrInvalidWorkbook
		}
		sheetName := reader.SheetName()
		if len(sheetName) == 0 || len(sheetName) > maxSheetNameBytes {
			return Summary{}, ErrLimitExceeded
		}
		var sheetRows uint64
		for reader.Read() {
			if err := ctx.Err(); err != nil {
				return Summary{}, err
			}
			rowNumber := reader.RowIndex()
			if rowNumber < 0 || uint64(rowNumber) >= maxRowsPerSheet {
				return Summary{}, ErrLimitExceeded
			}
			sheetRows++
			if sheetRows > maxRowsPerSheet {
				return Summary{}, ErrLimitExceeded
			}

			sourceCells := reader.Row()
			if len(sourceCells) > maxCellsPerRow {
				return Summary{}, ErrLimitExceeded
			}
			row := Row{
				Sheet:      sheetName,
				SheetIndex: sheetIndex,
				Number:     rowNumber,
				Cells:      make([]Cell, 0, len(sourceCells)),
			}
			var rowBytes uint64
			for column, cell := range sourceCells {
				if cell.Value == nil && cell.Raw == "" {
					continue
				}
				if len(cell.Raw) > maxCellBytes || rowBytes > maxRowBytes-uint64(len(cell.Raw)) {
					return Summary{}, ErrLimitExceeded
				}
				rowBytes += uint64(len(cell.Raw))
				row.Cells = append(row.Cells, Cell{Column: column, Raw: cell.Raw})
			}
			if uint64(len(row.Cells)) > maxCellsPerRow || summary.Cells > maxCellsPerWorkbook-uint64(len(row.Cells)) {
				return Summary{}, ErrLimitExceeded
			}
			if summary.Rows == ^uint64(0) {
				return Summary{}, ErrLimitExceeded
			}
			if summary.Rows >= maxRowsPerWorkbook {
				return Summary{}, ErrLimitExceeded
			}
			if err := accept(ctx, row); err != nil {
				if ctx.Err() != nil {
					return Summary{}, ctx.Err()
				}
				return Summary{}, errors.Join(ErrRowHandler, err)
			}
			summary.Rows++
			summary.Cells += uint64(len(row.Cells))
		}
		if err := reader.Err(); err != nil {
			if isLegacyLimit(err) {
				return Summary{}, ErrLimitExceeded
			}
			return Summary{}, classifyError(ctx, err)
		}
	}
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}
	if err := reader.Err(); err != nil {
		if isLegacyLimit(err) {
			return Summary{}, ErrLimitExceeded
		}
		return Summary{}, classifyError(ctx, err)
	}
	return summary, nil
}

func isLegacyLimit(err error) bool {
	return errors.Is(err, legacyxls.ErrColumnLimit) || errors.Is(err, legacyxls.ErrSheetLimit) || errors.Is(err, legacyxls.ErrStyleLimit)
}

func classifyError(ctx context.Context, err error) error {
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}
	if errors.Is(err, ErrLimitExceeded) {
		return ErrLimitExceeded
	}
	return errors.Join(ErrInvalidWorkbook, err)
}

type contextReaderAt struct {
	ctx    context.Context
	source io.ReaderAt
}

func (reader contextReaderAt) ReadAt(buffer []byte, offset int64) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	return reader.source.ReadAt(buffer, offset)
}
