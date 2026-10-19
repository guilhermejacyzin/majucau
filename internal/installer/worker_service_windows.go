//go:build windows

package installer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	WorkerServiceName        = "MajucauWorker"
	workerServiceDisplayName = "Majucau Background Worker"
	workerServiceAccount     = `NT SERVICE\MajucauWorker`
	workerSecretsDirName     = "WorkerSecrets"
	workerLogsDirName        = "WorkerLogs"
	workerServiceTimeout     = 45 * time.Second
)

var (
	ErrWorkerServiceConfig = errors.New("worker service configuration is invalid")
	ErrWorkerServiceState  = errors.New("worker service state transition failed")
)

type workerServiceFailure struct {
	stage string
	cause error
}

func (failure workerServiceFailure) Error() string { return failure.cause.Error() }
func (failure workerServiceFailure) Unwrap() error { return failure.cause }

// WorkerServiceFailureCode returns a stable stage code without exposing local
// paths, account names, or native Windows error strings to callers.
func WorkerServiceFailureCode(err error) string {
	var failure workerServiceFailure
	if errors.As(err, &failure) {
		return "WORKER_SERVICE_" + failure.stage
	}
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

// WorkerServiceOptions contains local paths only. It never accepts secrets.
type WorkerServiceOptions struct {
	InstallDir string
	WorkerPath string
	DataDir    string
}

// InstallOrUpdateWorkerService installs the worker under its virtual service
// account, prepares protected directories before startup, and rolls a failed
// update back to the last registered binary and service configuration.
func InstallOrUpdateWorkerService(options WorkerServiceOptions) (returnErr error) {
	failureStage := "PATH_VALIDATION"
	defer func() {
		if returnErr != nil {
			returnErr = workerServiceFailure{stage: failureStage, cause: returnErr}
		}
	}()
	installDir, workerPath, dataDir, err := validateWorkerServicePaths(options)
	if err != nil {
		return err
	}
	failureStage = "INSTALL_LOCK"
	installLock, err := AcquireInstallLock(DefaultInstallLockName)
	if err != nil {
		return ErrInstallerLockFailed
	}
	defer func() { returnErr = errors.Join(returnErr, installLock.Release()) }()

	failureStage = "READ_INTERFACE_IDENTITY"
	uiSID, err := currentProcessSID()
	if err != nil {
		return ErrWorkerServiceConfig
	}
	failureStage = "CONNECT_SCM"
	manager, err := mgr.Connect()
	if err != nil {
		return ErrWorkerServiceConfig
	}
	defer manager.Disconnect()

	failureStage = "OPEN_OR_CREATE_SERVICE"
	service, err := manager.OpenService(WorkerServiceName)
	created := false
	if err != nil {
		if !errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return ErrWorkerServiceConfig
		}
		// mgr.CreateService quotes exepath internally; pass the raw filesystem path.
		service, err = manager.CreateService(WorkerServiceName, workerPath, mgr.Config{
			ServiceType:      windows.SERVICE_WIN32_OWN_PROCESS,
			StartType:        mgr.StartManual,
			ErrorControl:     mgr.ErrorNormal,
			ServiceStartName: workerServiceAccount,
			DisplayName:      workerServiceDisplayName,
			Description:      "Executes approved Majucau background synchronizations.",
			SidType:          windows.SERVICE_SID_TYPE_RESTRICTED,
		})
		if err != nil {
			return ErrWorkerServiceConfig
		}
		created = true
	}
	defer func() { returnErr = errors.Join(returnErr, service.Close()) }()

	failureStage = "VALIDATE_EXISTING_SERVICE"
	previousConfig, err := service.Config()
	if err != nil || !validExistingWorkerService(previousConfig, installDir) {
		if created {
			_ = service.Delete()
		}
		return ErrWorkerServiceConfig
	}
	failureStage = "QUERY_SERVICE_STATE"
	previousStatus, err := service.Query()
	if err != nil {
		if created {
			_ = service.Delete()
		}
		return ErrWorkerServiceState
	}
	wasRunning := previousStatus.State != svc.Stopped

	failureStage = "RESOLVE_SERVICE_SID"
	serviceSID, _, _, err := windows.LookupSID("", workerServiceAccount)
	if err != nil || serviceSID == nil {
		if created {
			_ = service.Delete()
		}
		return ErrWorkerServiceConfig
	}
	serviceSIDText := serviceSID.String()
	secretRoot := filepath.Join(dataDir, workerSecretsDirName)
	logRoot := filepath.Join(dataDir, workerLogsDirName)
	failureStage = "PREPARE_SECRET_DIRECTORY"
	if err := ensureServicePrivateDirectory(secretRoot, serviceSIDText); err != nil {
		if created {
			_ = service.Delete()
		}
		return ErrWorkerServiceConfig
	}
	failureStage = "PREPARE_LOG_DIRECTORY"
	if err := ensureServicePrivateDirectory(logRoot, serviceSIDText); err != nil {
		if created {
			_ = service.Delete()
		}
		return ErrWorkerServiceConfig
	}
	failureStage = "PREPARE_PARAMETERS_ACL"
	if err := ensureWorkerServiceParametersACL(serviceSIDText); err != nil {
		if created {
			_ = service.Delete()
		}
		return ErrWorkerServiceConfig
	}

	failureStage = "READ_PREVIOUS_UI_SID"
	previousUISID, hadPreviousUISID, err := readWorkerServiceUISID()
	if err != nil {
		if created {
			_ = service.Delete()
		}
		return ErrWorkerServiceConfig
	}
	previousConfigCopy := previousConfig
	failureStage = "READ_RECOVERY_CONFIG"
	previousRecoveryActions, err := service.RecoveryActions()
	if err != nil {
		return ErrWorkerServiceConfig
	}
	previousRecoveryResetPeriod, err := service.ResetPeriod()
	if err != nil {
		return ErrWorkerServiceConfig
	}
	previousRecoveryOnNonCrash, err := service.RecoveryActionsOnNonCrashFailures()
	if err != nil {
		return ErrWorkerServiceConfig
	}
	previousRecoveryCommand, err := service.RecoveryCommand()
	if err != nil {
		return ErrWorkerServiceConfig
	}
	rollback := func() {
		if created {
			_ = stopWorkerService(service)
			_ = service.Delete()
			_ = writeWorkerServiceUISID(previousUISID, hadPreviousUISID)
			return
		}
		_ = stopWorkerService(service)
		_ = service.UpdateConfig(previousConfigCopy)
		_ = writeWorkerServiceUISID(previousUISID, hadPreviousUISID)
		if len(previousRecoveryActions) == 0 {
			_ = service.ResetRecoveryActions()
		} else {
			_ = service.SetRecoveryActions(previousRecoveryActions, previousRecoveryResetPeriod)
		}
		_ = service.SetRecoveryActionsOnNonCrashFailures(previousRecoveryOnNonCrash)
		_ = service.SetRecoveryCommand(previousRecoveryCommand)
		if wasRunning {
			_ = service.Start()
			_ = waitForWorkerServiceState(service, svc.Running)
		}
	}

	if wasRunning {
		failureStage = "STOP_EXISTING_SERVICE"
		if err := stopWorkerService(service); err != nil {
			if created {
				_ = service.Delete()
			}
			return ErrWorkerServiceState
		}
	}

	newConfig := mgr.Config{
		ServiceType:      windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:        mgr.StartAutomatic,
		ErrorControl:     mgr.ErrorNormal,
		BinaryPathName:   quoteWorkerServiceImage(workerPath),
		ServiceStartName: workerServiceAccount,
		DisplayName:      workerServiceDisplayName,
		Description:      "Executes approved Majucau background synchronizations.",
		SidType:          windows.SERVICE_SID_TYPE_RESTRICTED,
		DelayedAutoStart: true,
	}
	failureStage = "WRITE_UI_SID"
	if err := writeWorkerServiceUISID(uiSID, true); err != nil {
		rollback()
		return ErrWorkerServiceConfig
	}
	failureStage = "UPDATE_SERVICE_CONFIG"
	if err := service.UpdateConfig(newConfig); err != nil {
		rollback()
		return ErrWorkerServiceConfig
	}
	failureStage = "SET_RECOVERY_ACTIONS"
	if err := service.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: time.Minute},
		{Type: mgr.ServiceRestart, Delay: 5 * time.Minute},
		{Type: mgr.ServiceRestart, Delay: 15 * time.Minute},
	}, 24*60*60); err != nil {
		rollback()
		return ErrWorkerServiceConfig
	}
	failureStage = "SET_NONCRASH_RECOVERY"
	if err := service.SetRecoveryActionsOnNonCrashFailures(true); err != nil {
		rollback()
		return ErrWorkerServiceConfig
	}
	failureStage = "CLEAR_RECOVERY_COMMAND"
	if err := service.SetRecoveryCommand(""); err != nil {
		rollback()
		return ErrWorkerServiceConfig
	}
	failureStage = "START_SERVICE"
	if err := service.Start(); err != nil && !errors.Is(err, windows.ERROR_SERVICE_ALREADY_RUNNING) {
		rollback()
		return ErrWorkerServiceState
	}
	failureStage = "WAIT_FOR_RUNNING"
	if err := waitForWorkerServiceState(service, svc.Running); err != nil {
		rollback()
		return ErrWorkerServiceState
	}
	return nil
}

