//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// isProcessElevated returns true if the current process has administrator privileges.
func isProcessElevated() bool {
	token := windows.GetCurrentProcessToken()
	elevated := token.IsElevated()
	return elevated
}

// startCoreProcess launches a new instance of this executable as the core server.
// If requireElevated is true and the process is not elevated, it uses UAC to elevate.
func startCoreProcess(execPath string, requireElevated bool) error {
	if requireElevated && !isProcessElevated() {
		verb, _ := syscall.UTF16PtrFromString("runas")
		file, _ := syscall.UTF16PtrFromString(execPath)
		args, _ := syscall.UTF16PtrFromString("--core")
		err := windows.ShellExecute(0, verb, file, args, nil, windows.SW_HIDE)
		if err != nil {
			return fmt.Errorf("ShellExecute runas failed: %w", err)
		}
		return nil
	}

	cmd := exec.Command(execPath, "--core")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

// recoverBrokenSingleInstance attempts to clean up a broken single-instance lock
// left from a crashed previous run.
func recoverBrokenSingleInstance(uniqueID string) {
	// On Windows, Wails v3 single-instance uses a named mutex.
	// If a previous instance crashed, the mutex may still be held.
	// We attempt to open and release it.
	name, _ := syscall.UTF16PtrFromString("Global\\" + uniqueID)
	h, err := windows.OpenMutex(windows.SYNCHRONIZE|windows.MUTEX_MODIFY_STATE, false, name)
	if err != nil {
		return
	}
	_, _ = windows.ReleaseMutex(h)
	_ = windows.CloseHandle(h)
}

// getExecPath returns the path of the current executable for single-instance detection.
func getExecPath() string {
	p, _ := os.Executable()
	return p
}
