// Package sanitize strips ANSI escape codes and control characters from user input.
// This is a security-critical package -- guestbook input MUST pass through here
// before storage or display to prevent terminal escape injection.
package sanitize
