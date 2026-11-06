// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"math"
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

func TestTextEntryHasSelectionHostCopyRouting(t *testing.T) {
	var clip []string
	e := NewTextEntry(TextEntryOptions{})
	bindTextEntryClipboardForTest(e, func(text string) { clip = append(clip, text) })
	hostActions := 0
	routeCopy := func() {
		if e.HasSelection() {
			if r := entryKey(e, "ctrl-c"); !r.Consumed {
				t.Fatal("selected copy was not consumed")
			}
		} else {
			hostActions++
		}
	}
	if e.HasSelection() {
		t.Fatal("empty entry has a selection")
	}
	routeCopy()
	typeEntryText(e, "hello")
	if e.HasSelection() {
		t.Fatal("typed text has a selection")
	}
	routeCopy()
	entryKey(e, "shift-left")
	if !e.HasSelection() {
		t.Fatal("shift-left did not select text")
	}
	routeCopy()
	if want := []string{"o"}; !slices.Equal(clip, want) {
		t.Fatalf("clipboard = %q, want %q", clip, want)
	}
	entryKey(e, "right")
	if e.HasSelection() {
		t.Fatal("collapsed selection still reported")
	}
	routeCopy()
	entryKey(e, "ctrl-a")
	e.SetValue("")
	if e.HasSelection() {
		t.Fatal("cleared entry still has a selection")
	}
	routeCopy()
	if hostActions != 4 || len(clip) != 1 {
		t.Fatalf("host actions = %d, clipboard = %q", hostActions, clip)
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

func TestTextEntryMultiLinePlaceholderDrawsFirstLineOnly(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", Placeholder: "first line\nsecond line"})
	c := NewCanvas(30, 2)
	e.Draw(c, Rect{W: 30, H: 2})
	row0 := strings.TrimRight(strings.Join(entryScreenRow(c, 0)[2:], ""), " ")
	if row0 != "first line" {
		t.Fatalf("placeholder row 0 = %q, want only the first line %q", row0, "first line")
	}
	if got := strings.TrimSpace(strings.Join(entryScreenRow(c, 1), "")); got != "" {
		t.Fatalf("placeholder spilled onto row 1: %q", got)
	}
	for x := 2; x < 12; x++ {
		if !c.Get(x, 0).Style.Dim {
			t.Fatalf("placeholder column %d is not dim", x)
		}
	}
}

func TestTextEntryMultiLinePlaceholderClipsToWidth(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", Placeholder: "first\r\nsecond"})
	c := NewCanvas(6, 1)
	e.Draw(c, Rect{W: 6, H: 1})
	// The prompt takes columns 0 and 1, so the placeholder gets four cells.
	if got := strings.Join(entryScreenRow(c, 0)[2:6], ""); got != "firs" {
		t.Fatalf("clipped placeholder = %q, want %q", got, "firs")
	}
}

func TestTextEntryWidePlaceholderClipsWithoutSplittingRune(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", Placeholder: "界界界"})
	c := NewCanvas(5, 1)
	// Columns 2-3 hold the first rune; the second would need columns 4-5 and stops at width 5.
	e.Draw(c, Rect{W: 5, H: 1})
	if got := entryColumnOf(c, 0, "界"); got != 2 {
		t.Fatalf("first wide placeholder rune at column %d, want 2", got)
	}
	if n := strings.Count(strings.Join(entryScreenRow(c, 0), ""), "界"); n != 1 {
		t.Fatalf("%d wide placeholder runes drawn in width 5, want 1", n)
	}
}

func TestTextEntrySetPlaceholderRedrawsWithoutOnChange(t *testing.T) {
	calls := 0
	e := NewTextEntry(TextEntryOptions{Placeholder: "old", OnChange: func(string) { calls++ }})
	if e.Placeholder() != "old" {
		t.Fatalf("Placeholder() = %q, want %q", e.Placeholder(), "old")
	}
	e.SetPlaceholder("Try: summarise the diff")
	if e.Placeholder() != "Try: summarise the diff" {
		t.Fatalf("Placeholder() after set = %q", e.Placeholder())
	}
	if calls != 0 {
		t.Fatalf("SetPlaceholder notified OnChange %d times, want 0", calls)
	}
	if e.Value() != "" {
		t.Fatalf("placeholder leaked into Value(): %q", e.Value())
	}
	c := NewCanvas(30, 1)
	e.Draw(c, Rect{W: 30, H: 1})
	if got := strings.TrimRight(strings.Join(entryScreenRow(c, 0), ""), " "); got != "Try: summarise the diff" {
		t.Fatalf("redrawn placeholder = %q", got)
	}
	if !c.Get(0, 0).Style.Dim {
		t.Fatal("placeholder set at runtime must be dim")
	}
}

func TestTextEntryPlaceholderHiddenByTypingCaretAtStart(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", Placeholder: "hint"})
	e.SetFocus(true)
	c := NewCanvas(20, 1)
	e.Draw(c, Rect{W: 20, H: 1})
	if c.CursorX != 2 || c.CursorY != 0 {
		t.Fatalf("caret over placeholder at (%d,%d), want (2,0)", c.CursorX, c.CursorY)
	}
	typeEntryText(e, "x")
	c = NewCanvas(20, 1)
	e.Draw(c, Rect{W: 20, H: 1})
	if strings.Contains(strings.Join(entryScreenRow(c, 0), ""), "hint") {
		t.Fatal("placeholder still drawn after typing")
	}
	if e.Value() != "x" {
		t.Fatalf("Value() = %q, want %q", e.Value(), "x")
	}
}

func TestTextEntryAcceptPlaceholderInsertsOnceAndUndoes(t *testing.T) {
	var seen []string
	e := NewTextEntry(TextEntryOptions{Placeholder: "summarise the diff", OnChange: func(v string) { seen = append(seen, v) }})
	if !e.AcceptPlaceholder() {
		t.Fatal("AcceptPlaceholder on an empty entry with a placeholder returned false")
	}
	if e.Value() != "summarise the diff" {
		t.Fatalf("accepted value = %q", e.Value())
	}
	if !slices.Equal(seen, []string{"summarise the diff"}) {
		t.Fatalf("OnChange values after accept = %q, want one call with the text", seen)
	}
	if e.CursorOffset() != len("summarise the diff") {
		t.Fatalf("caret at byte %d, want the end of the inserted text", e.CursorOffset())
	}
	if e.Placeholder() != "summarise the diff" {
		t.Fatal("accepting must keep the placeholder set")
	}
	entryKey(e, "ctrl-z")
	if e.Value() != "" {
		t.Fatalf("undo after accept gave %q, want empty", e.Value())
	}
}

func TestTextEntryAcceptPlaceholderInsertsAllLines(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Placeholder: "a\nb"})
	if !e.AcceptPlaceholder() {
		t.Fatal("AcceptPlaceholder returned false")
	}
	if e.Value() != "a\nb" {
		t.Fatalf("accepted value = %q, want the whole placeholder", e.Value())
	}
}

