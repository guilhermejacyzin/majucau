package ipc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

var (
	ErrFrameTooLarge        = errors.New("ipc frame exceeds maximum size")
	ErrFrameEmpty           = errors.New("ipc frame is empty")
	ErrUnauthorizedClient   = errors.New("ipc client is not authorized")
	ErrTransportUnavailable = errors.New("ipc transport unavailable")
)

const frameHeaderSize = 4

type Handler interface {
	Handle(context.Context, Request) (Response, error)
}
type HandlerFunc func(context.Context, Request) (Response, error)

func (f HandlerFunc) Handle(ctx context.Context, req Request) (Response, error) { return f(ctx, req) }

type Client interface {
	Call(context.Context, Request) (Response, error)
	Close() error
}
type Server interface {
	Serve(context.Context, Handler) error
	Close() error
}

func writeFrame(w io.Writer, payload []byte) error {
	if len(payload) == 0 {
		return ErrFrameEmpty
	}
	if len(payload) > MaxMessageSize {
		return ErrFrameTooLarge
	}
	header := make([]byte, frameHeaderSize)
	binary.BigEndian.PutUint32(header, uint32(len(payload)))
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}
func readFrame(r io.Reader) ([]byte, error) {
	header := make([]byte, frameHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(header)
	if n == 0 {
		return nil, ErrFrameEmpty
	}
	if n > MaxMessageSize {
		return nil, ErrFrameTooLarge
	}
	payload := make([]byte, n)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func callOverStream(ctx context.Context, rw io.ReadWriter, req Request) (Response, error) {
	encoded, err := EncodeRequest(req)
	if err != nil {
		return Response{}, err
	}
	if err := writeFrame(rw, encoded); err != nil {
		return Response{}, err
	}
	payload, err := readFrame(rw)
	if err != nil {
		return Response{}, err
	}
	var response Response
	if err := json.Unmarshal(payload, &response); err != nil {
		return Response{}, ErrInvalidRequest
	}
	if response.Version != ProtocolVersion || response.RequestID != req.RequestID {
		return Response{}, ErrInvalidRequest
	}
	return response, nil
}
