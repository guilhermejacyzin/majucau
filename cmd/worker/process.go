package main

import (
	"context"
	"encoding/json"
	"fmt"

	"majucau.local/financial-intelligence/internal/ipc"
)

func runWorker(ctx context.Context) error {
	handler := newWorkerHandler()
	defer handler.Close()
	health := handler.health.CheckHealth(ctx)
	if payload, err := json.Marshal(health); err == nil {
		fmt.Println(string(payload))
	}
	return runWorkerPipe(ctx, handler)
}

func (h workerHandler) Close() {
	if h.receiptImporter != nil {
		h.receiptImporter.Close()
	}
}

var _ ipc.Handler = workerHandler{}
