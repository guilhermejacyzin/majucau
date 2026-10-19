//go:build !windows

package installer

import "errors"

var ErrWorkerServiceUnsupported = errors.New("Windows worker service management is unsupported on this platform")

type WorkerServiceOptions struct {
	InstallDir string
	WorkerPath string
	DataDir    string
}

func InstallOrUpdateWorkerService(WorkerServiceOptions) error { return ErrWorkerServiceUnsupported }
func RemoveWorkerService(string) error                        { return ErrWorkerServiceUnsupported }
