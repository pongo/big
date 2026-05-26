//go:build windows

package fs

import (
	"path/filepath"
	"testing"
)

func TestExplorerSelectArgumentQuotesCleanPath(t *testing.T) {
	path := filepath.Join(`C:\scan root`, `.lostpixel 2026-03-14.7z`)

	got := explorerSelectArgument(path)
	want := `/select,"` + filepath.Clean(path) + `"`
	if got != want {
		t.Fatalf("explorerSelectArgument(%q) = %q, want %q", path, got, want)
	}
}

func TestExplorerRevealCommandLineHandlesUnicodeAndSpaces(t *testing.T) {
	t.Setenv("WINDIR", `C:\Windows`)
	path := filepath.Join(`C:\scan root`, `Утиный Тест-2.ogg`)

	got := explorerRevealCommandLine(path)
	want := `"C:\Windows\explorer.exe" /select,"` + filepath.Clean(path) + `"`
	if got != want {
		t.Fatalf("explorerRevealCommandLine(%q) = %q, want %q", path, got, want)
	}
}
