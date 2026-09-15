package ipc

import "context"

// FakeClient exercises the same request/response contract without opening a
// handle. It is intentionally opt-in and never used by the worker runtime.
type FakeClient struct {
	Handler    Handler
	Authorized bool
}

func NewFakeClient(handler Handler) *FakeClient {
	return &FakeClient{Handler: handler, Authorized: true}
}
func (c *FakeClient) Call(ctx context.Context, req Request) (Response, error) {
	if !c.Authorized {
		return Response{}, ErrUnauthorizedClient
	}
	if err := req.Validate(); err != nil {
		return Response{}, err
	}
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	if c.Handler == nil {
		return Response{}, ErrTransportUnavailable
	}
	return c.Handler.Handle(ctx, req)
}
func (*FakeClient) Close() error { return nil }
