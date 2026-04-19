package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// PostsModel renders the posts list and individual post detail.
type PostsModel struct{}

func newPostsModel() PostsModel { return PostsModel{} }

func (m PostsModel) Init() tea.Cmd { return nil }

func (m PostsModel) Update(msg tea.Msg) (PostsModel, tea.Cmd) { return m, nil }

func (m PostsModel) View() string {
	return theme.SectionTitleStyle.Render("POSTS") + "\n\n" +
		theme.SubtleStyle.Render("Content coming soon.")
}
