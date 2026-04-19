package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// View enumerates all possible screen states.
type View int

const (
	ViewBanner View = iota
	ViewMenu
	ViewAbout
	ViewProjects
	ViewProjectDetail
	ViewNow
	ViewPosts
	ViewPostDetail
	ViewGuestbook
	ViewGuestbookCompose
	ViewContact
	ViewHelp
	ViewStats
	ViewQuit // sentinel used by menu to signal quit
)

// navigateMsg is sent by sub-models to request a view transition.
type navigateMsg struct{ to View }

// navigateTo returns a Cmd that sends a navigateMsg.
func navigateTo(v View) tea.Cmd {
	return func() tea.Msg { return navigateMsg{to: v} }
}

// Config carries per-session configuration for the root model.
type Config struct {
	Term   string
	Width  int
	Height int
}

// RootModel is the top-level Bubble Tea model. It owns the current view enum
// and delegates rendering and input to sub-models.
type RootModel struct {
	cfg         Config
	currentView View
	keys        KeyMap
	showHelp    bool

	// Sub-models.
	banner   BannerModel
	menu     MenuModel
	about    AboutModel
	projects ProjectsModel
	now      NowModel
	posts    PostsModel
	guestbook GuestbookModel
	contact  ContactModel
}

// NewRootModel constructs the root model for a new SSH session.
func NewRootModel(cfg Config) RootModel {
	keys := DefaultKeyMap()
	return RootModel{
		cfg:         cfg,
		currentView: ViewBanner,
		keys:        keys,
		banner:      newBannerModel(cfg.Width, cfg.Height),
		menu:        newMenuModel(keys),
		about:       newAboutModel(),
		projects:    newProjectsModel(),
		now:         newNowModel(),
		posts:       newPostsModel(),
		guestbook:   newGuestbookModel(),
		contact:     newContactModel(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.cfg.Width = msg.Width
		m.cfg.Height = msg.Height
		m.banner.width = msg.Width
		m.banner.height = msg.Height
		return m, nil

	case navigateMsg:
		m.currentView = msg.to
		m.showHelp = false
		return m, nil

	case tea.KeyMsg:
		// Global: hard quit.
		if key.Matches(msg, m.keys.HardQuit) {
			return m, tea.Quit
		}
		// Global: toggle help overlay (except on banner).
		if m.currentView != ViewBanner && key.Matches(msg, m.keys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}
		// Global: back from any content view goes to menu.
		if m.currentView != ViewBanner && m.currentView != ViewMenu && !m.showHelp {
			if key.Matches(msg, m.keys.Back) {
				m.currentView = ViewMenu
				return m, nil
			}
		}
		// If help overlay is open, close it on any key except help toggle.
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
	}

	// Delegate to the active sub-model.
	return m.delegateUpdate(msg)
}

func (m RootModel) delegateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentView {
	case ViewBanner:
		m.banner, cmd = m.banner.Update(msg)
	case ViewMenu:
		m.menu, cmd = m.menu.Update(msg)
	case ViewAbout:
		m.about, cmd = m.about.Update(msg)
	case ViewProjects, ViewProjectDetail:
		m.projects, cmd = m.projects.Update(msg)
	case ViewNow:
		m.now, cmd = m.now.Update(msg)
	case ViewPosts, ViewPostDetail:
		m.posts, cmd = m.posts.Update(msg)
	case ViewGuestbook, ViewGuestbookCompose:
		m.guestbook, cmd = m.guestbook.Update(msg)
	case ViewContact:
		m.contact, cmd = m.contact.Update(msg)
	}
	return m, cmd
}

func (m RootModel) View() tea.View {
	var content string

	if m.cfg.Width < 40 {
		msg := theme.SubtleStyle.Render("terminal too narrow (need at least 40 cols)")
		content = lipgloss.Place(m.cfg.Width, m.cfg.Height, lipgloss.Center, lipgloss.Center, msg)
	} else if m.currentView == ViewBanner {
		content = m.banner.View()
	} else if m.showHelp {
		content = m.helpView()
	} else if m.cfg.Width < 80 {
		content = m.narrowView(m.activeContentView())
	} else {
		content = m.standardView(m.activeContentView())
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m RootModel) activeContentView() string {
	switch m.currentView {
	case ViewAbout:
		return m.about.View()
	case ViewProjects, ViewProjectDetail:
		return m.projects.View()
	case ViewNow:
		return m.now.View()
	case ViewPosts, ViewPostDetail:
		return m.posts.View()
	case ViewGuestbook, ViewGuestbookCompose:
		return m.guestbook.View()
	case ViewContact:
		return m.contact.View()
	default:
		return ""
	}
}

func (m RootModel) standardView(content string) string {
	sidebarHeight := m.cfg.Height - 2 // leave room for footer
	sidebar := lipgloss.NewStyle().
		Width(theme.SidebarWidth).
		Height(sidebarHeight).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingRight(1).
		Render(m.menu.View())

	contentWidth := m.cfg.Width - theme.SidebarWidth - 3
	contentPane := lipgloss.NewStyle().
		Width(contentWidth).
		Height(sidebarHeight).
		PaddingLeft(2).
		Render(content)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, contentPane)
	footer := theme.SubtleStyle.
		Width(m.cfg.Width).
		Render(m.keys.ShortHelp())

	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

func (m RootModel) narrowView(content string) string {
	nav := theme.SubtleStyle.Render(m.currentViewName() + " -- q to go back")
	footer := theme.SubtleStyle.Width(m.cfg.Width).Render(m.keys.ShortHelp())
	body := lipgloss.NewStyle().
		Width(m.cfg.Width).
		Height(m.cfg.Height - 3).
		Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, nav, body, footer)
}

func (m RootModel) helpView() string {
	var b strings.Builder
	b.WriteString(theme.SectionTitleStyle.Render("KEYBINDINGS") + "\n\n")
	rows := [][]string{
		{"j / k", "navigate up / down"},
		{"h / l", "navigate panels"},
		{"enter", "select"},
		{"esc / q", "go back"},
		{"?", "toggle this help"},
		{"ctrl+c", "hard quit"},
		{"pgup / pgdn", "scroll content"},
	}
	labelStyle := theme.SubtleStyle.Copy().Width(14)
	for _, row := range rows {
		b.WriteString(labelStyle.Render(row[0]) + row[1] + "\n")
	}
	b.WriteString("\n" + theme.SubtleStyle.Render("press any key to close"))
	return lipgloss.Place(m.cfg.Width, m.cfg.Height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m RootModel) currentViewName() string {
	switch m.currentView {
	case ViewAbout:
		return "about"
	case ViewProjects, ViewProjectDetail:
		return "projects"
	case ViewNow:
		return "now"
	case ViewPosts, ViewPostDetail:
		return "posts"
	case ViewGuestbook, ViewGuestbookCompose:
		return "guestbook"
	case ViewContact:
		return "contact"
	default:
		return ""
	}
}
