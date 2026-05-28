package scan

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestSortRootEntriesMixedThresholdBehavior(t *testing.T) {
	entries := []RootEntry{
		{Name: "small-z", HasSize: true, Size: 100},
		{Name: "large-c", HasSize: true, Size: 3 * 1024 * 1024},
		{Name: "small-a", HasSize: true, Size: 900 * 1024},
		{Name: "exact-1m", HasSize: true, Size: 1024 * 1024},
		{Name: "large-b", HasSize: true, Size: 5 * 1024 * 1024},
	}

	SortRootEntries(entries)

	got := []string{entries[0].Name, entries[1].Name, entries[2].Name, entries[3].Name, entries[4].Name}
	want := []string{"large-b", "large-c", "exact-1m", "small-a", "small-z"}
	if !equalStrings(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortRootEntriesBelowThresholdOrderedByName(t *testing.T) {
	entries := []RootEntry{
		{Name: "c", HasSize: true, Size: 900 * 1024},
		{Name: "a", HasSize: true, Size: 1023 * 1024},
		{Name: "b", HasSize: true, Size: 1},
	}

	SortRootEntries(entries)

	got := []string{entries[0].Name, entries[1].Name, entries[2].Name}
	want := []string{"a", "b", "c"}
	if !equalStrings(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortRootEntriesSameSizeLargeOrderedByName(t *testing.T) {
	entries := []RootEntry{
		{Name: "z", HasSize: true, Size: 2 * 1024 * 1024},
		{Name: "a", HasSize: true, Size: 2 * 1024 * 1024},
		{Name: "m", HasSize: true, Size: 2 * 1024 * 1024},
	}

	SortRootEntries(entries)

	got := []string{entries[0].Name, entries[1].Name, entries[2].Name}
	want := []string{"a", "m", "z"}
	if !equalStrings(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortRootEntriesNoSizeRemainLastAndOrderedByName(t *testing.T) {
	entries := []RootEntry{
		{Name: "link-z", HasSize: false},
		{Name: "big", HasSize: true, Size: 4 * 1024 * 1024},
		{Name: "link-a", HasSize: false},
		{Name: "small", HasSize: true, Size: 1},
	}

	SortRootEntries(entries)

	got := []string{entries[0].Name, entries[1].Name, entries[2].Name, entries[3].Name}
	want := []string{"big", "small", "link-a", "link-z"}
	if !equalStrings(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortRootEntriesFilesAndFoldersFollowSameThresholdRule(t *testing.T) {
	entries := []RootEntry{
		{Name: "folder-a", Kind: EntryFolder, HasSize: true, Size: 900 * 1024},
		{Name: "file-z", Kind: EntryFile, HasSize: true, Size: 500 * 1024},
		{Name: "folder-z", Kind: EntryFolder, HasSize: true, Size: 3 * 1024 * 1024},
		{Name: "file-a", Kind: EntryFile, HasSize: true, Size: 2 * 1024 * 1024},
	}

	SortRootEntries(entries)

	got := []string{entries[0].Name, entries[1].Name, entries[2].Name, entries[3].Name}
	want := []string{"folder-z", "file-a", "file-z", "folder-a"}
	if !equalStrings(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSizeRankingDrivesDisplayVisibility(t *testing.T) {
	cases := []struct {
		name  string
		entry RootEntry
		want  bool
	}{
		{
			name:  "below ranking threshold",
			entry: RootEntry{HasSize: true, Size: 900 * 1024},
			want:  false,
		},
		{
			name:  "at ranking threshold",
			entry: RootEntry{HasSize: true, Size: 1024 * 1024},
			want:  true,
		},
		{
			name:  "without size",
			entry: RootEntry{},
			want:  false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShowsSize(tt.entry); got != tt.want {
				t.Fatalf("ShowsSize() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestRecursiveFolderSizing(t *testing.T) {
	fsys := newFakeFS()
	fsys.addDir("root")
	fsys.addDir(filepath.Join("root", "folder"))
	fsys.addDir(filepath.Join("root", "folder", "nested"))
	fsys.addFile(filepath.Join("root", "folder", "a.txt"), 120)
	fsys.addFile(filepath.Join("root", "folder", "nested", "b.txt"), 30)

	scanner := NewScanner(fsys)
	entries, err := scanner.ScanRoot("root")
	if err != nil {
		t.Fatalf("ScanRoot returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 root entry, got %d", len(entries))
	}
	if entries[0].Name != "folder" {
		t.Fatalf("unexpected root entry name: %s", entries[0].Name)
	}
	if entries[0].Size != 150 {
		t.Fatalf("unexpected folder size: got %d want %d", entries[0].Size, 150)
	}
}

func TestHiddenRootFilesAreNotExcluded(t *testing.T) {
	fsys := newFakeFS()
	fsys.addDir("root")
	fsys.addFile(filepath.Join("root", ".hidden"), 1)
	fsys.addFile(filepath.Join("root", "visible"), 2)

	scanner := NewScanner(fsys)
	entries, err := scanner.ScanRoot("root")
	if err != nil {
		t.Fatalf("ScanRoot returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 root entries, got %d", len(entries))
	}

	names := []string{entries[0].Name, entries[1].Name}
	if !containsString(names, ".hidden") {
		t.Fatalf("hidden root file is missing, names: %v", names)
	}
}

func TestScanRootTrimsCommandLineQuotes(t *testing.T) {
	fsys := newFakeFS()
	root := "root with space"
	fsys.addDir(root)
	fsys.addFile(filepath.Join(root, "visible"), 1)

	cases := []string{
		`"root with space"`,
		`root with space"`,
	}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			scanner := NewScanner(fsys)
			entries, err := scanner.ScanRoot(input)
			if err != nil {
				t.Fatalf("ScanRoot returned error: %v", err)
			}
			if len(entries) != 1 || entries[0].Name != "visible" {
				t.Fatalf("entries = %#v, want visible entry", entries)
			}
		})
	}
}

func TestLinksAreAfterSizedEntriesAndSortedByName(t *testing.T) {
	fsys := newFakeFS()
	fsys.addDir("root")
	fsys.addFile(filepath.Join("root", "big.dat"), 100)
	fsys.addSymlink(filepath.Join("root", "z-link"), "C:\\target-z")
	fsys.addSymlink(filepath.Join("root", "a-link"), "C:\\target-a")

	scanner := NewScanner(fsys)
	entries, err := scanner.ScanRoot("root")
	if err != nil {
		t.Fatalf("ScanRoot returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 root entries, got %d", len(entries))
	}

	if entries[0].Name != "big.dat" || !entries[0].HasSize {
		t.Fatalf("expected first entry to be sized file, got %+v", entries[0])
	}
	if entries[1].Name != "a-link" || entries[1].HasSize {
		t.Fatalf("expected second entry to be no-size link a-link, got %+v", entries[1])
	}
	if entries[2].Name != "z-link" || entries[2].HasSize {
		t.Fatalf("expected third entry to be no-size link z-link, got %+v", entries[2])
	}
	if entries[1].LinkTarget != "C:\\target-a" {
		t.Fatalf("unexpected link target for a-link: %q", entries[1].LinkTarget)
	}
}

func TestUnreadableNestedContentsDoNotContributeSize(t *testing.T) {
	fsys := newFakeFS()
	fsys.addDir("root")
	fsys.addDir(filepath.Join("root", "folder"))
	fsys.addFile(filepath.Join("root", "folder", "ok.txt"), 10)
	fsys.addDir(filepath.Join("root", "folder", "blocked"))
	fsys.addFile(filepath.Join("root", "folder", "blocked", "nope.txt"), 1000)
	fsys.readDirErr[filepath.Join("root", "folder", "blocked")] = fs.ErrPermission

	scanner := NewScanner(fsys)
	entries, err := scanner.ScanRoot("root")
	if err != nil {
		t.Fatalf("ScanRoot returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 root entry, got %d", len(entries))
	}
	if entries[0].Size != 10 {
		t.Fatalf("unexpected folder size: got %d want %d", entries[0].Size, 10)
	}
}

func TestRealSymlinkShownWithoutSizeAndNotFollowed(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific symlink/junction behavior")
	}

	tempDir := t.TempDir()
	root := filepath.Join(tempDir, "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	targetFile := filepath.Join(root, "target.txt")
	if err := os.WriteFile(targetFile, []byte("0123456789"), 0o644); err != nil {
		t.Fatalf("write target file: %v", err)
	}

	linkPath := filepath.Join(root, "link.txt")
	if err := os.Symlink(targetFile, linkPath); err != nil {
		t.Skipf("os does not allow creating symlink in this environment: %v", err)
	}

	scanner := NewScanner(nil)
	entries, err := scanner.ScanRoot(root)
	if err != nil {
		t.Fatalf("ScanRoot returned error: %v", err)
	}

	var linkEntry *RootEntry
	for idx := range entries {
		if entries[idx].Name == "link.txt" {
			linkEntry = &entries[idx]
			break
		}
	}
	if linkEntry == nil {
		t.Fatalf("expected symlink root entry to be present")
	}
	if linkEntry.HasSize {
		t.Fatalf("symlink must not have size")
	}
	if linkEntry.LinkTarget == "" {
		t.Fatalf("symlink target should be rendered when readable")
	}
}

type fakeFS struct {
	infos       map[string]fs.FileInfo
	dirChildren map[string][]string
	readDirErr  map[string]error
	readlinkMap map[string]string
}

func newFakeFS() *fakeFS {
	return &fakeFS{
		infos:       map[string]fs.FileInfo{},
		dirChildren: map[string][]string{},
		readDirErr:  map[string]error{},
		readlinkMap: map[string]string{},
	}
}

func (f *fakeFS) addDir(path string) {
	f.infos[path] = fakeFileInfo{name: filepath.Base(path), mode: fs.ModeDir}
	if _, ok := f.dirChildren[path]; !ok {
		f.dirChildren[path] = []string{}
	}

	parent := filepath.Dir(path)
	if parent != path {
		if _, ok := f.dirChildren[parent]; ok {
			f.dirChildren[parent] = append(f.dirChildren[parent], filepath.Base(path))
		}
	}
}

func (f *fakeFS) addFile(path string, size int64) {
	f.infos[path] = fakeFileInfo{name: filepath.Base(path), size: size}
	parent := filepath.Dir(path)
	f.dirChildren[parent] = append(f.dirChildren[parent], filepath.Base(path))
}

func (f *fakeFS) addSymlink(path string, target string) {
	f.infos[path] = fakeFileInfo{name: filepath.Base(path), mode: fs.ModeSymlink}
	f.readlinkMap[path] = target
	parent := filepath.Dir(path)
	f.dirChildren[parent] = append(f.dirChildren[parent], filepath.Base(path))
}

func (f *fakeFS) Lstat(name string) (fs.FileInfo, error) {
	info, ok := f.infos[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return info, nil
}

func (f *fakeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if err, ok := f.readDirErr[name]; ok {
		return nil, err
	}
	children, ok := f.dirChildren[name]
	if !ok {
		return nil, errors.New("not a directory")
	}

	out := make([]fs.DirEntry, 0, len(children))
	for _, child := range children {
		fullPath := filepath.Join(name, child)
		info := f.infos[fullPath]
		out = append(out, fakeDirEntry{info: info})
	}
	return out, nil
}

func (f *fakeFS) Readlink(name string) (string, error) {
	target, ok := f.readlinkMap[name]
	if !ok {
		return "", fs.ErrNotExist
	}
	return target, nil
}

type fakeFileInfo struct {
	name string
	size int64
	mode fs.FileMode
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return f.size }
func (f fakeFileInfo) Mode() fs.FileMode  { return f.mode }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f fakeFileInfo) Sys() interface{}   { return nil }

type fakeDirEntry struct {
	info fs.FileInfo
}

func (f fakeDirEntry) Name() string               { return f.info.Name() }
func (f fakeDirEntry) IsDir() bool                { return f.info.IsDir() }
func (f fakeDirEntry) Type() fs.FileMode          { return f.info.Mode().Type() }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return f.info, nil }

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

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
