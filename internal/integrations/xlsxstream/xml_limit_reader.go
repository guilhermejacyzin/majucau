package xlsxstream

import (
	"context"
	"errors"
	"io"
)

var errXMLTokenLimit = errors.New("xlsx XML token exceeds configured limit")

type xmlLimitMode uint8

const (
	xmlLimitText xmlLimitMode = iota
	xmlLimitAfterLessThan
	xmlLimitTag
	xmlLimitBang
	xmlLimitComment
	xmlLimitCDATA
	xmlLimitProcessingInstruction
	xmlLimitDeclaration
)

// xmlTokenLimitReader bounds a single XML token before encoding/xml can grow
// its internal buffer to hold that token. The source is still consumed in
// fixed-size chunks, and only bytes that remain within the limit reach the
// decoder.
type xmlTokenLimitReader struct {
	ctx    context.Context
	reader io.Reader
	limit  int

	mode       xmlLimitMode
	tokenBytes int
	quote      byte
	brackets   int
	previous   byte
	tail       [2]byte
	tailBytes  int
	bang       [7]byte
	bangBytes  int
	exceeded   bool

	input     [32 << 10]byte
	inputPos  int
	inputSize int
	inputErr  error
}

func newXMLTokenLimitReader(ctx context.Context, reader io.Reader, limit int) *xmlTokenLimitReader {
	return &xmlTokenLimitReader{ctx: ctx, reader: reader, limit: limit}
}

func (reader *xmlTokenLimitReader) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	if reader.exceeded {
		return 0, errXMLTokenLimit
	}
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}

	written := 0
	for written < len(buffer) {
		if reader.inputPos == reader.inputSize {
			if reader.inputErr != nil {
				err := reader.inputErr
				reader.inputErr = nil
				if written > 0 {
					return written, nil
				}
				return 0, err
			}
			n, err := reader.reader.Read(reader.input[:])
			if n == 0 {
				if err != nil {
					if written > 0 {
						return written, nil
					}
					return 0, err
				}
				return written, nil
			}
			reader.inputPos = 0
			reader.inputSize = n
			reader.inputErr = err
		}

		value := reader.input[reader.inputPos]
		if reader.advance(value) {
			reader.exceeded = true
			if written > 0 {
				return written, nil
			}
			return 0, errXMLTokenLimit
		}
		reader.inputPos++
		buffer[written] = value
		written++
	}
	return written, nil
}

func (reader *xmlTokenLimitReader) advance(value byte) bool {
	if reader.mode == xmlLimitText {
		if value == '<' {
			reader.tokenBytes = 0
			reader.mode = xmlLimitAfterLessThan
			return reader.addTokenByte()
		}
		return reader.addTokenByte()
	}

	if reader.addTokenByte() {
		return true
	}
	switch reader.mode {
	case xmlLimitAfterLessThan:
		switch value {
		case '?':
			reader.mode = xmlLimitProcessingInstruction
			reader.previous = 0
		case '!':
			reader.mode = xmlLimitBang
			reader.bangBytes = 0
		case '"', '\'':
			reader.mode = xmlLimitTag
			reader.quote = value
		case '>':
			reader.finishToken()
		default:
			reader.mode = xmlLimitTag
		}
	case xmlLimitTag:
		reader.advanceQuotedTag(value)
	case xmlLimitBang:
		if reader.advanceBang(value) {
			reader.finishToken()
		}
	case xmlLimitComment:
		if reader.hasSuffix(value, '-', '-') {
			reader.finishToken()
		}
	case xmlLimitCDATA:
		if reader.hasSuffix(value, ']', ']') {
			reader.finishToken()
		}
	case xmlLimitProcessingInstruction:
		if reader.previous == '?' && value == '>' {
			reader.finishToken()
		} else {
			reader.previous = value
		}
	case xmlLimitDeclaration:
		if reader.advanceDeclaration(value) {
			reader.finishToken()
		}
	}
	return false
}

func (reader *xmlTokenLimitReader) addTokenByte() bool {
	reader.tokenBytes++
	return reader.tokenBytes > reader.limit
}

func (reader *xmlTokenLimitReader) finishToken() {
	reader.mode = xmlLimitText
	reader.tokenBytes = 0
	reader.quote = 0
	reader.brackets = 0
	reader.previous = 0
	reader.tailBytes = 0
	reader.bangBytes = 0
}

func (reader *xmlTokenLimitReader) advanceQuotedTag(value byte) {
	if reader.quote != 0 {
		if value == reader.quote {
			reader.quote = 0
		}
		return
	}
	if value == '"' || value == '\'' {
		reader.quote = value
	} else if value == '>' {
		reader.finishToken()
	}
}

func (reader *xmlTokenLimitReader) advanceBang(value byte) bool {
	if reader.bangBytes == len(reader.bang) {
		reader.mode = xmlLimitDeclaration
		for _, prefixByte := range reader.bang[:reader.bangBytes] {
			if reader.advanceDeclaration(prefixByte) {
				return true
			}
		}
		return reader.advanceDeclaration(value)
	}
	reader.bang[reader.bangBytes] = value
	reader.bangBytes++
	prefix := reader.bang[:reader.bangBytes]
	if prefixMatches(prefix, "--") {
		reader.mode = xmlLimitComment
		reader.tailBytes = 0
		return false
	}
	if prefixMatches(prefix, "[CDATA[") {
		reader.mode = xmlLimitCDATA
		reader.tailBytes = 0
		return false
	}
	if prefixOf(prefix, "--") || prefixOf(prefix, "[CDATA[") {
		return false
	}
	reader.mode = xmlLimitDeclaration
	for _, prefixByte := range prefix {
		if reader.advanceDeclaration(prefixByte) {
			return true
		}
	}
	return false
}

func prefixMatches(prefix []byte, target string) bool {
	if len(prefix) != len(target) {
		return false
	}
	for index := range prefix {
		if prefix[index] != target[index] {
			return false
		}
	}
	return true
}

func prefixOf(prefix []byte, target string) bool {
	if len(prefix) > len(target) {
		return false
	}
	for index := range prefix {
		if prefix[index] != target[index] {
			return false
		}
	}
	return true
}

func (reader *xmlTokenLimitReader) advanceDeclaration(value byte) bool {
	if reader.quote != 0 {
		if value == reader.quote {
			reader.quote = 0
		}
		return false
	}
	switch value {
	case '"', '\'':
		reader.quote = value
	case '[':
		reader.brackets++
	case ']':
		if reader.brackets > 0 {
			reader.brackets--
		}
	case '>':
		return reader.brackets == 0
	}
	return false
}

func (reader *xmlTokenLimitReader) hasSuffix(value, first, second byte) bool {
	closed := reader.tailBytes == 2 && reader.tail[0] == first && reader.tail[1] == second && value == '>'
	if reader.tailBytes == 0 {
		reader.tail[0] = value
		reader.tailBytes = 1
	} else if reader.tailBytes == 1 {
		reader.tail[1] = value
		reader.tailBytes = 2
	} else {
		reader.tail[0] = reader.tail[1]
		reader.tail[1] = value
	}
	return closed
}
