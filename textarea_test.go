// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/syntax"
)

func taType(t *loom.TextArea, texts ...string) {
	for _, s := range texts {
		t.HandleKey(loom.KeyEvent{Text: s})
	}
}

func TestTextAreaValueRoundTrip(t *testing.T) {
	ta := loom.NewTextArea("line one\nline two")
	if ta.Value() != "line one\nline two" {
		t.Fatalf("round trip = %q", ta.Value())
	}
	if ta.LineCount() != 2 {
		t.Errorf("LineCount = %d, want 2", ta.LineCount())
	}
	r, col := ta.Caret()
	if r != 1 || col != len("line two") {
		t.Errorf("seed caret = (%d,%d), want (1,8)", r, col)
	}
}

func TestTextAreaContentDrivenHeightBounds(t *testing.T) {
	ta := loom.NewTextArea("one\ntwo\nthree")
	if got := ta.ContentHeight(); got != 3 {
		t.Fatalf("default content height = %d, want 3", got)
	}
	ta.HandleKey(loom.KeyEvent{Key: "enter"})
	if got := ta.Measure(20).Height; got != 4 {
		t.Fatalf("measured height after inserting a line = %d, want 4", got)
	}
	ta.MinHeight = 4
	if got := ta.Measure(20).Height; got != 4 {
		t.Fatalf("height with minimum = %d, want 4", got)
	}
	ta.MaxHeight = 5
	if got := ta.ContentHeight(); got != 4 {
		t.Fatalf("height with minimum and maximum = %d, want 4", got)
	}
	ta.MaxHeight = 2
	ta.MinHeight = 1
	if got := ta.ContentHeight(); got != 2 {
		t.Fatalf("height with maximum = %d, want 2", got)
	}
	ta.HandleKey(loom.KeyEvent{Text: "!"})
	if got := ta.ContentHeight(); got != 2 {
		t.Fatalf("height after editing at maximum = %d, want 2", got)
	}
	ta.SetValue("only")
	if got := ta.ContentHeight(); got != 1 {
		t.Fatalf("height after shrinking below minimum = %d, want 1", got)
	}
}

func TestTextAreaEnterSplitsLine(t *testing.T) {
	ta := loom.NewTextArea("abcd")
	ta.HandleKey(loom.KeyEvent{Key: "home"})
	ta.HandleKey(loom.KeyEvent{Key: "right"})
	ta.HandleKey(loom.KeyEvent{Key: "right"}) // caret after "ab"
	ta.HandleKey(loom.KeyEvent{Key: "enter"})
	if ta.Value() != "ab\ncd" {
		t.Errorf("enter split = %q, want \"ab\\ncd\"", ta.Value())
	}
	r, col := ta.Caret()
	if r != 1 || col != 0 {
		t.Errorf("caret after split = (%d,%d), want (1,0)", r, col)
	}
}

func TestTextAreaPasteInsertsMultilineText(t *testing.T) {
	ta := loom.NewTextArea("ab")
	ta.SetCaret(0, 1)
	if !loom.DispatchPasteEvent(ta, loom.PasteEvent{Text: "x\ny"}).Consumed {
		t.Fatal("paste was not consumed")
	}
	if got := ta.Value(); got != "ax\nyb" {
		t.Fatalf("paste value = %q, want %q", got, "ax\nyb")
	}
	if r, c := ta.Caret(); r != 1 || c != 1 {
		t.Fatalf("caret = (%d,%d), want (1,1)", r, c)
	}
}

func TestTextAreaBackspaceJoinsLines(t *testing.T) {
	ta := loom.NewTextArea("ab\ncd")
	ta.HandleKey(loom.KeyEvent{Key: "home"}) // start of "cd" (caret seeded on last line)
	ta.HandleKey(loom.KeyEvent{Key: "backspace"})
	if ta.Value() != "abcd" {
		t.Errorf("backspace join = %q, want abcd", ta.Value())
	}
	r, col := ta.Caret()
	if r != 0 || col != 2 {
		t.Errorf("caret after join = (%d,%d), want (0,2)", r, col)
	}
}

