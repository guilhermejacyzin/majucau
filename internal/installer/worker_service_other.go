//go:build !windows

package installer

import "errors"

var (
	ErrWorkerServiceUnsupported = errors.New("Windows worker service management is unsupported on this platform")
	ErrWorkerServiceConfig      = errors.New("worker service configuration is invalid")
	ErrWorkerServiceState       = errors.New("worker service state transition failed")
)

type WorkerServiceOptions struct {
	InstallDir string
	WorkerPath string
	DataDir    string
}

func InstallOrUpdateWorkerService(WorkerServiceOptions) error { return ErrWorkerServiceUnsupported }
func RemoveWorkerService(string) error                        { return ErrWorkerServiceUnsupported }
