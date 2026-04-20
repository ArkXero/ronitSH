package sanitize_test

import (
	"testing"

	"github.com/ArkXero/termfolio/internal/sanitize"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text passes through",
			input: "Hello, world!",
			want:  "Hello, world!",
		},
		{
			name:  "newline preserved",
			input: "line one\nline two",
			want:  "line one\nline two",
		},
		{
			name:  "ANSI color code stripped",
			input: "\x1b[31mred text\x1b[0m",
			want:  "red text",
		},
		{
			name:  "cursor movement stripped",
			input: "\x1b[2J\x1b[H clear screen",
			want:  "clear screen",
		},
		{
			name:  "OSC window title stripped",
			input: "\x1b]0;evil title\x07normal",
			want:  "normal",
		},
		{
			name:  "null bytes stripped",
			input: "hel\x00lo",
			want:  "hello",
		},
		{
			name:  "C0 control chars stripped except newline",
			input: "a\x01\x07\x08\x0d\x1fb",
			want:  "ab",
		},
		{
			name:  "DEL stripped",
			input: "a\x7fb",
			want:  "ab",
		},
		{
			name:  "C1 range stripped",
			input: "a\x80\x9fb",
			want:  "ab",
		},
		{
			name:  "CRLF normalised to nothing extra",
			input: "line\r\nend",
			want:  "line\nend",
		},
		{
			name:  "mixed attack payload",
			input: "\x1b[31m\x1b]0;pwned\x07\x1b[2Jhello\x00world",
			want:  "helloworld",
		},
		{
			name:  "unicode text preserved",
			input: "Cafe au lait and 500 yen",
			want:  "Cafe au lait and 500 yen",
		},
		{
			name:  "leading/trailing whitespace trimmed",
			input: "  hello  ",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitize.Sanitize(tt.input)
			if got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := sanitize.TruncateRunes("hello", 3); got != "hel" {
		t.Errorf("got %q, want %q", got, "hel")
	}
	if got := sanitize.TruncateRunes("hi", 10); got != "hi" {
		t.Errorf("got %q, want %q", got, "hi")
	}
}
