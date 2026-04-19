package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/viewport"
	"github.com/ArkXero/termfolio/internal/content"
	"github.com/ArkXero/termfolio/internal/theme"
)

// NowModel renders the "now" page from embedded markdown.
type NowModel struct {
	loader   *content.Loader
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

func newNowModel(loader *content.Loader) NowModel {
	return NowModel{loader: loader}
}

func (m NowModel) Init() tea.Cmd { return nil }

func (m NowModel) Update(msg tea.Msg) (NowModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.rebuild()
	}
	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m NowModel) View() string {
	if !m.ready {
		return theme.SubtleStyle.Render("loading...")
	}
	return m.viewport.View()
}

func (m NowModel) rebuild() NowModel {
	if m.width == 0 {
		return m
	}
	// Prepend "last updated" from frontmatter date.
	header := ""
	if m.loader.Now.Meta.Date != "" {
		header = fmt.Sprintf("*Last updated: %s*\n\n", m.loader.Now.Meta.Date)
	}
	rendered, err := content.RenderMarkdown(header+m.loader.Now.Body, m.width)
	if err != nil {
		rendered = header + m.loader.Now.Body
	}
	vp := viewport.New(viewport.WithWidth(m.width), viewport.WithHeight(m.height))
	vp.SetContent(rendered)
	m.viewport = vp
	m.ready = true
	return m
}
