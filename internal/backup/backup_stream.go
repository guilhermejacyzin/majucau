package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
	"github.com/tink-crypto/tink-go/v2/tink"
)

const (
	maxManifestBytes = 1 << 20
	maxKeysetBytes   = 64 << 10
	keysetAAD        = "MAJUCAU-BACKUP-2:keyset"
	wrappedKeysetHeader = "MAJUCAU-KEYSET-2"
)

type packageArchive struct {
	reader   *zip.ReadCloser
	entries  map[string]*zip.File
	manifest Manifest
}

func writeV2Package(ctx context.Context, path string, manifest Manifest, masterKey []byte, databasePath, globalsPath string) (Manifest, error) {
	if err := ctx.Err(); err != nil {
		return Manifest{}, err
	}
	handle, primitive, err := newStreamingPrimitive()
	if err != nil {
		return Manifest{}, fmt.Errorf("initialize backup streaming encryption: %w", err)
	}
	keysetBytes, err := encryptedKeysetBytes(handle, masterKey)
	if err != nil {
		return Manifest{}, fmt.Errorf("protect backup streaming keyset: %w", err)
	}
	defer clearBytes(keysetBytes)
	manifest.KeysetFile = fileRecord(streamKeysetEntry, keysetBytes)

	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return Manifest{}, fmt.Errorf("create backup package: %w", err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = output.Close()
		}
	}()
	archive := zip.NewWriter(output)
	keysetWriter, err := createStoredEntry(archive, streamKeysetEntry)
	if err != nil {
		_ = archive.Close()
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("write backup keyset entry: %w", err)
	}
	if _, err := keysetWriter.Write(keysetBytes); err != nil {
		_ = archive.Close()
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("write backup keyset entry: %w", err)
	}
	databaseRecord, err := encryptFileEntry(ctx, archive, primitive, databaseDumpEntry, databasePath)
	if err != nil {
		_ = archive.Close()
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("encrypt database dump: %w", err)
	}
	globalsRecord, err := encryptFileEntry(ctx, archive, primitive, globalsDumpEntry, globalsPath)
	if err != nil {
		_ = archive.Close()
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("encrypt globals dump: %w", err)
	}
	manifest.Files = []File{databaseRecord, globalsRecord}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		_ = archive.Close()
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("encode backup manifest: %w", err)
	}
	manifestBytes = append(manifestBytes, '\n')
	if len(manifestBytes) > maxManifestBytes {
		_ = archive.Close()
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("backup manifest exceeds the supported size")
	}
	manifestWriter, err := createStoredEntry(archive, manifestEntry)
	if err == nil {
		_, err = manifestWriter.Write(manifestBytes)
	}
	if err != nil {
		_ = archive.Close()
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("write backup manifest: %w", err)
	}
	if err := archive.Close(); err != nil {
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("finish backup archive: %w", err)
	}
	if err := output.Sync(); err != nil {
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("flush backup package: %w", err)
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(path)
		return Manifest{}, fmt.Errorf("close backup package: %w", err)
	}
	closed = true
	return manifest, nil
}

func createStoredEntry(archive *zip.Writer, name string) (io.Writer, error) {
	header := &zip.FileHeader{Name: name, Method: zip.Store}
	header.SetMode(0600)
	return archive.CreateHeader(header)
}

func encryptFileEntry(ctx context.Context, archive *zip.Writer, primitive tink.StreamingAEAD, entryName, sourcePath string) (File, error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return File{}, err
	}
	defer source.Close()
	entry, err := createStoredEntry(archive, entryName)
	if err != nil {
		return File{}, err
	}
	digest := &digestWriter{writer: entry, hash: sha256.New()}
	writer, err := primitive.NewEncryptingWriter(digest, payloadAAD(entryName))
	if err != nil {
		return File{}, err
	}
	if _, err := io.Copy(writer, contextReader{ctx: ctx, reader: source}); err != nil {
		_ = writer.Close()
		return File{}, err
	}
	if err := writer.Close(); err != nil {
		return File{}, err
	}
	return digest.file(entryName), nil
}

