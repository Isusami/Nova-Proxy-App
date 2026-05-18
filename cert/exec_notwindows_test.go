//go:build !windows

package cert

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHideWindow_NoOp(t *testing.T) {
	// On non-Windows, hideWindow is a no-op; just verify it doesn't panic
	cmd := exec.Command("echo", "test")
	hideWindow(cmd)
}

func TestRunHiddenCommand_Echo(t *testing.T) {
	err := runHiddenCommand("echo", "hello")
	if err != nil {
		t.Errorf("runHiddenCommand(echo, hello): %v", err)
	}
}

func TestRunHiddenCommand_InvalidBinary(t *testing.T) {
	err := runHiddenCommand("/nonexistent/binary/xyz123")
	if err == nil {
		t.Error("expected error for non-existent binary")
	}
}

func TestOutputHiddenCommand_Echo(t *testing.T) {
	out, err := outputHiddenCommand("echo", "hello")
	if err != nil {
		t.Fatalf("outputHiddenCommand: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestOutputHiddenCommand_Invalid(t *testing.T) {
	_, err := outputHiddenCommand("/nonexistent/binary/xyz123")
	if err == nil {
		t.Error("expected error for non-existent binary")
	}
}

func TestStartHiddenCommand_Echo(t *testing.T) {
	err := startHiddenCommand("echo", "hello")
	if err != nil {
		t.Errorf("startHiddenCommand(echo): %v", err)
	}
}

func TestStartVisibleCommand_Echo(t *testing.T) {
	err := startVisibleCommand("echo", "hello")
	if err != nil {
		t.Errorf("startVisibleCommand(echo): %v", err)
	}
}

func TestRunElevatedCommand_NoPkexec(t *testing.T) {
	// If pkexec is not available, we expect an error (not a panic)
	_, err := exec.LookPath("pkexec")
	if err != nil {
		err := runElevatedCommand("echo", "test")
		if err == nil {
			t.Error("expected error when pkexec is unavailable")
		}
		return
	}
	// pkexec exists but requires a UI for authorization in most environments
	// so just check the function doesn't panic — it will fail with auth error
	t.Log("pkexec available, skipping actual elevation (would need UI auth)")
}

func TestRunHiddenCommand_TrueCommand(t *testing.T) {
	// 'true' is available on all Unix systems
	err := runHiddenCommand("true")
	if err != nil {
		t.Errorf("runHiddenCommand(true): %v", err)
	}
}

func TestRunHiddenCommand_FalseCommand(t *testing.T) {
	// 'false' always exits with code 1
	err := runHiddenCommand("false")
	if err == nil {
		t.Error("expected non-zero exit from 'false'")
	}
}

func TestStartHiddenCommand_WithTempScript(t *testing.T) {
	// Write a minimal shell script to a temp file and run it
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}
	if err := startHiddenCommand(script); err != nil {
		t.Errorf("startHiddenCommand(script): %v", err)
	}
}
