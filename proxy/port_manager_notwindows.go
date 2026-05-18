//go:build !windows

package proxy

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// FindProcessByPort returns the PID listening on the specified port using /proc/net/tcp.
func FindProcessByPort(port int) (int, error) {
	f, err := os.Open("/proc/net/tcp")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	targetHex := fmt.Sprintf("%04X", port)
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		// local_address is field 1: hex "IP:PORT"
		addrPort := strings.Split(fields[1], ":")
		if len(addrPort) != 2 {
			continue
		}
		if strings.EqualFold(addrPort[1], targetHex) {
			inode, _ := strconv.ParseUint(fields[9], 10, 64)
			pid := findPIDByInode(inode)
			if pid > 0 {
				return pid, nil
			}
		}
	}
	return 0, nil
}

func findPIDByInode(inode uint64) int {
	if inode == 0 {
		return 0
	}
	target := fmt.Sprintf("socket:[%d]", inode)
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		fdDir := fmt.Sprintf("/proc/%d/fd", pid)
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(fmt.Sprintf("%s/%s", fdDir, fd.Name()))
			if err == nil && link == target {
				return pid
			}
		}
	}
	return 0
}

// GetProcessNameByPID returns the process name for a given PID via /proc.
func GetProcessNameByPID(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// KillProcessByPID sends SIGKILL to the specified PID.
func KillProcessByPID(pid int) error {
	return exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
}

// EnsurePortAvailable checks if the port is available; if occupied by selfNames, kills the process.
func EnsurePortAvailable(startPort int, selfNames []string) (int, error) {
	currentPort := startPort
	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		pid, err := FindProcessByPort(currentPort)
		if err != nil || pid == 0 {
			ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", currentPort))
			if err == nil {
				ln.Close()
				return currentPort, nil
			}
		} else {
			name, _ := GetProcessNameByPID(pid)
			isSelf := false
			for _, self := range selfNames {
				if strings.EqualFold(name, self) {
					isSelf = true
					break
				}
			}

			if isSelf {
				if err := KillProcessByPID(pid); err == nil {
					return currentPort, nil
				}
			}
		}

		currentPort++
	}

	return startPort, fmt.Errorf("could not find available port after %d attempts", maxAttempts)
}