func verifyPackage(ctx context.Context, path string, passphrase []byte, currentSchema string) Verification {
	result := Verification{}
	if ctx == nil || ctx.Err() != nil || !filepath.IsAbs(path) || strings.TrimSpace(currentSchema) == "" || len(passphrase) == 0 {
		result.Issues = append(result.Issues, "invalid verification arguments")
		return result
	}
	archive, err := openPackage(path)
	if err != nil {
		result.Issues = append(result.Issues, "backup package is unreadable or invalid")
		return result
	}
	defer archive.reader.Close()
	result.Manifest = archive.manifest
	if result.Manifest.SchemaVersion != currentSchema {
		result.Issues = append(result.Issues, ErrUnsupportedSchema.Error())
	}
	if !validPackageManifest(result.Manifest) {
		result.Issues = append(result.Issues, "backup manifest identity, version, or file list is invalid")
		return result
	}
	key := deriveKey(passphrase, mustDecodeSalt(result.Manifest.KDF.Salt), result.Manifest.KDF)
	defer clearBytes(key)
	var verifyErr error
	switch result.Manifest.FormatVersion {
	case legacyFormatVersion:
		verifyErr = verifyV1Payloads(ctx, archive, key)
	case FormatVersion:
		verifyErr = verifyV2Payloads(ctx, archive, key)
	default:
		verifyErr = ErrInvalidPackage
	}
	if verifyErr != nil {
		if errors.Is(verifyErr, ErrWrongPassphrase) {
			result.Issues = append(result.Issues, ErrWrongPassphrase.Error())
		} else {
			result.Issues = append(result.Issues, "backup payload authentication or checksum validation failed")
		}
	}
	result.Valid = len(result.Issues) == 0
	return result
}

func openPackage(path string) (*packageArchive, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	archive := &packageArchive{reader: reader, entries: make(map[string]*zip.File, len(reader.File))}
	if len(reader.File) < 3 || len(reader.File) > 4 {
		_ = reader.Close()
		return nil, ErrInvalidPackage
	}
	for _, entry := range reader.File {
		if _, exists := archive.entries[entry.Name]; exists {
			_ = reader.Close()
			return nil, ErrInvalidPackage
		}
		archive.entries[entry.Name] = entry
	}
	manifestEntryFile, ok := archive.entries[manifestEntry]
	if !ok || (manifestEntryFile.Method != zip.Store && manifestEntryFile.Method != zip.Deflate) || manifestEntryFile.UncompressedSize64 > maxManifestBytes {
		_ = reader.Close()
		return nil, ErrInvalidPackage
	}
	manifestBytes, err := readZipEntryBounded(manifestEntryFile, maxManifestBytes)
	if err != nil {
		_ = reader.Close()
		return nil, ErrInvalidPackage
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&archive.manifest); err != nil {
		_ = reader.Close()
		return nil, ErrInvalidPackage
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		_ = reader.Close()
		return nil, ErrInvalidPackage
	}
	allowed := expectedArchiveEntries(archive.manifest.FormatVersion)
	if len(allowed) != len(archive.entries) {
		_ = reader.Close()
		return nil, ErrInvalidPackage
	}
	for name, entry := range archive.entries {
		if !allowed[name] || !validEntryCompression(entry, archive.manifest.FormatVersion) {
			_ = reader.Close()
			return nil, ErrInvalidPackage
		}
	}
	return archive, nil
}

func validEntryCompression(entry *zip.File, version string) bool {
	if version == FormatVersion {
		return entry.Method == zip.Store
	}
	if version != legacyFormatVersion || (entry.Method != zip.Store && entry.Method != zip.Deflate) {
		return false
	}
	if entry.Method == zip.Store {
		return entry.UncompressedSize64 == entry.CompressedSize64
	}
	// V1 encrypted entries contain high-entropy ciphertext, so deflate should
	// not shrink them materially. This preserves old ZIPs while rejecting large
	// decompression bombs before a payload reader is opened.
	overhead := entry.UncompressedSize64 / 1000
	if overhead < 1<<20 {
		overhead = 1 << 20
	}
	return entry.CompressedSize64 <= ^uint64(0)-overhead && entry.UncompressedSize64 <= entry.CompressedSize64+overhead
}

func expectedArchiveEntries(version string) map[string]bool {
	switch version {
	case legacyFormatVersion:
		return map[string]bool{manifestEntry: true, legacyDatabaseDumpEntry: true, legacyGlobalsDumpEntry: true}
	case FormatVersion:
		return map[string]bool{manifestEntry: true, databaseDumpEntry: true, globalsDumpEntry: true, streamKeysetEntry: true}
	default:
		return nil
	}
}

