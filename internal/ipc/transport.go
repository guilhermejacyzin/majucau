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
// StreamingHandler is used for the dashboard read path. It writes one JSON
// document to dst; the Named Pipe transport splits those bytes into bounded,
// ordered frames.
type StreamingHandler interface {
	HandleStream(context.Context, Request, io.Writer) (Response, error)
}
type HandlerFunc func(context.Context, Request) (Response, error)

func (f HandlerFunc) Handle(ctx context.Context, req Request) (Response, error) { return f(ctx, req) }

type Client interface {
	Call(context.Context, Request) (Response, error)
	Close() error
}
type StreamClient interface {
	CallStream(context.Context, Request, io.Writer) (Response, error)
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
	if err := writeAll(w, header); err != nil {
		return err
	}
	return writeAll(w, payload)
}

func writeAll(w io.Writer, payload []byte) error {
	for len(payload) > 0 {
		count, err := w.Write(payload)
		if err != nil {
			return err
		}
		if count <= 0 || count > len(payload) {
			return io.ErrShortWrite
		}
		payload = payload[count:]
	}
	return nil
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

func callOverStreamChunks(ctx context.Context, rw io.ReadWriter, req Request, dst io.Writer) (Response, error) {
	encoded, err := EncodeRequest(req)
	if err != nil {
		return Response{}, err
	}
	if err := writeFrame(rw, encoded); err != nil {
		return Response{}, err
	}
	var sequence uint64
	for {
		if err := ctx.Err(); err != nil {
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
		if response.Version != ProtocolVersion || response.RequestID != req.RequestID || response.Stream == nil || len(response.Payload) != 0 {
			return Response{}, ErrInvalidRequest
		}
		frame := response.Stream
		if frame.Sequence != sequence || len(frame.Data) > StreamChunkSize || (!frame.Done && (!response.OK || response.Error != nil || sequence == ^uint64(0))) {
			return Response{}, ErrInvalidRequest
		}
		if len(frame.Data) > 0 {
			if err := writeAll(dst, frame.Data); err != nil {
				return Response{}, err
			}
		}
		sequence++
		if frame.Done {
			if response.OK && response.Error != nil || !response.OK && response.Error == nil {
				return Response{}, ErrInvalidRequest
			}
			return response, nil
		}
	}
}

type responseStreamWriter struct {
	pipe      io.Writer
	requestID string
	sequence  uint64
	buffer    []byte
	err       error
}

func newResponseStreamWriter(pipe io.Writer, requestID string) *responseStreamWriter {
	return &responseStreamWriter{pipe: pipe, requestID: requestID, buffer: make([]byte, 0, StreamChunkSize)}
}

func (w *responseStreamWriter) Write(payload []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	total := len(payload)
	for len(payload) > 0 {
		space := StreamChunkSize - len(w.buffer)
		if space == 0 {
			if err := w.flush(false, Response{}); err != nil {
				w.err = err
				return 0, err
			}
			space = StreamChunkSize
		}
		count := len(payload)
		if count > space {
			count = space
		}
		w.buffer = append(w.buffer, payload[:count]...)
		payload = payload[count:]
	}
	return total, nil
}

func (w *responseStreamWriter) finish(response Response) error {
	if w.err != nil {
		response = NewErrorResponse(w.requestID, "INTERNAL_ERROR", "worker request failed")
	}
	return w.flush(true, response)
}

func (w *responseStreamWriter) Flush() error {
	if len(w.buffer) == 0 {
		return nil
	}
	err := w.flush(false, Response{})
	if err != nil {
		w.err = err
	}
	return err
}

func (w *responseStreamWriter) flush(done bool, terminal Response) error {
	response := terminal
	response.Version = ProtocolVersion
	response.RequestID = w.requestID
	if !done {
		response.OK = true
		response.Error = nil
	}
	if response.Error != nil {
		response.OK = false
	}
	data := append([]byte(nil), w.buffer...)
	response.Payload = nil
	response.Stream = &StreamFrame{Sequence: w.sequence, Done: done, Data: data}
	encoded, err := json.Marshal(response)
	if err != nil {
		return err
	}
	if len(encoded) > MaxMessageSize {
		return ErrFrameTooLarge
	}
	if err := writeFrame(w.pipe, encoded); err != nil {
		return err
	}
	w.sequence++
	w.buffer = w.buffer[:0]
	return nil
}
