package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"majucau.local/financial-intelligence/internal/security"
)

func newWorkerSecretStore(ctx context.Context) security.SecretStore {
	programData := strings.TrimSpace(os.Getenv("ProgramData"))
	if usesWindowsServiceExecution(ctx) {
		if programData == "" {
			return nil
		}
		// The service uses its own machine-DPAPI vault. Do not allow a
		// per-process environment override to redirect persistent secrets.
		dir := filepath.Join(programData, "Majucau", "WorkerSecrets", "Vault")
		store, err := security.NewMachineDPAPIStore(dir)
		if err != nil {
			return nil
		}
		return store
	}
	dir := strings.TrimSpace(os.Getenv("MAJUCAU_SECRET_DIR"))
	if dir == "" {
		if programData == "" {
			return nil
		}
		dir = filepath.Join(programData, "Majucau", "Secrets")
	}
	newStore := security.NewDPAPIStore
	store, err := newStore(dir)
	if err != nil {
		return nil
	}
	return store
}
