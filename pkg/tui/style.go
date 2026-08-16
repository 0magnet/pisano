package tui

import "strings"

// ANSI styling, written out directly rather than through a styling library.
//
// The canvas renderers already emit raw SGR codes for the path colours, so the
// chrome using the same mechanism means one way of doing it rather than two.
// It also drops the dependency chain — lipgloss, and termenv underneath it —
// whose package initialisation queried the terminal for its background colour
// and blocked for five seconds when nothing answered.
//
// These are the standard bright ANSI slots rather than fixed RGB, so they keep
// following whatever palette the terminal is themed with.
const (
	sgrReset = "\x1b[0m"
	sgrTitle = "\x1b[1;94m" // bold bright blue
	sgrDim   = "\x1b[90m"   // bright black
	sgrKey   = "\x1b[96m"   // bright cyan
	sgrWarn  = "\x1b[93m"   // bright yellow
)

// paint wraps s unless colour is off. Nested spans reset back to the outer
// style rather than to plain, so a highlighted key inside a dim line leaves the
// rest of that line dim.
func (m Model) paint(code, s string) string {
	if !m.color || s == "" {
		return s
	}
	return code + s + sgrReset
}

func (m Model) title(s string) string { return m.paint(sgrTitle, s) }
func (m Model) dim(s string) string   { return m.paint(sgrDim, s) }
func (m Model) warn(s string) string  { return m.paint(sgrWarn, s) }

// key highlights a keystroke inside a dim run, so it has to restore the dim
// rather than reset to nothing.
func (m Model) key(s string) string {
	if !m.color {
		return s
	}
	return sgrKey + s + sgrReset + sgrDim
}

// height counts the lines a block occupies, which is what the layout needs and
// all that was ever wanted from lipgloss.Height.
func height(s string) int { return strings.Count(s, "\n") + 1 }
