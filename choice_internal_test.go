// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"testing"
	"time"
)

func TestChoiceMouseHitRegion(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		width      int
		wantStart  int
		wantEnd    int
		wantInside bool
	}{
		{name: "text", line: "  alpha", width: 10, wantStart: 2, wantEnd: 7, wantInside: true},
		{name: "trailing whitespace", line: "alpha   ", width: 10, wantStart: 0, wantEnd: 5, wantInside: true},
		{name: "empty", line: "       ", width: 10, wantInside: false},
		{name: "marker and gaps", line: "[ ] alpha  desc", width: 20, wantStart: 0, wantEnd: 15, wantInside: true},
		{name: "wide and combining", line: "中e\u0301 ", width: 10, wantStart: 0, wantEnd: 3, wantInside: true},
		{name: "truncated", line: "alpha…", width: 6, wantStart: 0, wantEnd: 6, wantInside: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, ok := choiceMouseHitRegion(tt.line, tt.width)
			if start != tt.wantStart || end != tt.wantEnd || ok != tt.wantInside {
				t.Fatalf("choiceMouseHitRegion(%q, %d) = (%d, %d, %v), want (%d, %d, %v)", tt.line, tt.width, start, end, ok, tt.wantStart, tt.wantEnd, tt.wantInside)
			}
		})
	}
}

func TestChoiceRowTextPinsDisplayWidthTruncation(t *testing.T) {
	c := NewChoice([]Item{{Name: "界界"}})
	c.sel = 0
	got := c.choiceRowText(0, 5)
	if got != "▶ 界…" {
		t.Fatalf("choiceRowText = %q, want wide-rune-safe ellipsis truncation", got)
	}
	if StringWidth(got) != 5 {
		t.Fatalf("choiceRowText width = %d, want 5", StringWidth(got))
	}
}

func TestChoiceSetItemsAndSelectIndex(t *testing.T) {
	c := NewChoice([]Item{{Name: "alpha"}, {Name: "beta"}})
	c.ConsumeKey(KeyEvent{Text: "beta"})
	if got := c.Query(); got != "beta" {
		t.Fatalf("query = %q, want beta", got)
	}

	c.SetItems([]Item{{Name: "one"}, {Name: "two"}, {Name: "three"}})
	if got := c.Query(); got != "" {
		t.Fatalf("query after SetItems = %q, want empty", got)
	}
	if got, ok := c.Selected(); !ok || got.Name != "one" {
		t.Fatalf("selection after SetItems = %+v, ok=%v; want one", got, ok)
	}

	c.SelectIndex(99)
	if got, ok := c.Selected(); !ok || got.Name != "three" {
		t.Fatalf("selection after high SelectIndex = %+v, ok=%v; want three", got, ok)
	}
	c.SelectIndex(-1)
	if got, ok := c.Selected(); !ok || got.Name != "one" {
		t.Fatalf("selection after low SelectIndex = %+v, ok=%v; want one", got, ok)
	}
}

func TestChoiceDrawAndMouseUseSameRowText(t *testing.T) {
	c := NewChoice([]Item{{Name: "alpha", Desc: "description"}})
	c.MouseTextOnly = true
	c.Draw(NewCanvas(20, 3), Rect{W: 20, H: 3})
	if got := c.choiceRowText(0, 20); got != "▶ alpha  description" {
		t.Fatalf("shared row text = %q", got)
	}
	start, end, ok := choiceMouseHitRegion(c.choiceRowText(0, 20), 20)
	if !ok || start != 0 || end != 20 {
		t.Fatalf("shared hit region = (%d, %d, %v)", start, end, ok)
	}
}

// makeChoice builds a Choice with n named items (item-00, item-01, …).
func makeChoice(n int) *Choice {
	items := make([]Item, n)
	for i := range items {
		items[i] = Item{Name: fmt.Sprintf("item-%02d", i)}
	}
	return NewChoice(items)
}

