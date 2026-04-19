package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	"github.com/ArkXero/termfolio/internal/content"
	"github.com/ArkXero/termfolio/internal/theme"
)

// PostsModel shows the post list or a single post detail.
type PostsModel struct {
	loader   *content.Loader
	keys     KeyMap
	cursor   int
	detail   bool
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

func newPostsModel(loader *content.Loader) PostsModel {
	return PostsModel{loader: loader, keys: DefaultKeyMap()}
}

func (m PostsModel) Init() tea.Cmd { return nil }

func (m PostsModel) Update(msg tea.Msg) (PostsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.detail {
			m = m.rebuildDetail()
		}
	case tea.KeyMsg:
		if m.detail {
			if key.Matches(msg, m.keys.Back) {
				m.detail = false
				m.ready = false
				return m, nil
			}
		} else {
			switch {
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.loader.Posts)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.keys.Select):
				m.detail = true
				m.ready = false
				m = m.rebuildDetail()
			}
			return m, nil
		}
	}
	if m.detail && m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m PostsModel) View() string {
	if len(m.loader.Posts) == 0 {
		return theme.SectionTitleStyle.Render("posts") + "\n\n" +
			theme.SubtleStyle.Render("No posts yet.")
	}
	if m.detail {
		return m.detailView()
	}
	return m.listView()
}

func (m PostsModel) listView() string {
	var b strings.Builder
	b.WriteString(theme.SectionTitleStyle.Render("posts") + "\n\n")
	for i, p := range m.loader.Posts {
		date := p.Meta.Date
		if date == "" {
			// Fall back to filename date prefix (YYYY-MM-DD).
			if len(p.Filename) >= 10 {
				date = p.Filename[:10]
			}
		}
		if i == m.cursor {
			title := theme.ActiveItemStyle.Render("> " + p.Meta.Title)
			b.WriteString(fmt.Sprintf("%s\n", title))
			b.WriteString(theme.SubtleStyle.Render("  "+date) + "\n\n")
		} else {
			title := theme.InactiveItemStyle.Render("  " + p.Meta.Title)
			b.WriteString(fmt.Sprintf("%s\n", title))
			b.WriteString(theme.SubtleStyle.Render("  "+date) + "\n\n")
		}
	}
	b.WriteString(theme.SubtleStyle.Render("enter: read  j/k: navigate"))
	return b.String()
}

func (m PostsModel) detailView() string {
	if !m.ready {
		return theme.SubtleStyle.Render("loading...")
	}
	return m.viewport.View()
}

func (m PostsModel) rebuildDetail() PostsModel {
	if m.width == 0 || m.cursor >= len(m.loader.Posts) {
		return m
	}
	p := m.loader.Posts[m.cursor]

	header := fmt.Sprintf("# %s\n\n*%s*\n\n---\n\n", p.Meta.Title, p.Meta.Date)
	rendered, err := content.RenderMarkdown(header+p.Body, m.width-4)
	if err != nil {
		rendered = header + p.Body
	}

	vp := viewport.New(viewport.WithWidth(m.width-4), viewport.WithHeight(m.height-2))
	vp.SetContent(rendered)
	m.viewport = vp
	m.ready = true
	return m
}
