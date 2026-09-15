//go:build !windows

package ipc

import (
	"context"
)

const DefaultPipeName = `\\.\pipe\Majucau-worker`

type NamedPipeConfig struct {
	Name              string
	AllowedClientSIDs []string
	ServiceSID        string
}
type NamedPipeServer struct{}
type NamedPipeClient struct{}

func DefaultNamedPipeConfig() NamedPipeConfig { return NamedPipeConfig{Name: DefaultPipeName} }
func NewNamedPipeServer(NamedPipeConfig) (*NamedPipeServer, error) {
	return nil, ErrTransportUnavailable
}
func (*NamedPipeServer) Serve(context.Context, Handler) error { return ErrTransportUnavailable }
func (*NamedPipeServer) Close() error                         { return nil }
func NewNamedPipeClient(NamedPipeConfig) (*NamedPipeClient, error) {
	return nil, ErrTransportUnavailable
}
func (*NamedPipeClient) Call(context.Context, Request) (Response, error) {
	return Response{}, ErrTransportUnavailable
}
func (*NamedPipeClient) Close() error { return nil }
