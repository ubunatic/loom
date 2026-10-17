// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

func (e *RichTextEdit) applyPopoverColor(palette string, color Color) {
	if !e.HasSelection || (palette != "#FG" && palette != "#BG") {
		e.popoverPalette = ""
		e.popoverSwatches = nil
		return
	}
	e.mutate(false, false, func() {
		lines := e.documentLines()
		from, to := e.selectionBounds()
		e.clampPosition(&from, lines)
		e.clampPosition(&to, lines)
		from, to = expandRangeForPills(lines, from, to)
		for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
			start, end := e.selectionLineBoundsForRange(lineIndex, from, to)
			lines[lineIndex].Spans = colorRichLine(lines[lineIndex], start, end, palette, color)
		}
	})
	e.popoverPalette = ""
	e.popoverSwatches = nil
	e.finishPopoverAction()
}

func colorRichLine(line RichLine, start, end int, field string, color Color) []RichSpan {
	var result []RichSpan
	offset := 0
	for _, span := range line.Spans {
		length := len([]rune(span.Text))
		spanEnd := offset + length
		if length == 0 || offset >= end || spanEnd <= start {
			result = append(result, span)
			offset = spanEnd
			continue
		}
		localStart, localEnd := max(0, start-offset), min(length, end-offset)
		prefix, rest := splitRichLine(RichLine{Spans: []RichSpan{span}}, localStart)
		selected, suffix := splitRichLine(RichLine{Spans: rest}, localEnd-localStart)
		result = append(result, prefix...)
		for i := range selected {
			if field == "#FG" {
				selected[i].Style.FG = color
			} else {
				selected[i].Style.BG = color
			}
		}
		result = append(result, selected...)
		result = append(result, suffix...)
		offset = spanEnd
	}
	return mergeRichSpans(result)
}

func (e *RichTextEdit) toggleAttribute(set func(*Style, bool), enabled func(Style) bool) {
	if e.HasSelection {
		from, to := e.selectionBounds()
		e.mutate(false, false, func() { e.toggleRange(from, to, set, enabled) })
		return
	}
	if from, to, ok := e.wordRange(e.Cursor); ok {
		e.mutate(false, false, func() { e.toggleRange(from, to, set, enabled) })
		return
	}
	set(&e.ActiveStyle, !enabled(e.ActiveStyle))
	e.activeStyleSet = true
}

func (e *RichTextEdit) toggleRange(from, to RichPosition, set func(*Style, bool), enabled func(Style) bool) {
	e.clampPosition(&from, e.Document.Lines)
	e.clampPosition(&to, e.Document.Lines)
	from, to = expandRangeForPills(e.Document.Lines, from, to)
	allEnabled := true
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		start, end := e.selectionLineBoundsForRange(lineIndex, from, to)
		offset := 0
		for _, span := range e.Document.Lines[lineIndex].Spans {
			spanEnd := offset + len([]rune(span.Text))
			if offset < end && spanEnd > start && !enabled(span.Style) {
				allEnabled = false
			}
			offset = spanEnd
		}
	}
	state := !allEnabled
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		start, end := e.selectionLineBoundsForRange(lineIndex, from, to)
		e.Document.Lines[lineIndex].Spans = styleRichLine(e.Document.Lines[lineIndex], start, end, state, set)
	}
}

func richSelectionLineBounds(line int, from, to RichPosition) (start, end int) {
	if line == from.Line {
		start = from.Offset
	}
	if line == to.Line {
		end = to.Offset
	} else {
		end = int(^uint(0) >> 1)
	}
	return start, end
}

func styleRichLine(line RichLine, start, end int, state bool, set func(*Style, bool)) []RichSpan {
	var result []RichSpan
	offset := 0
	for _, span := range line.Spans {
		length := len([]rune(span.Text))
		spanEnd := offset + length
		if length == 0 || offset >= end || spanEnd <= start {
			result = append(result, span)
			offset = spanEnd
			continue
		}
		localStart, localEnd := max(0, start-offset), min(length, end-offset)
		prefix, rest := splitRichLine(RichLine{Spans: []RichSpan{span}}, localStart)
		selected, suffix := splitRichLine(RichLine{Spans: rest}, localEnd-localStart)
		result = append(result, prefix...)
		for i := range selected {
			set(&selected[i].Style, state)
		}
		result = append(result, selected...)
		result = append(result, suffix...)
		offset = spanEnd
	}
	return mergeRichSpans(result)
}
