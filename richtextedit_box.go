// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"reflect"
	"strings"
	"ubunatic.com/loom/measure"
)

func (e *RichTextEdit) wrapSelectionInBox() {
	style := BoxBorderStyleSharp
	if strings.EqualFold(SpeccedDefaults.RichTextEdit.BoxStyleDefault, "rounded") {
		style = BoxBorderStyleRounded
	}
	e.wrapSelectionInBoxStyle(style)
}

func (e *RichTextEdit) wrapSelectionInBoxStyle(boxStyle BoxBorderStyle) {
	if !e.HasSelection || e.boxSelection != nil {
		return
	}
	e.mutate(false, false, func() {
		lines := e.documentLines()
		from, to := e.selectionBounds()
		e.clampPosition(&from, lines)
		e.clampPosition(&to, lines)
		from, to = expandRangeForPills(lines, from, to)
		if from == to {
			return
		}
		prefix, firstRest := splitRichLine(lines[from.Line], from.Offset)
		firstLength := richLineRuneCount(RichLine{Spans: firstRest})
		if from.Line == to.Line {
			firstLength = to.Offset - from.Offset
		}
		first, _ := splitRichLine(RichLine{Spans: firstRest}, firstLength)
		_, suffix := splitRichLine(lines[to.Line], to.Offset)
		content := make([]RichLine, 0, to.Line-from.Line+1)
		content = append(content, RichLine{Spans: first})
		for i := from.Line + 1; i < to.Line; i++ {
			content = append(content, RichLine{Spans: append([]RichSpan(nil), lines[i].Spans...)})
		}
		last, _ := splitRichLine(lines[to.Line], to.Offset)
		if to.Line != from.Line {
			content = append(content, RichLine{Spans: last})
		} else {
			content = []RichLine{{Spans: first}}
		}
		width := 0
		for _, line := range content {
			width = max(width, richLineColumn(line, richLineRuneCount(line)))
		}
		border, err := getBoxBorderGlyphs(boxStyle)
		if err != nil {
			return
		}
		style := Style{}
		if len(content) > 0 && len(content[0].Spans) > 0 {
			style = content[0].Spans[0].Style
		}
		boxed := []RichLine{{Spans: []RichSpan{{Text: border.TopLeft + strings.Repeat(border.Horizontal, width) + border.TopRight, Style: style}}}}
		for _, line := range content {
			padding := width - richLineColumn(line, richLineRuneCount(line))
			spans := []RichSpan{{Text: border.Vertical, Style: style}}
			spans = append(spans, line.Spans...)
			if padding > 0 {
				spans = append(spans, RichSpan{Text: strings.Repeat(" ", padding), Style: style})
			}
			spans = append(spans, RichSpan{Text: border.Vertical, Style: style})
			boxed = append(boxed, RichLine{Spans: mergeRichSpans(spans)})
		}
		boxed = append(boxed, RichLine{Spans: []RichSpan{{Text: border.BottomLeft + strings.Repeat(border.Horizontal, width) + border.BottomRight, Style: style}}})
		lastIndex := len(boxed) - 1
		updated := append([]RichLine(nil), lines[:from.Line]...)
		if richLineRuneCount(RichLine{Spans: prefix}) > 0 {
			updated = append(updated, RichLine{Spans: prefix})
		}
		boxLine := from.Line + len(updated) - len(lines[:from.Line])
		updated = append(updated, boxed...)
		if richLineRuneCount(RichLine{Spans: suffix}) > 0 {
			updated = append(updated, RichLine{Spans: suffix})
		}
		updated = append(updated, lines[to.Line+1:]...)
		e.Document.Lines = updated
		e.Cursor = RichPosition{Line: boxLine + lastIndex, Offset: richLineRuneCount(boxed[lastIndex])}
		e.ClearSelection()
	})
}