// TestConsumeMouseScrollOffset verifies that mouse hit-tests honor the list's
// scroll offset: once Draw has scrolled the view (viewOffset > 0), a click on a
// visible row must select the item actually shown there, not viewOffset rows
// above it. Regression test for the offset bug in Choice.ConsumeMouse.
func TestConsumeMouseScrollOffset(t *testing.T) {
	const total = 20
	const paneRows = 6 // r.H; itemRows = r.H-1 = 5 visible item rows

	c := makeChoice(total)
	// Select an item far down so clampView scrolls the view past the top.
	c.sel = 18
	cv := NewCanvas(40, paneRows)
	c.Draw(cv, Rect{X: 0, Y: 0, W: 40, H: paneRows})

	if c.viewOffset == 0 {
		t.Fatalf("expected the view to be scrolled (viewOffset>0), got 0")
	}

	// e.Y is 0-based canvas-absolute; row 0 is the first visible item, which is
	// c.filtered[c.viewOffset]. A click there must select that item.
	for row := 0; row < paneRows-1; row++ {
		want := c.viewOffset + row
		if want >= total {
			break
		}
		c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: row})
		if c.sel != want {
			t.Errorf("click at Y=%d (viewOffset=%d): sel=%d, want %d",
				row, c.viewOffset, c.sel, want)
		}
	}
}

// TestConsumeMouseClickSelects covers the unscrolled case: a top-of-list click
// selects the corresponding row, and a click past the last item is ignored.
func TestConsumeMouseClickSelects(t *testing.T) {
	c := makeChoice(3)
	cv := NewCanvas(40, 6)
	c.Draw(cv, Rect{X: 0, Y: 0, W: 40, H: 6})

	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1})
	if c.sel != 1 {
		t.Errorf("click at Y=2: sel=%d, want 1", c.sel)
	}

	// A click below the last item (Y beyond len) must not change the selection.
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 8})
	if c.sel != 1 {
		t.Errorf("out-of-range click changed sel to %d, want 1", c.sel)
	}
}

func TestChoiceMouseConfirmationCompletesWithoutQuittingHost(t *testing.T) {
	c := makeChoice(2)
	c.Draw(NewCanvas(40, 4), Rect{W: 40, H: 4})
	result := c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1})
	if !result.Done || result.Quit || !result.Consumed {
		t.Fatalf("mouse confirmation = %+v, want consumed completion without quit", result)
	}
}

func TestChoiceConfirmationPathsCompleteWithoutQuittingHost(t *testing.T) {
	for _, mode := range []string{"enter", "multi-select", "double-click"} {
		t.Run(mode, func(t *testing.T) {
			c := makeChoice(2)
			c.MultiSelect = mode == "multi-select"
			c.DoubleClickToActivate = mode == "double-click"
			c.Draw(NewCanvas(40, 4), Rect{W: 40, H: 4})
			var result EventResult
			if c.DoubleClickToActivate {
				now := time.Unix(0, 0)
				c.doubleClick = NewDoubleClickRecognizer(func() time.Time { return now })
				click := MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1}
				if first := c.ConsumeMouse(click); first != Handled() || c.done {
					t.Fatalf("first click = %+v, done=%t; want selection only", first, c.done)
				}
				c.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, Y: 1})
				now = now.Add(100 * time.Millisecond)
				result = c.ConsumeMouse(click)
			} else {
				result = c.ConsumeKey(KeyEvent{Key: "enter"})
			}
			if result != DoneResult() || !c.done || c.Aborted() {
				t.Fatalf("confirmation = %+v, done=%t, aborted=%t; want completion without quit", result, c.done, c.Aborted())
			}
		})
	}
}

func TestChoiceMouseWheelAndPromptHitTest(t *testing.T) {
	c := makeChoice(8)
	c.Draw(NewCanvas(20, 4), Rect{W: 20, H: 4})
	c.ConsumeMouse(MouseEvent{Action: MouseScrollDown, Y: 1})
	if c.sel != 1 {
		t.Fatalf("wheel down selected %d, want 1", c.sel)
	}
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 3}) // prompt row
	if c.sel != 1 || c.done {
		t.Fatal("clicking the prompt selected a file")
	}
	c.ConsumeMouse(MouseEvent{Action: MouseScrollUp, Y: 1})
	if c.sel != 0 {
		t.Fatalf("wheel up selected %d, want 0", c.sel)
	}
}

func TestChoiceSelectOnlyOnClick(t *testing.T) {
	c := makeChoice(3)
	c.SelectOnlyOnClick = true
	selected := false
	c.OnSelect = func(Item) { selected = true }
	c.Draw(NewCanvas(20, 4), Rect{W: 20, H: 4})
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1})
	if c.sel != 1 || selected {
		t.Fatalf("click selected index %d, invoked callback %v", c.sel, selected)
	}
	c.ConsumeKey(KeyEvent{Key: "enter"})
	if !selected {
		t.Fatal("Enter did not invoke OnSelect after click")
	}
}

