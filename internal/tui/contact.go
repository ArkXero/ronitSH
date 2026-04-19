package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// ContactModel renders the contact information page.
type ContactModel struct{}

func newContactModel() ContactModel { return ContactModel{} }

func (m ContactModel) Init() tea.Cmd { return nil }

func (m ContactModel) Update(msg tea.Msg) (ContactModel, tea.Cmd) { return m, nil }

func (m ContactModel) View() string {
	labelStyle := theme.SubtleStyle.Copy().Width(12)
	kv := func(label, value string) string {
		return labelStyle.Render(label+":") + value + "\n"
	}
	return theme.SectionTitleStyle.Render("CONTACT") + "\n\n" +
		kv("email", "ronit@example.com") +
		kv("github", "github.com/ArkXero") +
		kv("linkedin", "linkedin.com/in/ronitsingh")
}