func (e *RichTextEdit) restyleSelectedBox(boxStyle BoxBorderStyle) {
	if e.boxSelection == nil || (boxStyle != BoxBorderStyleSharp && boxStyle != BoxBorderStyleRounded) {
		return
	}
	border, err := getBoxBorderGlyphs(boxStyle)
	if err != nil {
		return
	}
	e.mutate(false, false, func() {
		selection := e.boxSelection
		corners := []struct {
			line, column int
			arms         BoxArms
			glyph        string
		}{
			{selection.top, selection.left, BoxArmRight | BoxArmDown, border.TopLeft},
			{selection.top, selection.right, BoxArmLeft | BoxArmDown, border.TopRight},
			{selection.bottom, selection.left, BoxArmRight | BoxArmUp, border.BottomLeft},
			{selection.bottom, selection.right, BoxArmLeft | BoxArmUp, border.BottomRight},
		}
		for _, corner := range corners {
			line := e.Document.Lines[corner.line]
			if BoxGlyphArms(richLineCell(line, corner.column)) != corner.arms {
				continue // Preserve crossing/shared-edge junction arms.
			}
			e.Document.Lines[corner.line] = richLineSetCell(line, corner.column, corner.glyph)
		}
	})
}

func (e *RichTextEdit) toggleBoxMode() {
	if !e.BoxMode {
		e.ClearSelection()
		e.boxStrokeFGSet = false
		lines := e.documentLines()
		line, column := e.cursorCoordinates(lines)
		if !e.cursorInVoid && line < len(lines) {
			if glyph := richBoxCell(lines[line], column); glyph != "" && BoxGlyphArms(glyph) != 0 {
				if style, ok := richLineStyleAtColumn(lines[line], column); ok {
					e.boxStrokeFG, e.boxStrokeFGSet = style.FG, true
				}
			}
		}
		e.BoxMode = true
		before := e.snapshot()
		e.boxStrokeBefore = &before
		e.typingRun = false
		return
	}
	e.BoxMode = false
	e.finishBoxStroke()
}

func (e *RichTextEdit) finishBoxStroke() {
	if e.boxStrokeBefore == nil {
		e.boxStrokeFGSet = false
		return
	}
	defer func() { e.boxStrokeFGSet = false }()
	before := *e.boxStrokeBefore
	e.boxStrokeBefore = nil
	if reflect.DeepEqual(before.lines, e.Document.Lines) {
		return
	}
	e.undoStack = append(e.undoStack, before)
	if len(e.undoStack) > SpeccedDefaults.RichTextEdit.UndoLimit {
		e.undoStack = e.undoStack[len(e.undoStack)-SpeccedDefaults.RichTextEdit.UndoLimit:]
	}
	e.redoStack = nil
}

func (e *RichTextEdit) drawBoxStep(dy, dx int) {
	e.boxSelection = nil
	e.materializeVoidCursor()
	e.ensureDocument()
	lines := e.Document.Lines
	fromLine, fromColumn := e.Cursor.Line, richLineColumn(lines[e.Cursor.Line], e.Cursor.Offset)
	toLine, toColumn := fromLine+dy, fromColumn+dx
	if toColumn < 0 {
		return
	}
	if toLine < 0 {
		prepend := make([]RichLine, -toLine)
		e.Document.Lines = append(prepend, lines...)
		fromLine += len(prepend)
		toLine = 0
	} else if toLine >= len(lines) {
		for len(e.Document.Lines) <= toLine {
			e.Document.Lines = append(e.Document.Lines, RichLine{})
		}
	}
	lines = e.Document.Lines
	for len(lines[fromLine].Spans) > 0 && richLineColumn(lines[fromLine], richLineRuneCount(lines[fromLine])) < fromColumn {
		lines[fromLine].Spans = append(lines[fromLine].Spans, RichSpan{Text: " "})
	}
	for len(lines[toLine].Spans) > 0 && richLineColumn(lines[toLine], richLineRuneCount(lines[toLine])) < toColumn {
		lines[toLine].Spans = append(lines[toLine].Spans, RichSpan{Text: " "})
	}
	fromArm, toArm := BoxArmRight, BoxArmLeft
	if dx < 0 {
		fromArm, toArm = BoxArmLeft, BoxArmRight
	} else if dy < 0 {
		fromArm, toArm = BoxArmUp, BoxArmDown
	} else if dy > 0 {
		fromArm, toArm = BoxArmDown, BoxArmUp
	}
	fromCell := richLineCell(lines[fromLine], fromColumn)
	fromMask := boxNeighbourArms(lines, fromLine, fromColumn) | fromArm
	lines[fromLine] = richLineSetCell(lines[fromLine], fromColumn, boxGlyphForStroke(fromCell, fromMask))
	if e.boxStrokeFGSet {
		lines[fromLine] = richLineSetCellForeground(lines[fromLine], fromColumn, e.boxStrokeFG)
	}
	toCell := richLineCell(lines[toLine], toColumn)
	toMask := boxNeighbourArms(lines, toLine, toColumn) | toArm
	lines[toLine] = richLineSetCell(lines[toLine], toColumn, boxGlyphForStroke(toCell, toMask))
	if e.boxStrokeFGSet {
		lines[toLine] = richLineSetCellForeground(lines[toLine], toColumn, e.boxStrokeFG)
	}
	e.Document.Lines = lines
	e.Cursor = RichPosition{Line: toLine, Offset: richLineOffsetAtColumn(lines[toLine], toColumn)}
	e.typingRun = false
}

