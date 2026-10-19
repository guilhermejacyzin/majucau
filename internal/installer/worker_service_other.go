//go:build !windows

package installer

import "errors"

var (
	ErrWorkerServiceUnsupported = errors.New("Windows worker service management is unsupported on this platform")
	ErrWorkerServiceConfig      = errors.New("worker service configuration is invalid")
	ErrWorkerServiceState       = errors.New("worker service state transition failed")
)

func WorkerServiceFailureCode(err error) string {
	switch {
	case errors.Is(err, ErrInstallerLockFailed):
		return "INSTALL_LOCK_FAILED"
	case errors.Is(err, ErrWorkerServiceState):
		return "WORKER_SERVICE_STATE_FAILED"
	case errors.Is(err, ErrWorkerServiceConfig):
		return "WORKER_SERVICE_CONFIG_INVALID"
	default:
		return "WORKER_SERVICE_INSTALL_FAILED"
	}
}

type WorkerServiceOptions struct {
	InstallDir string
	WorkerPath string
	DataDir    string
}

func InstallOrUpdateWorkerService(WorkerServiceOptions) error { return ErrWorkerServiceUnsupported }
func RemoveWorkerService(string) error                        { return ErrWorkerServiceUnsupported }