// RemoveWorkerService stops and removes only the Majucau service whose image
// is contained in installDir. Application data and protected secrets remain.
func RemoveWorkerService(installDir string) (returnErr error) {
	installDir, err := filepath.Abs(filepath.Clean(strings.TrimSpace(installDir)))
	if err != nil || installDir == "." {
		return ErrWorkerServiceConfig
	}
	programFiles := strings.TrimSpace(os.Getenv("ProgramFiles"))
	if programFiles == "" || !pathWithin(installDir, programFiles) || rejectReparsePath(installDir) != nil {
		return ErrWorkerServiceConfig
	}
	installLock, err := AcquireInstallLock(DefaultInstallLockName)
	if err != nil {
		return ErrInstallerLockFailed
	}
	defer func() { returnErr = errors.Join(returnErr, installLock.Release()) }()
	manager, err := mgr.Connect()
	if err != nil {
		return ErrWorkerServiceConfig
	}
	defer manager.Disconnect()
	service, err := manager.OpenService(WorkerServiceName)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return nil
	}
	if err != nil {
		return ErrWorkerServiceConfig
	}
	config, err := service.Config()
	if err != nil || config.ServiceStartName != workerServiceAccount || config.SidType != windows.SERVICE_SID_TYPE_RESTRICTED || !workerServiceImageWithin(config.BinaryPathName, installDir) {
		_ = service.Close()
		return ErrWorkerServiceConfig
	}
	if err := stopWorkerService(service); err != nil {
		_ = service.Close()
		return ErrWorkerServiceState
	}
	if err := service.Delete(); err != nil {
		_ = service.Close()
		return ErrWorkerServiceConfig
	}
	if err := service.Close(); err != nil {
		return ErrWorkerServiceConfig
	}
	deadline := time.Now().Add(workerServiceTimeout)
	for time.Now().Before(deadline) {
		probe, openErr := manager.OpenService(WorkerServiceName)
		if errors.Is(openErr, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return nil
		}
		if openErr == nil {
			_ = probe.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}
	return ErrWorkerServiceState
}

