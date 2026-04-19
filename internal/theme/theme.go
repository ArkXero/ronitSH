package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Color palette -- Civic Cycle teal/amber carried over for visual through-line.
// We use the dark-terminal variants since SSH portfolio visitors nearly always
// have dark terminals.
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
	Muted        = col("#555555")
)

// Base styles.
var (
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

	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Primary).
				PaddingBottom(1)

	// ChipShipped, ChipInProgress, ChipArchived used for project status badges.
	// Use symbols + color so they work without color support too.
	ChipShipped = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ChipInProgress = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	ChipArchived = lipgloss.NewStyle().
			Foreground(Muted)
)

// SidebarWidth is the column width of the left navigation pane.
const SidebarWidth = 25
