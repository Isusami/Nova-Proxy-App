//go:build windows

package proxy

import (
	"net"
	"strings"
	"syscall"
)

func isRSTError(err error) bool {
	if err == nil {
		return false
	}
	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(*syscall.Errno); ok {
			return *syscallErr == syscall.ECONNRESET || *syscallErr == syscall.WSAECONNRESET
		}
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection reset") || strings.Contains(errStr, "remote error: tls: internal error")
}
