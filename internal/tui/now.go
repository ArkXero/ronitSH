package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// NowModel renders the "now" page.
type NowModel struct{}

func newNowModel() NowModel { return NowModel{} }

func (m NowModel) Init() tea.Cmd { return nil }

func (m NowModel) Update(msg tea.Msg) (NowModel, tea.Cmd) { return m, nil }

func (m NowModel) View() string {
	return theme.SectionTitleStyle.Render("NOW") + "\n\n" +
		theme.SubtleStyle.Render("Content coming soon.")
}
