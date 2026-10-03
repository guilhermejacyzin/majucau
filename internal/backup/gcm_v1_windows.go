//go:build windows

package backup

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	bcryptAESAlgorithm       = "AES"
	bcryptChainingMode       = "ChainingMode"
	bcryptGCMMode            = "ChainingModeGCM"
	bcryptObjectLength       = "ObjectLength"
	bcryptChainCallsFlag     = uint32(0x00000001)
	bcryptAuthTagLength      = 16
	bcryptStreamChunkSize    = 64 * 1024
	legacyGCMNonceLength     = 12
	legacyGCMAuthenticationTagLength = 16
)

type bcryptAuthInfo struct {
	cbSize       uint32
	dwInfoVersion uint32
	pbNonce      *byte
	cbNonce      uint32
	pbAuthData   *byte
	cbAuthData   uint32
	pbTag        *byte
	cbTag        uint32
	pbMacContext *byte
	cbMacContext uint32
	cbAAD        uint32
	cbData       uint64
	dwFlags      uint32
}

var (
	bcryptDLL                  = windows.NewLazySystemDLL("bcrypt.dll")
	bcryptOpenAlgorithm        = bcryptDLL.NewProc("BCryptOpenAlgorithmProvider")
	bcryptCloseAlgorithm       = bcryptDLL.NewProc("BCryptCloseAlgorithmProvider")
	bcryptGetProperty          = bcryptDLL.NewProc("BCryptGetProperty")
	bcryptSetProperty          = bcryptDLL.NewProc("BCryptSetProperty")
	bcryptGenerateSymmetricKey = bcryptDLL.NewProc("BCryptGenerateSymmetricKey")
	bcryptDestroyKey           = bcryptDLL.NewProc("BCryptDestroyKey")
	bcryptDecrypt              = bcryptDLL.NewProc("BCryptDecrypt")
)

