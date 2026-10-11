package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"majucau.local/financial-intelligence/internal/ipc"
)

func runWorker(ctx context.Context) (returnErr error) {
	logger, closeLogger, logErr := newWorkerDiagnosticLogger()
	defer func() {
		if closeErr := closeLogger(); closeErr != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("close worker diagnostic log: %w", closeErr))
		}
	}()
	if logErr != nil {
		logger.Error("worker diagnostic file unavailable", slog.String("error_code", "DIAGNOSTICS_LOG_UNAVAILABLE"))
	}
	handler := newWorkerHandler()
	defer handler.Close()
	health := handler.health.CheckHealth(ctx)
	if payload, err := json.Marshal(health); err == nil {
		fmt.Println(string(payload))
	}
	return runWorkerPipe(ctx, diagnosticRequestHandler{next: handler, logger: logger})
}

func (h workerHandler) Close() {
	if h.receiptImporter != nil {
		h.receiptImporter.Close()
	}
}

var _ ipc.Handler = workerHandler{}
