package googledrive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type fakeServiceAccountCredentialAuditor struct {
	err              error
	calls            int
	replacedExisting bool
}

func (a *fakeServiceAccountCredentialAuditor) RecordGoogleDriveCredentialConfigured(_ context.Context, actorSID, correlationID string, replaced bool) error {
	a.calls++
	a.replacedExisting = replaced
	if actorSID != "S-1-5-21-1-2-3-1001" || correlationID != "b36dd2ab-2954-4fe9-95cc-f2d55595b8de" {
		return errors.New("invalid test audit context")
	}
	return a.err
}

const (
	testDriveActorSID  = "S-1-5-21-1-2-3-1001"
	testDriveRequestID = "b36dd2ab-2954-4fe9-95cc-f2d55595b8de"
)

func TestServiceAccountCredentialServiceStoresOnlyValidBoundedKey(t *testing.T) {
	_, keyJSON := serviceAccountFixture(t)
	store := newMemorySecretStore()
	auditor := &fakeServiceAccountCredentialAuditor{}
	service := NewServiceAccountCredentialService(store, auditor)
	folderIDs := testDriveFolderIDs()

	if err := service.Save(context.Background(), keyJSON, folderIDs, testDriveActorSID, testDriveRequestID); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if auditor.calls != 1 || auditor.replacedExisting {
		t.Fatalf("audit calls=%d replaced=%v", auditor.calls, auditor.replacedExisting)
	}
	stored, err := store.Get(context.Background(), ServiceAccountSecretRef)
	var bundle serviceAccountCredentialBundle
	if err != nil || json.Unmarshal(stored, &bundle) != nil || bundle.Format != serviceAccountCredentialBundleFormat || !bytes.Equal(bundle.ServiceAccountKey, keyJSON) {
		t.Fatalf("credential bundle missing or invalid: error=%v format=%q", err, bundle.Format)
	}
	provider, err := NewServiceAccountTokenProvider(context.Background(), store)
	if err != nil || provider == nil {
		t.Fatalf("stored credential cannot initialize provider: provider=%v error=%v", provider, err)
	}
	loadedFolderIDs, configured, err := ReadDriveFolderIDs(context.Background(), store)
	if err != nil || !configured || loadedFolderIDs[DriveFolderOperationalLosses] != folderIDs[DriveFolderOperationalLosses] {
		t.Fatalf("stored folder IDs=%v configured=%v error=%v", loadedFolderIDs, configured, err)
	}
}

func TestServiceAccountCredentialServiceRejectsInvalidInputWithoutReplacingKey(t *testing.T) {
	_, existing := serviceAccountFixture(t)
	store := newMemorySecretStore()
	if err := store.Put(context.Background(), ServiceAccountSecretRef, existing); err != nil {
		t.Fatal(err)
	}
	service := NewServiceAccountCredentialService(store, &fakeServiceAccountCredentialAuditor{})
	invalid := []byte(`{"type":"authorized_user","client_secret":"do-not-leak-this-value"}`)
	err := service.Save(context.Background(), invalid, testDriveFolderIDs(), testDriveActorSID, testDriveRequestID)
	if !errors.Is(err, ErrServiceAccountCredentialInvalid) || strings.Contains(err.Error(), "do-not-leak-this-value") {
		t.Fatalf("Save() error = %v; want sanitized invalid-credential error", err)
	}
	after, readErr := store.Get(context.Background(), ServiceAccountSecretRef)
	if readErr != nil || !bytes.Equal(after, existing) {
		t.Fatalf("existing credential changed after invalid input: matches=%v error=%v", bytes.Equal(after, existing), readErr)
	}
}