func boxGlyphForStroke(existing string, arms BoxArms) string {
	if existing == "╭" || existing == "╮" || existing == "╰" || existing == "╯" {
		if BoxGlyphArms(existing) == arms {
			return existing
		}
	}
	return BoxGlyph(arms)
}

func boxNeighbourArms(lines []RichLine, line, column int) BoxArms {
	var arms BoxArms
	if line > 0 && BoxGlyphArms(richLineCell(lines[line-1], column))&BoxArmDown != 0 {
		arms |= BoxArmUp
	}
	if line+1 < len(lines) && BoxGlyphArms(richLineCell(lines[line+1], column))&BoxArmUp != 0 {
		arms |= BoxArmDown
	}
	if column > 0 && BoxGlyphArms(richLineCell(lines[line], column-1))&BoxArmRight != 0 {
		arms |= BoxArmLeft
	}
	if BoxGlyphArms(richLineCell(lines[line], column+1))&BoxArmLeft != 0 {
		arms |= BoxArmRight
	}
	return arms | BoxGlyphArms(richLineCell(lines[line], column))
}

func richLineCell(line RichLine, column int) string {
	for _, unit := range richLineUnits(line) {
		start := richLineColumn(line, unit.start)
		if column >= start && column < start+measure.StringWidth(unit.text) {
			return unit.text
		}
	}
	return " "
}

func richLineStyleAtColumn(line RichLine, column int) (Style, bool) {
	cellX := 0
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			width := measure.StringWidth(cluster)
			if column >= cellX && column < cellX+width {
				return span.Style, span.PillData == nil
			}
			cellX += width
		}
	}
	return Style{}, false
}

func richLineOffsetAtColumn(line RichLine, column int) int {
	for _, unit := range richLineUnits(line) {
		start := richLineColumn(line, unit.start)
		if column < start+measure.StringWidth(unit.text) {
			return unit.start
		}
	}
	return richLineRuneCount(line)
}

func richLineSetCell(line RichLine, column int, glyph string) RichLine {
	var result []RichSpan
	cellX, replaced := 0, false
	for _, span := range line.Spans {
		for _, cluster := range measure.Clusters(span.Text) {
			width := measure.StringWidth(cluster)
			if width <= 0 {
				result = appendRichCellSpan(result, span, cluster)
				continue
			}
			if !replaced && column >= cellX && column < cellX+width {
				if span.PillData != nil {
					return line
				}
				before := strings.Repeat(" ", column-cellX)
				after := strings.Repeat(" ", cellX+width-column-1)
				result = appendRichCellSpan(result, span, before+glyph+after)
				replaced = true
			} else {
				result = appendRichCellSpan(result, span, cluster)
			}
			cellX += width
		}
	}
	if !replaced {
		if column > cellX {
			result = append(result, RichSpan{Text: strings.Repeat(" ", column-cellX)})
		}
		result = append(result, RichSpan{Text: glyph})
	}
	return RichLine{Spans: mergeRichSpans(result)}
}

func richLineSetCellForeground(line RichLine, column int, foreground Color) RichLine {
	for _, unit := range richLineUnits(line) {
		start := richLineColumn(line, unit.start)
		if column < start || column >= start+measure.StringWidth(unit.text) {
			continue
		}
		if unit.pill {
			return line
		}
		styled := styleRichLine(line, unit.start, unit.end, true, func(style *Style, _ bool) {
			style.FG = foreground
		})
		return RichLine{Spans: styled}
	}
	return line
}

func appendRichCellSpan(spans []RichSpan, source RichSpan, text string) []RichSpan {
	if text == "" {
		return spans
	}
	span := source
	span.Text = text
	span.PillData = nil
	return append(spans, span)
}
