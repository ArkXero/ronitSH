package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// AboutModel renders the about page.
type AboutModel struct{}

func newAboutModel() AboutModel { return AboutModel{} }

func (m AboutModel) Init() tea.Cmd { return nil }

func (m AboutModel) Update(msg tea.Msg) (AboutModel, tea.Cmd) { return m, nil }

func (m AboutModel) View() string {
	return theme.SectionTitleStyle.Render("ABOUT") + "\n\n" +
		theme.SubtleStyle.Render("Content coming soon -- Ronit is writing this.")
}
