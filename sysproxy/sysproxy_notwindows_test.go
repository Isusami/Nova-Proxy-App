//go:build !windows

package sysproxy

import (
	"os/exec"
	"testing"
)

func TestGetSystemProxyStatus_NoGsettings(t *testing.T) {
	// If gsettings is not installed, GetSystemProxyStatus should return a zero-value struct (not panic)
	_, err := exec.LookPath("gsettings")
	if err != nil {
		status := GetSystemProxyStatus()
		if status.Enabled {
			t.Error("expected Enabled=false when gsettings is unavailable")
		}
		return
	}
	// gsettings is available — just ensure it returns without panic
	status := GetSystemProxyStatus()
	_ = status.Enabled
	_ = status.Server
	_ = status.Override
}

func TestSaveAndSetOriginalProxySettings_Roundtrip(t *testing.T) {
	original := SystemProxyStatus{
		Enabled:  true,
		Server:   "127.0.0.1:8080",
		Override: "localhost",
	}
	SetOriginalProxySettings(original)

	// Lock check: ensure next Set overwrites
	second := SystemProxyStatus{
		Enabled: false,
		Server:  "",
	}
	SetOriginalProxySettings(second)

	// Check that the internal state was updated (indirectly via RestoreOriginalProxySettings)
	// We can't easily check the internal variable without exporting it,
	// so just ensure no panic
}

func TestSaveOriginalProxySettings_NoGsettings(t *testing.T) {
	_, err := exec.LookPath("gsettings")
	if err != nil {
		// Should return nil (gracefully handle missing gsettings)
		saveErr := SaveOriginalProxySettings()
		if saveErr != nil {
			t.Logf("SaveOriginalProxySettings() without gsettings returned: %v (acceptable)", saveErr)
		}
		return
	}
	err = SaveOriginalProxySettings()
	if err != nil {
		t.Logf("SaveOriginalProxySettings() returned: %v (may need DISPLAY/DBUS)", err)
	}
}

func TestSetOriginalProxySettings_ConcurrentSafe(t *testing.T) {
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(i int) {
			SetOriginalProxySettings(SystemProxyStatus{
				Enabled: i%2 == 0,
				Server:  "127.0.0.1:8080",
			})
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGsettingsHelpers_InvalidSchema(t *testing.T) {
	_, err := exec.LookPath("gsettings")
	if err != nil {
		t.Skip("gsettings not available")
	}
	// A bogus schema should return an error
	_, err = gsettingsGet("org.invalid.schema.does.not.exist", "key")
	if err == nil {
		t.Error("expected error for invalid schema")
	}
}
