package tui

import (
	"fmt"
	"strings"

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
	Quit   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		Down:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		PageUp: key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDn: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
		Home:   key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "first")),
		End:    key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "last")),
		Quit:   key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q/esc", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.PageUp, k.PageDn, k.Home, k.End, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDn},
		{k.Home, k.End, k.Quit},
	}
}

type Model struct {
	rootPath string
	entries  []scan.RootEntry

	sizeWidth int
	selected  int
	width     int
	height    int

	viewport viewport.Model
	help     help.Model
	keys     keyMap

	headerStyle   lipgloss.Style
	sizeStyle     lipgloss.Style
	nameStyle     lipgloss.Style
	selectedStyle lipgloss.Style
	footerStyle   lipgloss.Style
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
		rootPath: rootPath,
		entries:  entries,

		sizeWidth: sizeWidth,
		selected:  0,

		viewport: vp,
		help:     help.New(),
		keys:     keys,

		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")),
		sizeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")),
		nameStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		selectedStyle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("24")).
			Foreground(lipgloss.Color("255")),
		footerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
	}
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
		if len(m.entries) == 0 {
			return m, nil
		}
		switch {
		case key.Matches(typed, m.keys.Up):
			m.moveSelection(-1)
			return m, nil
		case key.Matches(typed, m.keys.Down):
			m.moveSelection(1)
			return m, nil
		case key.Matches(typed, m.keys.PageUp):
			step := max(1, m.viewport.Height())
			m.moveSelection(-step)
			return m, nil
		case key.Matches(typed, m.keys.PageDn):
			step := max(1, m.viewport.Height())
			m.moveSelection(step)
			return m, nil
		case key.Matches(typed, m.keys.Home):
			m.selected = 0
			m.keepSelectionVisible()
			m.refreshViewportContent()
			return m, nil
		case key.Matches(typed, m.keys.End):
			m.selected = len(m.entries) - 1
			m.keepSelectionVisible()
			m.refreshViewportContent()
			return m, nil
		case typed.String() == "enter":
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	header := m.headerStyle.Render(fmt.Sprintf("Big - %s", m.rootPath))
	footer := m.footerStyle.Render(m.help.View(m.keys))

	var content string
	if len(m.entries) == 0 {
		content = "No entries"
	} else {
		content = m.viewport.View()
	}

	body := strings.Join([]string{header, "", content, "", footer}, "\n")
	view := tea.NewView(body)
	view.AltScreen = true
	return view
}

func (m *Model) resize() {
	headerHeight := 1
	footerHeight := 1
	listHeight := m.height - headerHeight - footerHeight - 2
	if listHeight < 1 {
		listHeight = 1
	}

	m.viewport.SetWidth(m.width)
	m.viewport.SetHeight(listHeight)
	m.help.SetWidth(m.width)
}

func (m *Model) moveSelection(delta int) {
	next := m.selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.entries) {
		next = len(m.entries) - 1
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
	if len(m.entries) == 0 {
		m.viewport.SetContent("")
		return
	}

	lines := make([]string, 0, len(m.entries))
	for idx, entry := range m.entries {
		sizeCell := strings.Repeat(" ", m.sizeWidth)
		if entry.HasSize {
			sizeCell = fmt.Sprintf("%*s", m.sizeWidth, scan.FormatSize(entry.Size))
		}

		name := entry.Name
		if entry.LinkTarget != "" {
			name = fmt.Sprintf("%s -> %s", entry.Name, entry.LinkTarget)
		}

		row := fmt.Sprintf("%s  %s", m.sizeStyle.Render(sizeCell), m.nameStyle.Render(name))
		if idx == m.selected {
			row = m.selectedStyle.Render(row)
		}
		lines = append(lines, row)
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