func TestTextEntryAcceptPlaceholderNoOpWithoutPlaceholderOrWithValue(t *testing.T) {
	calls := 0
	e := NewTextEntry(TextEntryOptions{OnChange: func(string) { calls++ }})
	if e.AcceptPlaceholder() {
		t.Fatal("AcceptPlaceholder with no placeholder must return false")
	}
	e.SetPlaceholder("hint")
	typeEntryText(e, "a")
	calls = 0
	if e.AcceptPlaceholder() {
		t.Fatal("AcceptPlaceholder with a non-empty value must return false")
	}
	if e.Value() != "a" || calls != 0 {
		t.Fatalf("no-op accept changed value to %q with %d OnChange calls", e.Value(), calls)
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

func TestTextEntryCursorOffsetRoundTripsWideRunesAndClusters(t *testing.T) {
	// "a" is 1 byte and 1 cell, "界" is 3 bytes and 2 cells, "é" is one
	// 3-byte cluster of 1 cell, and "b" is 1 byte and 1 cell: 8 bytes in all.
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("a界éb")
	c := NewCanvas(20, 1)
	e.Draw(c, Rect{W: 20, H: 1})
	for _, tc := range []struct{ off, col int }{{0, 0}, {1, 1}, {4, 3}, {7, 4}, {8, 5}} {
		e.SetCursorOffset(tc.off)
		if got := e.CursorOffset(); got != tc.off {
			t.Fatalf("CursorOffset() after SetCursorOffset(%d) = %d", tc.off, got)
		}
		if _, col := e.Cursor(); col != tc.col {
			t.Fatalf("offset %d: caret column %d, want %d", tc.off, col, tc.col)
		}
	}
	// Offsets inside a cluster snap to its start; out-of-range offsets clamp.
	for _, tc := range []struct{ off, want int }{{2, 1}, {3, 1}, {5, 4}, {6, 4}, {-3, 0}, {99, 8}} {
		e.SetCursorOffset(tc.off)
		if got := e.CursorOffset(); got != tc.want {
			t.Fatalf("SetCursorOffset(%d) gives CursorOffset %d, want %d", tc.off, got, tc.want)
		}
	}
}

func TestTextEntrySetCursorOffsetClearsSelectionWithoutChange(t *testing.T) {
	var changes int
	e := NewTextEntry(TextEntryOptions{OnChange: func(string) { changes++ }})
	e.SetValue("abc")
	entryKey(e, "shift-left")
	if !e.hasSelection() {
		t.Fatal("shift-left did not create a selection")
	}
	e.SetCursorOffset(1)
	if e.hasSelection() {
		t.Fatal("SetCursorOffset must clear the selection")
	}
	if e.CursorOffset() != 1 || e.Value() != "abc" || changes != 0 {
		t.Fatalf("offset %d, value %q, OnChange calls %d; want 1, \"abc\", 0",
			e.CursorOffset(), e.Value(), changes)
	}
}

func TestTextEntryReplaceRangeIsOneUndoStepWithoutOnChange(t *testing.T) {
	var changes int
	e := NewTextEntry(TextEntryOptions{OnChange: func(string) { changes++ }})
	e.SetValue("hello world")
	e.ReplaceRange(6, 11, "there")
	if e.Value() != "hello there" {
		t.Fatalf("Value() = %q, want %q", e.Value(), "hello there")
	}
	if e.CursorOffset() != len("hello there") {
		t.Fatalf("caret at %d after ReplaceRange, want the end of the inserted text", e.CursorOffset())
	}
	if changes != 0 {
		t.Fatalf("ReplaceRange called OnChange %d times, want 0", changes)
	}
	entryKey(e, "ctrl-z")
	if e.Value() != "hello world" {
		t.Fatalf("one undo gave %q, want the text from before ReplaceRange", e.Value())
	}
	entryKey(e, "ctrl-y")
	if e.Value() != "hello there" {
		t.Fatalf("redo after undo gave %q, want %q", e.Value(), "hello there")
	}
}

func TestTextEntryReplaceRangeSwapsReversedAndClampsOutOfRange(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("hello world")
	e.ReplaceRange(11, 6, "X")
	if e.Value() != "hello X" {
		t.Fatalf("reversed range gave %q, want %q", e.Value(), "hello X")
	}
	e.SetValue("abc")
	e.ReplaceRange(-5, 100, "")
	if e.Value() != "" || e.CursorOffset() != 0 {
		t.Fatalf("clamped range gave %q at %d, want \"\" at 0", e.Value(), e.CursorOffset())
	}
	e.SetValue("abc")
	e.ReplaceRange(100, 200, "!")
	if e.Value() != "abc!" {
		t.Fatalf("range past the end gave %q, want %q", e.Value(), "abc!")
	}
}

func TestTextEntryReplaceRangeSnapsInsideClusters(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	// "a" is 0..1, "界" is 1..4 and "b" is 4..5.
	e.SetValue("a界b")
	e.ReplaceRange(2, 4, "") // start 2 snaps to 1, the start of 界
	if e.Value() != "ab" {
		t.Fatalf("start inside a cluster gave %q, want %q", e.Value(), "ab")
	}
	e.SetValue("a界b")
	e.ReplaceRange(0, 2, "") // end 2 snaps to 1, so only "a" goes
	if e.Value() != "界b" {
		t.Fatalf("end inside a cluster gave %q, want %q", e.Value(), "界b")
	}
}

func TestTextEntryHighlightColoursSpansAndSelectionWins(t *testing.T) {
	red, blue := ColorRGB(255, 0, 0), ColorRGB(0, 0, 255)
	e := NewTextEntry(TextEntryOptions{
		Prompt: "> ",
		Highlight: func(string) []TextEntrySpan {
			return []TextEntrySpan{
				{Start: 0, End: 3, Style: Style{FG: red}},
				{Start: 4, End: 7, Style: Style{FG: blue}},
			}
		},
	})
	e.SetValue("/go now")
	e.SetFocus(true)
	c := NewCanvas(20, 1)
	e.Draw(c, Rect{W: 20, H: 1})
	n := entryColumnOf(c, 0, "n")
	if n != 6 {
		t.Fatalf("\"n\" at column %d, want 6 after the two-cell prompt", n)
	}
	for x, want := range map[int]Color{2: red, 4: red, 5: ColorReset(), 6: blue, 8: blue} {
		if got := c.Get(x, 0).Style.FG; got != want {
			t.Fatalf("column %d FG = %v, want %v", x, got, want)
		}
	}

	// Select "/g" from the start. Selected cells take the selection style alone.
	e.SetCursorOffset(0)
	entryKey(e, "shift-right")
	entryKey(e, "shift-right")
	c = NewCanvas(20, 1)
	e.Draw(c, Rect{W: 20, H: 1})
	for x := 2; x <= 3; x++ {
		if cell := c.Get(x, 0).Style; cell != (Style{Invert: true}) {
			t.Fatalf("selected column %d style = %+v, want only Invert", x, cell)
		}
	}
	if got := c.Get(4, 0).Style; got.FG != red || got.Invert {
		t.Fatalf("unselected highlighted column 4 style = %+v, want red FG without Invert", got)
	}
}

func TestTextEntryHighlightBadSpansClampAndNeverPanic(t *testing.T) {
	red, green, blue := ColorRGB(255, 0, 0), ColorRGB(0, 255, 0), ColorRGB(0, 0, 255)
	e := NewTextEntry(TextEntryOptions{Highlight: func(string) []TextEntrySpan {
		return []TextEntrySpan{
			{Start: math.MinInt, End: math.MaxInt, Style: Style{FG: ColorIndex(1)}},
			{Start: -5, End: 2, Style: Style{FG: red}},     // clamps to [0, 2)
			{Start: 3, End: 1000, Style: Style{FG: green}}, // clamps to [3, 4)
			{Start: 4, End: 2, Style: Style{FG: blue}},     // reversed: ignored
			{Start: 1, End: 2, Style: Style{FG: blue}},     // overlaps red; later wins
		}
	}})
	e.SetValue("abcd")
	c := NewCanvas(10, 1)
	e.Draw(c, Rect{W: 10, H: 1})
	want := []Color{red, blue, ColorIndex(1), green}
	for x, fg := range want {
		if got := c.Get(x, 0).Style.FG; got != fg {
			t.Fatalf("column %d FG = %v, want %v", x, got, fg)
		}
	}
	empty := NewTextEntry(TextEntryOptions{Highlight: e.highlight})
	empty.Draw(NewCanvas(10, 1), Rect{W: 10, H: 1})
}

func TestTextEntryHighlightSnapsSpanToClusterStart(t *testing.T) {
	red := ColorRGB(255, 0, 0)
	e := NewTextEntry(TextEntryOptions{Highlight: func(string) []TextEntrySpan {
		return []TextEntrySpan{{Start: 2, End: 5, Style: Style{FG: red}}}
	}})
	e.SetValue("a界b")
	c := NewCanvas(10, 1)
	e.Draw(c, Rect{W: 10, H: 1})
	// Start 2 is inside 界 and snaps to its start, so 界 and b are red and a is not.
	if got := c.Get(0, 0).Style.FG; got == red {
		t.Fatal("column 0 (\"a\") is outside the span but is red")
	}
	if got := c.Get(1, 0).Style.FG; got != red {
		t.Fatalf("wide rune column FG = %v, want red", got)
	}
	if b := entryColumnOf(c, 0, "b"); b != 3 || c.Get(b, 0).Style.FG != red {
		t.Fatalf("\"b\" at column %d not red, want column 3 red", b)
	}
}

func TestTextEntryMouseSelectionOverHighlightUsesSelectionStyle(t *testing.T) {
	red := ColorRGB(255, 0, 0)
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", Highlight: func(string) []TextEntrySpan {
		return []TextEntrySpan{{Start: 0, End: 3, Style: Style{FG: red}}}
	}})
	e.SetValue("/go now")
	e.Draw(NewCanvas(20, 1), Rect{W: 20, H: 1})
	// Press on the "/" cell and drag to the "o" cell: selects "/g".
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 0})
	e.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 4, Y: 0})
	c := NewCanvas(20, 1)
	e.Draw(c, Rect{W: 20, H: 1})
	for x := 2; x <= 3; x++ {
		if cell := c.Get(x, 0).Style; cell != (Style{Invert: true}) {
			t.Fatalf("mouse-selected column %d style = %+v, want only Invert", x, cell)
		}
	}
	if got := c.Get(4, 0).Style; got.FG != red || got.Invert {
		t.Fatalf("column 4 after the drag = %+v, want red FG without Invert", got)
	}
}

