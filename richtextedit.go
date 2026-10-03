// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "ubunatic.com/loom/measure"

// RichPosition identifies a rune offset in a logical rich document line.
type RichPosition struct {
	Line   int
	Offset int
}

// RichTextEdit displays a RichDocument with a cursor, selection highlight, and
// scrolling viewport. Input handling is added in a later milestone.
type RichTextEdit struct {
	Document       *RichDocument
	Cursor         RichPosition
	SelectionFrom  RichPosition
	SelectionTo    RichPosition
	HasSelection   bool
	SelectionStyle Style
	ScrollX        int
	ScrollY        int
	ShowCursor     bool

	focused  bool
	lastRect Rect
}

var _ Widget = (*RichTextEdit)(nil)

// NewRichTextEdit creates a rich text view for doc. A nil document is treated
// as an empty document.
func NewRichTextEdit(doc *RichDocument) *RichTextEdit {
	return &RichTextEdit{Document: doc, ShowCursor: true, focused: true}
}

// Focused reports whether the editor currently has keyboard focus.
func (e *RichTextEdit) Focused() bool { return e.focused }

// SetFocus sets the editor focus state.
func (e *RichTextEdit) SetFocus(focused bool) { e.focused = focused }

// SetSelection sets a half-open selection range in rune offsets.
func (e *RichTextEdit) SetSelection(from, to RichPosition) {
	e.SelectionFrom, e.SelectionTo = from, to
	e.HasSelection = from != to
}

// ClearSelection removes the current selection.
func (e *RichTextEdit) ClearSelection() { e.HasSelection = false }

// Draw renders visible lines and styled spans, scrolls the cursor into view,
// paints selected text, and places the terminal cursor when focused.
func (e *RichTextEdit) Draw(c *Canvas, r Rect) {
	if c == nil {
		return
	}
	e.lastRect = r
	if r.W <= 0 || r.H <= 0 {
		return
	}
	c.PaintSurface(r, Style{})
	lines := e.documentLines()
	e.clampPosition(&e.Cursor, lines)
	cursorCol := richLineColumn(lines[e.Cursor.Line], e.Cursor.Offset)
	if e.Cursor.Line < e.ScrollY {
		e.ScrollY = e.Cursor.Line
	}
	if e.Cursor.Line >= e.ScrollY+r.H {
		e.ScrollY = e.Cursor.Line - r.H + 1
	}
	if cursorCol < e.ScrollX {
		e.ScrollX = cursorCol
	}
	if cursorCol >= e.ScrollX+r.W {
		e.ScrollX = cursorCol - r.W + 1
	}
	e.ScrollX = max(0, e.ScrollX)
	e.ScrollY = min(max(0, e.ScrollY), max(0, len(lines)-r.H))

	for row := 0; row < r.H && e.ScrollY+row < len(lines); row++ {
		e.drawLine(c, r, lines[e.ScrollY+row], e.ScrollY+row)
	}
	if e.ShowCursor && e.focused {
		x := cursorCol - e.ScrollX
		y := e.Cursor.Line - e.ScrollY
		if x >= 0 && x < r.W && y >= 0 && y < r.H {
			c.CursorX = r.X + x
			c.CursorY = r.Y + y
		}
	}
}

func (e *RichTextEdit) documentLines() []RichLine {
	if e.Document == nil || len(e.Document.Lines) == 0 {
		return []RichLine{{}}
	}
	return e.Document.Lines
}

func (e *RichTextEdit) clampPosition(pos *RichPosition, lines []RichLine) {
	pos.Line = min(max(0, pos.Line), len(lines)-1)
	lineRunes := 0
	for _, span := range lines[pos.Line].Spans {
		lineRunes += len([]rune(span.Text))
	}
	pos.Offset = min(max(0, pos.Offset), lineRunes)
}

func richLineColumn(line RichLine, offset int) int {
	col, consumed := 0, 0
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			clusterRunes := len([]rune(cluster))
			if consumed+clusterRunes > offset {
				return col
			}
			consumed += clusterRunes
			col += measure.StringWidth(cluster)
		}
	}
	return col
}

func (e *RichTextEdit) drawLine(c *Canvas, rect Rect, line RichLine, lineIndex int) {
	start, end := e.selectionBounds()
	col, offset := 0, 0
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			w := measure.StringWidth(cluster)
			runeCount := len([]rune(cluster))
			selected := e.HasSelection && richPositionBefore(RichPosition{Line: lineIndex, Offset: offset}, end) && richPositionBefore(start, RichPosition{Line: lineIndex, Offset: offset + runeCount})
			x := rect.X + col - e.ScrollX
			if w > 0 && x >= rect.X && x+w <= rect.X+rect.W {
				style := span.Style
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

func (e *RichTextEdit) selectionBounds() (RichPosition, RichPosition) {
	from, to := e.SelectionFrom, e.SelectionTo
	if richPositionBefore(to, from) {
		from, to = to, from
	}
	return from, to
}

func richPositionBefore(a, b RichPosition) bool {
	return a.Line < b.Line || a.Line == b.Line && a.Offset < b.Offset
}

func (e *RichTextEdit) selectionStyle(base Style) Style {
	selection := e.SelectionStyle
	if selection == (Style{}) {
		base.BG = ColorIndex(24)
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

// ConsumeKey is a stub until rich text editing commands are implemented.
func (e *RichTextEdit) ConsumeKey(KeyEvent) EventResult { return Ignored() }

// ConsumeMouse is a stub until rich text mouse selection is implemented.
func (e *RichTextEdit) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
