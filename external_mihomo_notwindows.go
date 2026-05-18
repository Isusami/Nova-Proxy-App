//go:build !windows

package main

import (
	"fmt"

	"novaproxy/proxy"
)

// externalMihomoManager is a stub for non-Windows platforms.
// TUN mode via external mihomo is only supported on Windows.
type externalMihomoManager struct{}

func newExternalMihomoManager() *externalMihomoManager {
	return &externalMihomoManager{}
}

func (m *externalMihomoManager) Start(cfg proxy.TUNConfig, listenPort string, logf func(string)) error {
	return fmt.Errorf("external mihomo TUN is only supported on Windows")
}

func (m *externalMihomoManager) Stop(logf func(string)) error {
	return nil
}

func (m *externalMihomoManager) RestartIfRunning(cfg proxy.TUNConfig, listenPort string, logf func(string)) error {
	return nil
}

func (m *externalMihomoManager) Status(cfg proxy.TUNConfig) proxy.TUNStatus {
	return proxy.TUNStatus{
		Supported: false,
		Running:   false,
		Message:   "TUN mode is not supported on this platform",
	}
}
