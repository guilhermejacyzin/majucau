package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"majucau.local/financial-intelligence/internal/integrations/googledrive"
	"majucau.local/financial-intelligence/internal/ipc"
)

func runDriveCredentialProvision(args []string, output io.Writer) int {
	flags := flag.NewFlagSet("provision-google-drive", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	filePath := flags.String("file", "", "absolute path to the downloaded service-account JSON file")
	folderIDsPath := flags.String("folder-ids-file", "", "absolute path to the approved Drive folder IDs JSON file")
	removeSource := flags.Bool("remove-source-after-success", false, "remove the source JSON file after it has been saved")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || strings.TrimSpace(*filePath) == "" || strings.TrimSpace(*folderIDsPath) == "" {
		_, _ = io.WriteString(output, "GOOGLE_DRIVE_CREDENTIAL_ARGUMENTS_INVALID\n")
		return 2
	}

	folderIDs, err := readDriveFolderIDsFile(*folderIDsPath)
	if err != nil {
		_, _ = io.WriteString(output, "GOOGLE_DRIVE_FOLDER_CONFIGURATION_INVALID\n")
		return 2
	}
	keyJSON, err := readDriveCredentialFile(*filePath)
	if err != nil {
		_, _ = io.WriteString(output, "GOOGLE_DRIVE_CREDENTIAL_FILE_INVALID\n")
		return 2
	}
	defer wipeDriveCredentialBytes(keyJSON)

	client, err := ipc.NewNamedPipeClient(ipc.DefaultNamedPipeConfig())
	if err != nil {
		_, _ = io.WriteString(output, "GOOGLE_DRIVE_WORKER_UNAVAILABLE\n")
		return 1
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	requestPayload, err := json.Marshal(struct {
		ServiceAccountKey json.RawMessage            `json:"service_account_key"`
		FolderIDs         googledrive.DriveFolderIDs `json:"folder_ids"`
	}{
		ServiceAccountKey: json.RawMessage(keyJSON),
		FolderIDs:         folderIDs,
	})
	if err != nil || len(requestPayload) > ipc.MaxMessageSize {
		_, _ = io.WriteString(output, "GOOGLE_DRIVE_CREDENTIAL_ARGUMENTS_INVALID\n")
		return 2
	}
	defer wipeDriveCredentialBytes(requestPayload)
	response, err := client.Call(ctx, ipc.Request{
		Version: ipc.ProtocolVersion, RequestID: uuid.NewString(), Method: ipc.MethodGoogleDriveCredentialSave,
		Payload: json.RawMessage(requestPayload),
	})
	if err != nil {
		_, _ = io.WriteString(output, "GOOGLE_DRIVE_WORKER_UNAVAILABLE\n")
		return 1
	}
	if !response.OK {
		_, _ = io.WriteString(output, driveCredentialProvisionErrorCode(response)+"\n")
		return 1
	}

	if *removeSource {
		if err := os.Remove(*filePath); err != nil {
			_, _ = io.WriteString(output, "GOOGLE_DRIVE_CREDENTIAL_SAVED_SOURCE_NOT_REMOVED\n")
			return 1
		}
		_, _ = io.WriteString(output, "GOOGLE_DRIVE_CREDENTIAL_SAVED_SOURCE_REMOVED\n")
		return 0
	}
	_, _ = io.WriteString(output, "GOOGLE_DRIVE_CREDENTIAL_SAVED_REMOVE_SOURCE_COPY\n")
	return 0
}

func driveCredentialProvisionErrorCode(response ipc.Response) string {
	if response.Error == nil {
		return "GOOGLE_DRIVE_CREDENTIAL_SAVE_FAILED"
	}
	switch response.Error.Code {
	case "GOOGLE_DRIVE_VAULT_UNAVAILABLE",
		"GOOGLE_DRIVE_CALLER_UNAVAILABLE",
		"GOOGLE_DRIVE_AUDIT_UNAVAILABLE",
		"GOOGLE_DRIVE_CREDENTIAL_INVALID",
		"GOOGLE_DRIVE_CREDENTIAL_CANCELLED",
		"GOOGLE_DRIVE_CREDENTIAL_SAVE_FAILED":
		return response.Error.Code
	default:
		return "GOOGLE_DRIVE_CREDENTIAL_SAVE_FAILED"
	}
}

func readDriveCredentialFile(path string) ([]byte, error) {
	if !filepath.IsAbs(path) || strings.HasPrefix(path, `\\`) || strings.ContainsRune(path, 0) || filepath.VolumeName(path) == "" {
		return nil, os.ErrInvalid
	}
	path = filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(path)
	resolvedPath, ok := normalizeResolvedDrivePath(resolved)
	if err != nil || !ok || !strings.EqualFold(filepath.VolumeName(resolvedPath), filepath.VolumeName(path)) {
		return nil, os.ErrInvalid
	}
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 || before.Size() < 1 || before.Size() > googledrive.MaxServiceAccountJSONBytes {
		return nil, os.ErrInvalid
	}
	resolvedInfo, err := os.Stat(resolvedPath)
	if err != nil || !resolvedInfo.Mode().IsRegular() || !os.SameFile(before, resolvedInfo) {
		return nil, os.ErrInvalid
	}
	file, err := os.Open(resolvedPath)
	if err != nil {
		return nil, err
	}
	after, statErr := file.Stat()
	if statErr != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errors.Join(os.ErrInvalid, statErr, file.Close())
	}
	data, readErr := io.ReadAll(io.LimitReader(file, googledrive.MaxServiceAccountJSONBytes+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil || len(data) == 0 || len(data) > googledrive.MaxServiceAccountJSONBytes {
		wipeDriveCredentialBytes(data)
		return nil, errors.Join(os.ErrInvalid, readErr, closeErr)
	}
	return data, nil
}

func readDriveFolderIDsFile(path string) (googledrive.DriveFolderIDs, error) {
	const maxFolderConfigBytes = 16 << 10
	if !filepath.IsAbs(path) || strings.HasPrefix(path, `\\`) || strings.ContainsRune(path, 0) || filepath.VolumeName(path) == "" {
		return nil, os.ErrInvalid
	}
	path = filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(path)
	resolvedPath, ok := normalizeResolvedDrivePath(resolved)
	if err != nil || !ok || !strings.EqualFold(filepath.VolumeName(resolvedPath), filepath.VolumeName(path)) {
		return nil, os.ErrInvalid
	}
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 || before.Size() < 1 || before.Size() > maxFolderConfigBytes {
		return nil, os.ErrInvalid
	}
	resolvedInfo, err := os.Stat(resolvedPath)
	if err != nil || !resolvedInfo.Mode().IsRegular() || !os.SameFile(before, resolvedInfo) {
		return nil, os.ErrInvalid
	}
	file, err := os.Open(resolvedPath)
	if err != nil {
		return nil, err
	}
	after, statErr := file.Stat()
	if statErr != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errors.Join(os.ErrInvalid, statErr, file.Close())
	}
	limited := &io.LimitedReader{R: file, N: maxFolderConfigBytes + 1}
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	var folderIDs googledrive.DriveFolderIDs
	decodeErr := decoder.Decode(&folderIDs)
	if decodeErr == nil {
		var trailing any
		if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
			decodeErr = os.ErrInvalid
		}
	}
	closeErr := file.Close()
	if decodeErr != nil || closeErr != nil || limited.N == 0 || googledrive.ValidateDriveFolderIDs(folderIDs) != nil {
		return nil, errors.Join(os.ErrInvalid, decodeErr, closeErr)
	}
	return folderIDs, nil
}

func normalizeResolvedDrivePath(path string) (string, bool) {
	path = filepath.Clean(path)
	if strings.HasPrefix(path, `\\?\`) {
		// EvalSymlinks may return the Windows extended-length form for an
		// ordinary local drive path. Compare it in the same form the caller
		// supplied, while still rejecting device and UNC paths.
		path = strings.TrimPrefix(path, `\\?\`)
		if len(path) < 3 || path[1] != ':' || path[2] != '\\' {
			return "", false
		}
	}
	if strings.HasPrefix(path, `\\`) || !filepath.IsAbs(path) || filepath.VolumeName(path) == "" {
		return "", false
	}
	return path, true
}

func wipeDriveCredentialBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
