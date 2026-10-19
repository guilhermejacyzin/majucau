package csvlimit

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNewTSVReaderEnforcesFieldCountBeforeCSVParsing(t *testing.T) {
	fields := make([]string, MaxFields+1)
	for index := range fields {
		fields[index] = "x"
	}
	_, err := io.Copy(io.Discard, NewTSVReader(strings.NewReader(strings.Join(fields, "\t"))))
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("TSV field limit error = %v, want %v", err, ErrLimitExceeded)
	}
}

func TestNewReaderKeepsSemicolonDelimiterAndEscapedFields(t *testing.T) {
	valid := `"left;inside";right` + "\n"
	if _, err := io.Copy(io.Discard, NewReader(strings.NewReader(valid))); err != nil {
		t.Fatalf("quoted semicolon was counted as a field delimiter: %v", err)
	}

	fields := make([]string, MaxFields+1)
	for index := range fields {
		fields[index] = "x"
	}
	_, err := io.Copy(io.Discard, NewReader(strings.NewReader(strings.Join(fields, ";"))))
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("CSV field limit error = %v, want %v", err, ErrLimitExceeded)
	}
}

func TestNewTSVReaderEnforcesRecordLimitAcrossQuotedLines(t *testing.T) {
	content := "\"" + strings.Repeat("x\n", MaxRecordBytes/2+1) + "\""
	_, err := io.Copy(io.Discard, NewTSVReader(strings.NewReader(content)))
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("TSV record limit error = %v, want %v", err, ErrLimitExceeded)
	}
}
