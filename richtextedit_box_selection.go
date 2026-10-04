// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "ubunatic.com/loom/measure"

type richBoxSlice struct {
	line       int
	start, end int
}

type richBoxSelection struct {
	top, left, bottom, right int
	rows                     []richBoxSlice
}

func cloneRichBoxSelection(source *richBoxSelection) *richBoxSelection {
	if source == nil {
		return nil
	}
	copy := *source
	copy.rows = append([]richBoxSlice(nil), source.rows...)
	return &copy
}

func (e *RichTextEdit) revalidateBoxSelection() {
	if e.boxSelection == nil {
		return
	}
	selection := e.boxSelection
	lines := e.documentLines()
	if selection.top < 0 || selection.left < 0 || selection.bottom >= len(lines) || selection.top >= selection.bottom || selection.left >= selection.right || !richClosedBox(lines, selection.top, selection.left, selection.bottom, selection.right) {
		e.ClearSelection()
		return
	}
	selection.rows = make([]richBoxSlice, 0, selection.bottom-selection.top+1)
	for row := selection.top; row <= selection.bottom; row++ {
		start := richLineOffsetAtColumn(lines[row], selection.left)
		end := richLineOffsetAtColumn(lines[row], selection.right+1)
		from, to := expandRangeForPills(lines, RichPosition{Line: row, Offset: start}, RichPosition{Line: row, Offset: end})
		selection.rows = append(selection.rows, richBoxSlice{line: row, start: from.Offset, end: to.Offset})
	}
	e.SelectionFrom = RichPosition{Line: selection.top, Offset: selection.rows[0].start}
	last := selection.rows[len(selection.rows)-1]
	e.SelectionTo = RichPosition{Line: last.line, Offset: last.end}
	e.HasSelection = true
}

func (e *RichTextEdit) selectBoxAtCursor() bool {
	lines := e.documentLines()
	line, col := e.Cursor.Line, richLineColumn(lines[e.Cursor.Line], e.Cursor.Offset)
	if richBoxCell(lines[line], col) == "" {
		return false
	}
	maxCol := 0
	for _, row := range lines {
		maxCol = max(maxCol, richLineColumn(row, richLineRuneCount(row)))
	}
	var best *richBoxSelection
	consider := func(top, left, bottom, right int) {
		if top < 0 || left < 0 || bottom >= len(lines) || right >= maxCol || top >= bottom || left >= right {
			return
		}
		if !((line == top || line == bottom) && col >= left && col <= right) &&
			!((col == left || col == right) && line >= top && line <= bottom) {
			return
		}
		area := (bottom - top + 1) * (right - left + 1)
		if best != nil {
			bestArea := (best.bottom - best.top + 1) * (best.right - best.left + 1)
			if area > bestArea || area == bestArea && !richBoxBefore(top, left, bottom, right, best) {
				return
			}
		}
		if !richBoxCorners(lines, top, left, bottom, right) || !richClosedBox(lines, top, left, bottom, right) {
			return
		}
		best = &richBoxSelection{top: top, left: left, bottom: bottom, right: right}
	}
	for other := line + 1; other < len(lines); other++ {
		for left := 0; left <= col; left++ {
			for right := col; right < maxCol; right++ {
				consider(line, left, other, right)
			}
		}
	}
	for other := 0; other < line; other++ {
		for left := 0; left <= col; left++ {
			for right := col; right < maxCol; right++ {
				consider(other, left, line, right)
			}
		}
	}
	for other := col + 1; other < maxCol; other++ {
		for top := 0; top <= line; top++ {
			for bottom := line; bottom < len(lines); bottom++ {
				consider(top, col, bottom, other)
			}
		}
	}
	for other := 0; other < col; other++ {
		for top := 0; top <= line; top++ {
			for bottom := line; bottom < len(lines); bottom++ {
				consider(top, other, bottom, col)
			}
		}
	}
	if best == nil {
		return false
	}
	e.boxSelection = best
	e.revalidateBoxSelection()
	e.selectionExtending = false
	return true
}

func richBoxBefore(top, left, bottom, right int, other *richBoxSelection) bool {
	if top != other.top {
		return top < other.top
	}
	if left != other.left {
		return left < other.left
	}
	if bottom != other.bottom {
		return bottom < other.bottom
	}
	return right < other.right
}

func richClosedBox(lines []RichLine, top, left, bottom, right int) bool {
	if !richBoxCorners(lines, top, left, bottom, right) {
		return false
	}
	for col := left + 1; col < right; col++ {
		if !richBoxHasArms(lines, top, col, BoxArmLeft|BoxArmRight) ||
			!richBoxHasArms(lines, bottom, col, BoxArmLeft|BoxArmRight) {
			return false
		}
	}
	for row := top + 1; row < bottom; row++ {
		if !richBoxHasArms(lines, row, left, BoxArmUp|BoxArmDown) ||
			!richBoxHasArms(lines, row, right, BoxArmUp|BoxArmDown) {
			return false
		}
	}
	return true
}

func richBoxCorners(lines []RichLine, top, left, bottom, right int) bool {
	return richBoxHasArms(lines, top, left, BoxArmRight|BoxArmDown) &&
		richBoxHasArms(lines, top, right, BoxArmLeft|BoxArmDown) &&
		richBoxHasArms(lines, bottom, left, BoxArmRight|BoxArmUp) &&
		richBoxHasArms(lines, bottom, right, BoxArmLeft|BoxArmUp)
}

func richBoxHasArms(lines []RichLine, row, col int, required BoxArms) bool {
	cell := richBoxCell(lines[row], col)
	return cell != "" && BoxGlyphArms(cell)&required == required
}

func richBoxCell(line RichLine, col int) string {
	for _, unit := range richLineUnits(line) {
		start := richLineColumn(line, unit.start)
		width := measure.StringWidth(unit.text)
		if col >= start && col < start+width {
			if unit.pill || width != 1 {
				return ""
			}
			return unit.text
		}
	}
	return ""
}
