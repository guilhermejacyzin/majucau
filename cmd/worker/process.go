package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"majucau.local/financial-intelligence/internal/ipc"
)

func runWorker(ctx context.Context) error {
	logger, closeLogger, logErr := newWorkerDiagnosticLogger()
	defer closeLogger()
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
