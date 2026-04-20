package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ArkXero/termfolio/internal/theme"
)

// blackhole is the 16-row × 52-col braille ASCII art.
const blackhole = `⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⢠⢀⡐⢄⢢⡐⢢⢁⠂⠄⠠⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⠀⠀⠀⠀⠀⡄⣌⠰⣘⣆⢧⡜⣮⣱⣎⠷⣌⡞⣌⡒⠤⣈⠠⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠒⠊⠀⠀⠀⠀⢀⠢⠱⡜⣞⣳⠝⣘⣭⣼⣾⣷⣶⣶⣮⣬⣥⣙⠲⢡⢂⠡⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠃⠀⠀⠀⠀⠀⠀⢀⠢⣑⢣⠝⣪⣵⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣯⣻⢦⣍⠢⢅⢂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⠀⢆⡱⠌⣡⢞⣵⣿⣿⣿⠿⠛⠛⠉⠉⠛⠛⠿⢷⣽⣻⣦⣎⢳⣌⠆⡱⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠂⠠⠌⢢⢃⡾⣱⣿⢿⡾⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⢻⣏⠻⣷⣬⡳⣤⡂⠜⢠⡀⣀⠀⠀⡀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠠⢀⠂⣌⢃⡾⢡⣿⢣⡏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹⡇⡊⣿⣿⣾⣽⣛⠶⣶⣬⣭⣥⣙⣚⢷⣶⠦⡤⢀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠠⢁⠂⠰⡌⡼⠡⣼⢃⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢻⣿⣿⣿⣿⣿⣿⣿⣿⣾⡿⠿⣛⣯⡴⢏⠳⠁⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠠⠑⡌⠀⣉⣾⣩⣼⣿⣾⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣀⣠⣤⣤⣿⣿⣿⣿⡿⢛⣛⣯⣭⠶⣞⠻⣉⠒⠀⠂⠀⠀⠀
⠀⠀⠀⠀⠀⠀⢀⣀⡶⢝⣢⣾⣿⣼⣿⣿⣿⣿⣿⣀⣼⣀⣀⣀⣤⣴⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⣿⠿⡛⠏⠍⠂⠁⢠⠁⠀⠀⠀⠀⠀⠀⠀
⠀⠠⢀⢥⣰⣾⣿⣯⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⣽⠟⣿⠐⠨⠑⡀⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⡐⢢⣟⣾⣿⣿⣟⣛⣿⣿⣿⣿⢿⣝⠻⠿⢿⣯⣛⢿⣿⣿⣿⡛⠻⠿⣛⠻⠛⡛⠩⢁⣴⡾⢃⣾⠇⢀⠡⠂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠈⠁⠊⠙⠉⠩⠌⠉⠢⠉⠐⠈⠂⠈⠁⠉⠂⠐⠉⣻⣷⣭⠛⠿⣶⣦⣤⣤⣴⣴⡾⠟⣫⣾⣿⡏⠀⠂⠐⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⢻⢿⢶⣤⣬⣉⣉⣭⣤⣴⣿⣿⡿⠃⠄⡈⠁⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠁⠘⢊⠳⠭⡽⣿⠿⠿⠟⠛⠉⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠀⠁⠈⠐⠀⠘⠀⠈⠀⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀`

// ronitArt is "RONIT" in 4-wide × 5-tall block pixel letters.
// Total: 5 rows × 24 cols. Centered in the footer's right pane.
const ronitArt = `███   ██  █  █ ████ ████
█  █ █  █ ██ █  ██   ██
███  █  █ █ ██  ██   ██
██   █  █ █  █  ██   ██
█  █  ██  █  █ ████  ██ `

type footerTickMsg struct{}

func footerTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return footerTickMsg{}
	})
}

// DecoFooterModel is the persistent animated footer shown on all views.
// Left: rainbow-animated ASCII blackhole. Right: "RONIT" in block letters.
type DecoFooterModel struct {
	frame int
	width int
}

// NewDecoFooterModel constructs the footer model with a default width.
func NewDecoFooterModel() DecoFooterModel {
	return DecoFooterModel{width: 80}
}

func (m DecoFooterModel) Init() tea.Cmd {
	return footerTick()
}

func (m DecoFooterModel) Update(msg tea.Msg) (DecoFooterModel, tea.Cmd) {
	switch msg := msg.(type) {
	case footerTickMsg:
		m.frame = (m.frame + 1) % 360
		return m, footerTick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

func (m DecoFooterModel) View() string {
	const bhWidth = 52

	// renderRainbow produces raw ANSI -- use it directly, no lipgloss wrapper.
	blackholePane := renderRainbow(blackhole, m.frame)

	rightWidth := m.width - bhWidth - 1
	if rightWidth < 10 {
		return blackholePane
	}

	namePane := lipgloss.Place(
		rightWidth, theme.DecoFooterHeight,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(theme.PrimaryLight).Render(ronitArt),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, blackholePane, namePane)
}

// renderRainbow applies per-character HSL color cycling to a braille string
// using raw ANSI 24-bit color escape sequences. We intentionally bypass
// lipgloss per-character Render, which in v2 can append trailing sequences
// (newlines, resets) that corrupt multi-line layout.
func renderRainbow(s string, frameOffset int) string {
	var sb strings.Builder
	charIndex := 0
	for _, r := range s {
		if r == '\n' {
			sb.WriteRune(r)
			continue
		}
		// Empty braille cell (U+2800) and space -- pass through uncolored.
		if r == '\u2800' || r == ' ' {
			sb.WriteRune(r)
			charIndex++
			continue
		}
		hue := math.Mod(float64(charIndex)*2.5+float64(frameOffset)*8, 360)
		ri, gi, bi := hslToRGB(hue, 90, 65)
		// Write: ESC[38;2;r;g;bm <char> ESC[0m
		fmt.Fprintf(&sb, "\033[38;2;%d;%d;%dm%s\033[0m", ri, gi, bi, string(r))
		charIndex++
	}
	return sb.String()
}

// hslToRGB converts HSL (h: 0-360, s: 0-100, l: 0-100) to 0-255 r,g,b values.
func hslToRGB(h, s, l float64) (int, int, int) {
	s /= 100
	l /= 100
	a := s * math.Min(l, 1-l)
	f := func(n float64) int {
		k := math.Mod(n+h/30, 12)
		c := l - a*math.Max(math.Min(k-3, math.Min(9-k, 1)), -1)
		return int(math.Round(255 * c))
	}
	return f(0), f(8), f(4)
}
