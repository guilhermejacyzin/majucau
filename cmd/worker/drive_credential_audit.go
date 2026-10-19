package main

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"majucau.local/financial-intelligence/internal/integrations/googledrive"
)

type postgresGoogleDriveCredentialAuditor struct {
	pool *pgxpool.Pool
}

func (a *postgresGoogleDriveCredentialAuditor) RecordGoogleDriveCredentialConfigured(ctx context.Context, actorWindowsSID, correlationID string, replacedExisting bool) error {
	if a == nil || a.pool == nil || ctx == nil || strings.TrimSpace(actorWindowsSID) == "" {
		return googledrive.ErrServiceAccountCredentialAudit
	}
	correlation, err := uuid.Parse(correlationID)
	if err != nil {
		return googledrive.ErrServiceAccountCredentialAudit
	}
	_, err = a.pool.Exec(ctx, `SELECT public.audit_google_drive_credential_configured($1, $2, $3)`, actorWindowsSID, correlation, replacedExisting)
	if err != nil {
		return errors.Join(googledrive.ErrServiceAccountCredentialAudit, err)
	}
	return nil
}
