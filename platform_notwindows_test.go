//go:build !windows

package main

import (
	"os"
	"testing"
)

func TestIsProcessElevated_ReturnsBool(t *testing.T) {
	// On most CI and dev machines this is false; just ensure it doesn't panic
	elevated := isProcessElevated()
	if elevated && os.Getuid() != 0 {
		t.Error("isProcessElevated() returned true but os.Getuid() != 0")
	}
	if !elevated && os.Getuid() == 0 {
		t.Error("isProcessElevated() returned false but os.Getuid() == 0 (running as root)")
	}
}

func TestStartCoreProcess_BadPath(t *testing.T) {
	err := startCoreProcess("/nonexistent/binary/path/novaproxy", false)
	if err == nil {
		t.Error("expected error for non-existent binary")
	}
}

func TestRecoverBrokenSingleInstance_NoOp(t *testing.T) {
	// Should not panic with any input
	recoverBrokenSingleInstance("")
	recoverBrokenSingleInstance("com.novaproxy.desktop")
	recoverBrokenSingleInstance("some-unique-id-12345")
}
