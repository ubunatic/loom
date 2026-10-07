// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

type richTextEditHelp struct {
	sections  []HelpSection
	scroll    int
	maxScroll int
}

// richTextEditHelpSections returns the spec-defined sections followed by the
// sections a host registered.
func richTextEditHelpSections(extra []HelpSection) []HelpSection {
	base := SpeccedDefaults.RichTextEdit.HelpSections
	return append(append([]HelpSection(nil), base...), extra...)
}

func newRichTextEditHelp(extra []HelpSection) *richTextEditHelp {
	return &richTextEditHelp{sections: richTextEditHelpSections(extra)}
}

// plainLines returns the table as plain text at the given width.
func (h *richTextEditHelp) plainLines(width int) []string {
	var lines []string
	for _, row := range KeyHelpSections(h.sections, width) {
		switch {
		case row.Header:
			lines = append(lines, row.Text)
		case row.Key == "" && row.Text == "":
			lines = append(lines, "")
		default:
			lines = append(lines, strings.TrimRight(row.Key+"  "+row.Text, " "))
		}
	}
	return lines
}

func (h *richTextEditHelp) ContentHeight() int { return len(KeyHelpSections(h.sections, 68)) + 1 }

func (h *richTextEditHelp) Draw(c *Canvas, r Rect) {
	if c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	rows := KeyHelpSections(h.sections, r.W)
	visible := max(0, r.H-1)
	maxScroll := max(0, len(rows)-visible)
	h.maxScroll = maxScroll
	h.scroll = min(max(0, h.scroll), maxScroll)
	for row := 0; row < visible; row++ {
		y := r.Y + row
		c.PaintSurface(Rect{X: r.X, Y: y, W: r.W, H: 1}, Style{})
		index := h.scroll + row
		if index >= len(rows) {
			continue
		}
		line := rows[index]
		if line.Header {
			c.Write(r.X, y, line.Text, Style{Bold: true})
			continue
		}
		c.Write(r.X, y, line.Key, Style{Bold: true})
		c.Write(r.X+StringWidth(line.Key)+2, y, TruncateText(line.Text, max(0, r.W-StringWidth(line.Key)-2), ""), Style{})
	}
	if r.H > 0 {
		y := r.Y + r.H - 1
		c.PaintSurface(Rect{X: r.X, Y: y, W: r.W, H: 1}, Style{})
		c.Write(r.X, y, TruncateText("↑/↓ Scroll · Esc Close", r.W, ""), Style{Dim: true})
	}
}

func (h *richTextEditHelp) ConsumeKey(e KeyEvent) EventResult {
	switch {
	case e.Is("up"):
		h.scroll = max(0, h.scroll-1)
		return Handled()
	case e.Is("down"):
		h.scroll++
		return Handled()
	case e.Is("pgup", "pageup"):
		h.scroll = max(0, h.scroll-5)
		return Handled()
	case e.Is("pgdn", "pgdown", "pagedown"):
		h.scroll = min(h.maxScroll, h.scroll+5)
		return Handled()
	case e.Is("home"):
		h.scroll = 0
		return Handled()
	case e.Is("end"):
		h.scroll = h.maxScroll
		return Handled()
	default:
		return Handled()
	}
}

func (h *richTextEditHelp) ConsumeMouse(e MouseEvent) EventResult {
	switch e.Action {
	case MouseScrollUp:
		h.scroll = max(0, h.scroll-1)
	case MouseScrollDown:
		h.scroll = min(h.maxScroll, h.scroll+1)
	default:
		return Ignored()
	}
	return Handled()
}

func wrapRichTextHelpLine(line string, width int) []string {
	if width <= 0 {
		return nil
	}
	var rows []string
	var row strings.Builder
	rowWidth := 0
	flush := func() {
		rows = append(rows, row.String())
		row.Reset()
		rowWidth = 0
	}
	for _, word := range strings.Fields(line) {
		wordWidth := StringWidth(word)
		if rowWidth > 0 && rowWidth+1+wordWidth <= width {
			row.WriteByte(' ')
			row.WriteString(word)
			rowWidth += 1 + wordWidth
			continue
		}
		if rowWidth > 0 {
			flush()
		}
		if wordWidth <= width {
			row.WriteString(word)
			rowWidth = wordWidth
			continue
		}
		var piece strings.Builder
		pieceWidth := 0
		for _, char := range word {
			charWidth := RuneWidth(char)
			if pieceWidth+charWidth > width && pieceWidth > 0 {
				rows = append(rows, piece.String())
				piece.Reset()
				pieceWidth = 0
			}
			piece.WriteRune(char)
			pieceWidth += charWidth
		}
		row.WriteString(piece.String())
		rowWidth = pieceWidth
	}
	if rowWidth > 0 || len(rows) == 0 {
		flush()
	}
	return rows
}

var _ Widget = (*richTextEditHelp)(nil)
var _ EventConsumer = (*richTextEditHelp)(nil)
var _ MouseConsumer = (*richTextEditHelp)(nil)
var _ ContentHeighter = (*richTextEditHelp)(nil)
