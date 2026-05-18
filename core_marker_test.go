package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteCoreMarker_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	writeCoreMarker(dir, "test_marker", "hello world")

	path := filepath.Join(dir, "data", "markers", "test_marker.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("marker file not created: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Errorf("marker file missing detail: %q", string(data))
	}
}

func TestWriteCoreMarker_AppendsLines(t *testing.T) {
	dir := t.TempDir()
	writeCoreMarker(dir, "append_test", "line one")
	writeCoreMarker(dir, "append_test", "line two")

	path := filepath.Join(dir, "data", "markers", "append_test.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("marker file not found: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "line one") {
		t.Error("missing 'line one'")
	}
	if !strings.Contains(content, "line two") {
		t.Error("missing 'line two'")
	}
}

func TestWriteCoreMarker_EmptyDirNoOp(t *testing.T) {
	// Must not panic with empty dir
	writeCoreMarker("", "noop", "detail")
}

func TestMarkerDetail_Formatting(t *testing.T) {
	tests := []struct {
		format string
		args   []any
		want   string
	}{
		{"hello %s", []any{"world"}, "hello world"},
		{"pid=%d", []any{1234}, "pid=1234"},
		{"no args", nil, "no args"},
	}
	for _, tc := range tests {
		got := markerDetail(tc.format, tc.args...)
		if got != tc.want {
			t.Errorf("markerDetail(%q, %v) = %q, want %q", tc.format, tc.args, got, tc.want)
		}
	}
}

func TestWriteCoreMarker_ContainsTimestamp(t *testing.T) {
	dir := t.TempDir()
	writeCoreMarker(dir, "ts_test", "check-timestamp")

	path := filepath.Join(dir, "data", "markers", "ts_test.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("marker file not found: %v", err)
	}
	// Timestamp format: "2006-01-02 15:04:05.000"
	content := string(data)
	if len(content) < 23 {
		t.Errorf("marker line too short to contain timestamp: %q", content)
	}
	// Year should start the line
	if content[0] != '2' {
		t.Errorf("expected line to start with year '2', got %q", content[:4])
	}
}
