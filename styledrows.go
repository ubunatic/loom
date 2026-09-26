// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"unicode/utf8"
)

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
// at its right edge. The allocated rectangle is cleared first, so a shorter
// redraw cannot leave old text or styling behind.
func (s *StyledRows) Draw(c *Canvas, r Rect) {
	if s == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			c.Set(x, y, Cell{Text: " ", Style: Reset, Claim: true})
		}
	}
	for rowIndex, line := range s.Lines {
		if rowIndex >= r.H {
			break
		}
		y := r.Y + rowIndex
		x := r.X
		for _, cell := range ParseANSI(styledRowSGR(line)) {
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

// styledRowSGR keeps complete SGR sequences and printable text, dropping
// other terminal controls. Incomplete controls consume the remainder so
// malformed input cannot turn escape syntax into visible text.
func styledRowSGR(line string) string {
	var out strings.Builder
	out.Grow(len(line))
	for i := 0; i < len(line); {
		if line[i] == 0x1b {
			if i+1 >= len(line) {
				break
			}
			switch line[i+1] {
			case '[':
				j := i + 2
				for j < len(line) && (line[j] < 0x40 || line[j] > 0x7e) {
					j++
				}
				if j == len(line) {
					return out.String()
				}
				if line[j] == 'm' {
					out.WriteString(line[i : j+1])
				}
				i = j + 1
			case ']':
				j := i + 2
				for j < len(line) {
					if line[j] == '\a' {
						j++
						break
					}
					if line[j] == 0x1b && j+1 < len(line) && line[j+1] == '\\' {
						j += 2
						break
					}
					j++
				}
				i = j
			default:
				i += 2
			}
			continue
		}
		r, size := utf8.DecodeRuneInString(line[i:])
		if r >= 0x20 && r != 0x7f {
			out.WriteString(line[i : i+size])
		}
		i += size
	}
	return out.String()
}

// HandleKey does not consume keyboard events.
func (s *StyledRows) HandleKey(KeyEvent) bool { return false }

// HandleMouse does not consume mouse events.
func (s *StyledRows) HandleMouse(MouseEvent) bool { return false }