func TestChoiceDoubleClickSelectsThenActivatesOnce(t *testing.T) {
	c := makeChoice(3)
	c.DoubleClickToActivate = true
	now := time.Unix(0, 0)
	c.doubleClick = NewDoubleClickRecognizer(func() time.Time { return now })
	activations := 0
	c.OnSelect = func(Item) { activations++ }
	c.Draw(NewCanvas(20, 4), Rect{W: 20, H: 4})
	click := MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1}

	c.ConsumeMouse(click)
	if c.sel != 1 || activations != 0 {
		t.Fatalf("first click selected %d and activated %d times; want selection 1 without activation", c.sel, activations)
	}
	c.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, Y: 1})
	now = now.Add(100 * time.Millisecond)
	c.ConsumeMouse(click)
	if c.sel != 1 || activations != 1 {
		t.Fatalf("second click selected %d and activated %d times; want selection 1 and one activation", c.sel, activations)
	}
	c.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, Y: 1})
	if activations != 1 {
		t.Fatalf("release repeated activation %d times", activations)
	}
}

func TestChoiceDoubleClickNeedsMatchingItem(t *testing.T) {
	c := makeChoice(3)
	c.DoubleClickToActivate = true
	now := time.Unix(0, 0)
	c.doubleClick = NewDoubleClickRecognizer(func() time.Time { return now })
	activations := 0
	c.OnSelect = func(Item) { activations++ }
	c.Draw(NewCanvas(20, 4), Rect{W: 20, H: 4})
	first := MouseEvent{Action: MousePress, Button: MouseLeft, Y: 0}
	second := MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1}

	c.ConsumeMouse(first)
	now = now.Add(10 * time.Millisecond)
	c.ConsumeMouse(second)
	if activations != 0 {
		t.Fatalf("clicks on different rows activated %d times", activations)
	}
	now = now.Add(10 * time.Millisecond)
	c.ConsumeMouse(second)
	if activations != 1 {
		t.Fatalf("second matching row click activated %d times; want 1", activations)
	}
}

func TestChoiceDoubleClickOverridesOnlySingleClickConfirmation(t *testing.T) {
	c := makeChoice(3)
	c.SelectOnlyOnClick = true
	c.DoubleClickToActivate = true
	now := time.Unix(0, 0)
	c.doubleClick = NewDoubleClickRecognizer(func() time.Time { return now })
	activations := 0
	c.OnSelect = func(Item) { activations++ }
	c.Draw(NewCanvas(20, 4), Rect{W: 20, H: 4})
	click := MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1}

	c.ConsumeMouse(click)
	if activations != 0 {
		t.Fatal("SelectOnlyOnClick allowed a single-click activation")
	}
	c.ConsumeKey(KeyEvent{Key: "enter"})
	if activations != 1 {
		t.Fatal("Enter no longer activates a SelectOnlyOnClick choice")
	}
	now = now.Add(time.Second)
	c.ConsumeMouse(click)
	if activations != 1 {
		t.Fatal("click after the timeout activated")
	}
	now = now.Add(10 * time.Millisecond)
	c.ConsumeMouse(click)
	if activations != 2 {
		t.Fatalf("recognized double-click activated %d times total; want Enter plus one double-click", activations)
	}
}

func TestChoiceSingleClickConfirmationRemainsDefault(t *testing.T) {
	c := makeChoice(3)
	activations := 0
	c.OnSelect = func(Item) { activations++ }
	c.Draw(NewCanvas(20, 4), Rect{W: 20, H: 4})
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1})
	if activations != 1 {
		t.Fatalf("default single click activated %d times, want 1", activations)
	}
}

