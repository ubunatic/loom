// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"slices"
	"strings"
	"testing"
	"time"
)

func typeEntryText(e *TextEntry, text string) {
	for _, r := range text {
		e.ConsumeKey(KeyEvent{Text: string(r)})
	}
}

func entryKey(e *TextEntry, key string) EventResult {
	return e.ConsumeKey(KeyEvent{Key: key})
}

// entryScreenRow returns row y with one entry per display column, so index
// equals column. A wide rune's trailing continuation column is "".
func entryScreenRow(c *Canvas, y int) []string {
	cols := make([]string, c.Cols())
	for x := range cols {
		cell := c.Get(x, y)
		if !cell.Continuation {
			cols[x] = cell.Text
		}
	}
	return cols
}

// entryColumnOf returns the display column where the cluster text starts on row y, or -1.
func entryColumnOf(c *Canvas, y int, text string) int {
	for x, cell := range entryScreenRow(c, y) {
		if cell == text {
			return x
		}
	}
	return -1
}

func TestTextEntryTypingNotifiesChange(t *testing.T) {
	var seen []string
	e := NewTextEntry(TextEntryOptions{OnChange: func(value string) { seen = append(seen, value) }})
	typeEntryText(e, "hi")
	if e.Value() != "hi" {
		t.Fatalf("Value() = %q, want %q", e.Value(), "hi")
	}
	if want := []string{"h", "hi"}; !slices.Equal(seen, want) {
		t.Fatalf("OnChange values = %q, want %q", seen, want)
	}
}

func TestTextEntryEnterSubmitsOrIsLeftToHost(t *testing.T) {
	bare := NewTextEntry(TextEntryOptions{})
	typeEntryText(bare, "draft")
	if r := entryKey(bare, "enter"); r.Consumed {
		t.Fatal("Enter without OnSubmit must be left to the host")
	}
	if bare.Value() != "draft" {
		t.Fatalf("Enter without OnSubmit changed the value to %q", bare.Value())
	}

	var submitted string
	e := NewTextEntry(TextEntryOptions{OnSubmit: func(value string) EventResult {
		submitted = value
		return Consumed()
	}})
	typeEntryText(e, "ab")
	if r := entryKey(e, "enter"); !r.Consumed {
		t.Fatal("Enter with OnSubmit must return the handler's consumed result")
	}
	if submitted != "ab" {
		t.Fatalf("OnSubmit received %q, want %q", submitted, "ab")
	}
}

func TestTextEntryNewlineKeysInsertLineBreak(t *testing.T) {
	for _, key := range []string{"shift-enter", "ctrl-j", "alt-enter"} {
		submitted := false
		e := NewTextEntry(TextEntryOptions{OnSubmit: func(string) EventResult {
			submitted = true
			return Consumed()
		}})
		typeEntryText(e, "a")
		entryKey(e, key)
		typeEntryText(e, "b")
		if submitted {
			t.Fatalf("%s submitted the entry", key)
		}
		if e.Value() != "a\nb" {
			t.Fatalf("%s: Value() = %q, want %q", key, e.Value(), "a\nb")
		}
	}
}

func TestTextEntryBindingsAcceptPageKeyAliases(t *testing.T) {
	var submitted string
	e := NewTextEntry(TextEntryOptions{
		SubmitKeys: []string{"pagedown"},
		OnSubmit: func(value string) EventResult {
			submitted = value
			return Consumed()
		},
	})
	typeEntryText(e, "go")
	if r := entryKey(e, "pgdown"); !r.Consumed || submitted != "go" {
		t.Fatalf("binding pagedown did not match decoder key pgdown: result %+v, submitted %q", r, submitted)
	}
}

func TestTextEntryHostKeysAreIgnored(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	typeEntryText(e, "keep")
	for _, key := range []string{"tab", "shift-tab", "ctrl-o", "ctrl-s", "ctrl-g", "ctrl-f", "esc", "f1"} {
		if r := entryKey(e, key); r.Consumed || r.Quit || r.Done {
			t.Fatalf("%s = %+v, want plain Ignored so the host keeps it", key, r)
		}
	}
	if e.Value() != "keep" {
		t.Fatalf("host keys changed the value to %q", e.Value())
	}
}

