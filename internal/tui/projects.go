package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/ArkXero/termfolio/internal/content"
	"github.com/ArkXero/termfolio/internal/theme"
)

// ProjectsModel shows either the project list or a single project detail.
type ProjectsModel struct {
	loader   *content.Loader
	keys     KeyMap
	cursor   int
	detail   bool // true = showing detail for projects[cursor]
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

func newProjectsModel(loader *content.Loader) ProjectsModel {
	return ProjectsModel{loader: loader, keys: DefaultKeyMap()}
}

func (m ProjectsModel) Init() tea.Cmd { return nil }

func (m ProjectsModel) Update(msg tea.Msg) (ProjectsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.detail {
			m = m.rebuildDetail()
		}
	case tea.KeyMsg:
		if m.detail {
			// In detail view: esc/q goes back to list.
			if key.Matches(msg, m.keys.Back) {
				m.detail = false
				m.ready = false
				return m, nil
			}
		} else {
			// In list view: j/k navigates, enter opens detail.
			switch {
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.loader.Projects)-1 {
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

func (m ProjectsModel) View() string {
	if len(m.loader.Projects) == 0 {
		return theme.SectionTitleStyle.Render("projects") + "\n\n" +
			theme.SubtleStyle.Render("No projects found.")
	}
	if m.detail {
		return m.detailView()
	}
	return m.listView()
}

func (m ProjectsModel) listView() string {
	var b strings.Builder
	b.WriteString(theme.SectionTitleStyle.Render("projects") + "\n\n")
	for i, p := range m.loader.Projects {
		statusBadge := statusChip(p.Meta.Status)
		if i == m.cursor {
			title := theme.ActiveItemStyle.Render("> " + p.Meta.Title)
			b.WriteString(fmt.Sprintf("%s  %s\n", title, statusBadge))
			b.WriteString(theme.SubtleStyle.Render("  "+p.Meta.Tagline) + "\n\n")
		} else {
			title := theme.InactiveItemStyle.Render("  " + p.Meta.Title)
			b.WriteString(fmt.Sprintf("%s  %s\n\n", title, statusBadge))
		}
	}
	b.WriteString(theme.SubtleStyle.Render("enter: open  j/k: navigate"))
	return b.String()
}

func (m ProjectsModel) detailView() string {
	if !m.ready {
		return theme.SubtleStyle.Render("loading...")
	}
	return m.viewport.View()
}

func (m ProjectsModel) rebuildDetail() ProjectsModel {
	if m.width == 0 || m.cursor >= len(m.loader.Projects) {
		return m
	}
	p := m.loader.Projects[m.cursor]

	// Build a header from the frontmatter fields.
	var header strings.Builder
	header.WriteString("# " + p.Meta.Title + "\n\n")
	if p.Meta.Tagline != "" {
		header.WriteString("*" + p.Meta.Tagline + "*\n\n")
	}
	header.WriteString("**Status:** " + p.Meta.Status + "\n\n")
	if len(p.Meta.Stack) > 0 {
		header.WriteString("**Stack:** " + strings.Join(p.Meta.Stack, ", ") + "\n\n")
	}
	if len(p.Meta.Links) > 0 {
		header.WriteString("**Links:**\n")
		for k, v := range p.Meta.Links {
			header.WriteString(fmt.Sprintf("- [%s](%s)\n", k, v))
		}
		header.WriteString("\n")
	}
	header.WriteString("---\n\n")

	rendered, err := content.RenderMarkdown(header.String()+p.Body, m.width-4)
	if err != nil {
		rendered = header.String() + p.Body
	}

	vp := viewport.New(viewport.WithWidth(m.width-4), viewport.WithHeight(m.height-2))
	vp.SetContent(rendered)
	m.viewport = vp
	m.ready = true
	return m
}

// statusChip returns a styled status badge using symbol+color (NO_COLOR safe).
func statusChip(status string) string {
	switch strings.ToLower(status) {
	case "shipped":
		return theme.ChipShipped.Render("[SHIPPED]")
	case "in-progress", "in_progress":
		return theme.ChipInProgress.Render("[IN PROGRESS]")
	case "archived":
		return theme.ChipArchived.Render("[ARCHIVED]")
	default:
		return lipgloss.NewStyle().Render("[" + strings.ToUpper(status) + "]")
	}
}
