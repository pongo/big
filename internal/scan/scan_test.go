package scan

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"time"
)

func TestSortRootEntriesBySizeThenName(t *testing.T) {
	entries := []RootEntry{
		{Name: "b", HasSize: true, Size: 10},
		{Name: "a", HasSize: true, Size: 10},
		{Name: "c", HasSize: true, Size: 20},
	}

	SortRootEntries(entries)

	got := []string{entries[0].Name, entries[1].Name, entries[2].Name}
	want := []string{"c", "a", "b"}
	if !equalStrings(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
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

type fakeFS struct {
	infos       map[string]fs.FileInfo
	dirChildren map[string][]string
	readDirErr  map[string]error
}

func newFakeFS() *fakeFS {
	return &fakeFS{
		infos:       map[string]fs.FileInfo{},
		dirChildren: map[string][]string{},
		readDirErr:  map[string]error{},
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
