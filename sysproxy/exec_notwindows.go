//go:build !windows

package sysproxy

import "os/exec"

func hideWindow(cmd *exec.Cmd) {
	// No-op on Linux: there are no hidden windows concept.
}

func startHiddenCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}
