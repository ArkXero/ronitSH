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

// BannerModel shows the neofetch-style intro screen.
// It transitions to the menu on any keypress.
type BannerModel struct {
	visitorCount int
	width        int
	height       int
}

func newBannerModel(w, h int) BannerModel {
	return BannerModel{
		visitorCount: 0,
		width:        w,
		height:       h,
	}
}

func (m BannerModel) Init() tea.Cmd {
	return nil
}

func (m BannerModel) Update(msg tea.Msg) (BannerModel, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		// Any key transitions to the menu.
		return m, navigateTo(ViewMenu)
	}
	return m, nil
}

func (m BannerModel) View() string {
	logo := theme.TitleStyle.Render(asciiLogo)

	kvStyle := lipgloss.NewStyle().PaddingLeft(2)
	labelStyle := theme.SubtleStyle.Copy().Width(10)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("default"))

	kv := func(label, value string) string {
		return labelStyle.Render(label+":") + valueStyle.Render(value) + "\n"
	}

	info := kvStyle.Render(
		kv("name", "Ronit Singh") +
			kv("school", "TJHSST '28") +
			kv("focus", "civic tech, full-stack dev") +
			kv("stack", "Go, TypeScript, Next.js") +
			kv("location", "Northern Virginia") +
			"\n" +
			theme.AccentStyle.Render(fmt.Sprintf("  visitor #%d", m.visitorCount+1)) + "\n",
	)

	banner := lipgloss.JoinHorizontal(lipgloss.Top, logo, info)

	prompt := theme.SubtleStyle.Render("\n  press any key to enter")

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		banner+prompt,
	)
}
