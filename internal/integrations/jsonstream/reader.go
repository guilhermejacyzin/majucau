// Package jsonstream validates untrusted JSON incrementally. It deliberately
// does not map fields to domain or financial meanings.
package jsonstream

import (
	"bufio"
	"context"
	"errors"
	"io"
	"unicode/utf8"
)

var (
	ErrInvalidJSON   = errors.New("JSON input is invalid or unsupported")
	ErrLimitExceeded = errors.New("JSON safety limit exceeded")
)

const (
	MaxFileBytes    int64 = 256 << 20
	MaxStringBytes        = 1 << 20
	maxNestingDepth       = 64
	maxValues             = 1_000_000
	maxEntries            = 1_000_000
	readBufferBytes       = 64 << 10
)

type Summary struct {
	FileBytes int64
	Values    int64
	Objects   int64
	Arrays    int64
	Strings   int64
	MaxDepth  int
}

// Validate reads and validates one JSON document without retaining its
// containers or values. Only a single bounded string token is buffered.
func Validate(ctx context.Context, source io.Reader, expectedSize int64) (Summary, error) {
	if ctx == nil || source == nil || expectedSize <= 0 {
		return Summary{}, ErrInvalidJSON
	}
	if expectedSize > MaxFileBytes {
		return Summary{}, ErrLimitExceeded
	}
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}

	limited := &limitedContextReader{ctx: ctx, source: source, limit: MaxFileBytes}
	p := parser{ctx: ctx, input: bufio.NewReaderSize(limited, readBufferBytes)}
	if err := p.skipWhitespace(); err != nil {
		return Summary{}, err
	}
	if err := p.parseValue(0); err != nil {
		return Summary{}, err
	}
	if err := p.skipWhitespace(); err != nil {
		return Summary{}, err
	}
	if _, err := p.peek(); !errors.Is(err, io.EOF) {
		if err != nil {
			return Summary{}, err
		}
		return Summary{}, ErrInvalidJSON
	}
	if limited.read != expectedSize {
		if limited.read > MaxFileBytes {
			return Summary{}, ErrLimitExceeded
		}
		return Summary{}, ErrInvalidJSON
	}
	p.summary.FileBytes = limited.read
	return p.summary, nil
}

type limitedContextReader struct {
	ctx    context.Context
	source io.Reader
	limit  int64
	read   int64
}

func (r *limitedContextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	if len(buffer) == 0 {
		return 0, nil
	}
	remaining := r.limit - r.read
	if remaining <= 0 {
		var extra [1]byte
		n, err := r.source.Read(extra[:])
		r.read += int64(n)
		if n > 0 {
			return 0, ErrLimitExceeded
		}
		return 0, err
	}
	if int64(len(buffer)) > remaining+1 {
		buffer = buffer[:remaining+1]
	}
	n, err := r.source.Read(buffer)
	r.read += int64(n)
	if r.read > r.limit {
		return 0, ErrLimitExceeded
	}
	return n, err
}

type parser struct {
	ctx     context.Context
	input   *bufio.Reader
	steps   uint64
	entries int64
	summary Summary
}

func (p *parser) peek() (byte, error) {
	if err := p.checkContext(); err != nil {
		return 0, err
	}
	value, err := p.input.Peek(1)
	if err != nil {
		return 0, err
	}
	return value[0], nil
}

func (p *parser) read() (byte, error) {
	p.steps++
	if p.steps&1023 == 0 {
		if err := p.checkContext(); err != nil {
			return 0, err
		}
	}
	return p.input.ReadByte()
}

