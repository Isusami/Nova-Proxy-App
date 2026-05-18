//go:build !windows

package cert

import "os/exec"

func hideWindow(cmd *exec.Cmd) {
	// No-op on Linux.
}

func runHiddenCommand(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

func outputHiddenCommand(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func startHiddenCommand(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

func startVisibleCommand(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

func runElevatedCommand(name string, args ...string) error {
	elevated := append([]string{name}, args...)
	return exec.Command("pkexec", elevated...).Run()
}