func TestTextEntryShiftSelectionCutAndPaste(t *testing.T) {
	var clip []string
	e := NewTextEntry(TextEntryOptions{})
	bindTextEntryClipboardForTest(e, func(text string) { clip = append(clip, text) })
	typeEntryText(e, "hello world")
	for range 5 {
		entryKey(e, "shift-left")
	}
	if r := entryKey(e, "ctrl-x"); !r.Consumed {
		t.Fatal("ctrl-x with a selection must cut")
	}
	if e.Value() != "hello " {
		t.Fatalf("after cut Value() = %q, want %q", e.Value(), "hello ")
	}
	if want := []string{"world"}; !slices.Equal(clip, want) {
		t.Fatalf("cut wrote %q to the clipboard, want %q", clip, want)
	}
	entryKey(e, "ctrl-v")
	if e.Value() != "hello world" {
		t.Fatalf("after paste Value() = %q, want %q", e.Value(), "hello world")
	}
}

func TestTextEntrySelectAllCopyAndNoSelectionCopyIgnored(t *testing.T) {
	var clip []string
	e := NewTextEntry(TextEntryOptions{})
	bindTextEntryClipboardForTest(e, func(text string) { clip = append(clip, text) })
	typeEntryText(e, "one\ntwo")
	if r := entryKey(e, "ctrl-c"); r.Consumed {
		t.Fatal("ctrl-c without a selection must be left to the host")
	}
	if len(clip) != 0 {
		t.Fatalf("copy without selection wrote %q", clip)
	}
	entryKey(e, "ctrl-a")
	entryKey(e, "ctrl-c")
	if want := []string{"one\ntwo"}; !slices.Equal(clip, want) {
		t.Fatalf("select-all copy wrote %q, want %q", clip, want)
	}
}

func TestTextEntryClipboardIsWiredLikePaneBinding(t *testing.T) {
	var clip []string
	e := NewTextEntry(TextEntryOptions{})
	bindClipboardTree(e, func(text string) { clip = append(clip, text) })
	typeEntryText(e, "osc")
	entryKey(e, "ctrl-a")
	entryKey(e, "ctrl-c")
	if want := []string{"osc"}; !slices.Equal(clip, want) {
		t.Fatalf("bindClipboardTree writer got %q, want %q", clip, want)
	}
}

func TestTextEntryUndoRedoCoalescesTyping(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	typeEntryText(e, "abc")
	entryKey(e, "backspace")
	if e.Value() != "ab" {
		t.Fatalf("backspace left %q, want %q", e.Value(), "ab")
	}
	entryKey(e, "ctrl-z")
	if e.Value() != "abc" {
		t.Fatalf("undo gave %q, want %q", e.Value(), "abc")
	}
	entryKey(e, "ctrl-z")
	if e.Value() != "" {
		t.Fatalf("second undo gave %q, want the typed run undone as one step", e.Value())
	}
	if r := entryKey(e, "ctrl-z"); r.Consumed {
		t.Fatal("undo with empty history must be left to the host")
	}
	entryKey(e, "ctrl-y")
	if e.Value() != "abc" {
		t.Fatalf("redo gave %q, want %q", e.Value(), "abc")
	}
}

func TestTextEntryPasteAndSetValueAreSingleUndoSteps(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	typeEntryText(e, "a")
	e.ConsumePaste(PasteEvent{Text: "x\r\ny"})
	if e.Value() != "ax\ny" {
		t.Fatalf("paste gave %q, want CRLF normalised to LF", e.Value())
	}
	entryKey(e, "ctrl-z")
	if e.Value() != "a" {
		t.Fatalf("undo after paste gave %q, want %q", e.Value(), "a")
	}

	e.SetValue("fresh")
	if e.Value() != "fresh" {
		t.Fatalf("SetValue gave %q", e.Value())
	}
	entryKey(e, "ctrl-z")
	if e.Value() != "a" {
		t.Fatalf("undo after SetValue gave %q, want %q", e.Value(), "a")
	}
}