// decryptV1Stream reads the historical single-message AES-GCM envelope in
// bounded blocks. It may write unauthenticated plaintext to dst before the
// final tag is checked; callers must keep dst private and must not consume it
// unless this function returns nil.
func decryptV1Stream(source io.Reader, encryptedSize int64, key, associatedData []byte, dst io.Writer) error {
	if encryptedSize < int64(len(encryptedHeader)+1+legacyGCMNonceLength+legacyGCMAuthenticationTagLength) {
		return ErrWrongPassphrase
	}
	var header [len(encryptedHeader)]byte
	if _, err := io.ReadFull(source, header[:]); err != nil || string(header[:]) != encryptedHeader {
		return ErrWrongPassphrase
	}
	var nonceLength [1]byte
	if _, err := io.ReadFull(source, nonceLength[:]); err != nil || int(nonceLength[0]) != legacyGCMNonceLength {
		return ErrWrongPassphrase
	}
	nonce := make([]byte, legacyGCMNonceLength)
	if _, err := io.ReadFull(source, nonce); err != nil {
		return ErrWrongPassphrase
	}
	ciphertextSize := encryptedSize - int64(len(encryptedHeader)+1+legacyGCMNonceLength+legacyGCMAuthenticationTagLength)
	if ciphertextSize < 0 {
		return ErrWrongPassphrase
	}
	if ciphertextSize == 0 {
		// Windows versions older than 2025 do not reliably authenticate an
		// empty BCryptDecrypt input. Real pg_dump and pg_dumpall outputs are
		// non-empty, so rejecting this edge avoids accepting an unchecked tag.
		return ErrInvalidPackage
	}

	algorithmName, _ := syscallUTF16(bcryptAESAlgorithm)
	var algorithm uintptr
	status, _, _ := bcryptOpenAlgorithm.Call(uintptr(unsafe.Pointer(&algorithm)), uintptr(unsafe.Pointer(algorithmName)), 0, 0)
	if status != 0 {
		return fmt.Errorf("open Windows AES provider (NTSTATUS 0x%x)", status)
	}
	defer bcryptCloseAlgorithm.Call(algorithm, 0)

	modeName, _ := syscallUTF16(bcryptChainingMode)
	mode, _ := syscallUTF16(bcryptGCMMode)
	status, _, _ = bcryptSetProperty.Call(algorithm, uintptr(unsafe.Pointer(modeName)), uintptr(unsafe.Pointer(&mode[0])), uintptr(len(mode)*2), 0)
	if status != 0 {
		return fmt.Errorf("configure Windows AES-GCM provider (NTSTATUS 0x%x)", status)
	}
	objectName, _ := syscallUTF16(bcryptObjectLength)
	objectLengthBytes := make([]byte, 4)
	var propertyBytes uint32
	status, _, _ = bcryptGetProperty.Call(algorithm, uintptr(unsafe.Pointer(objectName)), uintptr(unsafe.Pointer(&objectLengthBytes[0])), uintptr(len(objectLengthBytes)), uintptr(unsafe.Pointer(&propertyBytes)), 0)
	if status != 0 || propertyBytes != 4 {
		return fmt.Errorf("query Windows AES key object size (NTSTATUS 0x%x)", status)
	}
	keyObject := make([]byte, binary.LittleEndian.Uint32(objectLengthBytes))
	var symmetricKey uintptr
	status, _, _ = bcryptGenerateSymmetricKey.Call(algorithm, uintptr(unsafe.Pointer(&symmetricKey)), uintptr(unsafe.Pointer(&keyObject[0])), uintptr(len(keyObject)), uintptr(unsafe.Pointer(&key[0])), uintptr(len(key)), 0)
	if status != 0 {
		return fmt.Errorf("initialize Windows AES key (NTSTATUS 0x%x)", status)
	}
	defer bcryptDestroyKey.Call(symmetricKey)

	var tag [legacyGCMAuthenticationTagLength]byte
	chained := ciphertextSize > bcryptStreamChunkSize
	var macContext []byte
	var macContextPointer *byte
	var macContextLength uint32
	initialFlags := uint32(0)
	if chained {
		macContext = make([]byte, bcryptAuthTagLength)
		macContextPointer = &macContext[0]
		macContextLength = uint32(len(macContext))
		initialFlags = bcryptChainCallsFlag
	}
	aadPointer := bytePointer(associatedData)
	info := bcryptAuthInfo{
		cbSize: uint32(unsafe.Sizeof(bcryptAuthInfo{})), dwInfoVersion: 1,
		pbNonce: bytePointer(nonce), cbNonce: uint32(len(nonce)),
		pbAuthData: aadPointer, cbAuthData: uint32(len(associatedData)),
		pbTag: &tag[0], cbTag: uint32(len(tag)),
		pbMacContext: macContextPointer, cbMacContext: macContextLength,
		dwFlags: initialFlags,
	}

	decryptChunk := func(ciphertext []byte, final bool) error {
		if final {
			info.dwFlags &^= bcryptChainCallsFlag
		} else {
			info.dwFlags |= bcryptChainCallsFlag
		}
		var empty byte
		input := bytePointer(ciphertext)
		if input == nil {
			input = &empty
		}
		plaintext := make([]byte, len(ciphertext))
		var emptyOutput byte
		output := bytePointer(plaintext)
		outputLength := uintptr(len(plaintext))
		if output == nil {
			output = &emptyOutput
			outputLength = 1
		}
		var written uint32
		status, _, _ := bcryptDecrypt.Call(
			symmetricKey, uintptr(unsafe.Pointer(input)), uintptr(len(ciphertext)), uintptr(unsafe.Pointer(&info)),
			0, 0, uintptr(unsafe.Pointer(output)), outputLength, uintptr(unsafe.Pointer(&written)), 0,
		)
		if status != 0 {
			if final {
				return ErrWrongPassphrase
			}
			return fmt.Errorf("decrypt V1 backup block (NTSTATUS 0x%x)", status)
		}
		if written != uint32(len(ciphertext)) {
			return errors.New("Windows AES-GCM returned an unexpected plaintext size")
		}
		if len(plaintext) > 0 {
			writtenBytes, err := dst.Write(plaintext)
			if err != nil {
				return err
			}
			if writtenBytes != len(plaintext) {
				return io.ErrShortWrite
			}
		}
		return nil
	}

	remaining := ciphertextSize
	var pending []byte
	if remaining > 0 {
		chunkLength := int64(bcryptStreamChunkSize)
		if remaining < chunkLength {
			chunkLength = remaining
		}
		pending = make([]byte, int(chunkLength))
		if _, err := io.ReadFull(source, pending); err != nil {
			return ErrWrongPassphrase
		}
		remaining -= chunkLength
		for remaining > 0 {
			if err := decryptChunk(pending, false); err != nil {
				return err
			}
			chunkLength = int64(bcryptStreamChunkSize)
			if remaining < chunkLength {
				chunkLength = remaining
			}
			pending = make([]byte, int(chunkLength))
			if _, err := io.ReadFull(source, pending); err != nil {
				return ErrWrongPassphrase
			}
			remaining -= chunkLength
		}
	}
	if _, err := io.ReadFull(source, tag[:]); err != nil {
		return ErrWrongPassphrase
	}
	if err := decryptChunk(pending, true); err != nil {
		return err
	}
	var extra [1]byte
	if count, err := io.ReadFull(source, extra[:]); count != 0 || err != io.EOF {
		return ErrWrongPassphrase
	}
	return nil
}

func bytePointer(value []byte) *byte {
	if len(value) == 0 {
		return nil
	}
	return &value[0]
}

func syscallUTF16(value string) (*uint16, error) {
	pointer, err := windows.UTF16PtrFromString(value)
	return pointer, err
}
