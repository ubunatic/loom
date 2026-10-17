// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// copySelectionOrWord stores the selection, or the word at the cursor, in the
// internal clipboard. It reports whether anything was copied; a word copy
// selects the word so that cut can delete it.
func (e *RichTextEdit) copySelectionOrWord() bool {
	if !e.HasSelection && !e.selectWord() {
		return false
	}
	from, to := e.selectionBounds()
	e.clampPosition(&from, e.Document.Lines)
	e.clampPosition(&to, e.Document.Lines)
	from, to = expandRangeForPills(e.Document.Lines, from, to)
	if from == to {
		return false
	}
	e.clipboard = nil
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		line := e.Document.Lines[lineIndex]
		start, end := e.selectionLineBounds(lineIndex)
		end = min(end, richLineRuneCount(line))
		_, rest := splitRichLine(line, start)
		part, _ := splitRichLine(RichLine{Spans: rest}, end-start)
		e.clipboard = append(e.clipboard, RichLine{Spans: part})
	}
	return true
}

// insertRich inserts styled lines at the cursor, replacing any selection.
func (e *RichTextEdit) insertRich(clip []RichLine) {
	if len(clip) == 0 {
		return
	}
	e.materializeVoidCursor()
	e.deleteSelection()
	e.Cursor = e.normalizePosition(e.Cursor, 0)
	lines := e.Document.Lines
	left, right := splitRichLine(lines[e.Cursor.Line], e.Cursor.Offset)
	last := len(clip) - 1
	origin := e.Cursor.Line
	updated := make([]RichLine, 0, len(lines)+last)
	updated = append(updated, lines[:origin]...)
	first := append(append([]RichSpan(nil), left...), clip[0].Spans...)
	if last == 0 {
		e.Cursor.Offset += richLineRuneCount(clip[0])
		updated = append(updated, RichLine{Spans: mergeRichSpans(append(first, right...))})
	} else {
		updated = append(updated, RichLine{Spans: mergeRichSpans(first)})
		for _, line := range clip[1:last] {
			updated = append(updated, RichLine{Spans: append([]RichSpan(nil), line.Spans...)})
		}
		tail := append(append([]RichSpan(nil), clip[last].Spans...), right...)
		updated = append(updated, RichLine{Spans: mergeRichSpans(tail)})
		e.Cursor.Line += last
		e.Cursor.Offset = richLineRuneCount(clip[last])
	}
	updated = append(updated, lines[origin+1:]...)
	e.Document.Lines = updated
}
