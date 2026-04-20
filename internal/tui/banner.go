package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

const asciiLogo = `
 _ __ ___  _ __ (_) |_
| '__/ _ \| '_ \| | __|
| | | (_) | | | | | |_
|_|  \___/|_| |_|_|\__|
`

// visitorCountMsg carries the result of the async visitor count query.
type visitorCountMsg struct {
	count int
	err   error
}

// BannerModel shows the neofetch-style intro screen.
// Transitions to the menu on any keypress.
type BannerModel struct {
	visitorCount int
	loaded       bool
	width        int
	height       int
	db           interface {
		GetVisitorCount() (int, error)
	}
}

func newBannerModel(w, h int, db interface{ GetVisitorCount() (int, error) }) BannerModel {
	return BannerModel{width: w, height: h, db: db}
}

func (m BannerModel) Init() tea.Cmd {
	if m.db == nil {
		return nil
	}
	return m.fetchVisitorCount()
}

// fetchVisitorCount returns a Cmd that queries the DB off the hot path.
func (m BannerModel) fetchVisitorCount() tea.Cmd {
	return func() tea.Msg {
		n, err := m.db.GetVisitorCount()
		return visitorCountMsg{count: n, err: err}
	}
}

func (m BannerModel) Update(msg tea.Msg) (BannerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case visitorCountMsg:
		if msg.err == nil {
			m.visitorCount = msg.count
		}
		m.loaded = true
		return m, nil
	case tea.KeyMsg:
		return m, navigateTo(ViewMenu)
	}
	return m, nil
}

func (m BannerModel) View() string {
	logo := theme.TitleStyle.Render(asciiLogo)

	kvStyle := lipgloss.NewStyle().PaddingLeft(2)
	labelStyle := theme.SubtleStyle.Copy().Width(10)
	valueStyle := lipgloss.NewStyle()

	kv := func(label, value string) string {
		return labelStyle.Render(label+":") + valueStyle.Render(value) + "\n"
	}

	visitorLine := fmt.Sprintf("visitor #%d", m.visitorCount+1)
	if !m.loaded {
		visitorLine = "visitor #..."
	}

	info := kvStyle.Render(
		kv("name", "Ronit Singh") +
			kv("school", "TJHSST '28") +
			kv("focus", "civic tech, full-stack dev") +
			kv("stack", "Go, TypeScript, Next.js") +
			kv("location", "Northern Virginia") +
			"\n" +
			theme.AccentStyle.Render("  "+visitorLine) + "\n",
	)

	banner := lipgloss.JoinHorizontal(lipgloss.Top, logo, info)
	prompt := theme.SubtleStyle.Render("\n  press any key to enter")

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		banner+prompt,
	)
}
