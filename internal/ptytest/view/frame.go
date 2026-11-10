// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package view

import (
	"fmt"
	"strings"

	"ubunatic.com/loom/vt"
)

// Frame holds a virtual terminal and a set of static chrome lines that appear above it.
type Frame struct {
	Child  *vt.VT
	Chrome []string
}

// isTrailingBlank returns true if a cell has no visible character and no custom styling.
func isTrailingBlank(c vt.Cell) bool {
	if c.Text != "" && c.Text != " " {
		return false
	}
	if c.Rune != ' ' && c.Rune != 0 {
		return false
	}
	if c.Style.Reverse {
		return false
	}
	_, effBG := c.Style.Effective()
	if effBG.Type != vt.ColorDefault {
		return false
	}
	if c.Style.Underline || c.Style.Strike {
		return false
	}
	return true
}

// Render returns an ANSI escape sequence string that draws the chrome lines
// and the terminal screen.
func (f *Frame) Render() string {
	var b strings.Builder

	// Switch to alt screen buffer if child uses it
	if f.Child.AltScreen() {
		b.WriteString("\x1b[?1049h")
		b.WriteString("\x1b[0m\x1b[H\x1b[2J") // Reset, Home, Clear screen
	} else {
		b.WriteString("\x1b[?1049l")
		b.WriteString("\x1b[0m\x1b[H\x1b[2J") // Reset, Home, Clear screen
	}

	chromeOffset := 0

	if !f.Child.AltScreen() {
		// Draw Chrome only on main screen
		for y, line := range f.Chrome {
			fmt.Fprintf(&b, "\x1b[%d;1H%s", y+1, line)
		}
		chromeOffset = len(f.Chrome)
	}

	cells := f.Child.Cells()
	for y, row := range cells {
		last := -1
		for x := len(row) - 1; x >= 0; x-- {
			if !isTrailingBlank(row[x]) {
				last = x
				break
			}
		}
		if last >= 0 {
			fmt.Fprintf(&b, "\x1b[%d;1H", y+1+chromeOffset)
			curStyle := vt.Style{}
			for x := 0; x <= last; x++ {
				cell := row[x]
				if false {
					continue
				}
				if cell.Style != curStyle {
					b.WriteString(cell.Style.SGR())
					curStyle = cell.Style
				}
				if cell.Text != "" {
					b.WriteString(cell.Text)
				} else if cell.Rune != 0 {
					b.WriteRune(cell.Rune)
				} else {
					b.WriteByte(' ')
				}
			}
			if curStyle != (vt.Style{}) {
				b.WriteString("\x1b[0m")
			}
		}
	}

	cx, cy := f.Child.Cursor()
	fmt.Fprintf(&b, "\x1b[%d;%dH", cy+1+chromeOffset, cx+1)

	if f.Child.CursorVisible {
		b.WriteString("\x1b[?25h")
	} else {
		b.WriteString("\x1b[?25l")
	}

	if pen := f.Child.Pen(); pen != (vt.Style{}) {
		b.WriteString(pen.SGR())
	}

	return b.String()
}