func TestTextEntryNoOpEditsDoNotNotify(t *testing.T) {
	calls := 0
	e := NewTextEntry(TextEntryOptions{OnChange: func(string) { calls++ }})
	if r := entryKey(e, "backspace"); r.Consumed {
		t.Fatal("backspace at the start must be a no-op left to the host")
	}
	if calls != 0 {
		t.Fatalf("no-op backspace notified %d times", calls)
	}
}

func TestTextEntryMeasureGrowsWithSoftWrapAndCapsAtMaxHeight(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "> "})
	if got := e.Measure(12).Height; got != 1 {
		t.Fatalf("empty entry height = %d, want 1", got)
	}
	e.SetValue(strings.Repeat("a", 28))
	// Width 12 minus the 2-cell prompt and one reserved caret cell leaves 9 cells per row.
	if got := e.Measure(12).Height; got != 4 {
		t.Fatalf("28 cells at width 12 wrap to %d rows, want 4", got)
	}
	capped := NewTextEntry(TextEntryOptions{Prompt: "> ", MaxHeight: 2})
	capped.SetValue(strings.Repeat("a", 28))
	if got := capped.Measure(12).Height; got != 2 {
		t.Fatalf("MaxHeight 2 gave height %d", got)
	}
}

func TestTextEntryCursorFollowsSoftWrap(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "> "})
	e.SetValue(strings.Repeat("a", 28))
	e.Measure(12)
	if row, col := e.Cursor(); row != 3 || col != 3 {
		t.Fatalf("end of wrapped text cursor = (%d,%d), want (3,3)", row, col)
	}
	for range 10 {
		entryKey(e, "left")
	}
	if row, col := e.Cursor(); row != 2 || col != 2 {
		t.Fatalf("caret at 18 cursor = (%d,%d), want (2,2)", row, col)
	}
	for range 9 {
		entryKey(e, "left")
	}
	if row, col := e.Cursor(); row != 1 || col != 2 {
		t.Fatalf("caret at a wrap boundary cursor = (%d,%d), want start of the next row (1,2)", row, col)
	}
}

func TestTextEntryUpDownMoveAcrossWrappedRows(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "> "})
	e.SetValue(strings.Repeat("a", 28))
	e.Measure(12)
	entryKey(e, "up")
	if row, _ := e.Cursor(); row != 2 {
		t.Fatalf("up from the last row landed on row %d, want 2", row)
	}
	if e.caret != 19 {
		t.Fatalf("up kept goal column 1 at caret %d, want 19", e.caret)
	}
	entryKey(e, "down")
	if e.caret != 28 {
		t.Fatalf("down returned to caret %d, want 28", e.caret)
	}
	entryKey(e, "up")
	entryKey(e, "up")
	entryKey(e, "up")
	if r := entryKey(e, "up"); r.Consumed {
		t.Fatal("up on the first row must be left to the host for history")
	}
	if r := entryKey(e, "up"); r.Consumed {
		t.Fatal("repeated up on the first row must stay with the host")
	}
}

func TestTextEntryDownOnLastRowIsLeftToHost(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("only")
	e.Measure(20)
	if r := entryKey(e, "down"); r.Consumed {
		t.Fatal("down on the last row must be left to the host")
	}
}

func TestTextEntryWideRunesWrapByDisplayWidth(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("日本語日本語")
	// Width 10 minus the reserved caret cell leaves 9 cells, so four wide runes fit per row.
	if got := e.Measure(10).Height; got != 2 {
		t.Fatalf("six wide runes at width 10 wrap to %d rows, want 2", got)
	}
	e.Measure(10)
	e.caret = 5
	if row, col := e.Cursor(); row != 1 || col != 2 {
		t.Fatalf("caret after five wide runes = (%d,%d), want (1,2)", row, col)
	}
}

