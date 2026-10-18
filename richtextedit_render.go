// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"ubunatic.com/loom/measure"
)

// Draw renders visible lines and styled spans, scrolls the cursor into view,
// paints selected text, and places the terminal cursor when focused.
func (e *RichTextEdit) Draw(c *Canvas, r Rect) {
	if c == nil {
		return
	}
	if e.ViewMode {
		e.clearPopoverKeyboardState(false)
	}
	fullRect := r
	if e.ShowFileBar && r.H > 0 {
		r.H--
	}
	e.lastRect = r
	e.popoverButtons = nil
	if r.W <= 0 || fullRect.H <= 0 {
		return
	}
	if r.H <= 0 {
		if e.ShowFileBar {
			e.ensureFileBar().Draw(c, fullRect)
		}
		return
	}
	c.PaintSurface(r, Style{})
	lines := e.documentLines()
	if !e.cursorInVoid {
		e.clampPosition(&e.Cursor, lines)
		e.Cursor = e.normalizePosition(e.Cursor, 0)
	}
	cursorLine, cursorCol := e.cursorCoordinates(lines)
	if cursorLine < e.ScrollY {
		e.ScrollY = cursorLine
	}
	if cursorLine >= e.ScrollY+r.H {
		e.ScrollY = cursorLine - r.H + 1
	}
	if cursorCol < e.ScrollX {
		e.ScrollX = cursorCol
	}
	if cursorCol >= e.ScrollX+r.W {
		e.ScrollX = cursorCol - r.W + 1
	}
	e.ScrollX = max(0, e.ScrollX)
	maxScrollY := max(0, len(lines)-r.H)
	if e.cursorInVoid {
		maxScrollY = max(maxScrollY, cursorLine-r.H+1)
	}
	e.ScrollY = min(max(0, e.ScrollY), maxScrollY)

	for row := 0; row < r.H && e.ScrollY+row < len(lines); row++ {
		e.drawLine(c, r, lines[e.ScrollY+row], e.ScrollY+row)
	}
	e.drawPopover(c, r, lines)
	if e.ShowCursor && e.focused && !e.ViewMode {
		x := cursorCol - e.ScrollX
		y := cursorLine - e.ScrollY
		if x >= 0 && x < r.W && y >= 0 && y < r.H {
			c.CursorX = r.X + x
			c.CursorY = r.Y + y
			c.CursorShape = e.CursorShape()
			c.CursorShapeSet = true
		}
	}
	if e.ShowFileBar {
		e.ensureFileBar().Draw(c, fullRect)
	}
	if e.savePopup != nil {
		e.savePopup.Width = min(SpeccedDefaults.RichTextEdit.SavePopupMaxWidth, max(4, r.W-4))
		e.savePopup.Height = min(SpeccedDefaults.RichTextEdit.SavePopupMaxHeight, max(3, r.H-2))
		e.savePopup.Draw(c, r)
	}
	if e.helpPopup != nil {
		e.helpPopup.Width = min(SpeccedDefaults.RichTextEdit.HelpMaxWidth, max(4, r.W-2))
		e.helpPopup.Height = min(SpeccedDefaults.RichTextEdit.HelpMaxHeight, max(3, r.H))
		e.helpPopup.Draw(c, r)
	}
}

// RichTextHighlight marks a half-open range in document rune coordinates.
// Style controls highlighted cells without changing saved content or selection.
type RichTextHighlight struct {
	From, To RichPosition
	Style    Style
}

func (e *RichTextEdit) drawLine(c *Canvas, rect Rect, line RichLine, lineIndex int) {
	start, end := e.selectionLineBounds(lineIndex)
	col, offset := 0, 0
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			w := measure.StringWidth(cluster)
			runeCount := len([]rune(cluster))
			selected := e.HasSelection && offset < end && offset+runeCount > start
			x := rect.X + col - e.ScrollX
			if w > 0 && x >= rect.X && x+w <= rect.X+rect.W {
				style := e.spanStyle(span)
				for _, highlight := range e.Highlights {
					from, to := highlight.From, highlight.To
					if (lineIndex > from.Line || lineIndex == from.Line && offset+runeCount > from.Offset) &&
						(lineIndex < to.Line || lineIndex == to.Line && offset < to.Offset) {
						style = richOverlayStyle(style, highlight.Style)
					}
				}
				if selected {
					style = e.selectionStyle(style)
				}
				c.Set(x, rect.Y+lineIndex-e.ScrollY, Cell{Text: cluster, Style: style})
			}
			col += w
			offset += runeCount
		}
	}
}

// spanStyle returns the draw style of span; link spans get the spec link style
// for any attribute the span leaves unset, in edit and view mode, focused or not.
func (e *RichTextEdit) spanStyle(span RichSpan) Style {
	style := span.Style
	if span.Link == "" {
		return style
	}
	defs := SpeccedDefaults.RichTextEdit
	if style.FG == ColorReset() {
		style.FG = ColorIndex(uint8(defs.LinkFG))
	}
	style.Underline = style.Underline || defs.LinkUnderline
	return style
}

func (e *RichTextEdit) selectionStyle(base Style) Style {
	return richOverlayStyle(base, e.SelectionStyle)
}

func richOverlayStyle(base, selection Style) Style {
	if selection == (Style{}) {
		base.BG = ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SelectionBG))
		return base
	}
	if selection.FG != ColorReset() {
		base.FG = selection.FG
	}
	if selection.BG != ColorReset() {
		base.BG = selection.BG
	}
	base.Bold = base.Bold || selection.Bold
	base.Dim = base.Dim || selection.Dim
	base.Italic = base.Italic || selection.Italic
	base.Underline = base.Underline || selection.Underline
	base.Strike = base.Strike || selection.Strike
	base.Invert = base.Invert || selection.Invert
	return base
}
