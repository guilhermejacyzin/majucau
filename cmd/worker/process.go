package main

import (
	"context"
	"encoding/json"
	"fmt"

	"majucau.local/financial-intelligence/internal/ipc"
)

func runWorker(ctx context.Context) error {
	handler := newWorkerHandler()
	health := handler.health.CheckHealth(ctx)
	if payload, err := json.Marshal(health); err == nil {
		fmt.Println(string(payload))
	}
	return runWorkerPipe(ctx, handler)
}

var _ ipc.Handler = workerHandler{}
