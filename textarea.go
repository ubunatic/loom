// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"

	"codeberg.org/ubunatic/loom/syntax"
)

// TextArea is a multi-line text editor — the sibling of TextInput for bodies
// that span several lines (e.g. a git commit message body). It holds the text
// as a slice of rune lines with a (row, col) caret, scrolls vertically when the
// content exceeds the visible height, and commits nothing itself: the host
// decides when editing ends (e.g. Ctrl-D) and reads Value.
//
// Enter inserts a newline (splitting the current line); Backspace at column 0
// joins with the previous line; arrows/Home/End move the caret. HandleKey
// returns consumed=true for keys it acted on so the host keeps its own end key.
type TextArea struct {
	Placeholder   string // dim hint shown when the whole buffer is empty
	StyleResolver syntax.StyleResolver

	lines       [][]rune
	row         int // caret line index
	col         int // caret column (rune offset within lines[row])
	scroll      int // first visible line index
	lastH       int // visible height from the last Draw
	highlighter syntax.Engine
}

// NewTextArea creates a TextArea seeded with value (split on "\n").
func NewTextArea(value string) *TextArea {
	t := &TextArea{}
	t.SetValue(value)
	return t
}

// SetHighlighter attaches a syntax engine to the text area.
func (t *TextArea) SetHighlighter(h syntax.Engine) {
	t.highlighter = h
	if t.highlighter != nil {
		_ = t.highlighter.Parse([]byte(t.Value()))
	}
}

// Highlighter returns the attached syntax engine, or nil.
func (t *TextArea) Highlighter() syntax.Engine {
	return t.highlighter
}

// Value returns the buffer as a single newline-joined string.
func (t *TextArea) Value() string {
	parts := make([]string, len(t.lines))
	for i, ln := range t.lines {
		parts[i] = string(ln)
	}
	return strings.Join(parts, "\n")
}

// SetValue replaces the buffer and puts the caret at the end of the last line.
func (t *TextArea) SetValue(s string) {
	raw := strings.Split(s, "\n")
	t.lines = make([][]rune, len(raw))
	for i, ln := range raw {
		t.lines[i] = []rune(ln)
	}
	if len(t.lines) == 0 {
		t.lines = [][]rune{{}}
	}
	t.row = len(t.lines) - 1
	t.col = len(t.lines[t.row])
	t.scroll = 0
	if t.highlighter != nil {
		_ = t.highlighter.Parse([]byte(s))
	}
}

// Caret returns the caret position as (row, col), both rune offsets.
func (t *TextArea) Caret() (row, col int) { return t.row, t.col }

// LineCount returns the number of text lines.
func (t *TextArea) LineCount() int { return len(t.lines) }

// HandleKey applies an editing key, returning consumed=true when it acted.
func (t *TextArea) HandleKey(e KeyEvent) (consumed bool) {
	switch e.Key {
	case "left":
		t.moveLeft()
		return true
	case "right":
		t.moveRight()
		return true
	case "up":
		if t.row > 0 {
			t.row--
			t.clampCol()
		}
		return true
	case "down":
		if t.row < len(t.lines)-1 {
			t.row++
			t.clampCol()
		}
		return true
	case "home":
		t.col = 0
		return true
	case "end":
		t.col = len(t.lines[t.row])
		return true
	case "enter":
		t.splitLine()
		return true
	case "backspace":
		t.backspace()
		return true
	case "delete":
		t.deleteForward()
		return true
	}
	if e.Text != "" {
		t.insert([]rune(e.Text))
		return true
	}
	return false
}

func (t *TextArea) moveLeft() {
	switch {
	case t.col > 0:
		t.col--
	case t.row > 0:
		t.row--
		t.col = len(t.lines[t.row])
	}
}

func (t *TextArea) moveRight() {
	switch {
	case t.col < len(t.lines[t.row]):
		t.col++
	case t.row < len(t.lines)-1:
		t.row++
		t.col = 0
	}
}

// clampCol keeps the caret column within the current line after a vertical move.
func (t *TextArea) clampCol() {
	if t.col > len(t.lines[t.row]) {
		t.col = len(t.lines[t.row])
	}
}

func (t *TextArea) caretByteOffset(row, col int) int {
	offset := 0
	for r := 0; r < row && r < len(t.lines); r++ {
		offset += len(string(t.lines[r])) + 1
	}
	if row >= 0 && row < len(t.lines) {
		if col > len(t.lines[row]) {
			col = len(t.lines[row])
		}
		offset += len(string(t.lines[row][:col]))
	}
	return offset
}

func (t *TextArea) caretColBytes(row, col int) int {
	if row >= 0 && row < len(t.lines) {
		if col > len(t.lines[row]) {
			col = len(t.lines[row])
		}
		return len(string(t.lines[row][:col]))
	}
	return 0
}

