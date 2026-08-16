//go:build windows

package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
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
	handles map[syscall.Handle]struct{}
	closed  bool
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
	return &NamedPipeServer{config: config, handles: make(map[syscall.Handle]struct{})}, nil
}
func (s *NamedPipeServer) Close() error {
	s.mu.Lock()
	s.closed = true
	handles := make([]syscall.Handle, 0, len(s.handles))
	for h := range s.handles {
		handles = append(handles, h)
	}
	s.handles = make(map[syscall.Handle]struct{})
	s.mu.Unlock()
	for _, h := range handles {
		go closePipeHandle(h)
	}
	return nil
}
func (s *NamedPipeServer) track(h syscall.Handle) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.handles[h] = struct{}{}
	return true
}
func (s *NamedPipeServer) untrack(h syscall.Handle) { s.mu.Lock(); delete(s.handles, h); s.mu.Unlock() }

func (s *NamedPipeServer) Serve(ctx context.Context, handler Handler) error {
	if handler == nil {
		return errors.New("ipc handler is required")
	}
	security, release, err := makePipeSecurity(s.config)
	if err != nil {
		return err
	}
	defer release()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		h, err := createPipe(s.config.Name, security)
		if err != nil {
			return err
		}
		if !s.track(h) {
			syscall.CloseHandle(h)
			return context.Canceled
		}
		connected := make(chan error, 1)
		go func() { connected <- connectPipe(h) }()
		select {
		case err = <-connected:
		case <-ctx.Done():
			go closePipeHandle(h)
			s.untrack(h)
			return ctx.Err()
		}
		if err == nil {
			if authErr := authorizePipeClient(h, s.config); authErr == nil {
				_ = s.handleConnection(ctx, h, handler)
			}
		}
		go closePipeHandle(h)
		s.untrack(h)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}
func (s *NamedPipeServer) handleConnection(ctx context.Context, h syscall.Handle, handler Handler) error {
	f := os.NewFile(uintptr(h), "majucau-pipe")
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
		return c.file.Close()
	}
	return nil
}
func (c *NamedPipeClient) Call(ctx context.Context, request Request) (Response, error) {
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
		f, err := openPipeFile(c.name)
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
			_ = f.Close()
			result <- struct {
				response Response
				err      error
			}{err: ErrTransportUnavailable}
			return
		}
		c.file = f
		c.mu.Unlock()
		resp, err := callOverStream(callCtx, f, request)
		_ = f.Close()
		c.mu.Lock()
		c.file = nil
		c.mu.Unlock()
		result <- struct {
			response Response
			err      error
		}{resp, err}
	}()
	select {
	case out := <-result:
		return out.response, out.err
	case <-callCtx.Done():
		_ = c.Close()
		return Response{}, callCtx.Err()
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
	pipeTypeMessage                              = 0x00000004
	pipeReadModeMessage                          = 0x00000002
	pipeWait                                     = 0x00000000
	pipeUnlimitedInstances                       = 255
	genericRead                                  = 0x80000000
	genericWrite                                 = 0x40000000
	openExisting                                 = 3
	processQueryLimitedInformation               = 0x1000
	tokenQuery                                   = 0x0008
	tokenUser                                    = 1
	errorPipeConnected             syscall.Errno = 535
	errorInsufficientBuffer        syscall.Errno = 122
)

func openPipeFile(name string) (*os.File, error) {
	n, _ := syscall.UTF16PtrFromString(name)
	r, _, err := createFileW.Call(uintptr(unsafe.Pointer(n)), genericRead|genericWrite, 0, 0, openExisting, 0, 0)
	if syscall.Handle(r) == syscall.InvalidHandle {
		return nil, err
	}
	return os.NewFile(r, "majucau-pipe-client"), nil
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
	r, _, err := createNamedPipeW.Call(uintptr(unsafe.Pointer(n)), pipeAccessDuplex, pipeWait, pipeUnlimitedInstances, MaxMessageSize+frameHeaderSize, MaxMessageSize+frameHeaderSize, 0, uintptr(unsafe.Pointer(&attrs)))
	h := syscall.Handle(r)
	if h == syscall.InvalidHandle {
		return 0, err
	}
	return h, nil
}
func connectPipe(h syscall.Handle) error {
	r, _, err := connectNamedPipe.Call(uintptr(h), 0)
	if r != 0 {
		return nil
	}
	if errno, ok := err.(syscall.Errno); ok && errno == errorPipeConnected {
		return nil
	}
	return err
}
func disconnectPipe(h syscall.Handle)  { disconnectNamedPipe.Call(uintptr(h)) }
func closePipeHandle(h syscall.Handle) { disconnectPipe(h); syscall.CloseHandle(h) }