func TestServiceAccountCredentialServiceRejectsInvalidFoldersWithoutReplacingKey(t *testing.T) {
	_, existing := serviceAccountFixture(t)
	store := newMemorySecretStore()
	if err := store.Put(context.Background(), ServiceAccountSecretRef, existing); err != nil {
		t.Fatal(err)
	}
	invalidFolders := testDriveFolderIDs()
	invalidFolders[DriveFolderNuvemShipping] = invalidFolders[DriveFolderNuvemPagoReceipts]
	err := NewServiceAccountCredentialService(store, &fakeServiceAccountCredentialAuditor{}).Save(context.Background(), existing, invalidFolders, testDriveActorSID, testDriveRequestID)
	if !errors.Is(err, ErrServiceAccountCredentialInvalid) {
		t.Fatalf("Save() error = %v; want invalid configuration", err)
	}
	after, readErr := store.Get(context.Background(), ServiceAccountSecretRef)
	if readErr != nil || !bytes.Equal(after, existing) {
		t.Fatalf("existing credential changed after invalid folders: matches=%v error=%v", bytes.Equal(after, existing), readErr)
	}
}

func TestServiceAccountCredentialServiceHonorsCancellationAndMissingStore(t *testing.T) {
	if err := NewServiceAccountCredentialService(nil, &fakeServiceAccountCredentialAuditor{}).Save(context.Background(), []byte(`{}`), testDriveFolderIDs(), testDriveActorSID, testDriveRequestID); !errors.Is(err, ErrServiceAccountCredentialStore) {
		t.Fatalf("Save() without store error = %v", err)
	}
	store := newMemorySecretStore()
	service := NewServiceAccountCredentialService(store, &fakeServiceAccountCredentialAuditor{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := service.Save(ctx, []byte(`{}`), testDriveFolderIDs(), testDriveActorSID, testDriveRequestID); !errors.Is(err, context.Canceled) {
		t.Fatalf("Save() with cancelled context error = %v", err)
	}
	if value, _ := store.Get(context.Background(), ServiceAccountSecretRef); len(value) != 0 {
		t.Fatalf("cancelled request wrote credential: length=%d", len(value))
	}
}

func TestServiceAccountCredentialServiceRejectsOversizedInputBeforeStore(t *testing.T) {
	store := newMemorySecretStore()
	service := NewServiceAccountCredentialService(store, &fakeServiceAccountCredentialAuditor{})
	if err := service.Save(context.Background(), bytes.Repeat([]byte("x"), maxServiceAccountJSON+1), testDriveFolderIDs(), testDriveActorSID, testDriveRequestID); !errors.Is(err, ErrServiceAccountCredentialInvalid) {
		t.Fatalf("Save() oversized input error = %v", err)
	}
	if value, _ := store.Get(context.Background(), ServiceAccountSecretRef); len(value) != 0 {
		t.Fatalf("oversized request wrote credential: length=%d", len(value))
	}
}

func TestServiceAccountCredentialServiceDoesNotPublishWhenAuditFails(t *testing.T) {
	_, existing := serviceAccountFixture(t)
	store := newMemorySecretStore()
	if err := store.Put(context.Background(), ServiceAccountSecretRef, existing); err != nil {
		t.Fatal(err)
	}
	_, replacement := serviceAccountFixture(t)
	auditor := &fakeServiceAccountCredentialAuditor{err: errors.New("database offline")}
	err := NewServiceAccountCredentialService(store, auditor).Save(context.Background(), replacement, testDriveFolderIDs(), testDriveActorSID, testDriveRequestID)
	if !errors.Is(err, ErrServiceAccountCredentialAudit) {
		t.Fatalf("Save() error = %v; want audit failure", err)
	}
	after, readErr := store.Get(context.Background(), ServiceAccountSecretRef)
	if readErr != nil || !bytes.Equal(after, existing) {
		t.Fatalf("existing credential changed after audit failure: matches=%v error=%v", bytes.Equal(after, existing), readErr)
	}
}

func testDriveFolderIDs() DriveFolderIDs {
	return DriveFolderIDs{
		DriveFolderOperationalLosses: "losses-folder",
		DriveFolderNuvemPagoReceipts: "receipts-folder",
		DriveFolderNuvemShipping:     "shipping-folder",
	}
}