func TestTextAreaDeleteMergesNextLine(t *testing.T) {
	ta := loom.NewTextArea("ab\ncd")
	ta.HandleKey(loom.KeyEvent{Key: "up"})  // to line 0
	ta.HandleKey(loom.KeyEvent{Key: "end"}) // end of "ab"
	ta.HandleKey(loom.KeyEvent{Key: "delete"})
	if ta.Value() != "abcd" {
		t.Errorf("delete merge = %q, want abcd", ta.Value())
	}
}

func TestTextAreaVerticalCaretClamp(t *testing.T) {
	ta := loom.NewTextArea("longline\nx")
	// Caret seeded at (1,1) on "x"; moving up must clamp col to len("longline")? No —
	// up keeps col then clamps to the shorter target line. Here target is longer.
	ta.HandleKey(loom.KeyEvent{Key: "up"})
	_, col := ta.Caret()
	if col != 1 {
		t.Errorf("col after up = %d, want 1 (preserved)", col)
	}
	// Move to end of the long line, then down onto the short line clamps col.
	ta.HandleKey(loom.KeyEvent{Key: "end"})
	ta.HandleKey(loom.KeyEvent{Key: "down"})
	r, col := ta.Caret()
	if r != 1 || col != 1 {
		t.Errorf("caret after down-clamp = (%d,%d), want (1,1)", r, col)
	}
}

func TestTextAreaScrollKeepsCaretVisible(t *testing.T) {
	ta := loom.NewTextArea("l0\nl1\nl2\nl3\nl4") // 5 lines; caret on l4
	c := loom.NewCanvas(10, 3)                   // only 3 rows visible
	ta.Draw(c, c.Bounds(), true)
	// Caret is on the last line; it must be visible (cursor set within bounds).
	if c.CursorY < 0 || c.CursorY >= 3 {
		t.Errorf("caret not scrolled into view: CursorY=%d", c.CursorY)
	}
}

func TestTextAreaPlaceholder(t *testing.T) {
	ta := loom.NewTextArea("")
	ta.Placeholder = "body here"
	c := loom.NewCanvas(20, 3)
	ta.Draw(c, c.Bounds(), true)
	if got := strings.TrimRight(rowText(c, 0), " "); got != "body here" {
		t.Errorf("placeholder not drawn, row0 = %q", got)
	}
	// Typing replaces the placeholder.
	taType(ta, "x")
	c.Clear()
	ta.Draw(c, c.Bounds(), true)
	if got := strings.TrimRight(rowText(c, 0), " "); got != "x" {
		t.Errorf("after typing, row0 = %q, want \"x\"", got)
	}
}

func TestTextAreaSyntaxHighlighting(t *testing.T) {
	ta := loom.NewTextArea("package main\nfunc hello() {\n\treturn\n}")
	lexer := syntax.NewLexicalEngine("go")
	ta.SetHighlighter(lexer)
	if ta.Highlighter() != lexer {
		t.Fatalf("Highlighter() = %v, want %v", ta.Highlighter(), lexer)
	}

	c := loom.NewCanvas(40, 4)
	ta.Draw(c, c.Bounds(), false)

	// Check line 0: "package" should be styled as keyword (bold, magenta: ColorIndex(5))
	cellP := c.Get(0, 0)
	if cellP.Text != "p" {
		t.Errorf("cell (0,0) text = %q, want \"p\"", cellP.Text)
	}
	if !cellP.Style.Bold || cellP.Style.FG != loom.ColorIndex(5) {
		t.Errorf("cell (0,0) style = %+v, want Bold: true, FG: ColorIndex(5)", cellP.Style)
	}

	// Space after "package" should have default style
	cellSpace := c.Get(7, 0)
	if cellSpace.Text != " " || cellSpace.Style.Bold {
		t.Errorf("cell (7,0) space style = %+v, want default style", cellSpace.Style)
	}

	// Line 1: "func" should be styled as keyword
	cellF := c.Get(0, 1)
	if cellF.Text != "f" || !cellF.Style.Bold || cellF.Style.FG != loom.ColorIndex(5) {
		t.Errorf("cell (0,1) \"func\" style = %+v, want Bold: true, FG: ColorIndex(5)", cellF.Style)
	}
}

