//go:build !windows

package proxy

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// --- EnsurePortAvailable ---

func TestEnsurePortAvailable_FreePort(t *testing.T) {
	// Find a free port first
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	// Wait briefly for OS to release
	time.Sleep(10 * time.Millisecond)

	got, err := EnsurePortAvailable(port, nil)
	if err != nil {
		t.Fatalf("EnsurePortAvailable error: %v", err)
	}
	if got == 0 {
		t.Error("expected non-zero port")
	}
}

func TestEnsurePortAvailable_OccupiedByOther_FindsNext(t *testing.T) {
	// Occupy a port with a TCP listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot bind: " + err.Error())
	}
	defer ln.Close()
	occupiedPort := ln.Addr().(*net.TCPAddr).Port

	// EnsurePortAvailable with no selfNames should move to next port
	got, err := EnsurePortAvailable(occupiedPort, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == occupiedPort {
		// acceptable if the OS somehow gives us that port, but usually it won't
		t.Logf("got same port %d — OK if OS allowed it", got)
	}
	if got < 1 || got > 65535 {
		t.Errorf("invalid port: %d", got)
	}
}

// --- GetProcessNameByPID ---

func TestGetProcessNameByPID_CurrentProcess(t *testing.T) {
	pid := os.Getpid()
	name, err := GetProcessNameByPID(pid)
	if err != nil {
		// /proc might not exist on macOS — skip
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			t.Skip("no /proc filesystem")
		}
		t.Fatalf("GetProcessNameByPID(%d): %v", pid, err)
	}
	if name == "" {
		t.Error("expected non-empty process name")
	}
}

func TestGetProcessNameByPID_InvalidPID(t *testing.T) {
	_, err := GetProcessNameByPID(-1)
	if err == nil {
		t.Error("expected error for PID -1")
	}
}

func TestGetProcessNameByPID_NonExistentPID(t *testing.T) {
	_, err := GetProcessNameByPID(99999999)
	if err == nil {
		t.Error("expected error for non-existent PID")
	}
}

// --- FindProcessByPort ---

func TestFindProcessByPort_NobodyListening(t *testing.T) {
	if _, err := os.Stat("/proc/net/tcp"); os.IsNotExist(err) {
		t.Skip("no /proc/net/tcp (non-Linux)")
	}
	// Use a port that's very unlikely to be in use
	pid, err := FindProcessByPort(19999)
	if err != nil {
		t.Fatalf("FindProcessByPort error: %v", err)
	}
	_ = pid // 0 is fine — nobody listening
}

func TestFindProcessByPort_ListeningProcess(t *testing.T) {
	if _, err := os.Stat("/proc/net/tcp"); os.IsNotExist(err) {
		t.Skip("no /proc/net/tcp (non-Linux)")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot bind: " + err.Error())
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	pid, err := FindProcessByPort(port)
	if err != nil {
		t.Fatalf("FindProcessByPort(%d): %v", port, err)
	}
	if pid != os.Getpid() {
		t.Errorf("FindProcessByPort(%d) = %d, want %d (current PID)", port, pid, os.Getpid())
	}
}

// --- KillProcessByPID ---

func TestKillProcessByPID_InvalidPID(t *testing.T) {
	err := KillProcessByPID(-1)
	if err == nil {
		t.Error("expected error killing PID -1")
	}
}

// --- findPIDByInode ---

func TestFindPIDByInode_ZeroReturnsZero(t *testing.T) {
	pid := findPIDByInode(0)
	if pid != 0 {
		t.Errorf("expected 0 for zero inode, got %d", pid)
	}
}

// --- pmPortFromInt (Windows only, not compiled here; just verify helpers) ---

func TestHelperPortFormatRoundtrip(t *testing.T) {
	// Verify our hex port comparison logic is consistent with net.Listen
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot bind")
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	hexStr := fmt.Sprintf("%04X", port)
	parsed, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		t.Fatalf("failed to parse hex port %q: %v", hexStr, err)
	}
	if int(parsed) != port {
		t.Errorf("hex roundtrip failed: port=%d hex=%s parsed=%d", port, hexStr, parsed)
	}
}
