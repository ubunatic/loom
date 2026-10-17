// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"ubunatic.com/loom/measure"
	"unicode"
)

func (e *RichTextEdit) documentLines() []RichLine {
	if e.Document == nil || len(e.Document.Lines) == 0 {
		return []RichLine{{}}
	}
	return e.Document.Lines
}

func (e *RichTextEdit) cursorCoordinates(lines []RichLine) (int, int) {
	if e.cursorInVoid {
		return e.voidLine, e.voidColumn
	}
	return e.Cursor.Line, richLineColumn(lines[e.Cursor.Line], e.Cursor.Offset)
}

func (e *RichTextEdit) materializeVoidCursor() {
	if !e.cursorInVoid {
		return
	}
	e.ensureDocument()
	for len(e.Document.Lines) <= e.voidLine {
		e.Document.Lines = append(e.Document.Lines, RichLine{})
	}
	line := e.Document.Lines[e.voidLine]
	width := richLineColumn(line, richLineRuneCount(line))
	if width < e.voidColumn {
		line.Spans = mergeRichSpans(append(line.Spans, RichSpan{Text: strings.Repeat(" ", e.voidColumn-width)}))
	}
	e.Document.Lines[e.voidLine] = line
	e.Cursor = RichPosition{Line: e.voidLine, Offset: richLineOffsetAtColumn(line, e.voidColumn)}
	e.cursorInVoid = false
}