func TestTextEntryPageKeysMoveByVisibleRows(t *testing.T) {
	c := NewCanvas(20, 2)
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("a\nb\nc\nd")
	e.Draw(c, Rect{W: 20, H: 2})
	// Cursor() is relative to the scrolled view, so these checks read the caret's
	// document row. A page is the two visible rows.
	docRow := func() int {
		row, _ := e.locate(e.rows(e.layoutWidth), e.caret)
		return row
	}
	// The caret starts at the end of "d" on row 3.
	for _, key := range []string{"pgup", "pageup"} {
		e.SetValue("a\nb\nc\nd")
		entryKey(e, key)
		if row := docRow(); row != 1 {
			t.Fatalf("%s from the last row landed on row %d, want 1 (two visible rows)", key, row)
		}
	}
	entryKey(e, "pagedown")
	if row := docRow(); row != 3 {
		t.Fatalf("pagedown from row 1 landed on row %d, want 3", row)
	}
	if r := entryKey(e, "pgdn"); r.Consumed {
		t.Fatal("pgdn alias on the last row moves nothing and must be left to the host")
	}
}

func TestTextEntryMouseClickPlacesCaretAndTypesThere(t *testing.T) {
	c := NewCanvas(20, 3)
	e := NewTextEntry(TextEntryOptions{Prompt: "> "})
	e.SetValue("hello world")
	e.Draw(c, Rect{W: 20, H: 3})
	// Column 8 is the 'w' (prompt 2 cells plus "hello " 6 cells); a press lands before it.
	if r := e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 8, Y: 0}); !r.Consumed {
		t.Fatal("press inside the entry must be consumed")
	}
	typeEntryText(e, "X")
	if e.Value() != "hello Xworld" {
		t.Fatalf("typing after click gave %q, want %q", e.Value(), "hello Xworld")
	}
}

func TestTextEntryMouseDragSelectsRange(t *testing.T) {
	c := NewCanvas(20, 3)
	var clip []string
	e := NewTextEntry(TextEntryOptions{Prompt: "> "})
	bindTextEntryClipboardForTest(e, func(text string) { clip = append(clip, text) })
	e.SetValue("hello world")
	e.Draw(c, Rect{W: 20, H: 3})
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 0})
	if r := e.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 7, Y: 0}); !r.Consumed {
		t.Fatal("drag during a press must be consumed")
	}
	e.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 7, Y: 0})
	entryKey(e, "ctrl-c")
	if want := []string{"hello"}; !slices.Equal(clip, want) {
		t.Fatalf("drag selection copied %q, want %q", clip, want)
	}
	if r := e.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 9, Y: 0}); r.Consumed {
		t.Fatal("drag after release must not be consumed")
	}
}

func TestTextEntryMouseDoubleAndTripleClickSelectWordAndLine(t *testing.T) {
	c := NewCanvas(20, 3)
	var clip []string
	e := NewTextEntry(TextEntryOptions{Prompt: "> "})
	bindTextEntryClipboardForTest(e, func(text string) { clip = append(clip, text) })
	clock := time.Unix(1000, 0)
	e.now = func() time.Time { return clock }
	e.SetValue("hello world")
	e.Draw(c, Rect{W: 20, H: 3})

	// Column 8 lies in "world" (2 cells prompt, "hello " six cells).
	press := MouseEvent{Action: MousePress, Button: MouseLeft, X: 8, Y: 0}
	e.ConsumeMouse(press)
	e.ConsumeMouse(press)
	entryKey(e, "ctrl-c")
	if want := []string{"world"}; !slices.Equal(clip, want) {
		t.Fatalf("double click copied %q, want %q", clip, want)
	}
	e.ConsumeMouse(press)
	entryKey(e, "ctrl-c")
	if want := []string{"world", "hello world"}; !slices.Equal(clip, want) {
		t.Fatalf("triple click copied %q, want the whole line; got %q", clip, clip)
	}
}

func TestTextEntryMouseWheelScrollsWithoutSnappingBack(t *testing.T) {
	c := NewCanvas(20, 2)
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("l0\nl1\nl2\nl3\nl4\nl5")
	e.Draw(c, Rect{W: 20, H: 2})
	if c.Get(1, 0).Text != "4" {
		t.Fatalf("caret at end should show lines 4 and 5, row 0 = %q", entryScreenRow(c, 0))
	}
	if r := e.ConsumeMouse(MouseEvent{Action: MouseScrollUp, X: 0, Y: 0}); !r.Consumed {
		t.Fatal("wheel over overflowing content must be consumed")
	}
	e.Draw(c, Rect{W: 20, H: 2})
	if c.Get(1, 0).Text != "3" {
		t.Fatalf("after wheel up row 0 = %q, want line 3 to stay in view", entryScreenRow(c, 0))
	}
}