func validPackageManifest(manifest Manifest) bool {
	if manifest.Source != BackupSource || manifest.TokensExported || manifest.TokenPolicy != TokenExportPolicy || !validKDF(manifest.KDF) || len(manifest.Files) != 2 {
		return false
	}
	var databaseName, globalsName string
	switch manifest.FormatVersion {
	case legacyFormatVersion:
		if manifest.ManifestSchema != legacyManifestSchemaVersion || manifest.Cipher != "" || manifest.KeysetFile != (File{}) {
			return false
		}
		databaseName, globalsName = legacyDatabaseDumpEntry, legacyGlobalsDumpEntry
	case FormatVersion:
		if manifest.ManifestSchema != ManifestSchemaVersion || manifest.Cipher != streamCipher || manifest.KeysetFile.Path != streamKeysetEntry || !validFileRecord(manifest.KeysetFile) || manifest.KeysetFile.SizeBytes > maxKeysetBytes {
			return false
		}
		databaseName, globalsName = databaseDumpEntry, globalsDumpEntry
	default:
		return false
	}
	seen := map[string]bool{}
	for _, record := range manifest.Files {
		if !validFileRecord(record) || (record.Path != databaseName && record.Path != globalsName) || seen[record.Path] {
			return false
		}
		seen[record.Path] = true
	}
	return seen[databaseName] && seen[globalsName]
}

func validFileRecord(record File) bool {
	if strings.TrimSpace(record.Path) == "" || record.SizeBytes <= 0 || len(record.SHA256) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(record.SHA256)
	return err == nil && len(decoded) == sha256.Size
}

func readZipEntryBounded(entry *zip.File, limit uint64) ([]byte, error) {
	if entry.UncompressedSize64 > limit {
		return nil, ErrInvalidPackage
	}
	reader, err := entry.Open()
	if err != nil {
		return nil, err
	}
	payload, readErr := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if uint64(len(payload)) != entry.UncompressedSize64 || uint64(len(payload)) > limit {
		return nil, ErrInvalidPackage
	}
	return payload, nil
}

func verifyV1Payloads(ctx context.Context, archive *packageArchive, key []byte) error {
	for _, record := range archive.manifest.Files {
		if err := ctx.Err(); err != nil {
			return err
		}
		entry := archive.entries[record.Path]
		if entry == nil || entry.UncompressedSize64 != uint64(record.SizeBytes) {
			return ErrInvalidPackage
		}
		source, err := entry.Open()
		if err != nil {
			return ErrInvalidPackage
		}
		digest := &digestReader{reader: contextReader{ctx: ctx, reader: source}, hash: sha256.New()}
		decryptErr := decryptV1Stream(digest, record.SizeBytes, key, []byte(record.Path), io.Discard)
		closeErr := source.Close()
		if decryptErr != nil {
			if errors.Is(decryptErr, ErrWrongPassphrase) {
				return ErrWrongPassphrase
			}
			return ErrInvalidPackage
		}
		if closeErr != nil || !digest.matches(record) {
			return ErrInvalidPackage
		}
	}
	return nil
}


func verifyV2Payloads(ctx context.Context, archive *packageArchive, masterKey []byte) error {
	keysetRecordBytes, err := readZipEntryBounded(archive.entries[streamKeysetEntry], maxKeysetBytes)
	if err != nil || !validByteRecord(archive.manifest.KeysetFile, keysetRecordBytes) {
		return ErrInvalidPackage
	}
	primitive, err := openStreamingPrimitive(masterKey, keysetRecordBytes)
	clearBytes(keysetRecordBytes)
	if err != nil {
		return ErrWrongPassphrase
	}
	for _, record := range archive.manifest.Files {
		if err := authenticateV2Entry(ctx, archive.entries[record.Path], record, primitive, io.Discard); err != nil {
			return err
		}
	}
	return nil
}

