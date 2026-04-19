package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// GuestbookModel renders the guestbook list and compose form.
type GuestbookModel struct{}

func newGuestbookModel() GuestbookModel { return GuestbookModel{} }

func (m GuestbookModel) Init() tea.Cmd { return nil }

func (m GuestbookModel) Update(msg tea.Msg) (GuestbookModel, tea.Cmd) { return m, nil }

func (m GuestbookModel) View() string {
	return theme.SectionTitleStyle.Render("GUESTBOOK") + "\n\n" +
		theme.SubtleStyle.Render("No entries yet. Be the first!") + "\n\n" +
		theme.AccentStyle.Render("  [w] write a message")
}
