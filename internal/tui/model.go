package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pathfs "big/internal/fs"
	"big/internal/scan"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	PageUp key.Binding
	PageDn key.Binding
	Home   key.Binding
	End    key.Binding
	Open   key.Binding
	Reveal key.Binding
	Delete key.Binding
	Quit   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑/↓", "navigate")),
		Down:   key.NewBinding(key.WithKeys("down")),
		PageUp: key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup/pgdn", "page scroll")),
		PageDn: key.NewBinding(key.WithKeys("pgdown")),
		Home:   key.NewBinding(key.WithKeys("home"), key.WithHelp("home/end", "first/last")),
		End:    key.NewBinding(key.WithKeys("end")),
		Open:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter/e", "open/reveal")),
		Reveal: key.NewBinding(key.WithKeys("e")),
		Delete: key.NewBinding(key.WithKeys("delete"), key.WithHelp("del", "trash")),
		Quit:   key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q/esc", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.PageUp, k.Home, k.Open, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.PageUp, k.Home},
		{k.Open, k.Delete, k.Quit},
	}
}

type pathAction func(string) error

type pathActionFinishedMsg struct {
	verb string
	path string
	err  error
}

type entryView struct {
	name    string
	entries []scan.RootEntry
}

type Model struct {
	rootPath string
	header   string

	entryViews        []entryView
	selectedEntryView int

	sizeWidth int
	selected  int
	width     int
	height    int

	status string

	viewport viewport.Model
	help     help.Model
	keys     keyMap

	openPath   pathAction
	revealPath pathAction
	trashPath  pathAction

	trashedPaths map[string]struct{}

	headerStyle  lipgloss.Style
	sizeStyle    lipgloss.Style
	fileStyle    lipgloss.Style
	folderStyle  lipgloss.Style
	linkStyle    lipgloss.Style
	trashedStyle lipgloss.Style
	selectedName lipgloss.Style
	statusStyle  lipgloss.Style
	footerStyle  lipgloss.Style
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

	vp := viewport.New()
	keys := defaultKeyMap()

	model := Model{
		rootPath:   rootPath,
		header:     scanRootHeader(rootPath),
		entryViews: buildCompatibilityEntryViews(entries),

		sizeWidth: sizeWidth,
		selected:  0,

		viewport: vp,
		help:     help.New(),
		keys:     keys,

		openPath:   pathfs.OpenPath,
		revealPath: pathfs.RevealPath,
		trashPath:  pathfs.TrashPath,

		trashedPaths: make(map[string]struct{}),

		headerStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("170")).Padding(0, 1).
			Foreground(lipgloss.Color("255")),
		sizeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		fileStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		folderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("153")),
		linkStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("109")),
		trashedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")),
		selectedName: lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")),
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")),
		footerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
	}
	model.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	model.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	model.help.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	model.help.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	model.help.Styles.Ellipsis = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	model.refreshViewportContent()
	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		m.resize()
		m.refreshViewportContent()
		return m, nil
	case tea.KeyMsg:
		if key.Matches(typed, m.keys.Quit) {
			return m, tea.Quit
		}
		if len(m.activeEntries()) == 0 {
			return m, nil
		}
		switch {
		case key.Matches(typed, m.keys.Up):
			m.clearStatus()
			m.moveSelection(-1)
			return m, nil
		case key.Matches(typed, m.keys.Down):
			m.clearStatus()
			m.moveSelection(1)
			return m, nil
		case key.Matches(typed, m.keys.PageUp):
			m.clearStatus()
			step := max(1, m.viewport.Height())
			m.moveSelection(-step)
			return m, nil
		case key.Matches(typed, m.keys.PageDn):
			m.clearStatus()
			step := max(1, m.viewport.Height())
			m.moveSelection(step)
			return m, nil
		case key.Matches(typed, m.keys.Home):
			m.clearStatus()
			m.selected = 0
			m.keepSelectionVisible()
			m.refreshViewportContent()
			return m, nil
		case key.Matches(typed, m.keys.End):
			m.clearStatus()
			m.selected = len(m.activeEntries()) - 1
			m.keepSelectionVisible()
			m.refreshViewportContent()
			return m, nil
		case key.Matches(typed, m.keys.Open):
			return m, m.runPathAction("Open", m.openPath)
		case key.Matches(typed, m.keys.Reveal):
			return m, m.runPathAction("Reveal", m.revealPath)
		case key.Matches(typed, m.keys.Delete):
			path := m.selectedEntryPath()
			if m.isTrashed(path) {
				return m, nil
			}
			return m, m.runPathAction("Delete", m.trashPath)
		}
	case pathActionFinishedMsg:
		if typed.err == nil {
			if typed.verb == "Delete" {
				m.trashedPaths[typed.path] = struct{}{}
				if m.selected+1 < len(m.activeEntries()) {
					m.selected++
				}
				m.keepSelectionVisible()
				m.refreshViewportContent()
			}
			m.clearStatus()
			return m, nil
		}
		m.status = fmt.Sprintf("%s failed: %v", typed.verb, typed.err)
		m.resize()
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	header := m.headerStyle.Render(m.header)
	footer := m.footerStyle.Render(m.help.View(m.keys))

	var content string
	if len(m.activeEntries()) == 0 {
		content = "No entries"
	} else {
		content = m.viewport.View()
	}

	parts := []string{header, "", content, ""}
	if m.status != "" {
		parts = append(parts, m.statusStyle.Render(m.status))
	}
	parts = append(parts, footer)
	body := strings.Join(parts, "\n")
	view := tea.NewView(body)
	view.AltScreen = true
	return view
}

