//go:build windows

package main

import (
	"context"
	"os"
	"strings"

	"majucau.local/financial-intelligence/internal/ipc"
)

func runWorkerPipe(ctx context.Context, handler ipc.Handler) error {
	cfg := ipc.DefaultNamedPipeConfig()
	// The installer supplies the exact desktop user's SID. Omitting it is
	// allowed only for development and binds the ACL to the service process SID;
	// production installation must configure MAJUCAU_UI_SID explicitly.
	if sid := strings.TrimSpace(os.Getenv("MAJUCAU_UI_SID")); sid != "" {
		cfg.AllowedClientSIDs = []string{sid}
	}
	if sid := strings.TrimSpace(os.Getenv("MAJUCAU_SERVICE_SID")); sid != "" {
		cfg.ServiceSID = sid
	}
	server, err := ipc.NewNamedPipeServer(cfg)
	if err != nil {
		return err
	}
	defer server.Close()
	return server.Serve(ctx, handler)
}
