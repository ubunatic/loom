// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"ubunatic.com/loom/measure"
)

// Draw renders visible lines and styled spans, scrolls the cursor (or, while
// unfocused, a revealed highlight) into view, paints selected text, and places the terminal cursor when focused.
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
	// The view follows the cursor, or a revealed highlight while unfocused.
	followLine, followCol := cursorLine, cursorCol
	if line, col, ok := e.revealCoordinates(lines); ok {
		followLine, followCol = line, col
	}
	if followLine < e.ScrollY {
		e.ScrollY = followLine
	}
	if followLine >= e.ScrollY+r.H {
		e.ScrollY = followLine - r.H + 1
	}
	if followCol < e.ScrollX {
		e.ScrollX = followCol
	}
	if followCol >= e.ScrollX+r.W {
		e.ScrollX = followCol - r.W + 1
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
	if e.DeferPopover {
		// DrawPopover rebuilds the hit areas; none may outlive this frame.
		e.resetPopoverHitAreas()
	} else {
		e.drawPopover(c, r, lines)
	}
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

// DrawPopover paints the format popover, its submenus and its pointer. Hosts
// that set DeferPopover call it after painting their own overlays, with the
// rectangle they gave Draw, so the popover is drawn above everything it
// captures clicks over. It does nothing for a default editor, whose Draw
// already painted the popover, while a help or save popup is open (those are
// drawn above the popover), or when no popover is open.
func (e *RichTextEdit) DrawPopover(c *Canvas, r Rect) {
	if e == nil || c == nil || !e.DeferPopover || e.helpPopup != nil || e.savePopup != nil {
		return
	}
	if e.ShowFileBar && r.H > 0 {
		r.H--
	}
	if r.W <= 0 || r.H <= 0 {
		e.resetPopoverHitAreas()
		return
	}
	e.drawPopover(c, r, e.documentLines())
}

func (e *RichTextEdit) resetPopoverHitAreas() {
	e.popoverButtons = nil
	e.popoverBoxChoices = nil
	e.popoverHamburgerChoices = nil
	e.popoverColorRect = Rect{}
}

// RichTextHighlight marks a half-open range in document rune coordinates.
// Style controls highlighted cells without changing saved content or selection.
// Reveal asks Draw to keep the start of the range in view while the editor is
// unfocused, instead of the cursor; the cursor itself never moves.
type RichTextHighlight struct {
	From, To RichPosition
	Style    Style
	Reveal   bool
}

// revealCoordinates returns the display line and column of the first
// highlight that asks to be revealed. Only an unfocused editor follows it: a
// focused editor always keeps its cursor in view. The position is clamped to
// the document, and the cursor is never touched.
func (e *RichTextEdit) revealCoordinates(lines []RichLine) (line, col int, ok bool) {
	if e.focused || len(lines) == 0 {
		return 0, 0, false
	}
	for _, highlight := range e.Highlights {
		if !highlight.Reveal {
			continue
		}
		line = min(max(0, highlight.From.Line), len(lines)-1)
		offset := min(max(0, highlight.From.Offset), richLineRuneCount(lines[line]))
		return line, richLineColumn(lines[line], offset), true
	}
	return 0, 0, false
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
						style = richOverlayStyle(style, highlight.Style, e.toolbarTheme().Selection.Color())
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
	return richOverlayStyle(base, e.SelectionStyle, e.toolbarTheme().Selection.Color())
}

func richOverlayStyle(base, selection Style, selectionBG Color) Style {
	if selection == (Style{}) {
		base.BG = selectionBG
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
