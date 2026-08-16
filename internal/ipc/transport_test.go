package ipc

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"testing"
)

func TestFramingRejectsEmptyAndOversizedPayloads(t *testing.T) {
	if err := writeFrame(&bytes.Buffer{}, nil); !errors.Is(err, ErrFrameEmpty) {
		t.Fatalf("empty frame: %v", err)
	}
	var header bytes.Buffer
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, MaxMessageSize+1)
	_, _ = header.Write(b)
	if _, err := readFrame(&header); !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("oversized frame: %v", err)
	}
}
func TestFakeTransportAuthorizationAndCancellation(t *testing.T) {
	h := HandlerFunc(func(_ context.Context, req Request) (Response, error) {
		return NewResponse(req.RequestID, map[string]string{"ok": "yes"})
	})
	c := NewFakeClient(h)
	req := Request{Version: ProtocolVersion, RequestID: "r1", Method: MethodHealth}
	if _, err := c.Call(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	c.Authorized = false
	if _, err := c.Call(context.Background(), req); !errors.Is(err, ErrUnauthorizedClient) {
		t.Fatalf("unauthorized client: %v", err)
	}
	c.Authorized = true
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Call(ctx, req); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled call: %v", err)
	}
}
