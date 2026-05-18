//go:build !windows

package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
)

type cpuPlatformState struct {
	mu       sync.Mutex
	prevIdle uint64
	prevAll  uint64
}

func getSystemRAM() (totalMB, usedPercent float64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	var total, available uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			total = val
		case "MemAvailable:":
			available = val
		}
	}
	if total == 0 {
		return 0, 0
	}
	totalMB = float64(total) / 1024
	used := total - available
	usedPercent = float64(used) / float64(total) * 100.0
	return
}

func (a *App) getCPUPercent() float64 {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			break
		}
		vals := make([]uint64, len(fields)-1)
		for i, s := range fields[1:] {
			vals[i], _ = strconv.ParseUint(s, 10, 64)
		}
		idle := vals[3]
		var all uint64
		for _, v := range vals {
			all += v
		}

		a.cpuPlatState.mu.Lock()
		prevIdle := a.cpuPlatState.prevIdle
		prevAll := a.cpuPlatState.prevAll
		a.cpuPlatState.prevIdle = idle
		a.cpuPlatState.prevAll = all
		a.cpuPlatState.mu.Unlock()

		if prevAll == 0 {
			return 0
		}
		deltaAll := all - prevAll
		deltaIdle := idle - prevIdle
		if deltaAll == 0 {
			return 0
		}
		return float64(deltaAll-deltaIdle) / float64(deltaAll) * 100.0
	}
	return 0
}