func TestTextEntryDrawKeepsCaretVisibleWithMaxHeight(t *testing.T) {
	c := NewCanvas(20, 2)
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", MaxHeight: 2})
	e.SetFocus(true)
	e.SetValue("one\ntwo\nthree\nfour\nfive")
	e.Draw(c, Rect{W: 20, H: 2})
	if row, _ := e.Cursor(); row != 1 {
		t.Fatalf("caret after five lines is row %d inside a 2-row view, want 1", row)
	}
	// Prompt 2 cells plus "five" 4 cells puts the caret at column 6 of row 1.
	if c.CursorY != 1 || c.CursorX != 2+4 {
		t.Fatalf("hardware cursor at (%d,%d), want (6,1)", c.CursorX, c.CursorY)
	}
	if got := entryScreenRow(c, 0)[2:6]; strings.Join(got, "") != "four" {
		t.Fatalf("visible row 0 = %q, want %q", strings.Join(got, ""), "four")
	}
}

func TestTextEntryDrawPromptPlaceholderAndSelectionCells(t *testing.T) {
	c := NewCanvas(20, 1)
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", Placeholder: "type"})
	e.Draw(c, Rect{W: 20, H: 1})
	if c.Get(0, 0).Text != ">" || c.Get(2, 0).Text != "t" {
		t.Fatalf("prompt and placeholder columns wrong: %q", entryScreenRow(c, 0)[:6])
	}
	if !c.Get(2, 0).Style.Dim {
		t.Fatal("placeholder must be dim")
	}

	e.SetValue("ab cd")
	e.SetFocus(true)
	entryKey(e, "shift-left")
	entryKey(e, "shift-left")
	c = NewCanvas(20, 1)
	e.Draw(c, Rect{W: 20, H: 1})
	// Text starts at column 2; "c" and "d" are units 3 and 4, so columns 5 and 6 are selected.
	for x := 2; x < 9; x++ {
		selected := x == 5 || x == 6
		if got := c.Get(x, 0).Style.Invert; got != selected {
			t.Fatalf("column %d invert = %v, want %v", x, got, selected)
		}
	}
	if c.CursorX != 5 || c.CursorY != 0 {
		t.Fatalf("cursor at (%d,%d) with caret on the selection start, want (5,0)", c.CursorX, c.CursorY)
	}
}

func TestTextEntryWideRuneCellsAndCursorColumns(t *testing.T) {
	c := NewCanvas(10, 1)
	e := NewTextEntry(TextEntryOptions{})
	e.SetFocus(true)
	e.SetValue("日a")
	e.Draw(c, Rect{W: 10, H: 1})
	if got := entryColumnOf(c, 0, "日"); got != 0 {
		t.Fatalf("wide rune starts at column %d, want 0", got)
	}
	if got := entryColumnOf(c, 0, "a"); got != 2 {
		t.Fatalf("letter after wide rune at column %d, want 2 (display width)", got)
	}
	if c.CursorX != 3 {
		t.Fatalf("end caret after 日a at column %d, want 3", c.CursorX)
	}
}

func TestTextEntryCursorOnlyWhenFocused(t *testing.T) {
	c := NewCanvas(10, 1)
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("x")
	e.Draw(c, Rect{W: 10, H: 1})
	if c.CursorX != -1 {
		t.Fatalf("unfocused entry set cursor column %d", c.CursorX)
	}
	e.SetFocus(true)
	if !e.Focused() {
		t.Fatal("SetFocus(true) did not focus the entry")
	}
	e.Draw(c, Rect{W: 10, H: 1})
	if c.CursorX != 1 || c.CursorY != 0 {
		t.Fatalf("focused cursor at (%d,%d), want (1,0)", c.CursorX, c.CursorY)
	}
}

// bindTextEntryClipboardForTest installs a clipboard writer the way Pane does,
// through bindClipboardTree, so every test exercises the production hook.
func bindTextEntryClipboardForTest(e *TextEntry, write func(string)) {
	bindClipboardTree(e, write)
}
