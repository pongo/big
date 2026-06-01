package scan

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var (
	ErrPathNotFound = errors.New("scan root does not exist")
	ErrNotDirectory = errors.New("scan root is not a directory")
)

type EntryKind int

const (
	EntryFile EntryKind = iota
	EntryFolder
	EntryOther
)

type RootEntry struct {
	Name       string
	Path       string
	Kind       EntryKind
	HasSize    bool
	Size       int64
	LinkTarget string
}

type SizeRankingGroup int

const (
	SizeRankingLarge SizeRankingGroup = iota
	SizeRankingSmall
	SizeRankingNoSize
)

const sizeRankingDisplayThreshold int64 = 1024 * 1024

func SizeRanking(entry RootEntry) SizeRankingGroup {
	switch {
	case !entry.HasSize:
		return SizeRankingNoSize
	case entry.Size >= sizeRankingDisplayThreshold:
		return SizeRankingLarge
	default:
		return SizeRankingSmall
	}
}

func ShowsSize(entry RootEntry) bool {
	return SizeRanking(entry) == SizeRankingLarge
}

type FS interface {
	Lstat(name string) (fs.FileInfo, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	Readlink(name string) (string, error)
}

type Scanner struct {
	fsys       FS
	MinAgeDays int
	Now        func() time.Time
}

func NewScanner(fsys FS) *Scanner {
	if fsys == nil {
		fsys = osFS{}
	}
	return &Scanner{fsys: fsys}
}

func (s *Scanner) ScanRoot(root string) ([]RootEntry, error) {
	root = filepath.Clean(root)

	info, err := s.fsys.Lstat(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrPathNotFound, root)
		}
		return nil, fmt.Errorf("failed to read scan root %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrNotDirectory, root)
	}

	entries, err := s.fsys.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to list scan root %q: %w", root, err)
	}

	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	boundary := now().AddDate(0, 0, -s.MinAgeDays)

	out := make([]RootEntry, 0, len(entries))
	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())
		entryInfo, infoErr := s.fsys.Lstat(fullPath)
		if infoErr != nil {
			// Entry vanished or became inaccessible between list and stat.
			continue
		}

		rootEntry := RootEntry{
			Name: entry.Name(),
			Path: fullPath,
		}

		ageTime := entryInfo.ModTime()
		switch {
		case entryInfo.Mode()&fs.ModeSymlink != 0:
			rootEntry.Kind = EntryOther
			target, targetErr := s.fsys.Readlink(fullPath)
			if targetErr == nil {
				rootEntry.LinkTarget = target
			}
		case entryInfo.Mode().IsRegular():
			rootEntry.Kind = EntryFile
			rootEntry.HasSize = true
			rootEntry.Size = entryInfo.Size()
		case entryInfo.IsDir():
			rootEntry.Kind = EntryFolder
			if created, ok := s.creationTime(fullPath, entryInfo); ok {
				ageTime = created
			}
			if !s.includesAge(ageTime, boundary) {
				continue
			}
			rootEntry.HasSize = true
			rootEntry.Size = s.dirSize(fullPath)
		default:
			// Placeholder for issue 02 behavior.
			rootEntry.Kind = EntryOther
		}

		if rootEntry.Kind != EntryFolder && !s.includesAge(ageTime, boundary) {
			continue
		}

		out = append(out, rootEntry)
	}

	SortRootEntries(out)
	return out, nil
}

func (s *Scanner) includesAge(ageTime time.Time, boundary time.Time) bool {
	return s.MinAgeDays == 0 || !ageTime.After(boundary)
}

func (s *Scanner) creationTime(path string, info fs.FileInfo) (time.Time, bool) {
	fsys, ok := s.fsys.(interface {
		CreationTime(string, fs.FileInfo) (time.Time, bool)
	})
	if !ok {
		return time.Time{}, false
	}
	return fsys.CreationTime(path, info)
}

func (s *Scanner) dirSize(path string) int64 {
	children, err := s.fsys.ReadDir(path)
	if err != nil {
		// Unreadable folder contents are ignored for now.
		return 0
	}

	var total int64
	for _, child := range children {
		childPath := filepath.Join(path, child.Name())
		childInfo, infoErr := s.fsys.Lstat(childPath)
		if infoErr != nil {
			continue
		}

		switch {
		case childInfo.Mode().IsRegular():
			total += childInfo.Size()
		case childInfo.IsDir():
			total += s.dirSize(childPath)
		}
	}

	return total
}

func SortRootEntries(entries []RootEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		left := entries[i]
		right := entries[j]

		leftGroup := SizeRanking(left)
		rightGroup := SizeRanking(right)
		if leftGroup != rightGroup {
			return leftGroup < rightGroup
		}

		if leftGroup == SizeRankingLarge && left.Size != right.Size {
			return left.Size > right.Size
		}

		return left.Name < right.Name
	})
}

type osFS struct{}

func (osFS) Lstat(name string) (fs.FileInfo, error) {
	return os.Lstat(name)
}

func (osFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (osFS) Readlink(name string) (string, error) {
	return os.Readlink(name)
}
