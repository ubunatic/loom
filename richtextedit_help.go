// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"regexp"
	"strings"
)

type richTextEditHelp struct {
	lines     []string
	scroll    int
	maxScroll int
}

var helpModifierCombo = regexp.MustCompile(`\b((?:Ctrl\+|Shift\+|Alt\+)+)(\w+)`)

// richTextEditHelpLines returns the help text; written modifier combos such as
// Ctrl+Shift+Z are rendered through KeyCap so help matches the hint bars.
func richTextEditHelpLines() []string {
	defs := SpeccedDefaults.RichTextEdit
	lines := []string{
		"F1: open this help",
		"File menu: Alt+F open; Left/Right choose menu; Up/Down choose action; Enter/Space run; Escape close",
		"File actions: " + KeyCap(defs.HotkeySaveBinding) + " Save; " + KeyCap(defs.HotkeySaveAsBinding) + " Save as",
		"Mode: F7 toggle View/Edit; F5 box selection or toggle box drawing",
		"Move: Left/Right/Up/Down; Ctrl+Left and Ctrl+Right move by word; Home/Ctrl+A start; End/Ctrl+E end",
		"Select: Shift+arrows; Shift+Home and Shift+End; Ctrl+Shift+A/E extend to line start/end",
		"Text: printable keys insert; Enter/Return newline; Backspace/Delete erase",
		"Style: Ctrl+B bold; Ctrl+I italic; Ctrl+U underline",
		"Format: Ctrl+Space opens selection formatting popover, separate from File actions",
		"Popover: Tab/Shift+Tab or Left/Right choose; Enter/Space applies; Escape closes",
		"Popover: B/I/U/S, Link, #FG/#BG colors, Box styles, and Draw are available",
		"Clipboard: Ctrl+C/Ctrl+Insert copy; Ctrl+X/Shift+Delete cut; Ctrl+V/Shift+Insert paste",
		"History: Ctrl+Z/Ctrl+Y undo; Ctrl+R/Ctrl+Shift+Y/Ctrl+Shift+Z redo",
		"Selection: mouse drag; double-click word; triple-click line",
		"Box drawing: arrows draw connected lines; Escape ends a stroke",
		"Save as: Tab/Shift+Tab switch search and filename; click either field",
		"Save as: type to search or name; arrows navigate; Enter opens folder or saves",
		"Save as: Backspace edits search or moves to parent; Escape cancels",
	}
	for i, line := range lines {
		lines[i] = helpModifierCombo.ReplaceAllStringFunc(line, func(combo string) string {
			return KeyCap(strings.ToLower(strings.ReplaceAll(combo, "+", "-")))
		})
	}
	return lines
}

func newRichTextEditHelp() *richTextEditHelp {
	return &richTextEditHelp{lines: richTextEditHelpLines()}
}

func (h *richTextEditHelp) ContentHeight() int { return len(h.lines) + 1 }

func (h *richTextEditHelp) Draw(c *Canvas, r Rect) {
	if c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	lines := make([]string, 0, len(h.lines))
	for _, line := range h.lines {
		lines = append(lines, wrapRichTextHelpLine(line, r.W)...)
	}
	visible := max(0, r.H-1)
	maxScroll := max(0, len(lines)-visible)
	h.maxScroll = maxScroll
	h.scroll = min(max(0, h.scroll), maxScroll)
	for row := 0; row < visible; row++ {
		y := r.Y + row
		c.PaintSurface(Rect{X: r.X, Y: y, W: r.W, H: 1}, Style{})
		if index := h.scroll + row; index < len(lines) {
			c.Write(r.X, y, lines[index], Style{})
		}
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

func (h *richTextEditHelp) ConsumeMouse(MouseEvent) EventResult { return Ignored() }

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
