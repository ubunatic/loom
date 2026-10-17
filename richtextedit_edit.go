// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"ubunatic.com/loom/measure"
)

func (e *RichTextEdit) ensureDocument() {
	if e.Document == nil {
		e.Document = &RichDocument{}
	}
	if len(e.Document.Lines) == 0 {
		e.Document.Lines = []RichLine{{}}
	}
}

func richLineRuneCount(line RichLine) int {
	n := 0
	for _, span := range line.Spans {
		n += len([]rune(span.Text))
	}
	return n
}

func richLineUnits(line RichLine) []richTextUnit {
	var units []richTextUnit
	offset := 0
	for _, span := range line.Spans {
		if span.PillData != nil {
			n := len([]rune(span.Text))
			if n > 0 {
				units = append(units, richTextUnit{start: offset, end: offset + n, text: span.Text, pill: true})
				offset += n
			}
			continue
		}
		for _, cluster := range measure.Clusters(span.Text) {
			n := len([]rune(cluster))
			units = append(units, richTextUnit{start: offset, end: offset + n, text: cluster})
			offset += n
		}
	}
	return units
}

func (e *RichTextEdit) insertText(text string) {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	if text == "" {
		return
	}
	e.materializeVoidCursor()
	e.deleteSelection()
	e.Cursor = e.normalizePosition(e.Cursor, 0)
	lines := e.Document.Lines
	left, right := splitRichLine(lines[e.Cursor.Line], e.Cursor.Offset)
	parts := strings.Split(text, "\n")
	style := e.insertionStyle(left, right)
	if len(parts) == 1 {
		spans := append([]RichSpan(nil), left...)
		if parts[0] != "" {
			spans = append(spans, RichSpan{Text: parts[0], Style: style})
		}
		spans = append(spans, right...)
		lines[e.Cursor.Line].Spans = mergeRichSpans(spans)
		e.Cursor.Offset += len([]rune(parts[0]))
		return
	}
	updated := make([]RichLine, 0, len(lines)+len(parts)-1)
	updated = append(updated, lines[:e.Cursor.Line]...)
	first := append([]RichSpan(nil), left...)
	if parts[0] != "" {
		first = append(first, RichSpan{Text: parts[0], Style: style})
	}
	updated = append(updated, RichLine{Spans: mergeRichSpans(first)})
	for _, part := range parts[1 : len(parts)-1] {
		var spans []RichSpan
		if part != "" {
			spans = append(spans, RichSpan{Text: part, Style: style})
		}
		updated = append(updated, RichLine{Spans: spans})
	}
	last := make([]RichSpan, 0, len(right)+1)
	if lastText := parts[len(parts)-1]; lastText != "" {
		last = append(last, RichSpan{Text: lastText, Style: style})
	}
	last = append(last, right...)
	updated = append(updated, RichLine{Spans: mergeRichSpans(last)})
	updated = append(updated, lines[e.Cursor.Line+1:]...)
	e.Document.Lines = updated
	e.Cursor.Line += len(parts) - 1
	e.Cursor.Offset = len([]rune(parts[len(parts)-1]))
}

func (e *RichTextEdit) insertionStyle(left, right []RichSpan) Style {
	if e.activeStyleSet {
		return e.ActiveStyle
	}
	if len(left) > 0 {
		return left[len(left)-1].Style
	}
	if len(right) > 0 {
		return right[0].Style
	}
	return e.ActiveStyle
}

func splitRichLine(line RichLine, offset int) (left, right []RichSpan) {
	consumed := 0
	for _, span := range line.Spans {
		length := len([]rune(span.Text))
		end := consumed + length
		if end <= offset {
			left = append(left, span)
		} else if consumed >= offset {
			right = append(right, span)
		} else if span.PillData != nil {
			if offset-consumed < end-offset {
				right = append(right, span)
			} else {
				left = append(left, span)
			}
		} else {
			cut := offset - consumed
			runes := []rune(span.Text)
			before, after := span, span
			before.Text, after.Text = string(runes[:cut]), string(runes[cut:])
			if before.Text != "" {
				left = append(left, before)
			}
			if after.Text != "" {
				right = append(right, after)
			}
		}
		consumed = end
	}
	return mergeRichSpans(left), mergeRichSpans(right)
}

func mergeRichSpans(spans []RichSpan) []RichSpan {
	var merged []RichSpan
	for _, span := range spans {
		if span.Text == "" {
			continue
		}
		n := len(merged)
		if n > 0 && merged[n-1].PillData == nil && span.PillData == nil && merged[n-1].Style == span.Style && merged[n-1].Link == span.Link && merged[n-1].Code == span.Code {
			merged[n-1].Text += span.Text
		} else {
			merged = append(merged, span)
		}
	}
	return merged
}

