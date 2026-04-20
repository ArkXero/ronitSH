package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/ArkXero/termfolio/internal/ratelimit"
	"github.com/ArkXero/termfolio/internal/sanitize"
	"github.com/ArkXero/termfolio/internal/storage"
	"github.com/ArkXero/termfolio/internal/theme"
)

// guestbookStore is the storage interface required by GuestbookModel.
type guestbookStore interface {
	AddGuestbookEntry(name, message, ipPrefix, keyFP string) (int64, error)
	ListGuestbookEntries(includeHidden bool) ([]storage.GuestbookEntry, error)
}

type guestbookMode int

const (
	gbModeList    guestbookMode = iota
	gbModeCompose               // huh form
)

// GuestbookModel renders the guestbook list and compose form.
type GuestbookModel struct {
	mode    guestbookMode
	db      guestbookStore
	rl      *ratelimit.Limiter
	ip      string
	keyFP   string
	entries []storage.GuestbookEntry
	flash   string // one-line success/error message
	width   int

	// compose form
	form    *huh.Form
	name    string
	message string
}

// guestbookDB wraps the broader Config DB interface so we can pass it as a
// guestbookStore. We use an adapter because Config.DB carries methods that
// guestbookStore doesn't need, and Go interfaces require exact method sets.
type gbDBAdapter struct {
	raw interface {
		AddGuestbookEntry(name, message, ipPrefix, keyFP string) (int64, error)
		ListGuestbookEntries(includeHidden bool) ([]storage.GuestbookEntry, error)
	}
}

func (a gbDBAdapter) AddGuestbookEntry(n, m, ip, fp string) (int64, error) {
	return a.raw.AddGuestbookEntry(n, m, ip, fp)
}
func (a gbDBAdapter) ListGuestbookEntries(h bool) ([]storage.GuestbookEntry, error) {
	return a.raw.ListGuestbookEntries(h)
}

// newGuestbookModel constructs the model. db may be nil (for testing / no DB).
func newGuestbookModel(db interface {
	AddGuestbookEntry(name, message, ipPrefix, keyFP string) (int64, error)
	ListGuestbookEntries(includeHidden bool) ([]storage.GuestbookEntry, error)
}, rl *ratelimit.Limiter, ip, keyFP string) GuestbookModel {
	m := GuestbookModel{
		rl:    rl,
		ip:    ip,
		keyFP: keyFP,
		width: 80,
	}
	if db != nil {
		m.db = gbDBAdapter{raw: db}
	}
	return m
}

// --- msgs ---

type gbEntriesMsg struct {
	entries []storage.GuestbookEntry
	err     error
}

type gbSubmitMsg struct {
	err error
}

// --- init / cmds ---

func (m GuestbookModel) Init() tea.Cmd {
	return m.loadEntries()
}

func (m GuestbookModel) loadEntries() tea.Cmd {
	if m.db == nil {
		return nil
	}
	return func() tea.Msg {
		entries, err := m.db.ListGuestbookEntries(false)
		return gbEntriesMsg{entries: entries, err: err}
	}
}

func (m GuestbookModel) submitEntry() tea.Cmd {
	name := sanitize.TruncateRunes(sanitize.Sanitize(m.name), 50)
	msg := sanitize.TruncateRunes(sanitize.Sanitize(m.message), 280)
	ip := m.ip
	fp := m.keyFP
	db := m.db
	return func() tea.Msg {
		if db == nil {
			return gbSubmitMsg{err: fmt.Errorf("storage unavailable")}
		}
		_, err := db.AddGuestbookEntry(name, msg, ip, fp)
		return gbSubmitMsg{err: err}
	}
}

func (m GuestbookModel) newForm() *huh.Form {
	m.name = ""
	m.message = ""
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("name").
				Placeholder("your name").
				Value(&m.name).
				CharLimit(50).
				Validate(huh.ValidateNotEmpty()),
			huh.NewText().
				Title("message").
				Placeholder("leave a note...").
				Value(&m.message).
				CharLimit(280).
				Lines(4).
				Validate(huh.ValidateNotEmpty()),
		),
	)
}

