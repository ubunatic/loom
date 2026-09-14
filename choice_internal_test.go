// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"testing"
)

// makeChoice builds a Choice with n named items (item-00, item-01, …).
func makeChoice(n int) *Choice {
	items := make([]Item, n)
	for i := range items {
		items[i] = Item{Name: fmt.Sprintf("item-%02d", i)}
	}
	return NewChoice(items)
}

// TestHandleMouseScrollOffset verifies that mouse hit-tests honor the list's
// scroll offset: once Draw has scrolled the view (viewOffset > 0), a click on a
// visible row must select the item actually shown there, not viewOffset rows
// above it. Regression test for the offset bug in Choice.HandleMouse.
func TestHandleMouseScrollOffset(t *testing.T) {
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

	// e.Y is 1-based pane-relative; row 1 is the first visible item, which is
	// c.filtered[c.viewOffset]. A click there must select that item.
	for row := 1; row <= paneRows-1; row++ {
		want := c.viewOffset + row - 1
		if want >= total {
			break
		}
		c.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: row})
		if c.sel != want {
			t.Errorf("click at Y=%d (viewOffset=%d): sel=%d, want %d",
				row, c.viewOffset, c.sel, want)
		}
	}
}

// TestHandleMouseClickSelects covers the unscrolled case: a top-of-list click
// selects the corresponding row, and a click past the last item is ignored.
func TestHandleMouseClickSelects(t *testing.T) {
	c := makeChoice(3)
	cv := NewCanvas(40, 6)
	c.Draw(cv, Rect{X: 0, Y: 0, W: 40, H: 6})

	c.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 2})
	if c.sel != 1 {
		t.Errorf("click at Y=2: sel=%d, want 1", c.sel)
	}

	// A click below the last item (Y beyond len) must not change the selection.
	c.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 9})
	if c.sel != 1 {
		t.Errorf("out-of-range click changed sel to %d, want 1", c.sel)
	}
}

func TestChoiceMouseWheelAndPromptHitTest(t *testing.T) {
	c := makeChoice(8)
	c.Draw(NewCanvas(20, 4), Rect{W: 20, H: 4})
	c.HandleMouse(MouseEvent{Action: MouseScrollDown, Y: 1})
	if c.sel != 1 {
		t.Fatalf("wheel down selected %d, want 1", c.sel)
	}
	c.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 4}) // prompt row
	if c.sel != 1 || c.done {
		t.Fatal("clicking the prompt selected a file")
	}
	c.HandleMouse(MouseEvent{Action: MouseScrollUp, Y: 1})
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
	c.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 2})
	if c.sel != 1 || selected {
		t.Fatalf("click selected index %d, invoked callback %v", c.sel, selected)
	}
	c.HandleKey(KeyEvent{Key: "enter"})
	if !selected {
		t.Fatal("Enter did not invoke OnSelect after click")
	}
}

// TestBackspaceEmptyQueryLeavesView verifies that pressing backspace with no
// filter text dismisses the view (aborts), mirroring Esc/`:back`, while
// backspace with a non-empty query only trims the filter.
func TestBackspaceEmptyQueryLeavesView(t *testing.T) {
	c := makeChoice(3)

	// With a filter query, backspace trims it and does not quit.
	c.query = "it"
	if quit := c.HandleKey(KeyEvent{Key: "backspace"}); quit {
		t.Fatal("backspace with a query should not quit")
	}
	if c.query != "i" {
		t.Errorf("backspace should trim query to %q, got %q", "i", c.query)
	}

	// Trimming to empty does not quit on the keystroke that empties it.
	if quit := c.HandleKey(KeyEvent{Key: "backspace"}); quit {
		t.Fatal("backspace that empties the query should not quit")
	}
	if c.query != "" {
		t.Errorf("query should be empty, got %q", c.query)
	}

	// The next backspace on the now-empty query leaves the view.
	if quit := c.HandleKey(KeyEvent{Key: "backspace"}); !quit {
		t.Fatal("backspace on empty query should quit the view")
	}
	if !c.Aborted() {
		t.Error("leaving via backspace should mark the choice aborted")
	}
}

// TestMultiSelectToggleAndChecked verifies Space toggles the caret item and
// Checked() returns the set in original list order regardless of toggle order.
func TestMultiSelectToggleAndChecked(t *testing.T) {
	c := makeChoice(4) // item-00..item-03
	c.MultiSelect = true

	// Move to item-02 and toggle it, then back to item-00 and toggle it.
	c.HandleKey(KeyEvent{Key: "down"})
	c.HandleKey(KeyEvent{Key: "down"})
	c.HandleKey(KeyEvent{Text: " "}) // check item-02
	c.HandleKey(KeyEvent{Key: "up"})
	c.HandleKey(KeyEvent{Key: "up"})
	c.HandleKey(KeyEvent{Text: " "}) // check item-00

	got := c.Checked()
	if len(got) != 2 || got[0].Name != "item-00" || got[1].Name != "item-02" {
		t.Fatalf("Checked() = %v, want [item-00 item-02] in list order", names(got))
	}

	// Toggling item-00 again unchecks it.
	c.HandleKey(KeyEvent{Key: "up"}) // already at top; stays/wraps — re-seat on item-00
	c.sel = 0
	c.HandleKey(KeyEvent{Text: " "})
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
	c.HandleKey(KeyEvent{Text: " "}) // check item-01

	// Filter to "item-03" — item-01 is no longer in the filtered list.
	for _, r := range "03" {
		c.HandleKey(KeyEvent{Text: string(r)})
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

	quit := c.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, Y: 1})
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
	c.HandleKey(KeyEvent{Text: " "}) // check item-00
	if got := c.HandleKey(KeyEvent{Key: "esc"}); !got {
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