func TestChoiceScrollbarTrackClick(t *testing.T) {
	c := makeChoice(30)
	c.SelectOnlyOnClick = true
	c.Draw(NewCanvas(25, 8), Rect{X: 2, Y: 1, W: 20, H: 6})
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 19, Y: 4})
	if c.viewOffset != 25 || c.sel != 25 {
		t.Fatalf("bottom track click: offset=%d sel=%d, want 25", c.viewOffset, c.sel)
	}
	c.Draw(NewCanvas(25, 8), Rect{X: 2, Y: 1, W: 20, H: 6})
	if c.viewOffset != 25 {
		t.Fatalf("redraw snapped viewport to %d", c.viewOffset)
	}
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 19, Y: 0})
	if c.viewOffset != 0 || c.sel != 4 {
		t.Fatalf("top track click: offset=%d sel=%d, want 0 and 4", c.viewOffset, c.sel)
	}
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 19, Y: 5}) // prompt
	if c.viewOffset != 0 || c.sel != 4 {
		t.Fatal("prompt-row click moved scrollbar")
	}
}

func TestChoiceScrollbarDragCancelAndCapture(t *testing.T) {
	c := makeChoice(40)
	c.sel = 20
	c.Draw(NewCanvas(25, 8), Rect{W: 20, H: 6})
	c.viewOffset = 5
	c.Draw(NewCanvas(25, 8), Rect{W: 20, H: 6})
	start := c.viewOffset
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 19, Y: 2})
	c.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 19, Y: 99})
	if c.viewOffset == start {
		t.Fatal("captured choice drag did not update outside the widget")
	}
	c.ConsumeKey(KeyEvent{Key: "esc"})
	if c.viewOffset != start || c.drag.active {
		t.Fatalf("choice escape cancel: offset=%d active=%v, want %d false", c.viewOffset, c.drag.active, start)
	}
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseRight, X: 19, Y: 2})
	if c.drag.active {
		t.Fatal("choice non-primary press started drag")
	}
}

func TestChoiceDrawsScrollbarTrackOutsidePrompt(t *testing.T) {
	thumb := SpeccedDefaults.Scrollbar.ForegroundChar
	track := SpeccedDefaults.Scrollbar.BackgroundChar
	c := makeChoice(20)
	c.Style.Scrollbar = ScrollbarStyle{
		Track: Style{FG: ColorIndex(33), BG: ColorIndex(27)},
		Thumb: Style{FG: ColorIndex(51), BG: ColorIndex(27)},
	}
	canvas := NewCanvas(10, 5)
	c.Draw(canvas, Rect{W: 10, H: 5})
	if canvas.Get(9, 0).Text != thumb {
		t.Fatal("top thumb missing")
	}
	if got := canvas.Get(9, 0).Style; got != c.Style.Scrollbar.Thumb {
		t.Fatalf("thumb style = %+v, want %+v", got, c.Style.Scrollbar.Thumb)
	}
	for y := 1; y < 4; y++ {
		if got := canvas.Get(9, y).Text; got != track {
			t.Fatalf("track row %d = %q, want %q", y, got, track)
		}
		if got := canvas.Get(9, y).Style; got != c.Style.Scrollbar.Track {
			t.Fatalf("track row %d style = %+v, want %+v", y, got, c.Style.Scrollbar.Track)
		}
	}
	if got := canvas.Get(9, 4).Text; got != " " {
		t.Fatalf("prompt row got scrollbar track %q", got)
	}
	c = makeChoice(2)
	c.Draw(canvas, Rect{W: 10, H: 5})
	if got := canvas.Get(9, 0).Text; got == track || got == thumb {
		t.Fatalf("non-scrollable choice has track %q", got)
	}
}

func TestChoiceScrollbarModeDefaultsToAutoAndCanBeDisabled(t *testing.T) {
	if got := SpeccedDefaults.Scrollbar.Mode; got != ScrollbarAuto {
		t.Fatalf("spec scrollbar mode = %q, want %q", got, ScrollbarAuto)
	}
	c := makeChoice(20)
	canvas := NewCanvas(10, 5)
	c.Draw(canvas, canvas.Bounds())
	if got := canvas.Get(9, 0).Text; got != SpeccedDefaults.Scrollbar.ForegroundChar {
		t.Fatalf("default overflow cell = %q, want scrollbar thumb", got)
	}

	c = makeChoice(2)
	canvas.Clear()
	c.Draw(canvas, canvas.Bounds())
	for y := 0; y < 4; y++ {
		if got := canvas.Get(9, y).Text; got == SpeccedDefaults.Scrollbar.ForegroundChar || got == SpeccedDefaults.Scrollbar.BackgroundChar {
			t.Fatalf("non-overflow row %d reserved scrollbar cell %q", y, got)
		}
	}
	c.ScrollbarMode = ScrollbarAlways
	c.Draw(canvas, canvas.Bounds())
	if got := canvas.Get(9, 0).Text; got != SpeccedDefaults.Scrollbar.ForegroundChar {
		t.Fatalf("always-mode cell = %q, want scrollbar thumb", got)
	}

	c = makeChoice(20)
	c.ScrollbarMode = ScrollbarNever
	c.SelectOnlyOnClick = true
	canvas.Clear()
	c.Draw(canvas, canvas.Bounds())
	for y := 0; y < 4; y++ {
		if got := canvas.Get(9, y).Text; got == SpeccedDefaults.Scrollbar.ForegroundChar || got == SpeccedDefaults.Scrollbar.BackgroundChar {
			t.Fatalf("opt-out row %d reserved scrollbar cell %q", y, got)
		}
	}
	c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 9, Y: 3})
	if c.viewOffset != 0 {
		t.Fatalf("opt-out track click changed view offset to %d", c.viewOffset)
	}
}

