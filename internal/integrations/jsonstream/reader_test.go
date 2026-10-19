package jsonstream

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestValidateStreamsNestedJSONAndReportsOnlyStructure(t *testing.T) {
	input := []byte(` {"rows":[{"name":"caf\u00e9","amount":-12.50e+2,"active":true},null],"empty":[]} `)
	summary, err := Validate(context.Background(), bytes.NewReader(input), int64(len(input)))
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if summary.FileBytes != int64(len(input)) || summary.Values != 8 || summary.Objects != 2 || summary.Arrays != 2 || summary.Strings != 6 || summary.MaxDepth != 3 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestValidateRejectsMalformedJSON(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte(" ")},
		{name: "trailing comma", data: []byte(`{"a":1,}`)},
		{name: "leading zero", data: []byte(`01`)},
		{name: "missing fraction digits", data: []byte(`1.`)},
		{name: "missing exponent digits", data: []byte(`1e+`)},
		{name: "invalid escape", data: []byte(`"\x"`)},
		{name: "unpaired high surrogate", data: []byte(`"\ud800"`)},
		{name: "unpaired low surrogate", data: []byte(`"\udc00"`)},
		{name: "unescaped control", data: []byte{'"', 'a', '\n', 'b', '"'}},
		{name: "invalid utf8", data: []byte{'"', 0xff, '"'}},
		{name: "trailing document", data: []byte(`true false`)},
		{name: "truncated array", data: []byte(`[1,2`)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Validate(context.Background(), bytes.NewReader(test.data), int64(len(test.data)))
			if !errors.Is(err, ErrInvalidJSON) {
				t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidJSON)
			}
		})
	}
}

func TestValidateEnforcesStringNestingAndFileLimits(t *testing.T) {
	tooLongString := []byte(`"` + strings.Repeat("x", MaxStringBytes+1) + `"`)
	tooDeep := []byte(strings.Repeat("[", maxNestingDepth+1) + "0" + strings.Repeat("]", maxNestingDepth+1))
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "string limit", data: tooLongString, want: ErrLimitExceeded},
		{name: "nesting limit", data: tooDeep, want: ErrLimitExceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Validate(context.Background(), bytes.NewReader(test.data), int64(len(test.data)))
			if !errors.Is(err, test.want) {
				t.Fatalf("Validate() error = %v, want %v", err, test.want)
			}
		})
	}
	if _, err := Validate(context.Background(), strings.NewReader("{}"), MaxFileBytes+1); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("oversized declared file error = %v, want %v", err, ErrLimitExceeded)
	}
}

func TestValidateRequiresExactSizeAndHonorsCancellation(t *testing.T) {
	if _, err := Validate(context.Background(), strings.NewReader("{}"), 3); !errors.Is(err, ErrInvalidJSON) {
		t.Fatalf("size mismatch error = %v, want %v", err, ErrInvalidJSON)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Validate(ctx, strings.NewReader("{}"), 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v, want %v", err, context.Canceled)
	}

	limited := &limitedContextReader{ctx: context.Background(), source: strings.NewReader(`"value"`), limit: 1}
	parser := parser{ctx: context.Background(), input: bufio.NewReaderSize(limited, 1)}
	if err := parser.parseValue(0); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("stream limit error = %v, want %v", err, ErrLimitExceeded)
	}
}