func (e *RichTextEdit) moveToVoid(line, column int, extend bool) {
	if !e.GhostCursorEnabled {
		return
	}
	if line < 0 || column < 0 {
		return
	}
	if !extend {
		e.ClearSelection()
	}
	e.cursorInVoid = true
	e.voidLine, e.voidColumn = line, column
	e.boxSelection = nil
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

type richTextUnit struct {
	start, end int
	text       string
	pill       bool
}

func richLineStops(line RichLine) []int {
	stops := []int{0}
	for _, unit := range richLineUnits(line) {
		stops = append(stops, unit.end)
	}
	return stops
}

func (e *RichTextEdit) normalizePosition(pos RichPosition, direction int) RichPosition {
	lines := e.documentLines()
	e.clampPosition(&pos, lines)
	line := lines[pos.Line]
	for _, unit := range richLineUnits(line) {
		if pos.Offset > unit.start && pos.Offset < unit.end {
			if direction < 0 {
				pos.Offset = unit.start
			} else if direction > 0 {
				pos.Offset = unit.end
			} else if pos.Offset-unit.start < unit.end-pos.Offset {
				pos.Offset = unit.start
			} else {
				pos.Offset = unit.end
			}
			break
		}
	}
	return pos
}

func (e *RichTextEdit) moveCursor(target RichPosition, extend bool) {
	e.cursorInVoid = false
	e.boxSelection = nil
	target = e.normalizePosition(target, 0)
	if extend {
		if !e.selectionExtending {
			e.selectionAnchor = e.Cursor
			e.selectionExtending = true
		}
		e.Cursor = target
		e.SelectionFrom, e.SelectionTo = e.selectionAnchor, target
		e.HasSelection = e.SelectionFrom != e.SelectionTo
		return
	}
	if e.HasSelection && target.Line == e.Cursor.Line && target.Offset == e.Cursor.Offset {
		from, to := e.selectionBounds()
		if target.Offset <= from.Offset {
			target = from
		} else if target.Offset >= to.Offset {
			target = to
		}
	}
	e.Cursor = target
	e.ClearSelection()
}

func (e *RichTextEdit) moveHorizontal(direction int, extend bool) {
	if e.cursorInVoid {
		line, column := e.voidLine, e.voidColumn+direction
		if column < 0 {
			return
		}
		lines := e.documentLines()
		if line < len(lines) && column <= richLineColumn(lines[line], richLineRuneCount(lines[line])) {
			e.moveCursor(RichPosition{Line: line, Offset: richLineOffsetAtColumn(lines[line], column)}, extend)
			return
		}
		e.moveToVoid(line, column, extend)
		return
	}
	if e.HasSelection && !extend {
		from, to := e.selectionBounds()
		if direction < 0 {
			e.moveCursor(from, false)
		} else {
			e.moveCursor(to, false)
		}
		return
	}
	lines := e.documentLines()
	stops := richLineStops(lines[e.Cursor.Line])
	if direction < 0 {
		for i := len(stops) - 1; i >= 0; i-- {
			if stops[i] < e.Cursor.Offset {
				e.moveCursor(RichPosition{Line: e.Cursor.Line, Offset: stops[i]}, extend)
				return
			}
		}
		if e.Cursor.Line > 0 {
			line := e.Cursor.Line - 1
			e.moveCursor(RichPosition{Line: line, Offset: richLineRuneCount(lines[line])}, extend)
		}
		return
	}
	for _, stop := range stops {
		if stop > e.Cursor.Offset {
			e.moveCursor(RichPosition{Line: e.Cursor.Line, Offset: stop}, extend)
			return
		}
	}
	if e.Cursor.Line+1 < len(lines) {
		e.moveCursor(RichPosition{Line: e.Cursor.Line + 1}, extend)
	}
}

func (e *RichTextEdit) moveVertical(direction int, extend bool) {
	lines := e.documentLines()
	line, column := e.cursorCoordinates(lines)
	targetLine := line + direction
	if targetLine < 0 {
		return
	}
	if !e.GhostCursorEnabled {
		if targetLine >= len(lines) {
			return
		}
		best, distance := 0, int(^uint(0)>>1)
		for _, stop := range richLineStops(lines[targetLine]) {
			d := richTextAbs(richLineColumn(lines[targetLine], stop) - column)
			if d < distance {
				best, distance = stop, d
			}
		}
		e.moveCursor(RichPosition{Line: targetLine, Offset: best}, extend)
		return
	}
	if targetLine >= len(lines) {
		e.moveToVoid(targetLine, column, extend)
		return
	}
	lineWidth := richLineColumn(lines[targetLine], richLineRuneCount(lines[targetLine]))
	if column > lineWidth {
		e.moveToVoid(targetLine, column, extend)
		return
	}
	best, distance := 0, int(^uint(0)>>1)
	for _, stop := range richLineStops(lines[targetLine]) {
		d := richTextAbs(richLineColumn(lines[targetLine], stop) - column)
		if d < distance {
			best, distance = stop, d
		}
	}
	e.moveCursor(RichPosition{Line: targetLine, Offset: best}, extend)
}

func (e *RichTextEdit) wordPosition(direction int) RichPosition {
	lines := e.documentLines()
	line := lines[e.Cursor.Line]
	units := richLineUnits(line)
	index := 0
	for index < len(units) && units[index].end <= e.Cursor.Offset {
		index++
	}
	if direction < 0 {
		if index == 0 {
			if e.Cursor.Line > 0 {
				previous := e.Cursor.Line - 1
				return RichPosition{Line: previous, Offset: richLineRuneCount(lines[previous])}
			}
			return RichPosition{}
		}
		index--
		for index >= 0 && richUnitSpace(units[index]) {
			index--
		}
		if index >= 0 && richUnitWord(units[index]) {
			for index > 0 && richUnitWord(units[index-1]) {
				index--
			}
		}
		if index < 0 {
			return RichPosition{Line: e.Cursor.Line}
		}
		return RichPosition{Line: e.Cursor.Line, Offset: units[index].start}
	}
	if index < len(units) && richUnitWord(units[index]) {
		for index < len(units) && richUnitWord(units[index]) {
			index++
		}
	} else if index < len(units) {
		index++
	}
	for index < len(units) && richUnitSpace(units[index]) {
		index++
	}
	if index == len(units) && e.Cursor.Line+1 < len(lines) {
		return RichPosition{Line: e.Cursor.Line + 1}
	}
	if index == len(units) {
		return RichPosition{Line: e.Cursor.Line, Offset: richLineRuneCount(line)}
	}
	return RichPosition{Line: e.Cursor.Line, Offset: units[index].start}
}

func richUnitSpace(unit richTextUnit) bool {
	for _, r := range unit.text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return unit.text != ""
}

func richUnitWord(unit richTextUnit) bool {
	if unit.pill {
		return true
	}
	for _, r := range unit.text {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_') {
			return false
		}
	}
	return unit.text != ""
}

func richTextAbs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
