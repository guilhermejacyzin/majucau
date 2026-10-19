// Package csvlimit enforces memory bounds while CSV bytes are being read.
package csvlimit

import (
	"errors"
	"io"
)

const (
	MaxFieldBytes  = 1 << 20
	MaxRecordBytes = 4 << 20
	MaxFields      = 256
)

var (
	ErrLimitExceeded    = errors.New("csv safety limit exceeded")
	ErrInvalidDelimiter = errors.New("csv delimiter is unsupported")
)

// Reader tracks CSV quoting before encoding/csv buffers a record. The limits
// apply to bytes, so multibyte text remains bounded in memory as well.
type Reader struct {
	source    io.Reader
	delimiter byte

	fieldBytes   int
	recordBytes  int
	fields       int
	inQuotes     bool
	quotePending bool
	fieldStart   bool
}

func NewReader(source io.Reader) *Reader {
	return newReader(source, ';')
}

// NewDelimitedReader applies the same bounded field, record and column checks
// to comma- or semicolon-delimited CSV. Callers that do not know the source
// contract may inspect a bounded header first, then choose the delimiter.
func NewDelimitedReader(source io.Reader, delimiter byte) (*Reader, error) {
	if source == nil || (delimiter != ',' && delimiter != ';') {
		return nil, ErrInvalidDelimiter
	}
	return newReader(source, delimiter), nil
}

// NewTSVReader applies the same bounded field and record checks to tab-delimited
// input. Keeping the delimiter in this reader ensures the field-count limit is
// enforced before encoding/csv allocates a row.
func NewTSVReader(source io.Reader) *Reader {
	return newReader(source, '\t')
}

func newReader(source io.Reader, delimiter byte) *Reader {
	return &Reader{source: source, delimiter: delimiter, fields: 1, fieldStart: true}
}

func (r *Reader) Read(buffer []byte) (int, error) {
	count, readErr := r.source.Read(buffer)
	for index := 0; index < count; index++ {
		if err := r.track(buffer[index]); err != nil {
			return index, err
		}
	}
	return count, readErr
}

func (r *Reader) track(value byte) error {
	r.recordBytes++
	r.fieldBytes++
	if r.recordBytes > MaxRecordBytes || r.fieldBytes > MaxFieldBytes {
		return ErrLimitExceeded
	}

	if r.inQuotes {
		if r.quotePending {
			if value == '"' {
				r.quotePending = false
				return nil
			}
			r.inQuotes = false
			r.quotePending = false
		} else if value == '"' {
			r.quotePending = true
			return nil
		} else {
			return nil
		}
	}

	if r.fieldStart {
		if value == ' ' || value == '\t' {
			return nil
		}
		if value == '"' {
			r.inQuotes = true
			r.fieldStart = false
			return nil
		}
	}

	switch value {
	case r.delimiter:
		r.fields++
		if r.fields > MaxFields {
			return ErrLimitExceeded
		}
		r.fieldBytes = 0
		r.fieldStart = true
	case '\n':
		r.fieldBytes = 0
		r.recordBytes = 0
		r.fields = 1
		r.fieldStart = true
	default:
		r.fieldStart = false
	}
	return nil
}
