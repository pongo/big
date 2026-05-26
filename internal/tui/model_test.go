package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanRootHeaderUsesCurrentDirectoryNameForDot(t *testing.T) {
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "big-tui-root-*")
	if err != nil {
		t.Fatalf("MkdirTemp returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("failed to remove temp directory: %v", err)
		}
	})

	root := filepath.Join(tempDir, "project")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	if got := scanRootHeader("."); got != "project" {
		t.Fatalf("scanRootHeader(\".\") = %q, want %q", got, "project")
	}
}

func TestScanRootHeaderUsesPathBaseForExplicitPath(t *testing.T) {
	root := filepath.Join("parent", "project")

	if got := scanRootHeader(root); got != "project" {
		t.Fatalf("scanRootHeader(%q) = %q, want %q", root, got, "project")
	}
}