func validateWorkerServicePaths(options WorkerServiceOptions) (string, string, string, error) {
	installDir, err := filepath.Abs(filepath.Clean(strings.TrimSpace(options.InstallDir)))
	if err != nil || installDir == "." {
		return "", "", "", ErrWorkerServiceConfig
	}
	workerPath, err := filepath.Abs(filepath.Clean(strings.TrimSpace(options.WorkerPath)))
	if err != nil || !workerExecutableWithin(workerPath, installDir) {
		return "", "", "", ErrWorkerServiceConfig
	}
	programFiles := strings.TrimSpace(os.Getenv("ProgramFiles"))
	if programFiles == "" || !pathWithin(installDir, programFiles) {
		return "", "", "", ErrWorkerServiceConfig
	}
	if err := rejectReparsePath(installDir); err != nil {
		return "", "", "", ErrWorkerServiceConfig
	}
	if err := rejectReparsePath(workerPath); err != nil {
		return "", "", "", ErrWorkerServiceConfig
	}
	info, err := os.Lstat(workerPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return "", "", "", ErrWorkerServiceConfig
	}
	dataDir, err := filepath.Abs(filepath.Clean(strings.TrimSpace(options.DataDir)))
	if err != nil || dataDir == "." {
		return "", "", "", ErrWorkerServiceConfig
	}
	programData := strings.TrimSpace(os.Getenv("ProgramData"))
	if programData == "" || !strings.EqualFold(dataDir, filepath.Join(programData, "Majucau")) {
		return "", "", "", ErrWorkerServiceConfig
	}
	if err := rejectReparsePath(dataDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", "", "", ErrWorkerServiceConfig
	}
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return "", "", "", ErrWorkerServiceConfig
	}
	return installDir, workerPath, dataDir, nil
}

func validExistingWorkerService(config mgr.Config, installDir string) bool {
	return config.ServiceStartName == workerServiceAccount &&
		config.SidType == windows.SERVICE_SID_TYPE_RESTRICTED &&
		workerServiceImageWithin(config.BinaryPathName, installDir)
}

