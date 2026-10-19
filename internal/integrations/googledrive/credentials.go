package googledrive

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"sync"

	"github.com/google/uuid"
	"majucau.local/financial-intelligence/internal/security"
)

var (
	ErrServiceAccountCredentialInvalid = errors.New("Google Drive service-account credential is invalid")
	ErrServiceAccountCredentialStore   = errors.New("Google Drive service-account credential could not be protected")
	ErrServiceAccountCredentialAudit   = errors.New("Google Drive service-account credential audit failed")
)

type ServiceAccountCredentialAuditor interface {
	RecordGoogleDriveCredentialConfigured(context.Context, string, string, bool) error
}

// ServiceAccountCredentialService validates and stores the read-only Drive
// service-account key. The caller supplies the protected SecretStore owned by
// the Windows worker; this service never writes credential bytes to the
// database or returns them to its caller.
type ServiceAccountCredentialService struct {
	store   security.SecretStore
	auditor ServiceAccountCredentialAuditor
	mu      sync.Mutex
}

type serviceAccountCredentialBundle struct {
	Format            string          `json:"format"`
	ServiceAccountKey json.RawMessage `json:"service_account_key"`
	FolderIDs         DriveFolderIDs  `json:"folder_ids"`
}

const serviceAccountCredentialBundleFormat = "majucau-google-drive-v1"

func NewServiceAccountCredentialService(store security.SecretStore, auditor ServiceAccountCredentialAuditor) *ServiceAccountCredentialService {
	return &ServiceAccountCredentialService{store: store, auditor: auditor}
}

// Save accepts only a bounded Google service-account key for the official
// token endpoint, the Drive read-only scope, and exactly the three approved
// folder IDs. One protected-store write publishes the key and folder mapping
// together, so a partial configuration cannot replace the previous one.
func (s *ServiceAccountCredentialService) Save(ctx context.Context, keyJSON []byte, folderIDs DriveFolderIDs, actorWindowsSID, correlationID string) error {
	if s == nil || s.store == nil {
		return ErrServiceAccountCredentialStore
	}
	if ctx == nil {
		return ErrServiceAccountCredentialInvalid
	}
	if s.auditor == nil || strings.TrimSpace(actorWindowsSID) == "" {
		return ErrServiceAccountCredentialAudit
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := uuid.Parse(correlationID); err != nil {
		return ErrServiceAccountCredentialInvalid
	}
	if ValidateDriveFolderIDs(folderIDs) != nil {
		return ErrServiceAccountCredentialInvalid
	}
	config, err := serviceAccountJWTConfig(keyJSON)
	if err != nil {
		return ErrServiceAccountCredentialInvalid
	}
	wipeBytes(config.PrivateKey)
	if err := ctx.Err(); err != nil {
		return err
	}
	previous, previousErr := s.store.Get(ctx, ServiceAccountSecretRef)
	previousExists := previousErr == nil
	if previousErr != nil && !errors.Is(previousErr, fs.ErrNotExist) {
		return ErrServiceAccountCredentialStore
	}
	defer wipeBytes(previous)
	bundle, err := json.Marshal(serviceAccountCredentialBundle{
		Format:            serviceAccountCredentialBundleFormat,
		ServiceAccountKey: json.RawMessage(keyJSON),
		FolderIDs:         cloneDriveFolderIDs(folderIDs),
	})
	if err != nil || len(bundle) > maxServiceAccountJSON {
		return ErrServiceAccountCredentialInvalid
	}
	defer wipeBytes(bundle)
	if err := s.store.Put(ctx, ServiceAccountSecretRef, bundle); err != nil {
		return ErrServiceAccountCredentialStore
	}
	if err := s.auditor.RecordGoogleDriveCredentialConfigured(ctx, actorWindowsSID, correlationID, previousExists); err != nil {
		var rollbackErr error
		if previousExists {
			rollbackErr = s.store.Put(context.Background(), ServiceAccountSecretRef, previous)
		} else {
			rollbackErr = s.store.Delete(context.Background(), ServiceAccountSecretRef)
		}
		if rollbackErr != nil {
			return errors.Join(ErrServiceAccountCredentialAudit, ErrServiceAccountCredentialStore)
		}
		return ErrServiceAccountCredentialAudit
	}
	return nil
}

// ReadDriveFolderIDs loads the folder IDs from the same protected bundle as
// the service-account key. Older raw-key entries return configured=false; the
// worker stays pending until the combined protected bundle is provisioned.
func ReadDriveFolderIDs(ctx context.Context, store security.SecretStore) (DriveFolderIDs, bool, error) {
	if ctx == nil || store == nil {
		return nil, false, ErrServiceAccountCredentialStore
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	value, err := store.Get(ctx, ServiceAccountSecretRef)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, ErrServiceAccountCredentialStore
	}
	defer wipeBytes(value)
	var bundle serviceAccountCredentialBundle
	decodeErr := json.Unmarshal(value, &bundle)
	defer wipeBytes(bundle.ServiceAccountKey)
	if decodeErr != nil || bundle.Format != serviceAccountCredentialBundleFormat {
		return nil, false, nil
	}
	if len(bundle.ServiceAccountKey) == 0 || ValidateDriveFolderIDs(bundle.FolderIDs) != nil {
		return nil, false, ErrServiceAccountCredentialInvalid
	}
	return cloneDriveFolderIDs(bundle.FolderIDs), true, nil
}

func cloneDriveFolderIDs(source DriveFolderIDs) DriveFolderIDs {
	clone := make(DriveFolderIDs, len(source))
	for scope, id := range source {
		clone[scope] = id
	}
	return clone
}
