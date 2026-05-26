package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestDeleteTrashesSelectedRootEntry(t *testing.T) {
	path := filepath.Join("root", "delete-me.txt")
	wantPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs returned error: %v", err)
	}
	model := NewModel("root", []scan.RootEntry{
		{Name: "keep.txt", Path: filepath.Join("root", "keep.txt"), Kind: scan.EntryFile, HasSize: true},
		{Name: "delete-me.txt", Path: path, Kind: scan.EntryFile, HasSize: true},
	})
	model.selected = 1

	var trashed string
	model.trashPath = func(path string) error {
		trashed = path
		return nil
	}

	_, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDelete}))
	if cmd == nil {
		t.Fatal("Update returned nil command")
	}
	gotMsg := cmd()
	msg, ok := gotMsg.(pathActionFinishedMsg)
	if !ok {
		t.Fatalf("command returned %T, want pathActionFinishedMsg", gotMsg)
	}

	if trashed != wantPath {
		t.Fatalf("trashed path = %q, want %q", trashed, wantPath)
	}
	if msg.verb != "Delete" || msg.path != wantPath || msg.err != nil {
		t.Fatalf("path action message = %#v, want successful Delete for %q", msg, wantPath)
	}
}

func TestDeleteSuccessMarksRowTrashedAndMovesSelection(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "first.txt", Path: filepath.Join("root", "first.txt"), Kind: scan.EntryFile, HasSize: true},
		{Name: "second.txt", Path: filepath.Join("root", "second.txt"), Kind: scan.EntryFile, HasSize: true},
	})
	model.selected = 0

	path := model.selectedEntryPath()
	updated, _ := model.Update(pathActionFinishedMsg{verb: "Delete", path: path})
	got := updated.(Model)

	if !got.isTrashed(path) {
		t.Fatalf("path %q is not marked as trashed", path)
	}
	if got.selected != 1 {
		t.Fatalf("selected index = %d, want %d", got.selected, 1)
	}
}

func TestDeleteSuccessOnLastRowKeepsSelection(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "first.txt", Path: filepath.Join("root", "first.txt"), Kind: scan.EntryFile, HasSize: true},
		{Name: "second.txt", Path: filepath.Join("root", "second.txt"), Kind: scan.EntryFile, HasSize: true},
	})
	model.selected = 1

	path := model.selectedEntryPath()
	updated, _ := model.Update(pathActionFinishedMsg{verb: "Delete", path: path})
	got := updated.(Model)

	if !got.isTrashed(path) {
		t.Fatalf("path %q is not marked as trashed", path)
	}
	if got.selected != 1 {
		t.Fatalf("selected index = %d, want %d", got.selected, 1)
	}
}

func TestDeleteFailureKeepsSelectionAndUntrashed(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "first.txt", Path: filepath.Join("root", "first.txt"), Kind: scan.EntryFile, HasSize: true},
		{Name: "second.txt", Path: filepath.Join("root", "second.txt"), Kind: scan.EntryFile, HasSize: true},
	})
	model.selected = 0
	path := model.selectedEntryPath()

	updated, _ := model.Update(pathActionFinishedMsg{verb: "Delete", path: path, err: errors.New("trash unavailable")})
	got := updated.(Model)

	if got.isTrashed(path) {
		t.Fatalf("path %q should not be marked as trashed", path)
	}
	if got.selected != 0 {
		t.Fatalf("selected index = %d, want %d", got.selected, 0)
	}
	if got.status != "Delete failed: trash unavailable" {
		t.Fatalf("status = %q, want %q", got.status, "Delete failed: trash unavailable")
	}
}

func TestDeleteFailureRollsBackOptimisticTrashedState(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "first.txt", Path: filepath.Join("root", "first.txt"), Kind: scan.EntryFile, HasSize: true},
	})
	model.trashPath = func(path string) error {
		return errors.New("trash unavailable")
	}

	updated, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDelete}))
	got := updated.(Model)
	path := got.selectedEntryPath()
	if !got.isTrashed(path) {
		t.Fatalf("path %q is not optimistically marked as trashed", path)
	}

	updated, _ = got.Update(cmd())
	got = updated.(Model)
	if got.isTrashed(path) {
		t.Fatalf("path %q should not stay trashed after delete failure", path)
	}
	if got.status != "Delete failed: trash unavailable" {
		t.Fatalf("status = %q, want %q", got.status, "Delete failed: trash unavailable")
	}
}

