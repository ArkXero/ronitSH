package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Color palette -- Civic Cycle teal/amber carried over for visual through-line.
// SSH visitors almost universally use dark terminals, so we use the dark variants.
func col(hex string) color.Color {
	return lipgloss.Color(hex)
}

var (
	Primary      = col("#1A8A9A")
	PrimaryLight = col("#2BBDD4")
	PrimaryDark  = col("#0D5E6B")
	Accent       = col("#F5A623")
	Subtle       = col("#888888")
	Success      = col("#4CAF50")
	Warning      = col("#F5A623")
	Muted        = col("#444444")
)

// letterSpace inserts a single space between each rune, giving the wide
// letter-spaced look used for section titles.
func letterSpace(s string) string {
	runes := []rune(strings.ToUpper(s))
	if len(runes) == 0 {
		return s
	}
	var b strings.Builder
	for i, r := range runes {
		b.WriteRune(r)
		if i < len(runes)-1 {
			b.WriteRune(' ')
		}
	}
	return b.String()
}

// Base styles.
var (
	// SectionTitleStyle is used for top-level section headings (about, projects, etc.)
	// Uppercase, bold, teal, letter-spaced per spec.
	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Primary).
				Transform(letterSpace).
				MarginBottom(1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	ActiveItemStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryLight)

	InactiveItemStyle = lipgloss.NewStyle()

	SubtleStyle = lipgloss.NewStyle().
			Foreground(Subtle)

	AccentStyle = lipgloss.NewStyle().
			Foreground(Accent)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Status badge styles -- use symbol AND color so NO_COLOR terminals still work.
	ChipShipped = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ChipInProgress = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	ChipArchived = lipgloss.NewStyle().
			Foreground(Muted)
)

// SidebarWidth is the column count of the left navigation pane.
const SidebarWidth = 25

// PreviewWidth is the column count of the right preview pane in wide mode.
const PreviewWidth = 30