func TestChoiceUsesPlaceholderStyleOnlyForPlaceholder(t *testing.T) {
	c := makeChoice(1)
	c.Prompt = "filter> "
	c.Placeholder = "type"
	c.Style.Prompt = Style{FG: ColorIndex(0), BG: ColorIndex(51)}
	c.Style.Placeholder = Style{FG: ColorIndex(240), BG: ColorIndex(51)}
	canvas := NewCanvas(20, 2)
	c.Draw(canvas, canvas.Bounds())

	promptY := canvas.Rows() - 1
	if got := canvas.Get(0, promptY).Style; got != c.Style.Prompt {
		t.Fatalf("prompt style = %+v, want %+v", got, c.Style.Prompt)
	}
	wantHint := c.Style.Placeholder
	wantHint.Dim = true // the hint is always dim
	if got := canvas.Get(StringWidth(c.Prompt), promptY).Style; got != wantHint {
		t.Fatalf("placeholder style = %+v, want %+v", got, wantHint)
	}
	if got := canvas.Get(StringWidth(c.Prompt)+StringWidth(c.Placeholder), promptY).Style; got != c.Style.Prompt {
		t.Fatalf("prompt-row fill style = %+v, want %+v", got, c.Style.Prompt)
	}
}

// TestBackspaceEmptyQueryDoesNotLeaveView verifies that pressing backspace with no
// filter text does not dismiss the view (abort), while backspace with a non-empty
// query trims the filter.
func TestBackspaceEmptyQueryDoesNotLeaveView(t *testing.T) {
	c := makeChoice(3)

	// With a filter query, backspace trims it and does not quit.
	c.query = "it"
	if quit := c.ConsumeKey(KeyEvent{Key: "backspace"}).Quit; quit {
		t.Fatal("backspace with a query should not quit")
	}
	if c.query != "i" {
		t.Errorf("backspace should trim query to %q, got %q", "i", c.query)
	}

	// Trimming to empty does not quit on the keystroke that empties it.
	if quit := c.ConsumeKey(KeyEvent{Key: "backspace"}).Quit; quit {
		t.Fatal("backspace that empties the query should not quit")
	}
	if c.query != "" {
		t.Errorf("query should be empty, got %q", c.query)
	}

	// The next backspace on the empty query is a no-op and does not leave the view.
	if quit := c.ConsumeKey(KeyEvent{Key: "backspace"}).Quit; quit {
		t.Fatal("backspace on empty query should not quit the view")
	}
	if c.Aborted() {
		t.Error("backspace on empty query should not mark choice as aborted")
	}
}

// TestChoicePgUpPgDnAndHomeEnd verifies that PgUp, PgDn, Home, and End navigate
// the list by page or boundary steps.
func TestChoicePgUpPgDnAndHomeEnd(t *testing.T) {
	c := makeChoice(30)
	c.itemRows = 10

	// Home / End
	c.ConsumeKey(KeyEvent{Key: "end"})
	if c.sel != 29 {
		t.Fatalf("End: sel = %d, want 29", c.sel)
	}
	c.ConsumeKey(KeyEvent{Key: "home"})
	if c.sel != 0 {
		t.Fatalf("Home: sel = %d, want 0", c.sel)
	}

	// PgDown jumps by max(1, itemRows-1) = 9
	c.ConsumeKey(KeyEvent{Key: "pgdown"})
	if c.sel != 9 {
		t.Fatalf("PgDown: sel = %d, want 9", c.sel)
	}
	c.ConsumeKey(KeyEvent{Key: "pgdn"})
	if c.sel != 18 {
		t.Fatalf("PgDn: sel = %d, want 18", c.sel)
	}

	// PgUp jumps back
	c.ConsumeKey(KeyEvent{Key: "pgup"})
	if c.sel != 9 {
		t.Fatalf("PgUp: sel = %d, want 9", c.sel)
	}
	c.ConsumeKey(KeyEvent{Key: "pageup"})
	if c.sel != 0 {
		t.Fatalf("PageUp: sel = %d, want 0", c.sel)
	}
}

