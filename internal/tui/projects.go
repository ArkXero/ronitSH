package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// ProjectsModel renders the projects list and individual project details.
type ProjectsModel struct{}

func newProjectsModel() ProjectsModel { return ProjectsModel{} }

func (m ProjectsModel) Init() tea.Cmd { return nil }

func (m ProjectsModel) Update(msg tea.Msg) (ProjectsModel, tea.Cmd) { return m, nil }

func (m ProjectsModel) View() string {
	return theme.SectionTitleStyle.Render("PROJECTS") + "\n\n" +
		theme.SubtleStyle.Render("Content coming soon.")
}
