//go:build windows

package ipc

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNamedPipeHealthRoundTripAndCancellation(t *testing.T) {
	name := fmt.Sprintf(`\\.\pipe\Majucau-test-%d`, time.Now().UnixNano())
	server, err := NewNamedPipeServer(NamedPipeConfig{Name: name, IOTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(ctx, HandlerFunc(func(ctx context.Context, req Request) (Response, error) {
			return NewResponse(req.RequestID, map[string]string{"state": "ok"})
		}))
	}()
	time.Sleep(50 * time.Millisecond)
	select {
	case runErr := <-serveErr:
		t.Fatalf("server exited before client: %v", runErr)
	default:
	}
	client, err := NewNamedPipeClient(NamedPipeConfig{Name: name, IOTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	callCtx, callCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer callCancel()
	response, err := client.Call(callCtx, Request{Version: ProtocolVersion, RequestID: "test", Method: MethodHealth})
	if err != nil {
		select {
		case runErr := <-serveErr:
			t.Fatalf("server: %v; client: %v", runErr, err)
		default:
			t.Error(err)
			return
		}
	}
	if !response.OK || string(response.Payload) != `{"state":"ok"}` {
		t.Fatalf("response: %#v", response)
	}
	cancel()
	_ = server.Close()
}
func TestNamedPipeRejectsUnauthorizedSID(t *testing.T) {
	name := fmt.Sprintf(`\\.\pipe\Majucau-unauthorized-%d`, time.Now().UnixNano())
	server, err := NewNamedPipeServer(NamedPipeConfig{Name: name, AllowedClientSIDs: []string{"S-1-5-21-999999999-999999999-999999999-9999"}, IOTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer server.Close()
	defer cancel()
	go func() {
		_ = server.Serve(ctx, HandlerFunc(func(context.Context, Request) (Response, error) { return Response{}, nil }))
	}()
	client, err := NewNamedPipeClient(NamedPipeConfig{Name: name, IOTimeout: 300 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	_, err = client.Call(context.Background(), Request{Version: ProtocolVersion, RequestID: "unauthorized", Method: MethodHealth})
	if err == nil || strings.Contains(strings.ToLower(err.Error()), "success") {
		t.Fatalf("unauthorized call unexpectedly succeeded: %v", err)
	}
}
