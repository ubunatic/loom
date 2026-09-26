// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// StyledRows renders rows of ANSI-formatted text while preserving SGR styles.
// Each input string occupies one canvas row; rows are clipped and never wrap.
type StyledRows struct {
	Lines []string
}

// NewStyledRows creates a widget for the supplied ANSI-formatted rows.
func NewStyledRows(lines ...string) *StyledRows {
	return &StyledRows{Lines: append([]string(nil), lines...)}
}

// Draw renders styled cells into r, clipping complete display-width clusters
// at its right edge. ParseANSI returns style on each cell, so a row cannot
// leave terminal style state behind for subsequent rows or canvas content.
func (s *StyledRows) Draw(c *Canvas, r Rect) {
	if s == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	for rowIndex, line := range s.Lines {
		if rowIndex >= r.H {
			break
		}
		y := r.Y + rowIndex
		x := r.X
		for _, cell := range ParseANSI(line) {
			if cell.Continuation {
				continue
			}
			width := StringWidth(cell.Text)
			if width <= 0 {
				continue
			}
			if x+width > r.X+r.W {
				break
			}
			c.Set(x, y, cell)
			x += width
		}
	}
}

// HandleKey does not consume keyboard events.
func (s *StyledRows) HandleKey(KeyEvent) bool { return false }

// HandleMouse does not consume mouse events.
func (s *StyledRows) HandleMouse(MouseEvent) bool { return false }
