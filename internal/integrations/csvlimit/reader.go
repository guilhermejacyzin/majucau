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

var ErrLimitExceeded = errors.New("csv safety limit exceeded")

// Reader tracks CSV quoting before encoding/csv buffers a record. The limits
// apply to bytes, so multibyte text remains bounded in memory as well.
type Reader struct {
	source io.Reader

	fieldBytes  int
	recordBytes int
	fields      int
	inQuotes    bool
	quotePending bool
	fieldStart  bool
}

func NewReader(source io.Reader) *Reader {
	return &Reader{source: source, fields: 1, fieldStart: true}
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
	case ';':
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