func (m *Model) resize() {
	headerHeight := 1
	footerHeight := 1
	statusHeight := 0
	if m.status != "" {
		statusHeight = 1
	}
	listHeight := m.height - headerHeight - footerHeight - statusHeight - 2
	if listHeight < 1 {
		listHeight = 1
	}

	m.viewport.SetWidth(m.width)
	m.viewport.SetHeight(listHeight)
	m.help.SetWidth(m.width)
}

func (m *Model) clearStatus() {
	if m.status == "" {
		return
	}
	m.status = ""
	m.resize()
}

func (m Model) runPathAction(verb string, action pathAction) tea.Cmd {
	path := m.selectedEntryPath()
	return func() tea.Msg {
		err := action(path)
		return pathActionFinishedMsg{
			verb: verb,
			path: path,
			err:  err,
		}
	}
}

func (m Model) selectedEntryPath() string {
	return m.entryPath(m.activeEntries()[m.selected])
}

func (m *Model) moveSelection(delta int) {
	next := m.selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.activeEntries()) {
		next = len(m.activeEntries()) - 1
	}
	m.selected = next
	m.keepSelectionVisible()
	m.refreshViewportContent()
}

func (m *Model) keepSelectionVisible() {
	if m.selected < m.viewport.YOffset() {
		m.viewport.SetYOffset(m.selected)
		return
	}

	bottom := m.viewport.YOffset() + m.viewport.Height() - 1
	if m.selected > bottom {
		m.viewport.SetYOffset(m.selected - m.viewport.Height() + 1)
	}
}

func (m *Model) refreshViewportContent() {
	entries := m.activeEntries()
	if len(entries) == 0 {
		m.viewport.SetContent("")
		return
	}

	lines := make([]string, 0, len(entries))
	for idx, entry := range entries {
		entryPath := m.entryPath(entry)
		isTrashed := m.isTrashed(entryPath)

		sizeStyle := m.sizeStyle
		if isTrashed {
			sizeStyle = m.trashedStyle.Inherit(sizeStyle)
		}
		if idx == m.selected {
			sizeStyle = m.selectedName.Inherit(sizeStyle)
		}

		sizeCell := strings.Repeat(" ", m.sizeWidth)
		if entry.HasSize {
			sizeCell = lipgloss.NewStyle().Width(m.sizeWidth).Align(lipgloss.Right).Render(scan.FormatSize(entry.Size))
		}

		name := entry.Name
		if entry.LinkTarget != "" {
			name = fmt.Sprintf("%s -> %s", entry.Name, entry.LinkTarget)
		}
		nameStyle := m.fileStyle
		switch entry.Kind {
		case scan.EntryFolder:
			nameStyle = m.folderStyle
		case scan.EntryOther:
			nameStyle = m.linkStyle
		}
		if isTrashed {
			nameStyle = m.trashedStyle.Inherit(nameStyle)
		}
		if idx == m.selected {
			nameStyle = m.selectedName.Inherit(nameStyle)
		}
		row := fmt.Sprintf("%s  %s", sizeStyle.Render(sizeCell), nameStyle.Render(name))
		lines = append(lines, row)
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
}

func (m Model) activeEntries() []scan.RootEntry {
	if len(m.entryViews) == 0 {
		return nil
	}
	return m.entryViews[m.selectedEntryView].entries
}

func buildCompatibilityEntryViews(entries []scan.RootEntry) []entryView {
	if len(entries) == 0 {
		return nil
	}
	return []entryView{
		{
			name:    "All",
			entries: entries,
		},
	}
}

func (m Model) entryPath(entry scan.RootEntry) string {
	absolute, err := filepath.Abs(entry.Path)
	if err != nil {
		return filepath.Clean(entry.Path)
	}
	return absolute
}

func (m Model) isTrashed(path string) bool {
	_, ok := m.trashedPaths[path]
	return ok
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func scanRootHeader(rootPath string) string {
	cleaned := filepath.Clean(rootPath)
	if cleaned == "." {
		wd, err := os.Getwd()
		if err == nil {
			cleaned = filepath.Clean(wd)
		}
	}
	base := filepath.Base(cleaned)
	if base == "." || base == string(filepath.Separator) || strings.HasSuffix(base, ":\\") {
		return cleaned
	}
	return base
}