// TestMultiSelectToggleAndChecked verifies Space toggles the caret item and
// Checked() returns the set in original list order regardless of toggle order.
func TestMultiSelectToggleAndChecked(t *testing.T) {
	c := makeChoice(4) // item-00..item-03
	c.MultiSelect = true

	// Move to item-02 and toggle it, then back to item-00 and toggle it.
	c.ConsumeKey(KeyEvent{Key: "down"})
	c.ConsumeKey(KeyEvent{Key: "down"})
	c.ConsumeKey(KeyEvent{Text: " "}) // check item-02
	c.ConsumeKey(KeyEvent{Key: "up"})
	c.ConsumeKey(KeyEvent{Key: "up"})
	c.ConsumeKey(KeyEvent{Text: " "}) // check item-00

	got := c.Checked()
	if len(got) != 2 || got[0].Name != "item-00" || got[1].Name != "item-02" {
		t.Fatalf("Checked() = %v, want [item-00 item-02] in list order", names(got))
	}

	// Toggling item-00 again unchecks it.
	c.ConsumeKey(KeyEvent{Key: "up"}) // already at top; stays/wraps — re-seat on item-00
	c.sel = 0
	c.ConsumeKey(KeyEvent{Text: " "})
	got = c.Checked()
	if len(got) != 1 || got[0].Name != "item-02" {
		t.Errorf("after untoggle, Checked() = %v, want [item-02]", names(got))
	}
}

// TestMultiSelectFilterPreservesChecks verifies checks are keyed by name and
// survive filtering the item out of view.
func TestMultiSelectFilterPreservesChecks(t *testing.T) {
	c := makeChoice(4)
	c.MultiSelect = true
	c.sel = 1
	c.ConsumeKey(KeyEvent{Text: " "}) // check item-01

	// Filter to "item-03" — item-01 is no longer in the filtered list.
	for _, r := range "03" {
		c.ConsumeKey(KeyEvent{Text: string(r)})
	}
	if len(c.filtered) != 1 || c.filtered[0].Name != "item-03" {
		t.Fatalf("filter setup wrong: %v", names(c.filtered))
	}
	// The check on the now-hidden item-01 is preserved.
	got := c.Checked()
	if len(got) != 1 || got[0].Name != "item-01" {
		t.Errorf("filtering dropped the check: Checked() = %v, want [item-01]", names(got))
	}
}

// TestMultiSelectMouseToggleHonorsScroll verifies a left-click toggles (never
// confirms) and respects the scroll offset.
func TestMultiSelectMouseToggleHonorsScroll(t *testing.T) {
	c := makeChoice(10)
	c.MultiSelect = true
	c.viewOffset = 3 // list scrolled so row 1 shows item-03

	quit := c.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 0}).Quit
	if quit {
		t.Error("multi-select click should not confirm/quit")
	}
	got := c.Checked()
	if len(got) != 1 || got[0].Name != "item-03" {
		t.Errorf("click toggled wrong item: Checked() = %v, want [item-03]", names(got))
	}
}

// TestMultiSelectEscAbortsToEmpty verifies Esc yields an empty set.
func TestMultiSelectEscAbortsToEmpty(t *testing.T) {
	c := makeChoice(3)
	c.MultiSelect = true
	c.ConsumeKey(KeyEvent{Text: " "}) // check item-00
	if got := c.ConsumeKey(KeyEvent{Key: "esc"}); !got.Quit {
		t.Fatal("Esc should quit")
	}
	if got := c.Checked(); got != nil {
		t.Errorf("aborted picker should return nil Checked(), got %v", names(got))
	}
}

// names extracts Item.Name for readable assertions.
func names(items []Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Name
	}
	return out
}
