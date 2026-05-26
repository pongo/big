package tui

import (
	"fmt"
	"strings"

	"big/internal/scan"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	rootPath    string
	entries     []scan.RootEntry
	sizeWidth   int
	headerStyle lipgloss.Style
	sizeStyle   lipgloss.Style
	nameStyle   lipgloss.Style
}

func NewModel(rootPath string, entries []scan.RootEntry) Model {
	sizeWidth := len("Size")
	for _, entry := range entries {
		if !entry.HasSize {
			continue
		}
		entrySizeWidth := len(scan.FormatSize(entry.Size))
		if entrySizeWidth > sizeWidth {
			sizeWidth = entrySizeWidth
		}
	}

	return Model{
		rootPath:  rootPath,
		entries:   entries,
		sizeWidth: sizeWidth,
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")),
		sizeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")),
		nameStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	var b strings.Builder
	b.WriteString(m.headerStyle.Render(fmt.Sprintf("Big - %s", m.rootPath)))
	b.WriteString("\n\n")

	if len(m.entries) == 0 {
		b.WriteString("No entries\n")
		view := tea.NewView(b.String())
		view.AltScreen = true
		return view
	}

	for _, entry := range m.entries {
		sizeCell := strings.Repeat(" ", m.sizeWidth)
		if entry.HasSize {
			sizeCell = fmt.Sprintf("%*s", m.sizeWidth, scan.FormatSize(entry.Size))
		}
		name := entry.Name
		if entry.LinkTarget != "" {
			name = fmt.Sprintf("%s -> %s", entry.Name, entry.LinkTarget)
		}
		row := fmt.Sprintf("%s  %s", m.sizeStyle.Render(sizeCell), m.nameStyle.Render(name))
		b.WriteString(row)
		b.WriteString("\n")
	}

	view := tea.NewView(b.String())
	view.AltScreen = true
	return view
}