func authenticateV2Entry(ctx context.Context, entry *zip.File, record File, primitive tink.StreamingAEAD, plaintext io.Writer) error {
	if entry == nil || entry.UncompressedSize64 != uint64(record.SizeBytes) {
		return ErrInvalidPackage
	}
	source, err := entry.Open()
	if err != nil {
		return ErrInvalidPackage
	}
	digest := &digestReader{reader: contextReader{ctx: ctx, reader: source}, hash: sha256.New()}
	decryptReader, err := primitive.NewDecryptingReader(digest, payloadAAD(record.Path))
	if err == nil {
		_, err = io.Copy(plaintext, decryptReader)
	}
	closeErr := source.Close()
	if err != nil {
		if errors.Is(err, ErrWrongPassphrase) {
			return ErrWrongPassphrase
		}
		return ErrInvalidPackage
	}
	if closeErr != nil || !digest.matches(record) {
		return ErrInvalidPackage
	}
	return nil
}

func decryptPackageToFiles(ctx context.Context, path string, passphrase []byte, expected Manifest, directory string) (string, string, error) {
	archive, err := openPackage(path)
	if err != nil {
		return "", "", ErrInvalidPackage
	}
	defer archive.reader.Close()
	if archive.manifest.FormatVersion != expected.FormatVersion || !validPackageManifest(archive.manifest) {
		return "", "", ErrInvalidPackage
	}
	masterKey := deriveKey(passphrase, mustDecodeSalt(archive.manifest.KDF.Salt), archive.manifest.KDF)
	defer clearBytes(masterKey)
	var primitive tink.StreamingAEAD
	if archive.manifest.FormatVersion == FormatVersion {
		wrapped, err := readZipEntryBounded(archive.entries[streamKeysetEntry], maxKeysetBytes)
		if err != nil || !validByteRecord(archive.manifest.KeysetFile, wrapped) {
			return "", "", ErrInvalidPackage
		}
		primitive, err = openStreamingPrimitive(masterKey, wrapped)
		clearBytes(wrapped)
		if err != nil {
			return "", "", ErrWrongPassphrase
		}
	}
	databasePath := filepath.Join(directory, "database.dump")
	globalsPath := filepath.Join(directory, "globals.sql")
	for _, record := range archive.manifest.Files {
		if err := ctx.Err(); err != nil {
			return "", "", err
		}
		entry := archive.entries[record.Path]
		if entry == nil || entry.UncompressedSize64 != uint64(record.SizeBytes) {
			return "", "", ErrInvalidPackage
		}
		output, err := os.OpenFile(filepath.Join(directory, restoreFileName(record.Path)), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return "", "", err
		}
		var decryptErr error
		if archive.manifest.FormatVersion == legacyFormatVersion {
			source, openErr := entry.Open()
			if openErr != nil {
				decryptErr = ErrInvalidPackage
			} else {
				digest := &digestReader{reader: contextReader{ctx: ctx, reader: source}, hash: sha256.New()}
				decryptErr = decryptV1Stream(digest, record.SizeBytes, masterKey, []byte(record.Path), output)
				if closeErr := source.Close(); decryptErr == nil && closeErr != nil {
					decryptErr = closeErr
				}
				if decryptErr == nil && !digest.matches(record) {
					decryptErr = ErrInvalidPackage
				}
			}
		} else {
			decryptErr = authenticateV2Entry(ctx, entry, record, primitive, output)
		}
		if decryptErr == nil {
			decryptErr = output.Sync()
		}
		closeErr := output.Close()
		if decryptErr == nil {
			decryptErr = closeErr
		}
		if decryptErr != nil {
			_ = os.Remove(output.Name())
			return "", "", decryptErr
		}
	}
	return databasePath, globalsPath, nil
}

func restoreFileName(entryName string) string {
	if strings.Contains(entryName, "database.dump") {
		return "database.dump"
	}
	return "globals.sql"
}

func newStreamingPrimitive() (*keyset.Handle, tink.StreamingAEAD, error) {
	handle, err := keyset.NewHandle(streamingaead.AES256GCMHKDF1MBKeyTemplate())
	if err != nil {
		return nil, nil, err
	}
	primitive, err := streamingaead.New(handle)
	if err != nil {
		return nil, nil, err
	}
	return handle, primitive, nil
}