func (p *parser) checkContext() error {
	if err := p.ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (p *parser) skipWhitespace() error {
	for {
		value, err := p.peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if value != ' ' && value != '\t' && value != '\n' && value != '\r' {
			return nil
		}
		if _, err := p.read(); err != nil {
			return err
		}
	}
}

func (p *parser) parseValue(depth int) error {
	if err := p.checkContext(); err != nil {
		return err
	}
	if depth > p.summary.MaxDepth {
		p.summary.MaxDepth = depth
	}
	if p.summary.Values >= maxValues {
		return ErrLimitExceeded
	}
	p.summary.Values++
	value, err := p.peek()
	if err != nil {
		return normalizeReadError(err)
	}
	switch value {
	case '{':
		if depth >= maxNestingDepth {
			return ErrLimitExceeded
		}
		return p.parseObject(depth + 1)
	case '[':
		if depth >= maxNestingDepth {
			return ErrLimitExceeded
		}
		return p.parseArray(depth + 1)
	case '"':
		_, err := p.parseString()
		if err == nil {
			p.summary.Strings++
		}
		return err
	case 't':
		return p.literal("true")
	case 'f':
		return p.literal("false")
	case 'n':
		return p.literal("null")
	default:
		if value == '-' || value >= '0' && value <= '9' {
			return p.number()
		}
		return ErrInvalidJSON
	}
}

func (p *parser) parseObject(depth int) error {
	if _, err := p.read(); err != nil { // opening brace
		return normalizeReadError(err)
	}
	p.summary.Objects++
	if err := p.skipWhitespace(); err != nil {
		return err
	}
	value, err := p.peek()
	if errors.Is(err, io.EOF) {
		return ErrInvalidJSON
	}
	if err != nil {
		return err
	}
	if value == '}' {
		if _, err := p.read(); err != nil {
			return normalizeReadError(err)
		}
		return nil
	}
	for {
		if err := p.checkEntryLimit(); err != nil {
			return err
		}
		value, err = p.peek()
		if err != nil {
			return normalizeReadError(err)
		}
		if value != '"' {
			return ErrInvalidJSON
		}
		if _, err := p.parseString(); err != nil {
			return err
		}
		p.summary.Strings++
		if err := p.skipWhitespace(); err != nil {
			return err
		}
		if value, err = p.read(); err != nil {
			return normalizeReadError(err)
		}
		if value != ':' {
			return ErrInvalidJSON
		}
		if err := p.skipWhitespace(); err != nil {
			return err
		}
		if err := p.parseValue(depth); err != nil {
			return err
		}
		if err := p.skipWhitespace(); err != nil {
			return err
		}
		value, err = p.read()
		if err != nil {
			return normalizeReadError(err)
		}
		if value == '}' {
			return nil
		}
		if value != ',' {
			return ErrInvalidJSON
		}
		if err := p.skipWhitespace(); err != nil {
			return err
		}
	}
}

func (p *parser) parseArray(depth int) error {
	if _, err := p.read(); err != nil { // opening bracket
		return normalizeReadError(err)
	}
	p.summary.Arrays++
	if err := p.skipWhitespace(); err != nil {
		return err
	}
	value, err := p.peek()
	if errors.Is(err, io.EOF) {
		return ErrInvalidJSON
	}
	if err != nil {
		return err
	}
	if value == ']' {
		if _, err := p.read(); err != nil {
			return normalizeReadError(err)
		}
		return nil
	}
	for {
		if err := p.checkEntryLimit(); err != nil {
			return err
		}
		if err := p.parseValue(depth); err != nil {
			return err
		}
		if err := p.skipWhitespace(); err != nil {
			return err
		}
		value, err = p.read()
		if err != nil {
			return normalizeReadError(err)
		}
		if value == ']' {
			return nil
		}
		if value != ',' {
			return ErrInvalidJSON
		}
		if err := p.skipWhitespace(); err != nil {
			return err
		}
	}
}

func (p *parser) checkEntryLimit() error {
	if p.entries >= maxEntries {
		return ErrLimitExceeded
	}
	p.entries++
	return nil
}

func (p *parser) parseString() ([]byte, error) {
	opening, err := p.read()
	if err != nil {
		return nil, normalizeReadError(err)
	}
	if opening != '"' {
		return nil, ErrInvalidJSON
	}
	value := make([]byte, 0, 64)
	escaped := false
	for {
		character, err := p.read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, ErrInvalidJSON
			}
			return nil, err
		}
		if !escaped && character == '"' {
			break
		}
		if !escaped && character < 0x20 {
			return nil, ErrInvalidJSON
		}
		if len(value) >= MaxStringBytes {
			return nil, ErrLimitExceeded
		}
		value = append(value, character)
		if escaped {
			escaped = false
		} else if character == '\\' {
			escaped = true
		}
	}
	if !utf8.Valid(value) || !validEscapes(value) {
		return nil, ErrInvalidJSON
	}
	return value, nil
}

