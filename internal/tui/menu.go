package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// MenuItem represents a single entry in the main menu.
type MenuItem struct {
	label       string
	description string
	view        View
}

var menuItems = []MenuItem{
	{label: "about", description: "who I am", view: ViewAbout},
	{label: "projects", description: "things I have built", view: ViewProjects},
	{label: "now", description: "what I am working on", view: ViewNow},
	{label: "posts", description: "writing and notes", view: ViewPosts},
	{label: "guestbook", description: "leave a message", view: ViewGuestbook},
	{label: "contact", description: "get in touch", view: ViewContact},
	{label: "help", description: "keybindings", view: ViewHelp},
	{label: "quit", description: "goodbye", view: ViewQuit},
}

// MenuModel is the main navigation menu.
type MenuModel struct {
	cursor int // index into menuItems
	keys   KeyMap
}

// Cursor returns the index of the currently highlighted menu item.
func (m MenuModel) Cursor() int { return m.cursor }

func newMenuModel(keys KeyMap) MenuModel {
	return MenuModel{
		cursor: 0,
		keys:   keys,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Select):
			item := menuItems[m.cursor]
			if item.view == ViewQuit {
				return m, tea.Quit
			}
			return m, navigateTo(item.view)
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	var s string
	for i, item := range menuItems {
		if i == m.cursor {
			cursor := theme.AccentStyle.Render(">")
			name := theme.ActiveItemStyle.Render(item.label)
			desc := theme.SubtleStyle.Render(item.description)
			s += cursor + " " + name + "\n"
			s += "  " + desc + "\n"
		} else {
			name := theme.InactiveItemStyle.Render("  " + item.label)
			s += name + "\n"
			s += "\n"
		}
	}
	return lipgloss.NewStyle().
		Width(theme.SidebarWidth).
		Render(s)
}