func serviceImageWithin(binaryPath, installDir string) bool {
	image := strings.TrimSpace(binaryPath)
	if strings.HasPrefix(image, `"`) {
		end := strings.Index(image[1:], `"`)
		if end < 0 || end+2 != len(image) {
			return false
		}
		image = image[1 : end+1]
	} else if strings.ContainsAny(image, " \t") {
		// A service image path containing spaces must be quoted. Reject
		// ambiguous unquoted paths and any appended command-line arguments.
		return false
	}
	absolute, err := filepath.Abs(filepath.Clean(image))
	return err == nil && pathWithin(absolute, installDir)
}

func workerServiceImageWithin(binaryPath, installDir string) bool {
	if !serviceImageWithin(binaryPath, installDir) {
		return false
	}
	image := strings.TrimSpace(binaryPath)
	if strings.HasPrefix(image, `"`) {
		image = image[1 : len(image)-1]
	}
	absolute, err := filepath.Abs(filepath.Clean(image))
	if err != nil || !strings.EqualFold(filepath.Base(absolute), "majucau-worker.exe") {
		return false
	}
	return pathWithin(filepath.Dir(absolute), filepath.Join(installDir, "worker"))
}

// workerExecutableWithin validates a filesystem path supplied to the helper.
// Unlike a Service Control Manager image command line, this path is not quoted;
// spaces in Program Files are valid and are quoted only when writing SCM config.
func workerExecutableWithin(workerPath, installDir string) bool {
	workerAbsolute, err := filepath.Abs(filepath.Clean(strings.TrimSpace(workerPath)))
	if err != nil {
		return false
	}
	installAbsolute, err := filepath.Abs(filepath.Clean(strings.TrimSpace(installDir)))
	if err != nil || !pathWithin(workerAbsolute, filepath.Join(installAbsolute, "worker")) {
		return false
	}
	return strings.EqualFold(filepath.Base(workerAbsolute), "majucau-worker.exe")
}

func quoteWorkerServiceImage(workerPath string) string {
	return `"` + workerPath + `"`
}