func makePipeSecurity(c NamedPipeConfig) (uintptr, func(), error) {
	sids := append([]string(nil), c.AllowedClientSIDs...)
	if len(sids) == 0 {
		sid, err := currentProcessSID()
		if err != nil {
			return 0, func() {}, err
		}
		sids = []string{sid}
	}
	if c.ServiceSID != "" {
		sids = append(sids, c.ServiceSID)
	}
	sddl := "D:P"
	for _, sid := range sids {
		if !validSID(sid) {
			return 0, func() {}, errors.New("invalid authorized SID")
		}
		sddl += "(A;;GA;;;" + sid + ")"
	}
	sddl += "(A;;GA;;;BA)(A;;GA;;;SY)"
	p, _ := syscall.UTF16PtrFromString(sddl)
	var descriptor uintptr
	var size uint32
	r, _, err := convertStringSecurityDescriptorW.Call(uintptr(unsafe.Pointer(p)), 1, uintptr(unsafe.Pointer(&descriptor)), uintptr(unsafe.Pointer(&size)))
	if r == 0 {
		return 0, func() {}, err
	}
	return descriptor, func() { localFree.Call(descriptor) }, nil
}
func validSID(s string) bool { return strings.HasPrefix(s, "S-") && len(s) < 184 }
func authorizePipeClient(h syscall.Handle, c NamedPipeConfig) error {
	var pid uint32
	r, _, err := getNamedPipeClientProcessID.Call(uintptr(h), uintptr(unsafe.Pointer(&pid)))
	if r == 0 {
		return ErrUnauthorizedClient
	}
	// Development fallback: an empty allow-list uses the DACL's current-process
	// SID and skips the second SID lookup. Production UI/service pairs must
	// configure the UI SID explicitly for cross-process authorization.
	if len(c.AllowedClientSIDs) == 0 {
		return nil
	}
	p, _, err := openProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if p == 0 {
		_ = err
		return ErrUnauthorizedClient
	}
	defer syscall.CloseHandle(syscall.Handle(p))
	var token syscall.Handle
	r, _, _ = openProcessToken.Call(p, tokenQuery, uintptr(unsafe.Pointer(&token)))
	if r == 0 {
		return ErrUnauthorizedClient
	}
	defer syscall.CloseHandle(token)
	sid, err := tokenSID(token)
	if err != nil {
		return ErrUnauthorizedClient
	}
	allowedSIDs := c.AllowedClientSIDs
	if len(allowedSIDs) == 0 {
		if own, ownErr := currentProcessSID(); ownErr == nil {
			allowedSIDs = []string{own}
		}
	}
	for _, allowed := range allowedSIDs {
		if sid == allowed {
			return nil
		}
	}
	return ErrUnauthorizedClient
}

type sidAndAttributes struct {
	sid        *byte
	attributes uint32
}
type tokenUserData struct{ user sidAndAttributes }

func tokenSID(token syscall.Handle) (string, error) {
	var size uint32
	getTokenInformation.Call(uintptr(token), tokenUser, 0, 0, uintptr(unsafe.Pointer(&size)))
	if size == 0 {
		return "", errors.New("token information unavailable")
	}
	buf := make([]byte, size)
	r, _, err := getTokenInformation.Call(uintptr(token), tokenUser, uintptr(unsafe.Pointer(&buf[0])), uintptr(size), uintptr(unsafe.Pointer(&size)))
	if r == 0 {
		return "", err
	}
	tu := (*tokenUserData)(unsafe.Pointer(&buf[0]))
	var text *uint16
	r, _, err = convertSidToStringSidW.Call(uintptr(unsafe.Pointer(tu.user.sid)), uintptr(unsafe.Pointer(&text)))
	if r == 0 {
		return "", err
	}
	defer localFree.Call(uintptr(unsafe.Pointer(text)))
	return syscall.UTF16ToString((*[1 << 15]uint16)(unsafe.Pointer(text))[:]), nil
}
func currentProcessSID() (string, error) {
	var token syscall.Handle
	pseudoCurrentProcess := ^uintptr(0)
	r, _, err := openProcessToken.Call(pseudoCurrentProcess, tokenQuery, uintptr(unsafe.Pointer(&token)))
	if r == 0 {
		return "", err
	}
	defer syscall.CloseHandle(token)
	return tokenSID(token)
}