// --- update ---

func (m GuestbookModel) Update(msg tea.Msg) (GuestbookModel, tea.Cmd) {
	switch msg := msg.(type) {
	case gbEntriesMsg:
		if msg.err == nil {
			m.entries = msg.entries
		}
		return m, nil

	case gbSubmitMsg:
		m.mode = gbModeList
		m.form = nil
		if msg.err != nil {
			m.flash = "error saving entry: " + msg.err.Error()
		} else {
			m.flash = "message saved -- thanks!"
		}
		return m, m.loadEntries()

	case navigateMsg:
		// Pass navigation to root.
		return m, nil

	case tea.KeyMsg:
		if m.mode == gbModeList {
			switch msg.String() {
			case "w":
				// Check rate limit.
				if m.rl != nil && !m.rl.Allow(m.ip) {
					rem := m.rl.Remaining(m.ip)
					mins := int(rem.Minutes()) + 1
					m.flash = fmt.Sprintf("rate limited -- try again in ~%d min", mins)
					return m, nil
				}
				m.flash = ""
				f := m.newForm()
				m.form = f
				m.mode = gbModeCompose
				return m, f.Init()
			}
			return m, nil
		}

		// Compose mode: esc cancels.
		if msg.String() == "esc" {
			m.mode = gbModeList
			m.form = nil
			return m, nil
		}
	}

	// In compose mode, forward all messages to the form.
	if m.mode == gbModeCompose && m.form != nil {
		updated, cmd := m.form.Update(msg)
		if f, ok := updated.(*huh.Form); ok {
			m.form = f
		}
		if m.form != nil && m.form.State == huh.StateCompleted {
			return m, tea.Batch(cmd, m.submitEntry())
		}
		if m.form != nil && m.form.State == huh.StateAborted {
			m.mode = gbModeList
			m.form = nil
			return m, nil
		}
		return m, cmd
	}

	return m, nil
}

// --- view ---

func (m GuestbookModel) View() string {
	if m.mode == gbModeCompose && m.form != nil {
		var b strings.Builder
		b.WriteString(theme.SectionTitleStyle.Render("GUESTBOOK") + "\n\n")
		b.WriteString(theme.SubtleStyle.Render("esc to cancel") + "\n\n")
		b.WriteString(m.form.View())
		return b.String()
	}

	var b strings.Builder
	b.WriteString(theme.SectionTitleStyle.Render("GUESTBOOK") + "\n\n")

	if m.flash != "" {
		style := theme.AccentStyle
		if strings.HasPrefix(m.flash, "error") || strings.HasPrefix(m.flash, "rate") {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#E74C3C"))
		}
		b.WriteString(style.Render(m.flash) + "\n\n")
	}

	if len(m.entries) == 0 {
		b.WriteString(theme.SubtleStyle.Render("No entries yet. Be the first!") + "\n\n")
	} else {
		nameStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.PrimaryLight)
		timeStyle := theme.SubtleStyle.Copy()
		msgStyle := lipgloss.NewStyle()
		sep := theme.SubtleStyle.Render(strings.Repeat("─", 40)) + "\n"

		for i, e := range m.entries {
			if i > 0 {
				b.WriteString(sep)
			}
			b.WriteString(nameStyle.Render(e.Name))
			b.WriteString("  " + timeStyle.Render(storage.RelativeTime(e.CreatedAt)) + "\n")
			b.WriteString(msgStyle.Render(e.Message) + "\n\n")
		}
	}

	b.WriteString(theme.AccentStyle.Render("  [w] leave a message"))

	if m.rl != nil {
		rem := m.rl.Remaining(m.ip)
		if rem > 0 {
			mins := int(rem.Minutes()) + 1
			b.WriteString(theme.SubtleStyle.Render(
				fmt.Sprintf("  (next entry in ~%d min)", mins),
			))
		}
	}

	return b.String()
}

// refreshEntries is called by root on navigate-to-guestbook to reload the list.
func (m *GuestbookModel) refreshEntries() tea.Cmd {
	return m.loadEntries()
}
