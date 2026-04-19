package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/viewport"
	"github.com/ArkXero/termfolio/internal/content"
	"github.com/ArkXero/termfolio/internal/theme"
)

// AboutModel renders the about page from embedded markdown.
type AboutModel struct {
	loader   *content.Loader
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

func newAboutModel(loader *content.Loader) AboutModel {
	return AboutModel{loader: loader}
}

func (m AboutModel) Init() tea.Cmd { return nil }

func (m AboutModel) Update(msg tea.Msg) (AboutModel, tea.Cmd) {
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

func (m AboutModel) View() string {
	if !m.ready {
		return theme.SubtleStyle.Render("loading...")
	}
	return m.viewport.View()
}

// initViewport sets up the viewport with rendered markdown content.
// Called lazily on first render if width/height are known.
func (m AboutModel) rebuild() AboutModel {
	if m.width == 0 {
		return m
	}
	rendered, err := content.RenderMarkdown(m.loader.About.Body, m.width)
	if err != nil {
		rendered = m.loader.About.Body
	}
	vp := viewport.New(viewport.WithWidth(m.width), viewport.WithHeight(m.height))
	vp.SetContent(rendered) // pointer receiver -- OK since vp is addressable
	m.viewport = vp
	m.ready = true
	return m
}

// SetSize is called by the root model when terminal size changes.
func (m *AboutModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	*m = m.rebuild()
}
