package tui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"big/internal/scan"
	tea "charm.land/bubbletea/v2"
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

func TestEnterOpensSelectedRootEntry(t *testing.T) {
	path := filepath.Join("root", "Утиный Тест-2.ogg")
	wantPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs returned error: %v", err)
	}
	model := NewModel("root", []scan.RootEntry{
		{Name: "other.txt", Path: filepath.Join("root", "other.txt"), Kind: scan.EntryFile, HasSize: true},
		{Name: "Утиный Тест-2.ogg", Path: path, Kind: scan.EntryFile, HasSize: true},
	})
	model.selected = 1

	var opened string
	model.openPath = func(path string) error {
		opened = path
		return nil
	}

	_, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if cmd == nil {
		t.Fatal("Update returned nil command")
	}
	gotMsg := cmd()
	msg, ok := gotMsg.(pathActionFinishedMsg)
	if !ok {
		t.Fatalf("command returned %T, want pathActionFinishedMsg", gotMsg)
	}

	if opened != wantPath {
		t.Fatalf("opened path = %q, want %q", opened, wantPath)
	}
	if msg.verb != "Open" || msg.path != wantPath || msg.err != nil {
		t.Fatalf("path action message = %#v, want successful Open for %q", msg, wantPath)
	}
}

func TestERevealsSelectedRootEntry(t *testing.T) {
	path := filepath.Join("root", ".lostpixel 2026-03-14.7z")
	wantPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs returned error: %v", err)
	}
	model := NewModel("root", []scan.RootEntry{
		{Name: ".lostpixel 2026-03-14.7z", Path: path, Kind: scan.EntryFile, HasSize: true},
	})

	var revealed string
	model.revealPath = func(path string) error {
		revealed = path
		return nil
	}

	_, cmd := model.Update(tea.KeyPressMsg(tea.Key{Text: "e", Code: 'e'}))
	if cmd == nil {
		t.Fatal("Update returned nil command")
	}
	gotMsg := cmd()
	msg, ok := gotMsg.(pathActionFinishedMsg)
	if !ok {
		t.Fatalf("command returned %T, want pathActionFinishedMsg", gotMsg)
	}

	if revealed != wantPath {
		t.Fatalf("revealed path = %q, want %q", revealed, wantPath)
	}
	if msg.verb != "Reveal" || msg.path != wantPath || msg.err != nil {
		t.Fatalf("path action message = %#v, want successful Reveal for %q", msg, wantPath)
	}
}

func TestPathActionErrorSetsStatus(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "broken.txt", Path: filepath.Join("root", "broken.txt"), Kind: scan.EntryFile, HasSize: true},
	})
	errOpen := errors.New("no association")

	updated, _ := model.Update(pathActionFinishedMsg{verb: "Open", path: "broken.txt", err: errOpen})
	got := updated.(Model).status

	if got != "Open failed: no association" {
		t.Fatalf("status = %q, want %q", got, "Open failed: no association")
	}
}

func TestNavigationClearsPathActionStatus(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "first.txt", Path: filepath.Join("root", "first.txt"), Kind: scan.EntryFile, HasSize: true},
		{Name: "second.txt", Path: filepath.Join("root", "second.txt"), Kind: scan.EntryFile, HasSize: true},
	})
	model.status = "Open failed: no association"

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	got := updated.(Model).status

	if got != "" {
		t.Fatalf("status = %q, want empty status", got)
	}
}

func TestHelpIncludesOpenRevealBinding(t *testing.T) {
	help := defaultKeyMap().Open.Help()

	if help.Key != "enter/e" || help.Desc != "open/reveal" {
		t.Fatalf("open help = %#v, want enter/e open/reveal", help)
	}
}
