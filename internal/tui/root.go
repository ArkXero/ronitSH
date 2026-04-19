package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/ArkXero/termfolio/internal/content"
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
	banner    BannerModel
	menu      MenuModel
	about     AboutModel
	projects  ProjectsModel
	now       NowModel
	posts     PostsModel
	guestbook GuestbookModel
	contact   ContactModel
}

// NewRootModel constructs the root model for a new SSH session.
// loader is shared across sessions (read-only after init).
func NewRootModel(cfg Config, loader *content.Loader) RootModel {
	keys := DefaultKeyMap()
	return RootModel{
		cfg:         cfg,
		currentView: ViewBanner,
		keys:        keys,
		banner:      newBannerModel(cfg.Width, cfg.Height),
		menu:        newMenuModel(keys),
		about:       newAboutModel(loader),
		projects:    newProjectsModel(loader),
		now:         newNowModel(loader),
		posts:       newPostsModel(loader),
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
		// Forward to sub-models that manage viewports.
		// Use a batch so all models get the resize concurrently.
		var cmds []tea.Cmd
		m.about, _ = m.about.Update(msg)
		m.now, _ = m.now.Update(msg)
		m.projects, _ = m.projects.Update(msg)
		m.posts, _ = m.posts.Update(msg)
		return m, tea.Batch(cmds...)

	case navigateMsg:
		m.currentView = msg.to
		m.showHelp = false
		return m, nil

	case tea.KeyMsg:
		// Global: hard quit.
		if key.Matches(msg, m.keys.HardQuit) {
			return m, tea.Quit
		}
		// Global: toggle help overlay (not on banner).
		if m.currentView != ViewBanner && key.Matches(msg, m.keys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}
		// Global: esc/q from a content view returns to menu.
		if m.currentView != ViewBanner && m.currentView != ViewMenu && !m.showHelp {
			if key.Matches(msg, m.keys.Back) {
				m.currentView = ViewMenu
				return m, nil
			}
		}
		// Any key closes the help overlay.
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
	}

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

// View renders the current state of the application.
// Layout depends on terminal width:
//
//	< 40 cols  : error message
//	< 80 cols  : narrow -- menu or top-bar + content
//	< 120 cols : standard -- sidebar (25) + content
//	>= 120 cols: wide -- sidebar (25) + content + preview (30)
func (m RootModel) View() tea.View {
	var content string

	switch {
	case m.cfg.Width < 40:
		content = m.tooNarrowView()
	case m.currentView == ViewBanner:
		content = m.banner.View()
	case m.showHelp:
		content = m.helpView()
	case m.cfg.Width < 80:
		content = m.narrowView()
	case m.cfg.Width >= 120:
		content = m.wideView()
	default:
		content = m.standardView()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// activeContentView returns the rendered string for the current non-menu section.
// Returns an empty string when in ViewMenu (callers handle that case separately).
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

// menuWelcomeView is shown in the content pane when the user is at the main menu.
func (m RootModel) menuWelcomeView() string {
	var b strings.Builder
	b.WriteString(theme.TitleStyle.Render("ronit.sh") + "\n\n")
	b.WriteString("Sophomore at TJHSST.\n")
	b.WriteString("Building civic tech and learning systems programming.\n\n")
	b.WriteString(theme.SubtleStyle.Render("Navigate with j/k, select with enter, ? for help."))
	return b.String()
}

// tooNarrowView is shown when the terminal is less than 40 columns wide.
func (m RootModel) tooNarrowView() string {
	msg := theme.SubtleStyle.Render("terminal too narrow\n(need >= 40 cols)")
	return lipgloss.Place(m.cfg.Width, m.cfg.Height, lipgloss.Center, lipgloss.Center, msg)
}

// standardView renders the 2-pane layout: sidebar (25) + content.
// Used when 80 <= width < 120.
func (m RootModel) standardView() string {
	bodyHeight := m.cfg.Height - 1 // one line for footer

	sidebarStyle := lipgloss.NewStyle().
		Width(theme.SidebarWidth).
		Height(bodyHeight).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingTop(1).
		PaddingRight(1)

	contentWidth := m.cfg.Width - theme.SidebarWidth - 2 // 2 for border + gap
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(bodyHeight).
		PaddingLeft(2).
		PaddingTop(1)

	var contentStr string
	if m.currentView == ViewMenu {
		contentStr = m.menuWelcomeView()
	} else {
		contentStr = m.activeContentView()
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		sidebarStyle.Render(m.menu.View()),
		contentStyle.Render(contentStr),
	)

	footer := lipgloss.NewStyle().
		Width(m.cfg.Width).
		Foreground(theme.Subtle).
		Render(m.keys.ShortHelp())

	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

// wideView renders the 3-pane layout: sidebar (25) + content + preview (30).
// Used when width >= 120.
func (m RootModel) wideView() string {
	bodyHeight := m.cfg.Height - 1

	sidebarStyle := lipgloss.NewStyle().
		Width(theme.SidebarWidth).
		Height(bodyHeight).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingTop(1).
		PaddingRight(1)

	previewWidth := theme.PreviewWidth
	contentWidth := m.cfg.Width - theme.SidebarWidth - previewWidth - 4

	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(bodyHeight).
		PaddingLeft(2).
		PaddingTop(1)

	previewStyle := lipgloss.NewStyle().
		Width(previewWidth).
		Height(bodyHeight).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingLeft(2).
		PaddingTop(1)

	var contentStr string
	if m.currentView == ViewMenu {
		contentStr = m.menuWelcomeView()
	} else {
		contentStr = m.activeContentView()
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		sidebarStyle.Render(m.menu.View()),
		contentStyle.Render(contentStr),
		previewStyle.Render(m.previewView()),
	)

	footer := lipgloss.NewStyle().
		Width(m.cfg.Width).
		Foreground(theme.Subtle).
		Render(m.keys.ShortHelp())

	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

// previewView returns content for the right preview pane in wide mode.
// Shows section hints based on what is currently selected in the menu.
func (m RootModel) previewView() string {
	previews := map[View]string{
		ViewAbout:     "Background, skills,\nand education.",
		ViewProjects:  "Civic Cycle, Botball\n2026, and this site.",
		ViewNow:       "What I am working\non this month.",
		ViewPosts:     "Writing and notes\nfrom building things.",
		ViewGuestbook: "Leave a message.\nRead what others wrote.",
		ViewContact:   "Email, GitHub,\nand LinkedIn.",
		ViewHelp:      "All keyboard\nshortcuts.",
		ViewQuit:      "See you next time.",
	}

	cursor := m.menu.Cursor()
	selected := menuItems[cursor].view
	hint, ok := previews[selected]
	if !ok {
		hint = ""
	}

	var b strings.Builder
	b.WriteString(theme.SubtleStyle.Render("preview") + "\n\n")
	b.WriteString(theme.ActiveItemStyle.Render(menuItems[cursor].label) + "\n\n")
	b.WriteString(hint)
	return b.String()
}

// narrowView renders the single-pane layout used when width < 80.
// In ViewMenu it shows the menu full-screen; otherwise shows the content with a back hint.
func (m RootModel) narrowView() string {
	footer := lipgloss.NewStyle().
		Width(m.cfg.Width).
		Foreground(theme.Subtle).
		Render(m.keys.ShortHelp())
	footerHeight := 1
	bodyHeight := m.cfg.Height - footerHeight

	if m.currentView == ViewMenu {
		// Full-screen menu -- no sidebar needed.
		menuStyle := lipgloss.NewStyle().
			Width(m.cfg.Width).
			Height(bodyHeight).
			PaddingTop(2).
			PaddingLeft(2)
		return lipgloss.JoinVertical(lipgloss.Left,
			menuStyle.Render(m.menu.View()),
			footer,
		)
	}

	// Content view: show section name at top, content, footer.
	header := theme.SubtleStyle.Render("< " + m.currentViewName() + "  (q: back)")
	headerHeight := 1
	contentStyle := lipgloss.NewStyle().
		Width(m.cfg.Width).
		Height(bodyHeight - headerHeight).
		PaddingLeft(1).
		PaddingTop(1)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		contentStyle.Render(m.activeContentView()),
		footer,
	)
}

func (m RootModel) helpView() string {
	var b strings.Builder
	b.WriteString(theme.SectionTitleStyle.Render("keybindings") + "\n\n")
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
