//go:build windows

package ipc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestNamedPipeHealthRoundTripAndCancellation(t *testing.T) {
	name := fmt.Sprintf(`\\.\pipe\Majucau-test-%d`, time.Now().UnixNano())
	server, err := NewNamedPipeServer(NamedPipeConfig{Name: name, IOTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
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
	response, err = client.Call(callCtx, Request{Version: ProtocolVersion, RequestID: "test-second-instance", Method: MethodHealth})
	if err != nil {
		t.Fatalf("second server pipe instance: %v", err)
	}
	if !response.OK || string(response.Payload) != `{"state":"ok"}` {
		t.Fatalf("second response: %#v", response)
	}
	cancel()
	_ = server.Close()
}

func TestNamedPipeConcurrentClientHealthRequests(t *testing.T) {
	name := fmt.Sprintf(`\\.\pipe\Majucau-concurrent-%d`, time.Now().UnixNano())
	server, err := NewNamedPipeServer(NamedPipeConfig{Name: name, IOTimeout: 3 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(ctx, HandlerFunc(func(_ context.Context, req Request) (Response, error) {
			return NewResponse(req.RequestID, map[string]string{"request_id": req.RequestID})
		}))
	}()

	const clientCount = 8
	type callResult struct {
		requestID string
		response  Response
		err       error
	}
	start := make(chan struct{})
	results := make(chan callResult, clientCount)
	for index := 0; index < clientCount; index++ {
		requestID := fmt.Sprintf("concurrent-%02d", index)
		go func(requestID string) {
			<-start
			client, err := NewNamedPipeClient(NamedPipeConfig{Name: name, IOTimeout: 8 * time.Second})
			if err != nil {
				results <- callResult{requestID: requestID, err: err}
				return
			}
			defer client.Close()
			response, err := client.Call(ctx, Request{Version: ProtocolVersion, RequestID: requestID, Method: MethodHealth})
			results <- callResult{requestID: requestID, response: response, err: err}
		}(requestID)
	}
	close(start)

	seen := make(map[string]bool, clientCount)
	for index := 0; index < clientCount; index++ {
		select {
		case result := <-results:
			if result.err != nil {
				t.Fatalf("concurrent IPC request %q failed: %v", result.requestID, result.err)
			}
			if seen[result.requestID] {
				t.Fatalf("duplicate IPC response for request %q", result.requestID)
			}
			seen[result.requestID] = true
			expected := fmt.Sprintf(`{"request_id":%q}`, result.requestID)
			if !result.response.OK || string(result.response.Payload) != expected {
				t.Fatalf("unexpected IPC response for %q: %#v", result.requestID, result.response)
			}
		case <-ctx.Done():
			t.Fatalf("timed out waiting for concurrent IPC requests: %v", ctx.Err())
		}
	}
	if len(seen) != clientCount {
		t.Fatalf("received %d distinct IPC responses; want %d", len(seen), clientCount)
	}

	cancel()
	_ = server.Close()
	select {
	case err := <-serveErr:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("server stopped with unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop after concurrent clients finished")
	}
}

func TestNamedPipeAppliesProtectedDACLToCreatedPipe(t *testing.T) {
	name := fmt.Sprintf(`\\.\pipe\Majucau-dacl-%d`, time.Now().UnixNano())
	clientSID, err := currentProcessSID()
	if err != nil {
		t.Fatalf("read current process SID: %v", err)
	}
	serviceSID := "S-1-5-19" // LocalService.
	security, release, err := makePipeSecurity(NamedPipeConfig{AllowedClientSIDs: []string{clientSID}, ServiceSID: serviceSID})
	if err != nil {
		t.Fatalf("build named pipe security descriptor: %v", err)
	}
	t.Cleanup(release)

	pipe, err := createPipe(name, security)
	if err != nil {
		t.Fatalf("create named pipe with explicit DACL: %v", err)
	}
	t.Cleanup(func() { closePipeHandle(pipe) })

	descriptor, err := windows.GetSecurityInfo(windows.Handle(pipe), windows.SE_KERNEL_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatalf("read created named pipe DACL: %v", err)
	}
	if descriptor == nil {
		t.Fatal("created named pipe has no security descriptor")
	}
	control, _, err := descriptor.Control()
	if err != nil {
		t.Fatalf("read named pipe security descriptor control: %v", err)
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatalf("named pipe DACL is not protected: %s", descriptor)
	}

	dacl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatalf("read named pipe DACL: %v", err)
	}
	if control&windows.SE_DACL_PRESENT == 0 || dacl == nil {
		t.Fatal("created named pipe has no DACL")
	}
	if dacl.AceCount != 4 {
		t.Fatalf("named pipe DACL contains %d ACEs; want only the client, service, administrators, and SYSTEM", dacl.AceCount)
	}
	// The pipe object maps GENERIC_ALL to FILE_ALL_ACCESS when applying its DACL.
	const fileAllAccessMask = windows.ACCESS_MASK(0x001F01FF)
	allowedSIDs := map[string]windows.ACCESS_MASK{
		clientSID:      windows.ACCESS_MASK(pipeClientAccessMask),
		serviceSID:     windows.ACCESS_MASK(pipeServerAccessMask),
		"S-1-5-32-544": fileAllAccessMask, // BUILTIN\Administrators
		"S-1-5-18":     fileAllAccessMask, // LocalSystem
	}
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil {
			t.Fatalf("read named pipe ACE %d: %v", index, err)
		}
		if ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			t.Fatalf("named pipe ACE %d is not an access-allowed ACE", index)
		}
		if ace.Header.AceFlags != 0 {
			t.Errorf("named pipe ACE %d unexpectedly has inheritance flags %#x", index, ace.Header.AceFlags)
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart)).String()
		wantMask, allowed := allowedSIDs[sid]
		if !allowed {
			t.Errorf("named pipe DACL contains unexpected trustee SID %q", sid)
		} else if ace.Mask != wantMask {
			t.Errorf("named pipe ACE %d for SID %q has mask %#x; want %#x", index, sid, ace.Mask, wantMask)
		}
		delete(allowedSIDs, sid)
	}
	if len(allowedSIDs) != 0 {
		t.Errorf("named pipe DACL is missing authorized SIDs: %v", allowedSIDs)
	}
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