func (t *TextArea) insert(ins []rune) {
	startByte := t.caretByteOffset(t.row, t.col)
	startColBytes := t.caretColBytes(t.row, t.col)
	insBytes := len(string(ins))

	line := t.lines[t.row]
	next := make([]rune, 0, len(line)+len(ins))
	next = append(next, line[:t.col]...)
	next = append(next, ins...)
	next = append(next, line[t.col:]...)
	t.lines[t.row] = next
	t.col += len(ins)

	if t.highlighter != nil {
		t.highlighter.NotifyEdit(syntax.Edit{
			StartByte:   startByte,
			OldEndByte:  startByte,
			NewEndByte:  startByte + insBytes,
			StartPoint:  syntax.Point{Row: t.row, Column: startColBytes},
			OldEndPoint: syntax.Point{Row: t.row, Column: startColBytes},
			NewEndPoint: syntax.Point{Row: t.row, Column: startColBytes + insBytes},
		})
		_ = t.highlighter.Parse([]byte(t.Value()))
	}
}

// splitLine breaks the current line at the caret, inserting a new line below.
func (t *TextArea) splitLine() {
	startByte := t.caretByteOffset(t.row, t.col)
	startColBytes := t.caretColBytes(t.row, t.col)

	line := t.lines[t.row]
	head := append([]rune{}, line[:t.col]...)
	tail := append([]rune{}, line[t.col:]...)
	t.lines[t.row] = head
	// Insert tail as a new line after row.
	t.lines = append(t.lines, nil)
	copy(t.lines[t.row+2:], t.lines[t.row+1:])
	t.lines[t.row+1] = tail
	t.row++
	t.col = 0

	if t.highlighter != nil {
		t.highlighter.NotifyEdit(syntax.Edit{
			StartByte:   startByte,
			OldEndByte:  startByte,
			NewEndByte:  startByte + 1,
			StartPoint:  syntax.Point{Row: t.row - 1, Column: startColBytes},
			OldEndPoint: syntax.Point{Row: t.row - 1, Column: startColBytes},
			NewEndPoint: syntax.Point{Row: t.row, Column: 0},
		})
		_ = t.highlighter.Parse([]byte(t.Value()))
	}
}

func (t *TextArea) backspace() {
	if t.col > 0 {
		delRune := t.lines[t.row][t.col-1]
		delBytes := len(string(delRune))
		startByte := t.caretByteOffset(t.row, t.col-1)
		startColBytes := t.caretColBytes(t.row, t.col-1)

		line := t.lines[t.row]
		t.lines[t.row] = append(line[:t.col-1], line[t.col:]...)
		t.col--

		if t.highlighter != nil {
			t.highlighter.NotifyEdit(syntax.Edit{
				StartByte:   startByte,
				OldEndByte:  startByte + delBytes,
				NewEndByte:  startByte,
				StartPoint:  syntax.Point{Row: t.row, Column: startColBytes},
				OldEndPoint: syntax.Point{Row: t.row, Column: startColBytes + delBytes},
				NewEndPoint: syntax.Point{Row: t.row, Column: startColBytes},
			})
			_ = t.highlighter.Parse([]byte(t.Value()))
		}
		return
	}
	if t.row > 0 {
		// Join the current line onto the end of the previous one.
		prev := t.lines[t.row-1]
		t.col = len(prev)
		startByte := t.caretByteOffset(t.row-1, len(prev))
		startColBytes := t.caretColBytes(t.row-1, len(prev))

		t.lines[t.row-1] = append(prev, t.lines[t.row]...)
		t.lines = append(t.lines[:t.row], t.lines[t.row+1:]...)
		t.row--

		if t.highlighter != nil {
			t.highlighter.NotifyEdit(syntax.Edit{
				StartByte:   startByte,
				OldEndByte:  startByte + 1,
				NewEndByte:  startByte,
				StartPoint:  syntax.Point{Row: t.row, Column: startColBytes},
				OldEndPoint: syntax.Point{Row: t.row + 1, Column: 0},
				NewEndPoint: syntax.Point{Row: t.row, Column: startColBytes},
			})
			_ = t.highlighter.Parse([]byte(t.Value()))
		}
	}
}

