// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// SetSelection sets a half-open selection range in rune offsets.
func (e *RichTextEdit) SetSelection(from, to RichPosition) {
	e.clearPopoverKeyboardState(false)
	e.popoverSuppressed = false
	e.popoverExplicit = false
	e.cursorInVoid = false
	e.boxSelection = nil
	e.SelectionFrom, e.SelectionTo = from, to
	e.HasSelection = from != to
	e.selectionExtending = false
}

// ClearSelection removes the current selection.
func (e *RichTextEdit) ClearSelection() {
	e.clearPopoverKeyboardState(false)
	e.popoverSuppressed = false
	e.popoverExplicit = false
	e.boxSelection = nil
	e.HasSelection = false
	e.selectionExtending = false
}

func (e *RichTextEdit) selectionLineBounds(line int) (int, int) {
	if e.boxSelection != nil {
		for _, slice := range e.boxSelection.rows {
			if slice.line == line {
				return slice.start, slice.end
			}
		}
		return 0, 0
	}
	from, to := e.selectionBounds()
	if line < from.Line || line > to.Line {
		return 0, 0
	}
	return richSelectionLineBounds(line, from, to)
}

func (e *RichTextEdit) selectionLineBoundsForRange(line int, from, to RichPosition) (int, int) {
	if e.boxSelection != nil {
		return e.selectionLineBounds(line)
	}
	if line < from.Line || line > to.Line {
		return 0, 0
	}
	return richSelectionLineBounds(line, from, to)
}

func (e *RichTextEdit) selectionBounds() (RichPosition, RichPosition) {
	from, to := e.SelectionFrom, e.SelectionTo
	if richPositionBefore(to, from) {
		from, to = to, from
	}
	return from, to
}

func (e *RichTextEdit) selectionUniformlyHas(active func(Style) bool) bool {
	lines := e.documentLines()
	from, to := e.selectionBounds()
	seen := false
	for line := from.Line; line <= to.Line && line < len(lines); line++ {
		start, end := e.selectionLineBounds(line)
		if end <= start {
			continue
		}
		offset := 0
		for _, span := range lines[line].Spans {
			n := len([]rune(span.Text))
			lo, hi := max(start, offset), min(end, offset+n)
			if lo < hi {
				seen = true
				if !active(e.spanStyle(span)) {
					return false
				}
			}
			offset += n
		}
	}
	return seen
}

func richPositionBefore(a, b RichPosition) bool {
	return a.Line < b.Line || a.Line == b.Line && a.Offset < b.Offset
}

// wordRange returns the word containing pos on its line. Whitespace, punctuation,
// and the line end have no word.
func (e *RichTextEdit) wordRange(pos RichPosition) (from, to RichPosition, ok bool) {
	lines := e.documentLines()
	e.clampPosition(&pos, lines)
	units := richLineUnits(lines[pos.Line])
	index := -1
	for i, unit := range units {
		if unit.start <= pos.Offset && pos.Offset < unit.end {
			index = i
			break
		}
	}
	if index < 0 || !richUnitWord(units[index]) {
		return pos, pos, false
	}
	first, last := index, index
	for first > 0 && richUnitWord(units[first-1]) {
		first--
	}
	for last+1 < len(units) && richUnitWord(units[last+1]) {
		last++
	}
	return RichPosition{Line: pos.Line, Offset: units[first].start}, RichPosition{Line: pos.Line, Offset: units[last].end}, true
}

func (e *RichTextEdit) selectWord() bool {
	from, to, ok := e.wordRange(e.Cursor)
	if !ok {
		return false
	}
	e.selectRange(from, to)
	return true
}

func (e *RichTextEdit) selectLine() {
	line := e.Cursor.Line
	e.selectRange(RichPosition{Line: line}, RichPosition{Line: line, Offset: richLineRuneCount(e.documentLines()[line])})
}

func (e *RichTextEdit) selectRange(from, to RichPosition) {
	e.SetSelection(from, to)
	e.selectionAnchor = from
	e.Cursor = to
}
