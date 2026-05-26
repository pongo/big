package tui

import (
	"path/filepath"
	"sort"
	"strings"

	"big/internal/scan"
)

type entryView struct {
	name    string
	entries []scan.RootEntry
}

type entryViewSet struct {
	views    []entryView
	active   int
	selected int
}

func newEntryViewSet(entries []scan.RootEntry) entryViewSet {
	return entryViewSet{views: buildEntryViews(entries)}
}

func (s entryViewSet) activeName() string {
	if len(s.views) == 0 {
		return ""
	}
	return s.views[s.active].name
}

func (s entryViewSet) activeEntries() []scan.RootEntry {
	if len(s.views) == 0 {
		return nil
	}
	return s.views[s.active].entries
}

func (s entryViewSet) selectedIndex() int {
	return s.selected
}

func (s entryViewSet) selectedEntry() (scan.RootEntry, bool) {
	entries := s.activeEntries()
	if len(entries) == 0 {
		return scan.RootEntry{}, false
	}
	return entries[s.selected], true
}

func (s *entryViewSet) moveSelection(delta int) {
	entries := s.activeEntries()
	if len(entries) == 0 {
		return
	}
	next := s.selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(entries) {
		next = len(entries) - 1
	}
	s.selected = next
}

func (s *entryViewSet) selectFirst() {
	s.selected = 0
}

func (s *entryViewSet) selectLast() {
	entries := s.activeEntries()
	if len(entries) == 0 {
		s.selected = 0
		return
	}
	s.selected = len(entries) - 1
}

func (s *entryViewSet) advanceSelection() {
	if s.selected+1 < len(s.activeEntries()) {
		s.selected++
	}
}

func (s *entryViewSet) switchView(delta int) bool {
	if len(s.views) == 0 {
		return false
	}
	next := s.active + delta
	if next < 0 {
		next = 0
	}
	if next >= len(s.views) {
		next = len(s.views) - 1
	}
	if next == s.active {
		return false
	}
	s.active = next
	s.selected = 0
	return true
}

func buildEntryViews(entries []scan.RootEntry) []entryView {
	if len(entries) == 0 {
		return nil
	}

	extensionCounts := make(map[string]int)
	for _, entry := range entries {
		if entry.Kind != scan.EntryFile {
			continue
		}
		ext := normalizedExtension(entry.Name)
		if ext == "" {
			continue
		}
		extensionCounts[ext]++
	}

	folders := make([]scan.RootEntry, 0)
	other := make([]scan.RootEntry, 0)
	extensionEntries := make(map[string][]scan.RootEntry)
	for _, entry := range entries {
		switch entry.Kind {
		case scan.EntryFolder:
			folders = append(folders, entry)
		case scan.EntryFile:
			ext := normalizedExtension(entry.Name)
			if ext == "" || extensionCounts[ext] < 5 {
				other = append(other, entry)
				continue
			}
			extensionEntries[ext] = append(extensionEntries[ext], entry)
		default:
			other = append(other, entry)
		}
	}

	views := make([]entryView, 0, len(extensionEntries)+2)
	if len(folders) > 0 {
		views = append(views, entryView{name: "Folders", entries: folders})
	}

	extensionNames := make([]string, 0, len(extensionEntries))
	for name := range extensionEntries {
		extensionNames = append(extensionNames, name)
	}
	sort.Slice(extensionNames, func(i, j int) bool {
		left := extensionNames[i]
		right := extensionNames[j]
		leftCount := extensionCounts[left]
		rightCount := extensionCounts[right]
		if leftCount != rightCount {
			return leftCount > rightCount
		}
		return left < right
	})
	for _, name := range extensionNames {
		views = append(views, entryView{name: name, entries: extensionEntries[name]})
	}

	if len(other) > 0 {
		views = append(views, entryView{name: "Other", entries: other})
	}

	return views
}

func normalizedExtension(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return ""
	}
	return strings.ToLower(ext)
}
