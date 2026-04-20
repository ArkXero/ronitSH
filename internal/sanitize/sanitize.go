// Package sanitize strips ANSI escape codes and control characters from
// user-provided strings before they are stored or displayed.
//
// Security note: raw guestbook input MUST pass through Sanitize before being
// stored in SQLite or rendered to any terminal. A crafted entry containing
// escape sequences could corrupt every future visitor's terminal session.
package sanitize

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ansiRE matches ANSI/VT escape sequences:
//   - CSI sequences: ESC [ ... final-byte
//   - OSC sequences: ESC ] ... ST (BEL or ESC \)
//   - Single-char escapes: ESC followed by one non-CSI/OSC byte
var ansiRE = regexp.MustCompile(
	`\x1b\[[0-9;?]*[A-Za-z]` + // CSI
		`|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)` + // OSC
		`|\x1b[^[\]]`, // other single-char
)

// Sanitize removes ANSI escape sequences, C0/C1 control characters, and null
// bytes from s. Newlines (\n) are preserved. Printable ASCII and Unicode
// letters/digits/punctuation are allowed through.
//
// The result is safe to store in SQLite and render to any terminal.
func Sanitize(s string) string {
	// 1. Strip ANSI sequences.
	s = ansiRE.ReplaceAllString(s, "")

	// 2. Strip remaining control characters (C0 except \n, and C1 0x80-0x9F).
	// We iterate with utf8.DecodeRuneInString so that invalid UTF-8 bytes
	// (which Go's range would silently promote to U+FFFD) are detected and
	// dropped -- they fall in the C1 range (0x80-0x9F) by definition.
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		switch {
		case r == utf8.RuneError && size == 1:
			// Invalid UTF-8 byte -- drop (covers raw C1 bytes 0x80-0x9F).
		case r == '\n':
			b.WriteRune(r)
		case r == '\r':
			// Normalise CRLF to LF; skip bare CR.
		case r == 0x00:
			// Null byte -- drop.
		case r < 0x20:
			// C0 control character -- drop.
		case r >= 0x7F && r <= 0x9F:
			// DEL and C1 control range -- drop.
		case unicode.IsControl(r):
			// Any other control character -- drop.
		default:
			b.WriteRune(r)
		}
	}

	return strings.TrimSpace(b.String())
}

// TruncateRunes truncates s to at most n runes.
func TruncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
