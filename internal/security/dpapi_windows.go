//go:build windows

package security

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"majucau.local/financial-intelligence/internal/securetemp"
)

const (
	// IPC frames are capped at 1 MiB; secret bundles are control data, not a bulk-data channel.
	maxDPAPISecretValueBytes = 1 << 20
	// Leave bounded room for the DPAPI ciphertext wrapper while rejecting corrupt oversized files.
	maxDPAPISecretBlobBytes  = maxDPAPISecretValueBytes + (64 << 10)
	dpapiEntropyBytes        = 32
	cryptProtectLocalMachine = 0x4
)

const (
	dpapiScopeCurrentUser = "CURRENT_USER"
	dpapiScopeMachine     = "MACHINE"
)

type DPAPIStore struct {
	dir             string
	entropy         []byte
	userSID         string
	protectionFlags uintptr
	scope           string
	mu              sync.Mutex
}

func NewDPAPIStore(dir string) (*DPAPIStore, error) {
	return newDPAPIStore(dir, dpapiScopeCurrentUser, 0)
}

// NewMachineDPAPIStore protects blobs with the Windows machine DPAPI scope.
// Files still have a strict ACL for the current service identity, SYSTEM, and
// Administrators; machine scope must never be used as a substitute for ACLs.
func NewMachineDPAPIStore(dir string) (*DPAPIStore, error) {
	return newDPAPIStore(dir, dpapiScopeMachine, cryptProtectLocalMachine)
}

