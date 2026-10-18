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
	if handler.databaseErrorCode != "" {
		logger.Warn("worker database-backed operations are unavailable", slog.String("error_code", handler.databaseErrorCode))
	}
	health := handler.health.CheckHealth(ctx)
	if payload, err := json.Marshal(health); err == nil {
		fmt.Println(string(payload))
	}
	workerCtx, cancelWorker := context.WithCancel(ctx)
	defer cancelWorker()
	dailyDone := startDailyBlingSchedule(workerCtx, handler.dailyBlingSync, logger)
	pipeErr := runWorkerPipe(workerCtx, diagnosticRequestHandler{next: handler, logger: logger})
	cancelWorker()
	<-dailyDone
	return pipeErr
}

func (h workerHandler) Close() {
	if h.resources != nil {
		h.resources.Close()
	}
}

var _ ipc.Handler = workerHandler{}