func pathWithin(path, root string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil || relative == "." || filepath.IsAbs(relative) {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func rejectReparsePath(path string) error {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil || filepath.VolumeName(absolute) == "" || strings.HasPrefix(absolute, `\\`) {
		return ErrWorkerServiceConfig
	}
	volume := filepath.VolumeName(absolute)
	current := volume + string(filepath.Separator)
	remainder := strings.TrimLeft(absolute[len(volume):], `/\`)
	for _, component := range strings.FieldsFunc(remainder, func(r rune) bool { return r == '/' || r == '\\' }) {
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, os.ErrNotExist) {
			return nil
		}
		if statErr != nil || info.Mode()&os.ModeSymlink != 0 {
			return ErrWorkerServiceConfig
		}
		pathPtr, pathErr := windows.UTF16PtrFromString(current)
		if pathErr != nil {
			return ErrWorkerServiceConfig
		}
		attributes, attrErr := windows.GetFileAttributes(pathPtr)
		if attrErr != nil || attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			return ErrWorkerServiceConfig
		}
	}
	return nil
}

func currentProcessSID() (string, error) {
	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		return "", err
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil || user == nil || user.User.Sid == nil {
		return "", errors.New("current process SID is unavailable")
	}
	return user.User.Sid.String(), nil
}

func ensureServicePrivateDirectory(path, serviceSID string) error {
	if !filepath.IsAbs(path) || strings.TrimSpace(serviceSID) == "" {
		return ErrWorkerServiceConfig
	}
	if err := rejectReparsePath(filepath.Dir(path)); err != nil {
		return ErrWorkerServiceConfig
	}
	info, err := os.Lstat(path)
	if err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || rejectReparsePath(path) != nil {
			return ErrWorkerServiceConfig
		}
		return validateServicePrivateDirectory(path, serviceSID)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return ErrWorkerServiceConfig
	}
	if err := createServicePrivateDirectory(path, serviceSID); err != nil {
		return ErrWorkerServiceConfig
	}
	return validateServicePrivateDirectory(path, serviceSID)
}

func createServicePrivateDirectory(path, serviceSID string) error {
	if _, err := windows.StringToSid(serviceSID); err != nil {
		return err
	}
	sddl := fmt.Sprintf("D:P(A;OICI;FA;;;%s)(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)", serviceSID)
	security, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}
	attributes := &windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: security,
	}
	pathPointer, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	err = windows.CreateDirectory(pathPointer, attributes)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return nil
	}
	return err
}

func validateServicePrivateDirectory(path, serviceSID string) error {
	security, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil || security == nil {
		return ErrWorkerServiceConfig
	}
	control, _, err := security.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return ErrWorkerServiceConfig
	}
	dacl, _, err := security.DACL()
	if err != nil || dacl == nil || dacl.AceCount != 3 {
		return ErrWorkerServiceConfig
	}
	allowed := map[string]bool{serviceSID: false, "S-1-5-18": false, "S-1-5-32-544": false}
	wantFlags := uint8(windows.OBJECT_INHERIT_ACE | windows.CONTAINER_INHERIT_ACE)
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil || ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return ErrWorkerServiceConfig
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart)).String()
		seen, ok := allowed[sid]
		if !ok || seen || ace.Mask != windows.ACCESS_MASK(0x001F01FF) || ace.Header.AceFlags != wantFlags {
			return ErrWorkerServiceConfig
		}
		allowed[sid] = true
	}
	for _, found := range allowed {
		if !found {
			return ErrWorkerServiceConfig
		}
	}
	return nil
}

func readWorkerServiceUISID() (string, bool, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\`+WorkerServiceName+`\Parameters`, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	defer key.Close()
	value, _, err := key.GetStringValue("UIUserSID")
	if errors.Is(err, registry.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if _, err := windows.StringToSid(value); err != nil {
		return "", false, ErrWorkerServiceConfig
	}
	return value, true, nil
}

func writeWorkerServiceUISID(value string, present bool) error {
	keyPath := `SYSTEM\CurrentControlSet\Services\` + WorkerServiceName + `\Parameters`
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer key.Close()
	if !present {
		err := key.DeleteValue("UIUserSID")
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return err
	}
	if _, err := windows.StringToSid(value); err != nil {
		return ErrWorkerServiceConfig
	}
	return key.SetStringValue("UIUserSID", value)
}

func ensureWorkerServiceParametersACL(serviceSID string) error {
	if _, err := windows.StringToSid(serviceSID); err != nil {
		return ErrWorkerServiceConfig
	}
	keyPath := `SYSTEM\CurrentControlSet\Services\` + WorkerServiceName + `\Parameters`
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE|registry.SET_VALUE|windows.READ_CONTROL|windows.WRITE_DAC)
	if err != nil {
		return err
	}
	defer key.Close()

	sddl := fmt.Sprintf("D:P(A;;KR;;;%s)(A;;KA;;;SY)(A;;KA;;;BA)", serviceSID)
	security, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}
	dacl, _, err := security.DACL()
	if err != nil || dacl == nil {
		return ErrWorkerServiceConfig
	}
	if err := windows.SetSecurityInfo(windows.Handle(key), windows.SE_REGISTRY_KEY, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil); err != nil {
		return err
	}
	actual, err := windows.GetSecurityInfo(windows.Handle(key), windows.SE_REGISTRY_KEY, windows.DACL_SECURITY_INFORMATION)
	if err != nil || actual == nil {
		return ErrWorkerServiceConfig
	}
	control, _, err := actual.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return ErrWorkerServiceConfig
	}
	actualDACL, _, err := actual.DACL()
	if err != nil || actualDACL == nil || actualDACL.AceCount != 3 {
		return ErrWorkerServiceConfig
	}
	allowed := map[string]windows.ACCESS_MASK{
		serviceSID:     0x00020019, // KEY_READ
		"S-1-5-18":     0x000F003F, // KEY_ALL_ACCESS
		"S-1-5-32-544": 0x000F003F, // KEY_ALL_ACCESS
	}
	for index := uint32(0); index < uint32(actualDACL.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(actualDACL, index, &ace); err != nil || ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != 0 {
			return ErrWorkerServiceConfig
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart)).String()
		mask, ok := allowed[sid]
		if !ok || ace.Mask != mask {
			return ErrWorkerServiceConfig
		}
		delete(allowed, sid)
	}
	if len(allowed) != 0 {
		return ErrWorkerServiceConfig
	}
	return nil
}

func stopWorkerService(service *mgr.Service) error {
	status, err := service.Query()
	if err != nil {
		return err
	}
	if status.State == svc.Stopped {
		return nil
	}
	_, err = service.Control(svc.Stop)
	if err != nil && !errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) && !errors.Is(err, windows.ERROR_SERVICE_CANNOT_ACCEPT_CTRL) {
		return err
	}
	return waitForWorkerServiceState(service, svc.Stopped)
}

func waitForWorkerServiceState(service *mgr.Service, state svc.State) error {
	deadline := time.Now().Add(workerServiceTimeout)
	for time.Now().Before(deadline) {
		status, err := service.Query()
		if err != nil {
			return err
		}
		if status.State == state {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return ErrWorkerServiceState
}
