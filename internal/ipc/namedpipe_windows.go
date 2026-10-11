//go:build windows

package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const DefaultPipeName = `\\.\pipe\Majucau-worker`

type NamedPipeConfig struct {
	Name              string
	AllowedClientSIDs []string
	ServiceSID        string
	IOTimeout         time.Duration
}

func DefaultNamedPipeConfig() NamedPipeConfig {
	return NamedPipeConfig{Name: DefaultPipeName, IOTimeout: 10 * time.Second}
}
func (c NamedPipeConfig) validate() error {
	if !strings.HasPrefix(c.Name, `\\.\pipe\`) || strings.ContainsRune(c.Name, 0) || len(c.Name) > 240 {
		return errors.New("invalid named pipe name")
	}
	if strings.Contains(c.Name[len(`\\.\pipe\`):], `\`) {
		return errors.New("invalid named pipe name")
	}
	if c.IOTimeout <= 0 {
		return errors.New("named pipe timeout must be positive")
	}
	return nil
}

type NamedPipeServer struct {
	config  NamedPipeConfig
	mu      sync.Mutex
	handles map[syscall.Handle]*serverPipeHandle
	closed  bool
}

type serverPipeHandle struct {
	handle syscall.Handle
	file   *os.File
	once   sync.Once
	err    error
}

func (p *serverPipeHandle) close() error {
	p.once.Do(func() {
		p.err = disconnectPipe(p.handle)
		if p.file != nil {
			p.err = errors.Join(p.err, p.file.Close())
			return
		}
		p.err = errors.Join(p.err, syscall.CloseHandle(p.handle))
	})
	return p.err
}

func NewNamedPipeServer(config NamedPipeConfig) (*NamedPipeServer, error) {
	if config.Name == "" {
		config.Name = DefaultPipeName
	}
	if config.IOTimeout == 0 {
		config.IOTimeout = 10 * time.Second
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return &NamedPipeServer{config: config, handles: make(map[syscall.Handle]*serverPipeHandle)}, nil
}
func (s *NamedPipeServer) Close() error {
	s.mu.Lock()
	s.closed = true
	handles := make([]*serverPipeHandle, 0, len(s.handles))
	for _, pipe := range s.handles {
		handles = append(handles, pipe)
	}
	s.handles = make(map[syscall.Handle]*serverPipeHandle)
	s.mu.Unlock()
	var closeErr error
	for _, pipe := range handles {
		closeErr = errors.Join(closeErr, pipe.close())
	}
	return closeErr
}
func (s *NamedPipeServer) track(pipe *serverPipeHandle) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.handles[pipe.handle] = pipe
	return true
}
func (s *NamedPipeServer) untrack(h syscall.Handle) { s.mu.Lock(); delete(s.handles, h); s.mu.Unlock() }

func (s *NamedPipeServer) Serve(ctx context.Context, handler Handler) (returnErr error) {
	if handler == nil {
		return errors.New("ipc handler is required")
	}
	security, release, err := makePipeSecurity(s.config)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, release()) }()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		h, err := createPipe(s.config.Name, security)
		if err != nil {
			return err
		}
		pipe := &serverPipeHandle{handle: h, file: os.NewFile(uintptr(h), "majucau-pipe-server")}
		if pipe.file == nil {
			return errors.Join(ErrTransportUnavailable, cleanupFailure("close named pipe", syscall.CloseHandle(h)))
		}
		if !s.track(pipe) {
			return errors.Join(context.Canceled, pipe.close())
		}
		err = connectPipe(ctx, h)
		var authErr error
		if err != nil {
			var cleanupErr *pipeResourceCleanupError
			if errors.As(err, &cleanupErr) {
				authErr = err
			}
		} else {
			var authorized bool
			authorized, authErr = authorizePipeClient(h, s.config)
			if authorized && authErr == nil {
				_ = s.handleConnection(ctx, pipe, handler)
			}
		}
		// Finish closing this instance before creating the next one so a client
		// cannot attach to a stale, disconnected pipe instance.
		closeErr := pipe.close()
		s.untrack(h)
		if closeErr != nil || authErr != nil {
			return errors.Join(closeErr, authErr)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}
func (s *NamedPipeServer) handleConnection(ctx context.Context, pipe *serverPipeHandle, handler Handler) error {
	f := pipe.file
	if f == nil {
		return ErrTransportUnavailable
	}
	read := make(chan struct {
		payload []byte
		err     error
	}, 1)
	go func() {
		p, err := readFrame(f)
		read <- struct {
			payload []byte
			err     error
		}{p, err}
	}()
	var frame struct {
		payload []byte
		err     error
	}
	select {
	case frame = <-read:
	case <-ctx.Done():
		return ctx.Err()
	}
	if frame.err != nil {
		return frame.err
	}
	request, err := DecodeRequest(frame.payload)
	if err != nil {
		return err
	}
	requestCtx, cancel := context.WithTimeout(ctx, s.config.IOTimeout)
	defer cancel()
	// The pipe is opened for overlapped I/O so this disconnect watcher can read
	// concurrently with response writes without blocking the same handle.
	go func() {
		var trailing [1]byte
		_, _ = f.Read(trailing[:])
		cancel()
	}()
	if request.Method == MethodDashboardSnapshot {
		if streamHandler, ok := handler.(StreamingHandler); ok {
			stream := newResponseStreamWriter(f, request.RequestID)
			response, streamErr := streamHandler.HandleStream(requestCtx, request, stream)
			if streamErr != nil {
				response = NewErrorResponse(request.RequestID, "INTERNAL_ERROR", "worker request failed")
			}
			if err := stream.finish(response); err != nil {
				return err
			}
			// DisconnectNamedPipe can discard buffered bytes; force delivery before
			// the single-request connection is closed.
			return f.Sync()
		}
	}
	response, err := handler.Handle(requestCtx, request)
	if err != nil {
		response = NewErrorResponse(request.RequestID, "INTERNAL_ERROR", "worker request failed")
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		return err
	}
	if err := writeFrame(f, encoded); err != nil {
		return err
	}
	// DisconnectNamedPipe can discard buffered bytes; force delivery before the
	// single-request connection is closed.
	return f.Sync()
}

type NamedPipeClient struct {
	name    string
	timeout time.Duration
	mu      sync.Mutex
	file    *os.File
	closed  bool
}

func NewNamedPipeClient(config NamedPipeConfig) (*NamedPipeClient, error) {
	if config.Name == "" {
		config.Name = DefaultPipeName
	}
	if config.IOTimeout == 0 {
		config.IOTimeout = 10 * time.Second
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return &NamedPipeClient{name: config.Name, timeout: config.IOTimeout}, nil
}
func (c *NamedPipeClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.file != nil {
		return closePipeFile(c.file)
	}
	return nil
}
func (c *NamedPipeClient) Call(ctx context.Context, request Request) (Response, error) {
	return c.call(ctx, request, nil)
}

func (c *NamedPipeClient) CallStream(ctx context.Context, request Request, dst io.Writer) (Response, error) {
	if dst == nil {
		return Response{}, ErrInvalidRequest
	}
	return c.call(ctx, request, dst)
}

func (c *NamedPipeClient) call(ctx context.Context, request Request, streamDst io.Writer) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	if err := request.Validate(); err != nil {
		return Response{}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	if deadline, ok := ctx.Deadline(); ok {
		cancel()
		callCtx, cancel = context.WithDeadline(ctx, deadline)
	}
	defer cancel()
	result := make(chan struct {
		response Response
		err      error
	}, 1)
	go func() {
		f, err := openPipeFile(callCtx, c.name)
		if err != nil {
			result <- struct {
				response Response
				err      error
			}{err: err}
			return
		}
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			closeErr := closePipeFile(f)
			result <- struct {
				response Response
				err      error
			}{err: errors.Join(ErrTransportUnavailable, closeErr)}
			return
		}
		c.file = f
		c.mu.Unlock()
		var resp Response
		if streamDst == nil {
			resp, err = callOverStream(callCtx, f, request)
		} else {
			resp, err = callOverStreamChunks(callCtx, f, request, streamDst)
		}
		closeErr := closePipeFile(f)
		c.mu.Lock()
		c.file = nil
		c.mu.Unlock()
		err = errors.Join(err, closeErr)
		result <- struct {
			response Response
			err      error
		}{resp, err}
	}()
	select {
	case out := <-result:
		return out.response, out.err
	case <-callCtx.Done():
		return Response{}, errors.Join(callCtx.Err(), c.Close())
	}
}

var (
	kernel32                         = syscall.NewLazyDLL("kernel32.dll")
	advapi32                         = syscall.NewLazyDLL("advapi32.dll")
	createNamedPipeW                 = kernel32.NewProc("CreateNamedPipeW")
	createFileW                      = kernel32.NewProc("CreateFileW")
	connectNamedPipe                 = kernel32.NewProc("ConnectNamedPipe")
	disconnectNamedPipe              = kernel32.NewProc("DisconnectNamedPipe")
	getNamedPipeClientProcessID      = kernel32.NewProc("GetNamedPipeClientProcessId")
	openProcess                      = kernel32.NewProc("OpenProcess")
	openProcessToken                 = advapi32.NewProc("OpenProcessToken")
	getTokenInformation              = advapi32.NewProc("GetTokenInformation")
	convertSidToStringSidW           = advapi32.NewProc("ConvertSidToStringSidW")
	convertStringSecurityDescriptorW = advapi32.NewProc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
	localFree                        = kernel32.NewProc("LocalFree")
)

const (
	pipeAccessDuplex                             = 0x00000003
	pipeOpenOverlapped                           = 0x40000000
	pipeTypeMessage                              = 0x00000004
	pipeReadModeMessage                          = 0x00000002
	pipeWait                                     = 0x00000000
	pipeUnlimitedInstances                       = 255
	pipeClientAccessMask                         = 0x00100083 // FILE_READ_DATA | FILE_WRITE_DATA | FILE_READ_ATTRIBUTES | SYNCHRONIZE.
	pipeServerAccessMask                         = 0x00100087 // Client I/O rights plus FILE_CREATE_PIPE_INSTANCE.
	pipeOpenRetryDelay                           = 10 * time.Millisecond
	openExisting                                 = 3
	processQueryLimitedInformation               = 0x1000
	tokenQuery                                   = 0x0008
	tokenUser                                    = 1
	errorPipeBusy                  syscall.Errno = 231
	errorFileNotFound              syscall.Errno = 2
	errorInsufficientBuffer        syscall.Errno = 122
)

func openPipeFile(ctx context.Context, name string) (*os.File, error) {
	n, _ := syscall.UTF16PtrFromString(name)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		r, _, err := createFileW.Call(uintptr(unsafe.Pointer(n)), pipeClientAccessMask, 0, 0, openExisting, 0, 0)
		if syscall.Handle(r) != syscall.InvalidHandle {
			return os.NewFile(r, "majucau-pipe-client"), nil
		}
		var errno syscall.Errno
		if !errors.As(err, &errno) || (errno != errorPipeBusy && errno != errorFileNotFound) {
			return nil, fmt.Errorf("open local IPC pipe: %w", err)
		}
		timer := time.NewTimer(pipeOpenRetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

type securityAttributes struct {
	length     uint32
	descriptor uintptr
	inherit    int32
}

func createPipe(name string, descriptor uintptr) (syscall.Handle, error) {
	n, _ := syscall.UTF16PtrFromString(name)
	attrs := securityAttributes{length: uint32(unsafe.Sizeof(securityAttributes{})), descriptor: descriptor}
	// Byte mode is deliberate: framing is provided by the explicit 4-byte
	// length prefix and remains correct even when a frame spans Win32 writes.
	r, _, err := createNamedPipeW.Call(uintptr(unsafe.Pointer(n)), pipeAccessDuplex|pipeOpenOverlapped, pipeWait, pipeUnlimitedInstances, MaxMessageSize+frameHeaderSize, MaxMessageSize+frameHeaderSize, 0, uintptr(unsafe.Pointer(&attrs)))
	h := syscall.Handle(r)
	if h == syscall.InvalidHandle {
		return 0, win32CallFailure("CreateNamedPipeW", err)
	}
	return h, nil
}
func connectPipe(ctx context.Context, h syscall.Handle) (returnErr error) {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return err
	}
	defer func() {
		returnErr = errors.Join(returnErr, cleanupFailure("close named pipe connect event", windows.CloseHandle(event)))
	}()
	overlapped := windows.Overlapped{HEvent: event}
	r, _, err := connectNamedPipe.Call(uintptr(h), uintptr(unsafe.Pointer(&overlapped)))
	if r != 0 {
		return nil
	}
	if errors.Is(err, windows.ERROR_PIPE_CONNECTED) {
		return nil
	}
	if !errors.Is(err, windows.ERROR_IO_PENDING) {
		return win32CallFailure("connect named pipe", err)
	}
	pending := true
	defer func() {
		if pending {
			cancelErr := windows.CancelIoEx(windows.Handle(h), &overlapped)
			if errors.Is(cancelErr, windows.ERROR_NOT_FOUND) {
				cancelErr = nil
			}
			var transferred uint32
			completionErr := windows.GetOverlappedResult(windows.Handle(h), &overlapped, &transferred, true)
			if ctx.Err() != nil && errors.Is(completionErr, windows.ERROR_OPERATION_ABORTED) {
				completionErr = nil
			}
			returnErr = errors.Join(returnErr,
				cleanupFailure("cancel named pipe connection", cancelErr),
				cleanupFailure("wait for named pipe connection completion", completionErr),
			)
		}
	}()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		waitResult, err := windows.WaitForSingleObject(event, 100)
		if err != nil {
			return err
		}
		if waitResult == uint32(windows.WAIT_TIMEOUT) {
			continue
		}
		if waitResult != windows.WAIT_OBJECT_0 {
			return errors.New("named pipe connect wait failed")
		}
		pending = false
		var transferred uint32
		if err := windows.GetOverlappedResult(windows.Handle(h), &overlapped, &transferred, false); err != nil {
			if errors.Is(err, windows.ERROR_PIPE_CONNECTED) {
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		return nil
	}
}
func closePipeFile(file *os.File) error {
	if err := file.Close(); errors.Is(err, os.ErrClosed) {
		return nil
	} else {
		return err
	}
}

func disconnectPipe(h syscall.Handle) error {
	r, _, err := disconnectNamedPipe.Call(uintptr(h))
	if r != 0 || errors.Is(err, windows.ERROR_PIPE_NOT_CONNECTED) {
		return nil
	}
	if err == nil || errors.Is(err, syscall.Errno(0)) {
		return errors.New("disconnect named pipe failed")
	}
	return fmt.Errorf("disconnect named pipe: %w", err)
}

func closePipeHandle(h syscall.Handle) error {
	return errors.Join(disconnectPipe(h), cleanupFailure("close named pipe handle", syscall.CloseHandle(h)))
}

type pipeResourceCleanupError struct {
	operation string
	err       error
}

func (e *pipeResourceCleanupError) Error() string { return e.operation + ": " + e.err.Error() }
func (e *pipeResourceCleanupError) Unwrap() error { return e.err }

func cleanupFailure(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &pipeResourceCleanupError{operation: operation, err: err}
}

func freeLocalMemory(operation string, address uintptr) error {
	if address == 0 {
		return nil
	}
	r, _, err := localFree.Call(address)
	if r == 0 {
		return nil
	}
	if err == nil || errors.Is(err, syscall.Errno(0)) {
		err = errors.New("LocalFree returned the allocation handle")
	}
	return cleanupFailure(operation, err)
}

func win32CallFailure(operation string, err error) error {
	if err == nil || errors.Is(err, syscall.Errno(0)) {
		return fmt.Errorf("%s failed", operation)
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func makePipeSecurity(c NamedPipeConfig) (uintptr, func() error, error) {
	noRelease := func() error { return nil }
	clientSIDs := append([]string(nil), c.AllowedClientSIDs...)
	serviceSID := c.ServiceSID
	if len(clientSIDs) == 0 {
		sid, err := currentProcessSID()
		if err != nil {
			return 0, noRelease, err
		}
		clientSIDs = []string{sid}
		if serviceSID == "" {
			// Development fallback: the same identity hosts and connects to the pipe.
			serviceSID = sid
		}
	}
	accessBySID := make(map[string]uint32, len(clientSIDs)+1)
	orderedSIDs := make([]string, 0, len(clientSIDs)+1)
	addSID := func(sid string, access uint32) error {
		if !validSID(sid) {
			return errors.New("invalid authorized SID")
		}
		if _, ok := accessBySID[sid]; !ok {
			orderedSIDs = append(orderedSIDs, sid)
		}
		accessBySID[sid] |= access
		return nil
	}
	for _, sid := range clientSIDs {
		if err := addSID(sid, pipeClientAccessMask); err != nil {
			return 0, noRelease, err
		}
	}
	if serviceSID != "" {
		if err := addSID(serviceSID, pipeServerAccessMask); err != nil {
			return 0, noRelease, err
		}
	}
	sddl := "D:P"
	for _, sid := range orderedSIDs {
		// The client receives only the rights needed for framed I/O and to open
		// the pipe. The service SID also receives FILE_CREATE_PIPE_INSTANCE.
		sddl += "(A;;0x" + strconv.FormatUint(uint64(accessBySID[sid]), 16) + ";;;" + sid + ")"
	}
	sddl += "(A;;GA;;;BA)(A;;GA;;;SY)"
	p, _ := syscall.UTF16PtrFromString(sddl)
	var descriptor uintptr
	var size uint32
	r, _, err := convertStringSecurityDescriptorW.Call(uintptr(unsafe.Pointer(p)), 1, uintptr(unsafe.Pointer(&descriptor)), uintptr(unsafe.Pointer(&size)))
	if r == 0 {
		return 0, noRelease, win32CallFailure("ConvertStringSecurityDescriptorToSecurityDescriptorW", err)
	}
	return descriptor, func() error { return freeLocalMemory("free named pipe security descriptor", descriptor) }, nil
}
func validSID(s string) bool { return strings.HasPrefix(s, "S-") && len(s) < 184 }
func authorizePipeClient(h syscall.Handle, c NamedPipeConfig) (authorized bool, returnErr error) {
	var pid uint32
	r, _, err := getNamedPipeClientProcessID.Call(uintptr(h), uintptr(unsafe.Pointer(&pid)))
	if r == 0 {
		return false, nil
	}
	// Development fallback: an empty allow-list uses the DACL's current-process
	// SID and skips the second SID lookup. Production UI/service pairs must
	// configure the UI SID explicitly for cross-process authorization.
	if len(c.AllowedClientSIDs) == 0 {
		return true, nil
	}
	p, _, err := openProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if p == 0 {
		return false, nil
	}
	defer func() {
		returnErr = errors.Join(returnErr, cleanupFailure("close named pipe client process handle", syscall.CloseHandle(syscall.Handle(p))))
	}()
	var token syscall.Handle
	r, _, _ = openProcessToken.Call(p, tokenQuery, uintptr(unsafe.Pointer(&token)))
	if r == 0 {
		return false, nil
	}
	defer func() {
		returnErr = errors.Join(returnErr, cleanupFailure("close named pipe client token", syscall.CloseHandle(token)))
	}()
	sid, err := tokenSID(token)
	if err != nil {
		var cleanupErr *pipeResourceCleanupError
		if errors.As(err, &cleanupErr) {
			return false, err
		}
		return false, nil
	}
	allowedSIDs := c.AllowedClientSIDs
	if len(allowedSIDs) == 0 {
		if own, ownErr := currentProcessSID(); ownErr == nil {
			allowedSIDs = []string{own}
		}
	}
	for _, allowed := range allowedSIDs {
		if sid == allowed {
			return true, nil
		}
	}
	return false, nil
}

type sidAndAttributes struct {
	sid        *byte
	attributes uint32
}
type tokenUserData struct{ user sidAndAttributes }

func tokenSID(token syscall.Handle) (sid string, returnErr error) {
	var size uint32
	getTokenInformation.Call(uintptr(token), tokenUser, 0, 0, uintptr(unsafe.Pointer(&size)))
	if size == 0 {
		return "", errors.New("token information unavailable")
	}
	buf := make([]byte, size)
	r, _, err := getTokenInformation.Call(uintptr(token), tokenUser, uintptr(unsafe.Pointer(&buf[0])), uintptr(size), uintptr(unsafe.Pointer(&size)))
	if r == 0 {
		return "", win32CallFailure("read client token information", err)
	}
	tu := (*tokenUserData)(unsafe.Pointer(&buf[0]))
	var text *uint16
	r, _, err = convertSidToStringSidW.Call(uintptr(unsafe.Pointer(tu.user.sid)), uintptr(unsafe.Pointer(&text)))
	if r == 0 {
		return "", win32CallFailure("convert client SID", err)
	}
	defer func() {
		returnErr = errors.Join(returnErr, freeLocalMemory("free converted client SID", uintptr(unsafe.Pointer(text))))
	}()
	return syscall.UTF16ToString((*[1 << 15]uint16)(unsafe.Pointer(text))[:]), nil
}
func currentProcessSID() (sid string, returnErr error) {
	var token syscall.Handle
	pseudoCurrentProcess := ^uintptr(0)
	r, _, err := openProcessToken.Call(pseudoCurrentProcess, tokenQuery, uintptr(unsafe.Pointer(&token)))
	if r == 0 {
		return "", win32CallFailure("open current process token", err)
	}
	defer func() {
		returnErr = errors.Join(returnErr, cleanupFailure("close current process token", syscall.CloseHandle(token)))
	}()
	return tokenSID(token)
}
