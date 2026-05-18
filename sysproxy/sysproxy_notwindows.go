//go:build !windows

package sysproxy

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// SystemProxyStatus holds the current system proxy configuration.
type SystemProxyStatus struct {
	Enabled  bool
	Server   string
	Override string
}

var (
	originalProxySettings   *SystemProxyStatus
	originalProxySettingsMu sync.Mutex
)

// GetSystemProxyStatus reads the current system proxy settings via gsettings.
func GetSystemProxyStatus() SystemProxyStatus {
	status := SystemProxyStatus{}

	mode, err := gsettingsGet("org.gnome.system.proxy", "mode")
	if err != nil {
		return status
	}
	mode = strings.Trim(strings.TrimSpace(mode), "'")

	if mode != "manual" {
		return status
	}
	status.Enabled = true

	host, _ := gsettingsGet("org.gnome.system.proxy.http", "host")
	port, _ := gsettingsGet("org.gnome.system.proxy.http", "port")
	host = strings.Trim(strings.TrimSpace(host), "'")
	port = strings.TrimSpace(port)
	if host != "" && port != "" && port != "0" {
		status.Server = fmt.Sprintf("%s:%s", host, port)
	}

	ignoreHosts, _ := gsettingsGet("org.gnome.system.proxy", "ignore-hosts")
	status.Override = strings.TrimSpace(ignoreHosts)

	return status
}

// EnableSystemProxy enables the system HTTP proxy via gsettings.
func EnableSystemProxy(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("[sysproxy] invalid port: %d", port)
	}
	if err := gsettingsSet("org.gnome.system.proxy", "mode", "manual"); err != nil {
		return err
	}
	if err := gsettingsSet("org.gnome.system.proxy.http", "host", "127.0.0.1"); err != nil {
		return err
	}
	if err := gsettingsSet("org.gnome.system.proxy.http", "port", fmt.Sprintf("%d", port)); err != nil {
		return err
	}
	if err := gsettingsSet("org.gnome.system.proxy.https", "host", "127.0.0.1"); err != nil {
		return err
	}
	if err := gsettingsSet("org.gnome.system.proxy.https", "port", fmt.Sprintf("%d", port)); err != nil {
		return err
	}
	_ = gsettingsSet("org.gnome.system.proxy", "ignore-hosts", "['localhost', '127.0.0.0/8', '::1']")
	return nil
}

// DisableSystemProxy disables the system proxy via gsettings.
func DisableSystemProxy() error {
	return gsettingsSet("org.gnome.system.proxy", "mode", "none")
}

// SaveOriginalProxySettings saves the current proxy state to be restored later.
func SaveOriginalProxySettings() error {
	status := GetSystemProxyStatus()
	originalProxySettingsMu.Lock()
	originalProxySettings = &status
	originalProxySettingsMu.Unlock()
	return nil
}

// SetOriginalProxySettings manually sets the saved original proxy state.
func SetOriginalProxySettings(status SystemProxyStatus) {
	copy := status
	originalProxySettingsMu.Lock()
	originalProxySettings = &copy
	originalProxySettingsMu.Unlock()
}

// RestoreOriginalProxySettings restores proxy settings to their saved state.
func RestoreOriginalProxySettings() error {
	originalProxySettingsMu.Lock()
	settings := originalProxySettings
	originalProxySettingsMu.Unlock()

	if settings == nil || !settings.Enabled {
		return DisableSystemProxy()
	}
	return EnableSystemProxy(0)
}

// SetSystemProxyManual opens system proxy settings UI (best-effort).
func SetSystemProxyManual() error {
	return exec.Command("xdg-open", "gnome-control-center network").Start()
}

func gsettingsGet(schema, key string) (string, error) {
	out, err := exec.Command("gsettings", "get", schema, key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gsettingsSet(schema, key, value string) error {
	return exec.Command("gsettings", "set", schema, key, value).Run()
}