func (t *TextArea) deleteForward() {
	line := t.lines[t.row]
	if t.col < len(line) {
		delRune := line[t.col]
		delBytes := len(string(delRune))
		startByte := t.caretByteOffset(t.row, t.col)
		startColBytes := t.caretColBytes(t.row, t.col)

		t.lines[t.row] = append(line[:t.col], line[t.col+1:]...)

		if t.highlighter != nil {
			t.highlighter.NotifyEdit(syntax.Edit{
				StartByte:   startByte,
				OldEndByte:  startByte + delBytes,
				NewEndByte:  startByte,
				StartPoint:  syntax.Point{Row: t.row, Column: startColBytes},
				OldEndPoint: syntax.Point{Row: t.row, Column: startColBytes + delBytes},
				NewEndPoint: syntax.Point{Row: t.row, Column: startColBytes},
			})
			_ = t.highlighter.Parse([]byte(t.Value()))
		}
		return
	}
	if t.row < len(t.lines)-1 {
		// Pull the next line up onto this one.
		startByte := t.caretByteOffset(t.row, len(line))
		startColBytes := t.caretColBytes(t.row, len(line))

		t.lines[t.row] = append(line, t.lines[t.row+1]...)
		t.lines = append(t.lines[:t.row+1], t.lines[t.row+2:]...)

		if t.highlighter != nil {
			t.highlighter.NotifyEdit(syntax.Edit{
				StartByte:   startByte,
				OldEndByte:  startByte + 1,
				NewEndByte:  startByte,
				StartPoint:  syntax.Point{Row: t.row, Column: startColBytes},
				OldEndPoint: syntax.Point{Row: t.row + 1, Column: 0},
				NewEndPoint: syntax.Point{Row: t.row, Column: startColBytes},
			})
			_ = t.highlighter.Parse([]byte(t.Value()))
		}
	}
}

func (t *TextArea) resolveStyle(capture string) Style {
	var sgr string
	for {
		if t.StyleResolver != nil {
			sgr = t.StyleResolver.Resolve(capture)
		} else {
			sgr = syntax.DefaultStyleResolver().Resolve(capture)
		}
		if sgr != "" || !strings.Contains(capture, ".") {
			break
		}
		idx := strings.LastIndex(capture, ".")
		capture = capture[:idx]
	}
	if sgr == "" {
		return Style{}
	}
	return applySGRSequence(Style{}, sgr)
}

// Draw renders the visible lines into r, scrolling so the caret stays in view,
// and — when focused — places the canvas cursor at the caret. The placeholder
// shows only when the entire buffer is empty.
func (t *TextArea) Draw(c *Canvas, r Rect, focused bool) {
	t.lastH = r.H
	// Scroll so the caret row is within [scroll, scroll+r.H).
	if t.row < t.scroll {
		t.scroll = t.row
	}
	if r.H > 0 && t.row >= t.scroll+r.H {
		t.scroll = t.row - r.H + 1
	}
	if t.scroll < 0 {
		t.scroll = 0
	}

	var spansByLine map[int][]syntax.Span
	if t.highlighter != nil && r.H > 0 {
		spansByLine = t.highlighter.HighlightViewport(t.scroll, t.scroll+r.H)
	}

	lineStartRune := 0
	for i := 0; i < t.scroll && i < len(t.lines); i++ {
		lineStartRune += len(t.lines[i]) + 1
	}

	empty := len(t.lines) == 1 && len(t.lines[0]) == 0
	for row := 0; row < r.H; row++ {
		y := r.Y + row
		c.PaintSurface(Rect{r.X, y, r.W, 1}, Style{})
		li := t.scroll + row
		if li >= len(t.lines) {
			continue
		}
		curLineStartRune := lineStartRune
		lineStartRune += len(t.lines[li]) + 1

		if row == 0 && empty && t.Placeholder != "" {
			c.Write(r.X, y, t.Placeholder, Style{Dim: true})
			continue
		}

		runes := t.lines[li]
		if len(runes) == 0 {
			continue
		}

		var spans []syntax.Span
		if spansByLine != nil {
			spans = spansByLine[li]
		}

		if len(spans) == 0 {
			line := TruncateText(string(runes), r.W, "")
			c.Write(r.X, y, line, Style{})
			continue
		}

		styles := make([]Style, len(runes))
		for _, span := range spans {
			colStart := span.StartRune - curLineStartRune
			colEnd := span.EndRune - curLineStartRune
			if span.StartRune < curLineStartRune && span.StartRune >= 0 && span.EndRune <= len(runes) {
				colStart = span.StartRune
				colEnd = span.EndRune
			}
			if colStart < 0 {
				colStart = 0
			}
			if colEnd > len(runes) {
				colEnd = len(runes)
			}
			if colStart >= colEnd {
				continue
			}
			st := t.resolveStyle(span.Capture)
			for k := colStart; k < colEnd; k++ {
				styles[k] = st
			}
		}

		col := r.X
		start := 0
		for start < len(runes) && col < r.X+r.W {
			end := start + 1
			for end < len(runes) && styles[end] == styles[start] {
				end++
			}
			remW := (r.X + r.W) - col
			runStr := TruncateText(string(runes[start:end]), remW, "")
			if len(runStr) == 0 {
				break
			}
			n := c.Write(col, y, runStr, styles[start])
			if n <= 0 {
				break
			}
			col += n
			start = end
		}
	}

	if focused {
		cy := t.row - t.scroll
		if cy >= 0 && cy < r.H {
			c.CursorX = r.X + StringWidth(string(t.lines[t.row][:t.col]))
			c.CursorY = r.Y + cy
		}
	}
}
