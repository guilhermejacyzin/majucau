//go:build !windows

package main

import (
	"context"

	"majucau.local/financial-intelligence/internal/ipc"
)

// Console fallback is intentionally a no-op off Windows: no socket or network
// listener is opened, while the worker contracts remain testable with FakeClient.
func runWorkerPipe(_ context.Context, _ ipc.Handler) error { return nil }
