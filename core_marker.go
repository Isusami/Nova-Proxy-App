package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// writeCoreMarker writes a debug marker file to help diagnose startup/shutdown issues.
func writeCoreMarker(dir, name, detail string) {
	if dir == "" {
		return
	}
	markerDir := filepath.Join(dir, "data", "markers")
	_ = os.MkdirAll(markerDir, 0755)
	path := filepath.Join(markerDir, name+".txt")
	content := fmt.Sprintf("%s %s\n", time.Now().Format("2006-01-02 15:04:05.000"), detail)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(content)
}

// markerDetail formats a detail string for writeCoreMarker.
func markerDetail(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