// bindTextEntryClipboardForTest installs a clipboard writer the way Pane does,
// through bindClipboardTree, so every test exercises the production hook.
func bindTextEntryClipboardForTest(e *TextEntry, write func(string)) {
	bindClipboardTree(e, write)
}

func TestTextEntryCtrlUVisualAndHardLines(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.SetValue("abcdefghij")
	e.Measure(5)
	e.SetCursorOffset(8)
	for _, want := range []struct {
		value  string
		offset int
	}{{"abcdij", 4}, {"ij", 0}} {
		if result := entryKey(e, "ctrl-u"); !result.Consumed || e.Value() != want.value || e.CursorOffset() != want.offset {
			t.Fatalf("ctrl-u = %+v, value %q offset %d, want %q offset %d", result, e.Value(), e.CursorOffset(), want.value, want.offset)
		}
	}
	e.SetValue("a\n\nb")
	e.SetCursorOffset(3)
	if result := entryKey(e, "ctrl-u"); !result.Consumed || e.Value() != "a\n\nb" || e.CursorOffset() != 2 {
		t.Fatalf("hard start: %+v %q %d", result, e.Value(), e.CursorOffset())
	}
}

func TestTextEntryCtrlUEdges(t *testing.T) {
	for _, tc := range []struct {
		name, value   string
		width, offset int
		want          string
		pos           int
		consumed      bool
	}{
		{"empty", "", 0, 0, "", 0, false},
		{"first", "abc", 0, 0, "abc", 0, false},
		{"unknown layout", "abc\ndef", 0, 6, "abc\nf", 4, true},
		{"trailing newline", "a\n", 0, 2, "a\n", 1, true},
		{"interior", "abc\ndef", 0, 6, "abc\nf", 4, true},
		{"wide", "界界xyz", 5, len("界界x"), "界界yz", len("界界"), true},
		{"clusters", "á👩‍💻z", 0, len("á👩‍💻"), "z", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewTextEntry(TextEntryOptions{})
			e.SetValue(tc.value)
			if tc.width > 0 {
				e.Measure(tc.width)
			}
			e.SetCursorOffset(tc.offset)
			got := entryKey(e, "ctrl-u")
			if got.Consumed != tc.consumed || got.Quit || e.Value() != tc.want || e.CursorOffset() != tc.pos {
				t.Fatalf("got %+v %q @%d, want %q @%d", got, e.Value(), e.CursorOffset(), tc.want, tc.pos)
			}
		})
	}
}