func TestDeleteOnAlreadyTrashedRowDoesNotInvokeAction(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "first.txt", Path: filepath.Join("root", "first.txt"), Kind: scan.EntryFile, HasSize: true},
	})
	path := model.selectedEntryPath()
	model.trashedPaths[path] = struct{}{}
	called := false
	model.trashPath = func(path string) error {
		called = true
		return nil
	}

	_, cmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDelete}))
	if cmd != nil {
		t.Fatal("Update returned non-nil command for already trashed row")
	}
	if called {
		t.Fatal("trash action was called for already trashed row")
	}
}

func TestRapidDeleteOnSameRowDoesNotInvokeActionTwice(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "first.txt", Path: filepath.Join("root", "first.txt"), Kind: scan.EntryFile, HasSize: true},
		{Name: "second.txt", Path: filepath.Join("root", "second.txt"), Kind: scan.EntryFile, HasSize: true},
	})
	model.selected = 0
	calls := 0
	model.trashPath = func(path string) error {
		calls++
		return nil
	}

	updated, firstCmd := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDelete}))
	if firstCmd == nil {
		t.Fatal("first Update returned nil command")
	}

	_, secondCmd := updated.(Model).Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDelete}))
	if secondCmd != nil {
		t.Fatal("second Update returned non-nil command while delete was in flight")
	}

	firstCmd()
	if calls != 1 {
		t.Fatalf("trash action calls = %d, want 1", calls)
	}
}

