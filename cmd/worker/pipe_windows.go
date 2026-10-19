//go:build windows

package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"majucau.local/financial-intelligence/internal/ipc"
)

func runWorkerPipe(ctx context.Context, handler ipc.Handler) (returnErr error) {
	cfg := ipc.DefaultNamedPipeConfig()
	if usesWindowsServiceExecution(ctx) {
		serviceSID, err := lookupInstalledServiceSID()
		if err != nil {
			return err
		}
		uiSID, err := readInstalledUIUserSID()
		if err != nil {
			return err
		}
		cfg.ServiceSID = serviceSID
		cfg.AllowedClientSIDs = []string{uiSID}
	} else {
		// These environment overrides are development-only. An installed service
		// derives its own service SID and reads the UI SID from its Parameters key.
		if sid := strings.TrimSpace(os.Getenv("MAJUCAU_UI_SID")); sid != "" {
			cfg.AllowedClientSIDs = []string{sid}
		}
		if sid := strings.TrimSpace(os.Getenv("MAJUCAU_SERVICE_SID")); sid != "" {
			cfg.ServiceSID = sid
		}
	}
	server, err := ipc.NewNamedPipeServer(cfg)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, server.Close()) }()
	return server.Serve(ctx, handler)
}

func lookupInstalledServiceSID() (string, error) {
	sid, _, _, err := windows.LookupSID("", `NT SERVICE\MajucauWorker`)
	if err != nil || sid == nil {
		return "", errors.New("installed worker service identity is unavailable")
	}
	return sid.String(), nil
}

func readInstalledUIUserSID() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\MajucauWorker\Parameters`, registry.QUERY_VALUE)
	if err != nil {
		return "", errors.New("installed desktop identity is unavailable")
	}
	defer key.Close()
	sidText, _, err := key.GetStringValue("UIUserSID")
	if err != nil {
		return "", errors.New("installed desktop identity is unavailable")
	}
	if _, err := windows.StringToSid(sidText); err != nil {
		return "", errors.New("installed desktop identity is invalid")
	}
	return sidText, nil
}