func encryptedKeysetBytes(handle *keyset.Handle, masterKey []byte) ([]byte, error) {
	wrappingAEAD, err := newKeysetWrappingAEAD(masterKey)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	writer := keyset.NewBinaryWriter(&buffer)
	if err := handle.WriteWithAssociatedData(writer, wrappingAEAD, []byte(keysetAAD)); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func openStreamingPrimitive(masterKey, encryptedKeyset []byte) (tink.StreamingAEAD, error) {
	wrappingAEAD, err := newKeysetWrappingAEAD(masterKey)
	if err != nil {
		return nil, err
	}
	reader := keyset.NewBinaryReader(bytes.NewReader(encryptedKeyset))
	handle, err := keyset.ReadWithAssociatedData(reader, wrappingAEAD, []byte(keysetAAD))
	if err != nil {
		return nil, err
	}
	return streamingaead.New(handle)
}

type keysetWrappingAEAD struct{ cipher.AEAD }

func newKeysetWrappingAEAD(key []byte) (keysetWrappingAEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return keysetWrappingAEAD{}, err
	}
	primitive, err := cipher.NewGCM(block)
	if err != nil {
		return keysetWrappingAEAD{}, err
	}
	return keysetWrappingAEAD{AEAD: primitive}, nil
}

func (aead keysetWrappingAEAD) Encrypt(plaintext, associatedData []byte) ([]byte, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := aead.Seal(nil, nonce, plaintext, associatedData)
	result := make([]byte, 0, len(wrappedKeysetHeader)+1+len(nonce)+len(sealed))
	result = append(result, wrappedKeysetHeader...)
	result = append(result, byte(len(nonce)))
	result = append(result, nonce...)
	result = append(result, sealed...)
	return result, nil
}

func (aead keysetWrappingAEAD) Decrypt(ciphertext, associatedData []byte) ([]byte, error) {
	minimum := len(wrappedKeysetHeader) + 1
	if len(ciphertext) < minimum || string(ciphertext[:len(wrappedKeysetHeader)]) != wrappedKeysetHeader {
		return nil, ErrWrongPassphrase
	}
	nonceLength := int(ciphertext[len(wrappedKeysetHeader)])
	if nonceLength != aead.NonceSize() || len(ciphertext) < minimum+nonceLength+aead.Overhead() {
		return nil, ErrWrongPassphrase
	}
	start := minimum
	return aead.Open(nil, ciphertext[start:start+nonceLength], ciphertext[start+nonceLength:], associatedData)
}

func payloadAAD(path string) []byte { return []byte("MAJUCAU-BACKUP-2:" + path) }

type digestWriter struct {
	writer io.Writer
	hash   hash.Hash
	count  int64
}

func (writer *digestWriter) Write(data []byte) (int, error) {
	count, err := writer.writer.Write(data)
	if count > 0 {
		_, _ = writer.hash.Write(data[:count])
		writer.count += int64(count)
	}
	if err == nil && count != len(data) {
		err = io.ErrShortWrite
	}
	return count, err
}

func (writer *digestWriter) file(path string) File {
	return File{Path: path, SHA256: hex.EncodeToString(writer.hash.Sum(nil)), SizeBytes: writer.count}
}

type digestReader struct {
	reader io.Reader
	hash   hash.Hash
	count  int64
}

func (reader *digestReader) Read(data []byte) (int, error) {
	count, err := reader.reader.Read(data)
	if count > 0 {
		_, _ = reader.hash.Write(data[:count])
		reader.count += int64(count)
	}
	return count, err
}

func (reader *digestReader) matches(expected File) bool {
	return reader.count == expected.SizeBytes && strings.EqualFold(hex.EncodeToString(reader.hash.Sum(nil)), expected.SHA256)
}

func validByteRecord(expected File, data []byte) bool {
	actual := fileRecord(expected.Path, data)
	return actual.SizeBytes == expected.SizeBytes && strings.EqualFold(actual.SHA256, expected.SHA256)
}

func clearBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}

func copyEncryptedPackage(ctx context.Context, sourcePath, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	information, err := source.Stat()
	if err != nil || !information.Mode().IsRegular() {
		return ErrInvalidPackage
	}
	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destination, contextReader{ctx: ctx, reader: source})
	if copyErr == nil {
		copyErr = destination.Sync()
	}
	closeErr := destination.Close()
	if copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		_ = os.Remove(destinationPath)
		return copyErr
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (reader contextReader) Read(data []byte) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	return reader.reader.Read(data)
}
