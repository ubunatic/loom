// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"unicode"

	"ubunatic.com/loom/measure"
)

// RichPosition identifies a rune offset in a logical rich document line.
type RichPosition struct {
	Line   int
	Offset int
}

// RichTextEdit edits a RichDocument with styled spans, selection, and a
// scrolling viewport.
type RichTextEdit struct {
	Document       *RichDocument
	Cursor         RichPosition
	SelectionFrom  RichPosition
	SelectionTo    RichPosition
	HasSelection   bool
	SelectionStyle Style
	ActiveStyle    Style
	ScrollX        int
	ScrollY        int
	ShowCursor     bool

	focused            bool
	lastRect           Rect
	selectionAnchor    RichPosition
	selectionExtending bool
	dragSelecting      bool
	activeStyleSet     bool
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
	e.selectionExtending = false
}

// ClearSelection removes the current selection.
func (e *RichTextEdit) ClearSelection() {
	e.HasSelection = false
	e.selectionExtending = false
}

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
	e.Cursor = e.normalizePosition(e.Cursor, 0)
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

// ConsumeKey applies navigation, text editing, selection, and inline formatting.
func (e *RichTextEdit) ConsumeKey(key KeyEvent) EventResult {
	e.ensureDocument()
	e.clampPosition(&e.Cursor, e.documentLines())
	e.Cursor = e.normalizePosition(e.Cursor, 0)
	switch {
	case key.Is("ctrl-b"):
		e.toggleAttribute(func(s *Style, on bool) { s.Bold = on }, func(s Style) bool { return s.Bold })
	case key.Is("ctrl-i"):
		e.toggleAttribute(func(s *Style, on bool) { s.Italic = on }, func(s Style) bool { return s.Italic })
	case key.Is("ctrl-u"):
		e.toggleAttribute(func(s *Style, on bool) { s.Underline = on }, func(s Style) bool { return s.Underline })
	case key.Is("left", "shift-left"):
		e.moveHorizontal(-1, key.Is("shift-left"))
	case key.Is("right", "shift-right"):
		e.moveHorizontal(1, key.Is("shift-right"))
	case key.Is("up", "shift-up"):
		e.moveVertical(-1, key.Is("shift-up"))
	case key.Is("down", "shift-down"):
		e.moveVertical(1, key.Is("shift-down"))
	case key.Is("home", "shift-home"):
		e.moveCursor(RichPosition{Line: e.Cursor.Line}, key.Is("shift-home"))
	case key.Is("end", "shift-end"):
		e.moveCursor(RichPosition{Line: e.Cursor.Line, Offset: richLineRuneCount(e.documentLines()[e.Cursor.Line])}, key.Is("shift-end"))
	case key.Is("ctrl-left"):
		if e.HasSelection {
			from, _ := e.selectionBounds()
			e.moveCursor(from, false)
		} else {
			e.moveCursor(e.wordPosition(-1), false)
		}
	case key.Is("ctrl-right"):
		if e.HasSelection {
			_, to := e.selectionBounds()
			e.moveCursor(to, false)
		} else {
			e.moveCursor(e.wordPosition(1), false)
		}
	case key.Is("enter", "return"):
		e.insertText("\n")
	case key.Is("backspace"):
		e.deleteDirection(-1)
	case key.Is("delete"):
		e.deleteDirection(1)
	default:
		if key.Text == "" {
			return Ignored()
		}
		e.insertText(key.Text)
	}
	return Handled()
}

// ConsumeMouse positions the cursor and supports click-drag selection.
func (e *RichTextEdit) ConsumeMouse(mouse MouseEvent) EventResult {
	e.ensureDocument()
	switch mouse.Action {
	case MouseScrollUp:
		e.ScrollY = max(0, e.ScrollY-1)
		return Handled()
	case MouseScrollDown:
		e.ScrollY++
		return Handled()
	case MousePress:
		if mouse.Button != MouseLeft || !e.mouseInBounds(mouse) {
			return Ignored()
		}
		position := e.mousePosition(mouse)
		e.Cursor, e.selectionAnchor = position, position
		e.dragSelecting = true
		e.ClearSelection()
		return Handled()
	case MouseDrag:
		if !e.dragSelecting {
			return Ignored()
		}
		e.extendMouseSelection(mouse)
		return Handled()
	case MouseRelease:
		if !e.dragSelecting || mouse.Button != MouseLeft {
			return Ignored()
		}
		e.extendMouseSelection(mouse)
		e.dragSelecting = false
		return Handled()
	default:
		return Ignored()
	}
}

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
	targetLine := e.Cursor.Line + direction
	if targetLine < 0 || targetLine >= len(lines) {
		return
	}
	column := richLineColumn(lines[e.Cursor.Line], e.Cursor.Offset)
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

func (e *RichTextEdit) insertText(text string) {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	if text == "" {
		return
	}
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
	merged := make([]RichSpan, 0, len(spans))
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

func (e *RichTextEdit) toggleAttribute(set func(*Style, bool), enabled func(Style) bool) {
	if !e.HasSelection {
		set(&e.ActiveStyle, !enabled(e.ActiveStyle))
		e.activeStyleSet = true
		return
	}
	from, to := e.selectionBounds()
	e.clampPosition(&from, e.Document.Lines)
	e.clampPosition(&to, e.Document.Lines)
	from, to = expandRangeForPills(e.Document.Lines, from, to)
	allEnabled := true
	for lineIndex := from.Line; lineIndex <= to.Line; lineIndex++ {
		start, end := richSelectionLineBounds(lineIndex, from, to)
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
		start, end := richSelectionLineBounds(lineIndex, from, to)
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

func (e *RichTextEdit) mouseInBounds(mouse MouseEvent) bool {
	return e.lastRect.W > 0 && e.lastRect.H > 0 && mouse.X >= 0 && mouse.X < e.lastRect.W && mouse.Y >= 0 && mouse.Y < e.lastRect.H
}

func (e *RichTextEdit) mousePosition(mouse MouseEvent) RichPosition {
	lines := e.documentLines()
	lineIndex := min(max(0, e.ScrollY+mouse.Y), len(lines)-1)
	targetColumn := e.ScrollX + mouse.X
	best, distance := 0, int(^uint(0)>>1)
	for _, stop := range richLineStops(lines[lineIndex]) {
		d := richTextAbs(richLineColumn(lines[lineIndex], stop) - targetColumn)
		if d < distance {
			best, distance = stop, d
		}
	}
	return RichPosition{Line: lineIndex, Offset: best}
}

func (e *RichTextEdit) extendMouseSelection(mouse MouseEvent) {
	if !e.mouseInBounds(mouse) {
		return
	}
	position := e.mousePosition(mouse)
	e.Cursor = position
	e.SelectionTo = position
	e.SelectionFrom = e.selectionAnchor
	e.HasSelection = e.SelectionFrom != e.SelectionTo
}
