package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"majucau.local/financial-intelligence/internal/security"
)

func newWorkerSecretStore(ctx context.Context) security.SecretStore {
	dir := strings.TrimSpace(os.Getenv("MAJUCAU_SECRET_DIR"))
	if dir == "" {
		programData := strings.TrimSpace(os.Getenv("ProgramData"))
		if programData == "" {
			return nil
		}
		dir = filepath.Join(programData, "Majucau", "Secrets")
	}
	newStore := security.NewDPAPIStore
	if usesWindowsServiceExecution(ctx) {
		newStore = security.NewMachineDPAPIStore
	}
	store, err := newStore(dir)
	if err != nil {
		return nil
	}
	return store
}