func expandRangeForPills(lines []RichLine, from, to RichPosition) (RichPosition, RichPosition) {
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		offset := 0
		for _, span := range lines[lineIndex].Spans {
			end := offset + len([]rune(span.Text))
			if span.PillData != nil {
				if lineIndex == from.Line && from.Offset > offset && from.Offset < end {
					from.Offset = offset
				}
				if lineIndex == to.Line && to.Offset > offset && to.Offset < end {
					to.Offset = end
				}
				if lineIndex == from.Line && from.Offset < end && to.Line > lineIndex {
					from.Offset = min(from.Offset, offset)
				}
				if lineIndex == to.Line && to.Offset > offset && from.Line < lineIndex {
					to.Offset = max(to.Offset, end)
				}
			}
			offset = end
		}
	}
	return from, to
}

func (e *RichTextEdit) deleteSelection() bool {
	if !e.HasSelection {
		return false
	}
	if e.boxSelection != nil {
		selection := e.boxSelection
		for _, row := range selection.rows {
			line := e.Document.Lines[row.line]
			left, rest := splitRichLine(line, row.start)
			_, right := splitRichLine(RichLine{Spans: rest}, row.end-row.start)
			e.Document.Lines[row.line].Spans = mergeRichSpans(append(left, right...))
		}
		e.Cursor = RichPosition{Line: selection.top, Offset: richLineOffsetAtColumn(e.Document.Lines[selection.top], selection.left)}
		e.ClearSelection()
		return true
	}
	from, to := e.selectionBounds()
	e.clampPosition(&from, e.Document.Lines)
	e.clampPosition(&to, e.Document.Lines)
	from, to = expandRangeForPills(e.Document.Lines, from, to)
	e.deleteRange(from, to)
	e.Cursor = from
	e.ClearSelection()
	return true
}

func (e *RichTextEdit) deleteRange(from, to RichPosition) {
	lines := e.Document.Lines
	if richPositionBefore(to, from) {
		from, to = to, from
	}
	if from.Line == to.Line {
		left, _ := splitRichLine(lines[from.Line], from.Offset)
		_, right := splitRichLine(lines[to.Line], to.Offset)
		lines[from.Line].Spans = mergeRichSpans(append(left, right...))
		return
	}
	left, _ := splitRichLine(lines[from.Line], from.Offset)
	_, right := splitRichLine(lines[to.Line], to.Offset)
	replacement := RichLine{Spans: mergeRichSpans(append(left, right...))}
	updated := make([]RichLine, 0, len(lines)-(to.Line-from.Line))
	updated = append(updated, lines[:from.Line]...)
	updated = append(updated, replacement)
	updated = append(updated, lines[to.Line+1:]...)
	e.Document.Lines = updated
}

func (e *RichTextEdit) deleteDirection(direction int) {
	if e.deleteSelection() {
		return
	}
	lines := e.Document.Lines
	if direction < 0 {
		stops := richLineStops(lines[e.Cursor.Line])
		for i := len(stops) - 1; i >= 0; i-- {
			if stops[i] < e.Cursor.Offset {
				e.deleteRange(RichPosition{Line: e.Cursor.Line, Offset: stops[i]}, e.Cursor)
				e.Cursor.Offset = stops[i]
				return
			}
		}
		if e.Cursor.Line > 0 {
			previous := e.Cursor.Line - 1
			offset := richLineRuneCount(lines[previous])
			lines[previous].Spans = mergeRichSpans(append(lines[previous].Spans, lines[e.Cursor.Line].Spans...))
			e.Document.Lines = append(lines[:e.Cursor.Line], lines[e.Cursor.Line+1:]...)
			e.Cursor = RichPosition{Line: previous, Offset: offset}
		}
		return
	}
	for _, stop := range richLineStops(lines[e.Cursor.Line]) {
		if stop > e.Cursor.Offset {
			e.deleteRange(e.Cursor, RichPosition{Line: e.Cursor.Line, Offset: stop})
			return
		}
	}
	if e.Cursor.Line+1 < len(lines) {
		line := &lines[e.Cursor.Line]
		line.Spans = mergeRichSpans(append(line.Spans, lines[e.Cursor.Line+1].Spans...))
		e.Document.Lines = append(lines[:e.Cursor.Line+1], lines[e.Cursor.Line+2:]...)
	}
}
