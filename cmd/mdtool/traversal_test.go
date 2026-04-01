package cmd

import (
	"path/filepath"
	"testing"
)

func TestPathCleanup(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"../../etc/passwd", "../../etc/passwd"}, // filepath.Clean doesn't prevent traversal alone
		{"test/../etc/passwd", "etc/passwd"},
		{"/abs/path/../../etc/passwd", "/etc/passwd"},
		{"./subdir/file.txt", "subdir/file.txt"},
	}

	for _, tt := range tests {
		got := filepath.Clean(tt.input)
		if got != tt.expected {
			t.Errorf("filepath.Clean(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
