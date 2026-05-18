//go:build !windows

package main

import (
	"os"
	"os/exec"
)

// isProcessElevated returns true if the current process is running as root.
func isProcessElevated() bool {
	return os.Getuid() == 0
}

// startCoreProcess launches a new instance of this executable as the core server.
// If requireElevated is true and the process is not root, it uses pkexec/sudo.
func startCoreProcess(execPath string, requireElevated bool) error {
	if requireElevated && !isProcessElevated() {
		cmd := exec.Command("pkexec", execPath, "--core")
		return cmd.Start()
	}
	cmd := exec.Command(execPath, "--core")
	return cmd.Start()
}

// recoverBrokenSingleInstance is a no-op on non-Windows platforms.
// The Wails v3 single-instance mechanism on Linux does not leave stale locks
// that need manual cleanup.
func recoverBrokenSingleInstance(uniqueID string) {}