func TestNamedPipeAcceptsAuthorizedClientProcess(t *testing.T) {
	name := fmt.Sprintf(`\\.\pipe\Majucau-process-%d`, time.Now().UnixNano())
	processSID, err := currentProcessSID()
	if err != nil {
		t.Fatalf("read current process SID: %v", err)
	}
	server, err := NewNamedPipeServer(NamedPipeConfig{
		Name:              name,
		AllowedClientSIDs: []string{processSID},
		ServiceSID:        processSID,
		IOTimeout:         3 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(ctx, HandlerFunc(func(_ context.Context, req Request) (Response, error) {
			return NewResponse(req.RequestID, map[string]string{"state": "authorized-child"})
		}))
	}()

	childCtx, childCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer childCancel()
	command := exec.CommandContext(childCtx, os.Args[0], "-test.run=^TestNamedPipeAuthorizedClientProcessHelper$")
	const helperEnv = "MAJUCAU_NAMED_PIPE_HELPER"
	command.Env = make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, helperEnv+"=") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, helperEnv+"="+name)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("authorized client process failed: %v\n%s", err, output)
	}

	cancel()
	_ = server.Close()
	select {
	case err := <-serveErr:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("server stopped with unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop after the authorized client process finished")
	}
}

func TestNamedPipeAuthorizedClientProcessHelper(t *testing.T) {
	name := os.Getenv("MAJUCAU_NAMED_PIPE_HELPER")
	if name == "" {
		return
	}
	client, err := NewNamedPipeClient(NamedPipeConfig{Name: name, IOTimeout: 3 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := client.Call(ctx, Request{Version: ProtocolVersion, RequestID: "authorized-child", Method: MethodHealth})
	if err != nil {
		t.Fatalf("health request from authorized client process: %v", err)
	}
	if !response.OK || string(response.Payload) != `{"state":"authorized-child"}` {
		t.Fatalf("unexpected health response from authorized client process: %#v", response)
	}
}