func TestDeleteHelpIncludedInFullHelp(t *testing.T) {
	found := false
	for _, section := range defaultKeyMap().FullHelp() {
		for _, binding := range section {
			help := binding.Help()
			if help.Key == "del" && help.Desc == "trash" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("full help does not include delete trash binding")
	}
}

func TestOpenUsesSelectedEntryFromActiveEntryView(t *testing.T) {
	firstPath := filepath.Join("root", "first.txt")
	secondPath := filepath.Join("root", "second.txt")
	wantPath, err := filepath.Abs(secondPath)
	if err != nil {
		t.Fatalf("filepath.Abs returned error: %v", err)
	}

	model := NewModel("root", []scan.RootEntry{
		{Name: "first.txt", Path: firstPath, Kind: scan.EntryFile, HasSize: true},
		{Name: "second.txt", Path: secondPath, Kind: scan.EntryFile, HasSize: true},
	})
	model.entryViews = []entryView{
		{
			name: "Filtered",
			entries: []scan.RootEntry{
				{Name: "second.txt", Path: secondPath, Kind: scan.EntryFile, HasSize: true},
			},
		},
	}
	model.selectedEntryView = 0
	model.selected = 0

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
	if msg.path != wantPath {
		t.Fatalf("path action message path = %q, want %q", msg.path, wantPath)
	}
}

func TestBuildEntryViewsGroupsAndOrdersViews(t *testing.T) {
	entries := []scan.RootEntry{
		{Name: "z-folder", Path: "z-folder", Kind: scan.EntryFolder, HasSize: true, Size: 300},
		{Name: "a.JPG", Path: "a.JPG", Kind: scan.EntryFile, HasSize: true, Size: 200},
		{Name: "x.bin", Path: "x.bin", Kind: scan.EntryFile, HasSize: true, Size: 190},
		{Name: "b.jpg", Path: "b.jpg", Kind: scan.EntryFile, HasSize: true, Size: 180},
		{Name: ".env", Path: ".env", Kind: scan.EntryFile, HasSize: true, Size: 170},
		{Name: "c.Jpg", Path: "c.Jpg", Kind: scan.EntryFile, HasSize: true, Size: 160},
		{Name: "d.TXT", Path: "d.TXT", Kind: scan.EntryFile, HasSize: true, Size: 150},
		{Name: "e.txt", Path: "e.txt", Kind: scan.EntryFile, HasSize: true, Size: 140},
		{Name: "f.TxT", Path: "f.TxT", Kind: scan.EntryFile, HasSize: true, Size: 130},
		{Name: "link", Path: "link", Kind: scan.EntryOther},
		{Name: "g.md", Path: "g.md", Kind: scan.EntryFile, HasSize: true, Size: 120},
		{Name: "h.tar.gz", Path: "h.tar.gz", Kind: scan.EntryFile, HasSize: true, Size: 110},
	}

	views := buildEntryViews(entries)
	gotNames := make([]string, 0, len(views))
	for _, view := range views {
		gotNames = append(gotNames, view.name)
	}
	wantNames := []string{"Folders", "Other"}
	if !equalStrings(gotNames, wantNames) {
		t.Fatalf("view names = %#v, want %#v", gotNames, wantNames)
	}

	if got := namesFromEntries(views[0].entries); !equalStrings(got, []string{"z-folder"}) {
		t.Fatalf("Folders entries = %#v, want %#v", got, []string{"z-folder"})
	}
	if got := namesFromEntries(views[1].entries); !equalStrings(got, []string{"a.JPG", "x.bin", "b.jpg", ".env", "c.Jpg", "d.TXT", "e.txt", "f.TxT", "link", "g.md", "h.tar.gz"}) {
		t.Fatalf("Other entries = %#v, want %#v", got, []string{"a.JPG", "x.bin", "b.jpg", ".env", "c.Jpg", "d.TXT", "e.txt", "f.TxT", "link", "g.md", "h.tar.gz"})
	}
}

func TestBuildEntryViewsOmitsEmptyViews(t *testing.T) {
	entries := []scan.RootEntry{
		{Name: "a.txt", Path: "a.txt", Kind: scan.EntryFile, HasSize: true, Size: 10},
		{Name: "b.txt", Path: "b.txt", Kind: scan.EntryFile, HasSize: true, Size: 9},
	}

	views := buildEntryViews(entries)
	if len(views) != 1 {
		t.Fatalf("views len = %d, want %d", len(views), 1)
	}
	if views[0].name != "Other" {
		t.Fatalf("single view name = %q, want %q", views[0].name, "Other")
	}
}

func TestBuildEntryViewsOrdersExtensionTiesByName(t *testing.T) {
	entries := []scan.RootEntry{
		{Name: "a.zzz", Path: "a.zzz", Kind: scan.EntryFile, HasSize: true, Size: 30},
		{Name: "a.aaa", Path: "a.aaa", Kind: scan.EntryFile, HasSize: true, Size: 29},
		{Name: "b.zzz", Path: "b.zzz", Kind: scan.EntryFile, HasSize: true, Size: 28},
		{Name: "b.aaa", Path: "b.aaa", Kind: scan.EntryFile, HasSize: true, Size: 27},
		{Name: "c.zzz", Path: "c.zzz", Kind: scan.EntryFile, HasSize: true, Size: 26},
		{Name: "c.aaa", Path: "c.aaa", Kind: scan.EntryFile, HasSize: true, Size: 25},
		{Name: "d.zzz", Path: "d.zzz", Kind: scan.EntryFile, HasSize: true, Size: 24},
		{Name: "d.aaa", Path: "d.aaa", Kind: scan.EntryFile, HasSize: true, Size: 23},
		{Name: "e.zzz", Path: "e.zzz", Kind: scan.EntryFile, HasSize: true, Size: 22},
		{Name: "e.aaa", Path: "e.aaa", Kind: scan.EntryFile, HasSize: true, Size: 21},
	}

	views := buildEntryViews(entries)
	gotNames := make([]string, 0, len(views))
	for _, view := range views {
		gotNames = append(gotNames, view.name)
	}
	wantNames := []string{".aaa", ".zzz"}
	if !equalStrings(gotNames, wantNames) {
		t.Fatalf("view names = %#v, want %#v", gotNames, wantNames)
	}
}

func TestLeftRightSwitchEntryViewsWithClampAndReset(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "folder", Path: "folder", Kind: scan.EntryFolder, HasSize: true, Size: 300},
		{Name: "a.txt", Path: "a.txt", Kind: scan.EntryFile, HasSize: true, Size: 200},
		{Name: "b.txt", Path: "b.txt", Kind: scan.EntryFile, HasSize: true, Size: 190},
		{Name: "c.txt", Path: "c.txt", Kind: scan.EntryFile, HasSize: true, Size: 180},
		{Name: "odd.bin", Path: "odd.bin", Kind: scan.EntryFile, HasSize: true, Size: 170},
	})
	model.status = "Open failed: test"
	model.selectedEntryView = 1
	model.selected = 2
	model.viewport.SetYOffset(3)

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft}))
	got := updated.(Model)
	if got.selectedEntryView != 0 {
		t.Fatalf("selected entry view = %d, want %d", got.selectedEntryView, 0)
	}
	if got.selected != 0 {
		t.Fatalf("selected row = %d, want %d", got.selected, 0)
	}
	if got.viewport.YOffset() != 0 {
		t.Fatalf("viewport y offset = %d, want %d", got.viewport.YOffset(), 0)
	}
	if got.status != "" {
		t.Fatalf("status = %q, want empty status", got.status)
	}

	updated, _ = got.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft}))
	got = updated.(Model)
	if got.selectedEntryView != 0 {
		t.Fatalf("selected entry view after clamp-left = %d, want %d", got.selectedEntryView, 0)
	}

	updated, _ = got.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))
	got = updated.(Model)
	if got.selectedEntryView != 1 {
		t.Fatalf("selected entry view after right = %d, want %d", got.selectedEntryView, 1)
	}
}