func newDPAPIStore(dir, scope string, protectionFlags uintptr) (*DPAPIStore, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("secret directory is required")
	}
	if !filepath.IsAbs(dir) {
		return nil, errors.New("secret directory must be absolute")
	}
	dir = filepath.Clean(dir)
	userSID, err := currentProcessSID()
	if err != nil {
		return nil, errors.New("current Windows identity is unavailable for the secret vault")
	}
	if err := preparePrivateSecretDirectory(dir, userSID); err != nil {
		return nil, err
	}
	if err := ensureDPAPIScope(dir, scope, userSID); err != nil {
		return nil, err
	}
	entropyPath := filepath.Join(dir, ".entropy")
	if _, statErr := os.Lstat(entropyPath); statErr == nil {
		if err := preparePrivateSecretFile(entropyPath, userSID); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(statErr) {
		return nil, errors.New("secret vault entropy file is unavailable")
	}
	e, err := readBoundedSecretFile(entropyPath, dpapiEntropyBytes)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if os.IsNotExist(err) {
		candidate := make([]byte, dpapiEntropyBytes)
		if _, err := rand.Read(candidate); err != nil {
			return nil, err
		}
		// The directory is already restricted to this process identity, SYSTEM,
		// and Administrators; the entropy file inherits no broader access.
		file, createErr := os.OpenFile(entropyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if createErr == nil {
			if _, writeErr := file.Write(candidate); writeErr != nil {
				createErr = writeErr
			}
			if createErr == nil {
				createErr = file.Sync()
			}
			createErr = errors.Join(createErr, file.Close())
			if createErr != nil {
				return nil, errors.Join(createErr, securetemp.RemoveFile(entropyPath))
			}
			if aclErr := preparePrivateSecretFile(entropyPath, userSID); aclErr != nil {
				return nil, errors.Join(aclErr, securetemp.RemoveFile(entropyPath))
			}
			e = candidate
		} else if os.IsExist(createErr) {
			// Another worker initialization won the race. Both instances must use
			// the persisted winner so they never encrypt with different entropy.
			if err := preparePrivateSecretFile(entropyPath, userSID); err != nil {
				return nil, err
			}
			e, err = readBoundedSecretFile(entropyPath, dpapiEntropyBytes)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, createErr
		}
	}
	if len(e) != dpapiEntropyBytes {
		return nil, errors.New("invalid DPAPI entropy")
	}
	return &DPAPIStore{dir: dir, entropy: e, userSID: userSID, protectionFlags: protectionFlags, scope: scope}, nil
}

func ensureDPAPIScope(dir, expectedScope, userSID string) error {
	if expectedScope != dpapiScopeCurrentUser && expectedScope != dpapiScopeMachine {
		return errors.New("invalid DPAPI scope")
	}
	path := filepath.Join(dir, ".scope")
	if _, err := os.Lstat(path); err == nil {
		if err := preparePrivateSecretFile(path, userSID); err != nil {
			return err
		}
		value, err := readBoundedSecretFile(path, len(expectedScope))
		if err != nil {
			return errors.New("DPAPI scope marker is unavailable")
		}
		if string(value) != expectedScope {
			return errors.New("secret vault DPAPI scope does not match the requested scope")
		}
		return nil
	} else if !os.IsNotExist(err) {
		return errors.New("DPAPI scope marker is unavailable")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return errors.New("secret vault contents are unavailable")
	}
	if expectedScope == dpapiScopeMachine && len(entries) != 0 {
		return errors.New("unmarked secret vault cannot be initialized for machine DPAPI scope")
	}
	if expectedScope == dpapiScopeCurrentUser {
		for _, entry := range entries {
			name := entry.Name()
			if name != ".entropy" && !strings.HasSuffix(name, ".dpapi") {
				return errors.New("unmarked secret vault contains an unexpected file")
			}
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err == nil {
		_, writeErr := file.Write([]byte(expectedScope))
		if writeErr == nil {
			writeErr = file.Sync()
		}
		writeErr = errors.Join(writeErr, file.Close())
		if writeErr != nil {
			return errors.Join(writeErr, securetemp.RemoveFile(path))
		}
		if aclErr := preparePrivateSecretFile(path, userSID); aclErr != nil {
			return errors.Join(aclErr, securetemp.RemoveFile(path))
		}
		return nil
	}
	if !os.IsExist(err) {
		return errors.New("could not initialize DPAPI scope marker")
	}
	if err := preparePrivateSecretFile(path, userSID); err != nil {
		return err
	}
	value, err := readBoundedSecretFile(path, len(expectedScope))
	if err != nil || string(value) != expectedScope {
		return errors.New("concurrent secret vault initialization selected another DPAPI scope")
	}
	return nil
}
func (s *DPAPIStore) path(ref string) string { return filepath.Join(s.dir, ref+".dpapi") }
func (s *DPAPIStore) Put(_ context.Context, ref string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ValidateSecretRef(ref); err != nil {
		return err
	}
	if len(value) > maxDPAPISecretValueBytes {
		return errors.New("secret value exceeds size limit")
	}
	blob, err := protect(value, s.entropy, s.protectionFlags)
	if err != nil {
		return err
	}
	defer zeroBytes(blob)
	path := s.path(ref)
	if _, err := os.Lstat(path); err == nil {
		if err := preparePrivateSecretFile(path, s.userSID); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return errors.New("secret vault entry is unavailable")
	}
	return atomicWrite(path, blob, 0600, s.userSID)
}
func (s *DPAPIStore) Get(_ context.Context, ref string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ValidateSecretRef(ref); err != nil {
		return nil, err
	}
	path := s.path(ref)
	if err := preparePrivateSecretFile(path, s.userSID); err != nil {
		return nil, err
	}
	b, err := readBoundedSecretFile(path, maxDPAPISecretBlobBytes)
	if err != nil {
		return nil, err
	}
	defer zeroBytes(b)
	return unprotect(b, s.entropy)
}

func readBoundedSecretFile(path string, limit int) ([]byte, error) {
	if limit < 0 {
		return nil, errors.New("invalid secret file size limit")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		for index := range data {
			data[index] = 0
		}
		return nil, errors.Join(readErr, closeErr)
	}
	if len(data) > limit {
		for index := range data {
			data[index] = 0
		}
		return nil, errors.New("secret file exceeds size limit")
	}
	return data, nil
}

func zeroBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}

func (s *DPAPIStore) Delete(_ context.Context, ref string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ValidateSecretRef(ref); err != nil {
		return err
	}
	path := s.path(ref)
	if err := preparePrivateSecretFile(path, s.userSID); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

type dataBlob struct {
	cbData uint32
	pbData *byte
}

var crypt32 = syscall.NewLazyDLL("crypt32.dll")
var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var cryptProtect = crypt32.NewProc("CryptProtectData")
var cryptUnprotect = crypt32.NewProc("CryptUnprotectData")
var localFree = kernel32.NewProc("LocalFree")
var moveFileEx = kernel32.NewProc("MoveFileExW")

const (
	moveFileReplaceExisting = 0x1
	moveFileWriteThrough    = 0x8
)

func atomicWrite(path string, value []byte, mode os.FileMode, userSID string) (returnErr error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".secret-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	temporaryClosed := false
	removeTemporary := true
	defer func() {
		if !temporaryClosed {
			returnErr = errors.Join(returnErr, temporary.Close())
		}
		if removeTemporary {
			returnErr = errors.Join(returnErr, securetemp.RemoveFile(temporaryPath))
		}
	}()
	if err := preparePrivateSecretFile(temporaryPath, userSID); err != nil {
		return err
	}

	if err := temporary.Chmod(mode); err != nil {
		return err
	}
	if _, err := temporary.Write(value); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		temporaryClosed = true
		return err
	}
	temporaryClosed = true

	from, err := syscall.UTF16PtrFromString(temporaryPath)
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	result, _, callErr := moveFileEx.Call(
		uintptr(unsafe.Pointer(from)),
		uintptr(unsafe.Pointer(to)),
		moveFileReplaceExisting|moveFileWriteThrough,
	)
	if result == 0 {
		return callErr
	}
	removeTemporary = false
	return nil
}

func protect(value, entropy []byte, flags uintptr) ([]byte, error) {
	return cryptCall(cryptProtect, value, entropy, flags)
}
func unprotect(value, entropy []byte) ([]byte, error) {
	return cryptCall(cryptUnprotect, value, entropy, 0)
}
func cryptCall(proc *syscall.LazyProc, value, entropy []byte, flags uintptr) (result []byte, returnErr error) {
	if len(value) == 0 {
		value = []byte{0}
	}
	in := dataBlob{uint32(len(value)), &value[0]}
	ent := dataBlob{}
	if len(entropy) > 0 {
		ent = dataBlob{uint32(len(entropy)), &entropy[0]}
	}
	out := dataBlob{}
	r, _, err := proc.Call(uintptr(unsafe.Pointer(&in)), 0, uintptr(unsafe.Pointer(&ent)), 0, 0, flags, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		if err == nil || errors.Is(err, syscall.Errno(0)) {
			err = errors.New("DPAPI returned failure without an error code")
		} else {
			err = fmt.Errorf("DPAPI operation: %w", err)
		}
		return nil, errors.Join(err, freeDPAPIOutput(out.pbData))
	}
	if out.pbData == nil || out.cbData == 0 {
		return nil, errors.Join(errors.New("DPAPI returned an empty output buffer"), freeDPAPIOutput(out.pbData))
	}
	if out.cbData > maxDPAPISecretBlobBytes {
		return nil, errors.Join(errors.New("DPAPI output exceeds the configured size limit"), freeDPAPIOutput(out.pbData))
	}
	defer func() {
		if err := freeDPAPIOutput(out.pbData); err != nil {
			for index := range result {
				result[index] = 0
			}
			result = nil
			returnErr = errors.Join(returnErr, err)
		}
	}()
	// DPAPI owns the returned buffer. Copy it before LocalFree runs; returning
	// the unsafe slice itself would leave the caller with freed memory.
	result = make([]byte, int(out.cbData))
	copy(result, unsafe.Slice(out.pbData, int(out.cbData)))
	return result, nil
}

func freeDPAPIOutput(output *byte) error {
	if output == nil {
		return nil
	}
	r, _, err := localFree.Call(uintptr(unsafe.Pointer(output)))
	if r == 0 {
		return nil
	}
	if err == nil || errors.Is(err, syscall.Errno(0)) {
		err = errors.New("LocalFree returned the allocated DPAPI buffer")
	}
	return fmt.Errorf("release DPAPI output buffer: %w", err)
}
