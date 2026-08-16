//go:build windows

package main

import (
	"context"
	"errors"
	"syscall"
	"unsafe"
)

const (
	serviceWin32OwnProcess                            = 0x00000010
	serviceStartPending                               = 0x00000002
	serviceRunning                                    = 0x00000004
	serviceStopped                                    = 0x00000001
	serviceAcceptStop                                 = 0x00000001
	serviceAcceptShutdown                             = 0x00000004
	serviceControlStop                                = 0x00000001
	serviceControlShutdown                            = 0x00000005
	errorFailedServiceControllerConnect syscall.Errno = 1063
)

type serviceStatus struct{ serviceType, currentState, controlsAccepted, win32ExitCode, serviceSpecificExitCode, checkPoint, waitHint uint32 }
type serviceTableEntry struct {
	serviceName *uint16
	serviceProc uintptr
}

var (
	advapiService                 = syscall.NewLazyDLL("advapi32.dll")
	startServiceCtrlDispatcherW   = advapiService.NewProc("StartServiceCtrlDispatcherW")
	registerServiceCtrlHandlerExW = advapiService.NewProc("RegisterServiceCtrlHandlerExW")
	setServiceStatus              = advapiService.NewProc("SetServiceStatus")
	serviceCancel                 context.CancelFunc
	serviceStatusHandle           uintptr
	serviceName                   string
	serviceRun                    func(context.Context) error
)

func tryRunAsService(name string, run func(context.Context) error) (bool, error) {
	serviceName, serviceRun = name, run
	n, _ := syscall.UTF16PtrFromString(name)
	table := [2]serviceTableEntry{{serviceName: n, serviceProc: syscall.NewCallback(serviceMain)}, {}}
	r, _, err := startServiceCtrlDispatcherW.Call(uintptr(unsafe.Pointer(&table[0])))
	if r == 0 {
		if errno, ok := err.(syscall.Errno); ok && errno == errorFailedServiceControllerConnect {
			return false, nil
		}
		return true, err
	}
	return true, nil
}
func serviceMain(_ uint32, _ **uint16) uintptr {
	n, _ := syscall.UTF16PtrFromString(serviceName)
	handler := syscall.NewCallback(serviceControlHandler)
	h, _, _ := registerServiceCtrlHandlerExW.Call(uintptr(unsafe.Pointer(n)), handler, 0)
	serviceStatusHandle = h
	if h == 0 {
		return 0
	}
	_ = publishServiceStatus(serviceStartPending, 0, 1, 5000)
	ctx, cancel := context.WithCancel(context.Background())
	serviceCancel = cancel
	_ = publishServiceStatus(serviceRunning, serviceAcceptStop|serviceAcceptShutdown, 0, 0)
	err := serviceRun(ctx)
	cancel()
	code := uint32(0)
	if err != nil {
		code = 1
	}
	_ = publishServiceStatus(serviceStopped, 0, code, 0)
	return 0
}
func serviceControlHandler(control, _ uint32, _ uintptr, _ uintptr) uintptr {
	if (control == serviceControlStop || control == serviceControlShutdown) && serviceCancel != nil {
		serviceCancel()
	}
	return 0
}
func publishServiceStatus(state, accepted, exitCode, waitHint uint32) error {
	if serviceStatusHandle == 0 {
		return errors.New("service status handle unavailable")
	}
	status := serviceStatus{serviceType: serviceWin32OwnProcess, currentState: state, controlsAccepted: accepted, win32ExitCode: exitCode, waitHint: waitHint}
	r, _, err := setServiceStatus.Call(serviceStatusHandle, uintptr(unsafe.Pointer(&status)))
	if r == 0 {
		return err
	}
	return nil
}