func TestTextAreaCustomStyleResolver(t *testing.T) {
	ta := loom.NewTextArea("package main")
	ta.SetHighlighter(syntax.NewLexicalEngine("go"))
	// Custom resolver: keyword is red (ANSI 31)
	ta.StyleResolver = syntax.StyleResolver{"keyword": "31"}

	c := loom.NewCanvas(30, 2)
	ta.Draw(c, c.Bounds(), false)

	cellP := c.Get(0, 0)
	if cellP.Style.Bold || cellP.Style.FG != loom.ColorIndex(1) {
		t.Errorf("custom resolver style = %+v, want Bold: false, FG: ColorIndex(1)", cellP.Style)
	}
}

func TestTextAreaEditIncrementalHighlighting(t *testing.T) {
	ta := loom.NewTextArea("func run() {}")
	ta.SetHighlighter(syntax.NewLexicalEngine("go"))

	c := loom.NewCanvas(30, 2)
	ta.Draw(c, c.Bounds(), false)
	if !c.Get(0, 0).Style.Bold {
		t.Fatalf("expected func to be bold keyword before edit")
	}

	// Move to start of line and type comment "// "
	ta.HandleKey(loom.KeyEvent{Key: "home"})
	taType(ta, "// ")

	c.Clear()
	ta.Draw(c, c.Bounds(), false)

	// Now line begins with "// func run() {}" which is a comment (ColorIndex(8))
	cellSlash := c.Get(0, 0)
	if cellSlash.Text != "/" || cellSlash.Style.FG != loom.ColorIndex(8) {
		t.Errorf("comment style after edit = %+v, want FG: ColorIndex(8)", cellSlash.Style)
	}
}

type mockViewportEngine struct {
	syntax.NullEngine
	lastStart, lastEnd int
	queryCount         int
}

func (m *mockViewportEngine) HighlightViewport(startLine, endLine int) map[int][]syntax.Span {
	m.lastStart = startLine
	m.lastEnd = endLine
	m.queryCount++
	return map[int][]syntax.Span{}
}

func TestTextAreaViewportBoundedHighlighting(t *testing.T) {
	ta := loom.NewTextArea("0\n1\n2\n3\n4\n5\n6\n7\n8\n9")
	for i := 0; i < 9; i++ {
		ta.HandleKey(loom.KeyEvent{Key: "up"})
	}
	mock := &mockViewportEngine{}
	ta.SetHighlighter(mock)

	c := loom.NewCanvas(20, 3)
	// Visible lines with scroll=0: [0, 3)
	ta.Draw(c, c.Bounds(), false)

	if mock.queryCount == 0 {
		t.Fatalf("HighlightViewport was never queried")
	}
	if mock.lastStart != 0 || mock.lastEnd != 3 {
		t.Errorf("query range = [%d, %d), want [0, 3)", mock.lastStart, mock.lastEnd)
	}

	// Move caret down so scrolling occurs
	for i := 0; i < 6; i++ {
		ta.HandleKey(loom.KeyEvent{Key: "down"})
	}
	mock.queryCount = 0
	ta.Draw(c, c.Bounds(), true)

	// Caret is at row 6, height is 3, scroll should be 4, range [4, 7)
	if mock.lastStart != 4 || mock.lastEnd != 7 {
		t.Errorf("scrolled query range = [%d, %d), want [4, 7)", mock.lastStart, mock.lastEnd)
	}
}
