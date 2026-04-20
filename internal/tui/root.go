package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/ArkXero/termfolio/internal/content"
	"github.com/ArkXero/termfolio/internal/ratelimit"
	"github.com/ArkXero/termfolio/internal/storage"
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
	IP     string // masked IP prefix for rate limiting
	KeyFP  string // truncated key fingerprint
	RL     *ratelimit.Limiter
	DB     interface {
		GetVisitorCount() (int, error)
		GetGuestbookCount() (int, error)
		IncrementCommandUsage(string) error
		AddGuestbookEntry(name, message, ipPrefix, keyFP string) (int64, error)
		ListGuestbookEntries(includeHidden bool) ([]storage.GuestbookEntry, error)
	}
}

// bodyHeight returns the number of rows available for the main body content,
// accounting for the deco footer and the keybinding strip.
func bodyHeight(totalHeight int) int {
	return totalHeight - theme.DecoFooterHeight - 1
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
	footer    DecoFooterModel
}

// NewRootModel constructs the root model for a new SSH session.
// loader is shared across sessions (read-only after init).
func NewRootModel(cfg Config, loader *content.Loader) RootModel {
	keys := DefaultKeyMap()
	bh := bodyHeight(cfg.Height)
	return RootModel{
		cfg:         cfg,
		currentView: ViewBanner,
		keys:        keys,
		banner:      newBannerModel(cfg.Width, bh, cfg.DB),
		menu:        newMenuModel(keys),
		about:       newAboutModel(loader),
		projects:    newProjectsModel(loader),
		now:         newNowModel(loader),
		posts:       newPostsModel(loader),
		guestbook:   newGuestbookModel(cfg.DB, cfg.RL, cfg.IP, cfg.KeyFP),
		contact:     newContactModel(),
		footer:      NewDecoFooterModel(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return tea.Batch(m.banner.Init(), m.footer.Init())
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.cfg.Width = msg.Width
		m.cfg.Height = msg.Height
		bh := bodyHeight(msg.Height)
		m.banner.width = msg.Width
		m.banner.height = bh
		m.footer, _ = m.footer.Update(msg)
		// Forward to sub-models that manage viewports.
		var cmds []tea.Cmd
		m.about, _ = m.about.Update(msg)
		m.now, _ = m.now.Update(msg)
		m.projects, _ = m.projects.Update(msg)
		m.posts, _ = m.posts.Update(msg)
		return m, tea.Batch(cmds...)

	case footerTickMsg:
		var cmd tea.Cmd
		m.footer, cmd = m.footer.Update(msg)
		return m, cmd

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
		// Skip when a sub-model is in its own detail view and should handle Back itself.
		submodelInDetail := (m.currentView == ViewProjects && m.projects.detail) ||
			(m.currentView == ViewPosts && m.posts.detail)
		if m.currentView != ViewBanner && m.currentView != ViewMenu && !m.showHelp && !submodelInDetail {
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
//	< 40 cols  : error message (no deco footer)
//	< 80 cols  : narrow -- menu or top-bar + content
//	< 120 cols : standard -- sidebar (25) + content
//	>= 120 cols: wide -- sidebar (25) + content + preview (30)
//
// All layouts except tooNarrow include the animated deco footer at the bottom.
func (m RootModel) View() tea.View {
	// Too narrow: show error only, no deco footer.
	if m.cfg.Width < 40 {
		v := tea.NewView(m.tooNarrowView())
		v.AltScreen = true
		return v
	}

	// Build body (the area above the deco footer).
	var body string
	switch {
	case m.currentView == ViewBanner:
		body = m.banner.View()
	case m.showHelp:
		body = m.helpView()
	case m.cfg.Width < 80:
		body = m.narrowView()
	case m.cfg.Width >= 120:
		body = m.wideView()
	default:
		body = m.standardView()
	}

	// Keybinding strip (1 row).
	keyFooter := lipgloss.NewStyle().
		Width(m.cfg.Width).
		Foreground(theme.Subtle).
		Render(m.keys.ShortHelp())

	content := lipgloss.JoinVertical(lipgloss.Left, body, m.footer.View(), keyFooter)

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
// Used when 80 <= width < 120. Returns body only (no keybinding footer).
func (m RootModel) standardView() string {
	bh := bodyHeight(m.cfg.Height)
	innerH := bh - 2 // subtract top+bottom border rows

	// Sidebar: inner width = SidebarWidth - 2 (borders), full border box
	sidebarInner := theme.SidebarWidth - 2
	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarInner).
		Height(innerH).
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingTop(1).
		PaddingLeft(1)

	// Content pane: remaining width minus sidebar total (SidebarWidth) - 1 gap
	contentTotal := m.cfg.Width - theme.SidebarWidth - 1
	contentInner := contentTotal - 2
	contentStyle := lipgloss.NewStyle().
		Width(contentInner).
		Height(innerH).
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingLeft(2).
		PaddingTop(1)

	var contentStr string
	if m.currentView == ViewMenu {
		contentStr = m.menuWelcomeView()
	} else {
		contentStr = m.activeContentView()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		sidebarStyle.Render(m.menu.View()),
		contentStyle.Render(contentStr),
	)
}

// wideView renders the 3-pane layout: sidebar (25) + content + preview (30).
// Used when width >= 120. Returns body only (no keybinding footer).
func (m RootModel) wideView() string {
	bh := bodyHeight(m.cfg.Height)
	innerH := bh - 2 // subtract top+bottom border rows

	sidebarInner := theme.SidebarWidth - 2
	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarInner).
		Height(innerH).
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingTop(1).
		PaddingLeft(1)

	previewTotal := theme.PreviewWidth
	previewInner := previewTotal - 2
	// Total columns used: sidebar + 1 gap + content + 1 gap + preview
	contentTotal := m.cfg.Width - theme.SidebarWidth - 1 - previewTotal - 1
	contentInner := contentTotal - 2

	contentStyle := lipgloss.NewStyle().
		Width(contentInner).
		Height(innerH).
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingLeft(2).
		PaddingTop(1)

	previewStyle := lipgloss.NewStyle().
		Width(previewInner).
		Height(innerH).
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingLeft(2).
		PaddingTop(1)

	var contentStr string
	if m.currentView == ViewMenu {
		contentStr = m.menuWelcomeView()
	} else {
		contentStr = m.activeContentView()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		sidebarStyle.Render(m.menu.View()),
		contentStyle.Render(contentStr),
		previewStyle.Render(m.previewView()),
	)
}

// previewView returns content for the right preview pane in wide mode.
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
// Returns body only (no keybinding footer).
func (m RootModel) narrowView() string {
	bh := bodyHeight(m.cfg.Height)
	innerH := bh - 2
	innerW := m.cfg.Width - 2

	boxStyle := lipgloss.NewStyle().
		Width(innerW).
		Height(innerH).
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryDark).
		PaddingLeft(1).
		PaddingTop(1)

	if m.currentView == ViewMenu {
		return boxStyle.Render(m.menu.View())
	}

	header := theme.SubtleStyle.Render("< " + m.currentViewName() + "  (q: back)") + "\n\n"
	return boxStyle.Render(header + m.activeContentView())
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
	return lipgloss.Place(m.cfg.Width, bodyHeight(m.cfg.Height), lipgloss.Center, lipgloss.Center, b.String())
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