func TestTextEntryCtrlUSelectionHistoryAndResize(t *testing.T) {
	changes := 0
	e := NewTextEntry(TextEntryOptions{Prompt: "> ", OnChange: func(string) { changes++ }})
	e.SetValue("abcdef\ngh")
	e.Measure(7)
	e.SetCursorOffset(5)
	entryKey(e, "shift-left")
	entryKey(e, "shift-left")
	entryKey(e, "ctrl-u")
	if e.Value() != "abcf\ngh" || changes != 1 {
		t.Fatalf("selection: %q changes %d", e.Value(), changes)
	}
	entryKey(e, "ctrl-z")
	if e.Value() != "abcdef\ngh" {
		t.Fatal("undo selection")
	}
	e.SetCursorOffset(7)
	entryKey(e, "ctrl-u")
	if changes != 2 || e.CursorOffset() != 6 {
		t.Fatalf("navigation notified or moved incorrectly: %d %d", changes, e.CursorOffset())
	}
	entryKey(e, "ctrl-y")
	if e.Value() != "abcf\ngh" {
		t.Fatal("navigation lost redo")
	}
	e.SetValue("abcdefghij")
	e.Measure(7)
	e.Measure(5)
	e.SetCursorOffset(6)
	entryKey(e, "ctrl-u")
	if e.Value() != "abcdghij" {
		t.Fatalf("resize/prompt: %q", e.Value())
	}
	e.SetValue("hello world")
	e.Draw(NewCanvas(20, 3), Rect{W: 20, H: 3})
	e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 0})
	e.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 7, Y: 0})
	e.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 7, Y: 0})
	entryKey(e, "ctrl-u")
	if e.Value() != " world" {
		t.Fatalf("mouse selection: %q", e.Value())
	}
}
