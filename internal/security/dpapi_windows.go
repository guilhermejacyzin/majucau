//go:build windows

package security

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

type DPAPIStore struct {
	dir     string
	entropy []byte
}

func NewDPAPIStore(dir string) (*DPAPIStore, error) {
	if dir == "" {
		return nil, errors.New("secret directory is required")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	entropyPath := filepath.Join(dir, ".entropy")
	e, err := os.ReadFile(entropyPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if os.IsNotExist(err) {
		candidate := make([]byte, 32)
		if _, err := rand.Read(candidate); err != nil {
			return nil, err
		}
		// The installer must apply the worker-service ACL to this file and its
		// directory. It is persisted so a worker restart can decrypt its vault.
		file, createErr := os.OpenFile(entropyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if createErr == nil {
			if _, createErr = file.Write(candidate); createErr == nil {
				createErr = file.Sync()
			}
			if closeErr := file.Close(); createErr == nil {
				createErr = closeErr
			}
			if createErr != nil {
				_ = os.Remove(entropyPath)
				return nil, createErr
			}
			e = candidate
		} else if os.IsExist(createErr) {
			// Another worker initialization won the race. Both instances must use
			// the persisted winner so they never encrypt with different entropy.
			e, err = os.ReadFile(entropyPath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, createErr
		}
	}
	if len(e) != 32 {
		return nil, errors.New("invalid DPAPI entropy")
	}
	return &DPAPIStore{dir: dir, entropy: e}, nil
}
func (s *DPAPIStore) path(ref string) string { return filepath.Join(s.dir, ref+".dpapi") }
func (s *DPAPIStore) Put(_ context.Context, ref string, value []byte) error {
	if err := ValidateSecretRef(ref); err != nil {
		return err
	}
	blob, err := protect(value, s.entropy)
	if err != nil {
		return err
	}
	return atomicWrite(s.path(ref), blob, 0600)
}
func (s *DPAPIStore) Get(_ context.Context, ref string) ([]byte, error) {
	if err := ValidateSecretRef(ref); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(s.path(ref))
	if err != nil {
		return nil, err
	}
	return unprotect(b, s.entropy)
}
func (s *DPAPIStore) Delete(_ context.Context, ref string) error {
	if err := ValidateSecretRef(ref); err != nil {
		return err
	}
	err := os.Remove(s.path(ref))
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

func atomicWrite(path string, value []byte, mode os.FileMode) (returnErr error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".secret-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		if returnErr != nil {
			_ = os.Remove(temporaryPath)
		}
	}()

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
		return err
	}

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
	return nil
}

func protect(value, entropy []byte) ([]byte, error) { return cryptCall(cryptProtect, value, entropy) }
func unprotect(value, entropy []byte) ([]byte, error) {
	return cryptCall(cryptUnprotect, value, entropy)
}
func cryptCall(proc *syscall.LazyProc, value, entropy []byte) ([]byte, error) {
	if len(value) == 0 {
		value = []byte{0}
	}
	in := dataBlob{uint32(len(value)), &value[0]}
	ent := dataBlob{}
	if len(entropy) > 0 {
		ent = dataBlob{uint32(len(entropy)), &entropy[0]}
	}
	out := dataBlob{}
	r, _, err := proc.Call(uintptr(unsafe.Pointer(&in)), 0, uintptr(unsafe.Pointer(&ent)), 0, 0, 0, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, err
	}
	defer localFree.Call(uintptr(unsafe.Pointer(out.pbData)))
	// DPAPI owns the returned buffer. Copy it before LocalFree runs; returning
	// the unsafe slice itself would leave the caller with freed memory.
	result := make([]byte, int(out.cbData))
	copy(result, unsafe.Slice(out.pbData, int(out.cbData)))
	return result, nil
}
