package main

import (
	"os"
	"path/filepath"
	"strings"

	"majucau.local/financial-intelligence/internal/security"
)

func newWorkerSecretStore() security.SecretStore {
	dir := strings.TrimSpace(os.Getenv("MAJUCAU_SECRET_DIR"))
	if dir == "" {
		programData := strings.TrimSpace(os.Getenv("ProgramData"))
		if programData == "" {
			return nil
		}
		dir = filepath.Join(programData, "Majucau", "Secrets")
	}
	store, err := security.NewDPAPIStore(dir)
	if err != nil {
		return nil
	}
	return store
}
