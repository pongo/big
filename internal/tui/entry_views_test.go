package tui

import (
	"testing"

	"big/internal/scan"
)

func TestEntryViewSetSwitchesActiveViewAndResetsSelection(t *testing.T) {
	views := entryViewSet{
		views: []entryView{
			{
				name: "Folders",
				entries: []scan.RootEntry{
					{Name: "folder", Kind: scan.EntryFolder},
				},
			},
			{
				name: "Other",
				entries: []scan.RootEntry{
					{Name: "first.txt", Kind: scan.EntryFile},
					{Name: "second.txt", Kind: scan.EntryFile},
				},
			},
		},
		selected: 1,
	}

	if !views.switchView(1) {
		t.Fatal("switchView returned false")
	}

	if got := views.activeName(); got != "Other" {
		t.Fatalf("active name = %q, want %q", got, "Other")
	}
	if got := views.selectedIndex(); got != 0 {
		t.Fatalf("selected index = %d, want %d", got, 0)
	}
}

func TestEntryViewSetClampsSelectionMovement(t *testing.T) {
	views := entryViewSet{
		views: []entryView{
			{
				name: "Other",
				entries: []scan.RootEntry{
					{Name: "first.txt", Kind: scan.EntryFile},
					{Name: "second.txt", Kind: scan.EntryFile},
				},
			},
		},
	}

	views.moveSelection(10)
	if got := views.selectedIndex(); got != 1 {
		t.Fatalf("selected index after moving down = %d, want %d", got, 1)
	}

	views.moveSelection(-10)
	if got := views.selectedIndex(); got != 0 {
		t.Fatalf("selected index after moving up = %d, want %d", got, 0)
	}
}

func TestEntryViewSetReturnsSelectedRootEntryFromActiveView(t *testing.T) {
	views := entryViewSet{
		views: []entryView{
			{
				name: "Other",
				entries: []scan.RootEntry{
					{Name: "first.txt", Kind: scan.EntryFile},
					{Name: "second.txt", Kind: scan.EntryFile},
				},
			},
		},
		selected: 1,
	}

	entry, ok := views.selectedEntry()
	if !ok {
		t.Fatal("selectedEntry returned false")
	}
	if entry.Name != "second.txt" {
		t.Fatalf("selected root entry = %q, want %q", entry.Name, "second.txt")
	}
}