func validEscapes(value []byte) bool {
	for index := 0; index < len(value); index++ {
		if value[index] != '\\' {
			continue
		}
		index++
		if index >= len(value) {
			return false
		}
		switch value[index] {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
		case 'u':
			if index+4 >= len(value) {
				return false
			}
			unit, ok := hexUnit(value[index+1 : index+5])
			if !ok {
				return false
			}
			index += 4
			if unit >= 0xD800 && unit <= 0xDBFF {
				if index+6 >= len(value) || value[index+1] != '\\' || value[index+2] != 'u' {
					return false
				}
				low, ok := hexUnit(value[index+3 : index+7])
				if !ok || low < 0xDC00 || low > 0xDFFF {
					return false
				}
				index += 6
			} else if unit >= 0xDC00 && unit <= 0xDFFF {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func hexUnit(value []byte) (uint16, bool) {
	if len(value) != 4 {
		return 0, false
	}
	var result uint16
	for _, digit := range value {
		result <<= 4
		switch {
		case digit >= '0' && digit <= '9':
			result |= uint16(digit - '0')
		case digit >= 'a' && digit <= 'f':
			result |= uint16(digit - 'a' + 10)
		case digit >= 'A' && digit <= 'F':
			result |= uint16(digit - 'A' + 10)
		default:
			return 0, false
		}
	}
	return result, true
}

func (p *parser) literal(expected string) error {
	for index := range expected {
		value, err := p.read()
		if err != nil {
			return normalizeReadError(err)
		}
		if value != expected[index] {
			return ErrInvalidJSON
		}
	}
	return nil
}

func (p *parser) number() error {
	value, err := p.peek()
	if err != nil {
		return normalizeReadError(err)
	}
	if value == '-' {
		_, _ = p.read()
		value, err = p.peek()
		if err != nil || value < '0' || value > '9' {
			if err != nil {
				return normalizeReadError(err)
			}
			return ErrInvalidJSON
		}
	}
	if value == '0' {
		_, _ = p.read()
		if next, err := p.peek(); err == nil && next >= '0' && next <= '9' {
			return ErrInvalidJSON
		} else if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	} else {
		if value < '1' || value > '9' {
			return ErrInvalidJSON
		}
		if err := p.consumeDigits(); err != nil {
			return err
		}
	}
	if value, err = p.peek(); err == nil && value == '.' {
		_, _ = p.read()
		next, peekErr := p.peek()
		if peekErr != nil || next < '0' || next > '9' {
			if peekErr != nil {
				return normalizeReadError(peekErr)
			}
			return ErrInvalidJSON
		}
		if err := p.consumeDigits(); err != nil {
			return err
		}
	} else if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if value, err = p.peek(); err == nil && (value == 'e' || value == 'E') {
		_, _ = p.read()
		if sign, signErr := p.peek(); signErr == nil && (sign == '+' || sign == '-') {
			_, _ = p.read()
		} else if signErr != nil && !errors.Is(signErr, io.EOF) {
			return signErr
		}
		next, peekErr := p.peek()
		if peekErr != nil || next < '0' || next > '9' {
			if peekErr != nil {
				return normalizeReadError(peekErr)
			}
			return ErrInvalidJSON
		}
		if err := p.consumeDigits(); err != nil {
			return err
		}
	} else if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func normalizeReadError(err error) error {
	if err == nil || errors.Is(err, io.EOF) {
		return ErrInvalidJSON
	}
	return err
}

func (p *parser) consumeDigits() error {
	for {
		value, err := p.peek()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if value < '0' || value > '9' {
			return nil
		}
		if _, err := p.read(); err != nil {
			return err
		}
	}
}