func TestHeaderRendersActiveEntryViewNameRightAligned(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "a.txt", Path: "a.txt", Kind: scan.EntryFile, HasSize: true, Size: 10},
		{Name: "b.txt", Path: "b.txt", Kind: scan.EntryFile, HasSize: true, Size: 9},
		{Name: "c.txt", Path: "c.txt", Kind: scan.EntryFile, HasSize: true, Size: 8},
	})
	model.width = 30

	header := model.renderHeaderContent()
	if !strings.HasPrefix(header, "root") {
		t.Fatalf("header %q does not start with %q", header, "root")
	}
	if !strings.HasSuffix(header, "Other") {
		t.Fatalf("header %q does not end with %q", header, "Other")
	}

	model.entryViews = append([]entryView{{name: "Folders", entries: nil}}, model.entryViews...)
	model.selectedEntryView = 1
	header = model.renderHeaderContent()
	if !strings.HasPrefix(header, "root") {
		t.Fatalf("header %q does not start with %q", header, "root")
	}
	if !strings.HasSuffix(header, "Other") {
		t.Fatalf("header %q does not end with %q", header, "Other")
	}
}

func TestEmptyRootHasNoActiveEntryViewName(t *testing.T) {
	model := NewModel("root", nil)
	if got := model.activeEntryViewName(); got != "" {
		t.Fatalf("active entry view name = %q, want empty", got)
	}
}

func TestDeleteInNonFirstEntryViewMarksTrashedInThatView(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "folder", Path: "folder", Kind: scan.EntryFolder, HasSize: true, Size: 300},
		{Name: "a.txt", Path: "a.txt", Kind: scan.EntryFile, HasSize: true, Size: 200},
		{Name: "b.txt", Path: "b.txt", Kind: scan.EntryFile, HasSize: true, Size: 190},
		{Name: "c.txt", Path: "c.txt", Kind: scan.EntryFile, HasSize: true, Size: 180},
	})
	model.selectedEntryView = 1
	model.selected = 0

	path := model.selectedEntryPath()
	updated, _ := model.Update(pathActionFinishedMsg{verb: "Delete", path: path})
	got := updated.(Model)

	if !got.isTrashed(path) {
		t.Fatalf("path %q is not marked as trashed", path)
	}
	if got.entryViews[1].entries[0].Path != "a.txt" {
		t.Fatalf("trashed entry disappeared from its view, first path = %q", got.entryViews[1].entries[0].Path)
	}
}

func TestViewportHidesSizeForEntriesBelowOneMiB(t *testing.T) {
	model := NewModel("root", []scan.RootEntry{
		{Name: "small.txt", Path: "small.txt", Kind: scan.EntryFile, HasSize: true, Size: 900 * 1024},
		{Name: "large.txt", Path: "large.txt", Kind: scan.EntryFile, HasSize: true, Size: 2 * 1024 * 1024},
	})
	model.width = 80
	model.height = 10
	model.resize()
	model.refreshViewportContent()

	view := model.viewport.View()
	if strings.Contains(view, "900 KB") {
		t.Fatalf("view should hide size below 1 MiB, got %q", view)
	}
	if !strings.Contains(view, "2 MB") {
		t.Fatalf("view should render size at or above 1 MiB, got %q", view)
	}
}

func namesFromEntries(entries []scan.RootEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name)
	}
	return names
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if left[idx] != right[idx] {
			return false
		}
	}
	return true
}
