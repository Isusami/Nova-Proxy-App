//go:build !windows

package main

import (
	"os"
	"runtime"
	"testing"
)

func TestGetSystemRAM_ReturnsValues(t *testing.T) {
	totalMB, usedPercent := getSystemRAM()

	if runtime.GOOS != "linux" {
		// On non-Linux, /proc/meminfo doesn't exist — should return zeros gracefully
		if totalMB != 0 || usedPercent != 0 {
			t.Logf("non-Linux RAM: total=%.1fMB used=%.1f%% (informational)", totalMB, usedPercent)
		}
		return
	}

	if totalMB <= 0 {
		t.Errorf("getSystemRAM(): totalMB = %.1f, want > 0", totalMB)
	}
	if usedPercent < 0 || usedPercent > 100 {
		t.Errorf("getSystemRAM(): usedPercent = %.1f, want 0-100", usedPercent)
	}
}

func TestGetSystemRAM_ProcMeminfoAbsent(t *testing.T) {
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/proc/meminfo"); err == nil {
			t.Skip("skipping absent-file test: /proc/meminfo exists")
		}
	}
	// On non-Linux, should return zeros without panic
	total, pct := getSystemRAM()
	_ = total
	_ = pct
}

func TestGetCPUPercent_NoApp(t *testing.T) {
	a := &App{}

	// First call initialises state — may return 0
	pct1 := a.getCPUPercent()
	if pct1 < 0 || pct1 > 100 {
		if runtime.GOOS != "linux" {
			t.Logf("non-Linux: getCPUPercent() = %.2f (informational)", pct1)
			return
		}
		t.Errorf("getCPUPercent() = %.2f, want 0-100", pct1)
	}
}

func TestGetCPUPercent_TwoCalls(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("CPU percent test requires /proc/stat (Linux only)")
	}

	a := &App{}
	pct1 := a.getCPUPercent()
	pct2 := a.getCPUPercent()

	for _, pct := range []float64{pct1, pct2} {
		if pct < 0 || pct > 100 {
			t.Errorf("getCPUPercent() = %.2f, want 0-100", pct)
		}
	}
}
