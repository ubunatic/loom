package editor

import (
	"fmt"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

// Issue 287 M4: a formatting popover open from a selection against the search
// panel. The popover captures mouse input through the modal contract, so it must
// also be painted above the search panel (RichTextEdit.DeferPopover).

// popover287 opens search, then a popover on a selection whose popover bar
// lands inside the search panel's columns and rows. Search stays open, with a
// query, and the editor holds the focus.
func popover287(t *testing.T) (*editView, *loom.Canvas) {
	t.Helper()
	long := strings.Repeat("0123456789", 10)
	view, _ := new287View(t, "first line\n"+long+"\n"+long+"\nfourth line\n")
	draw287(view, 100, 24)
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	type287(view, "line")
	view.focused = focusEditor
	draw287(view, 100, 24)
	if !view.showSearch || view.searchRect.W == 0 {
		t.Fatalf("setup: search panel not drawn: %+v", view.searchRect)
	}
	// The popover bar sits two rows above the selection: select on the third
	// document line so the bar lands on the editor's first row, which is also
	// the search panel's top border.
	x := view.searchRect.X - view.editorRect.X + 4
	view.edit.SetSelection(loom.RichPosition{Line: 2, Offset: x}, loom.RichPosition{Line: 2, Offset: x + 6})
	// A selection opens the popover by itself on the next draw; ctrl-space
	// would toggle it closed again.
	c := draw287(view, 100, 24)
	if !view.edit.ModalOpen() || !view.edit.HasSelection {
		t.Fatalf("setup: popover not open\n%s", canvasScreenText(c))
	}
	return view, c
}

// requirePopoverGoneSame287 checks that a click dismissed the popover without
// touching what the editor owns. Dismissing a popover clears its selection
// (docs/Widgets.md, pinned for backdrop clicks by Test304EditBackdrop...), so
// the selection must end up cleared, exactly like a backdrop click on the
// editor, and never moved or extended by the search panel.
func requirePopoverGoneSame287(t *testing.T, what string, want state287, v *editView) {
	t.Helper()
	got := snap287(v)
	if got.ansi != want.ansi {
		t.Fatalf("%s changed the document: %q, want %q", what, got.ansi, want.ansi)
	}
	if got.modified != want.modified {
		t.Fatalf("%s changed IsModified: %v, want %v", what, got.modified, want.modified)
	}
	if got.cursor != want.cursor {
		t.Fatalf("%s moved the editing cursor: %+v, want %+v", what, got.cursor, want.cursor)
	}
	if got.hasSelection {
		t.Fatalf("%s left a selection after dismissing the popover: %+v %+v", what, got.from, got.to)
	}
}

// Probe 1: a click on the search panel with the popover open belongs to the
// popover layer only. It is consumed, dismisses the popover and changes nothing
// in the search bar or the editor.
func Test287PopoverClickOnSearchPanelIsConsumedByPopover(t *testing.T) {
	// Relative to searchRect and clear of the popover's own cells (the bar spans
	// the panel's first row and its pointer sits on the second), so each click
	// reaches the popover only as a backdrop click.
	spots := []struct {
		name string
		x, y int
	}{
		{"border", 30, 0},
		{"query row", 10, 1},
		{"mode row", 3, 2},
		{"count row", 3, 3},
		{"hint row", 3, 4},
		{"right edge of query row", 47, 1},
	}
	for _, spot := range spots {
		view, _ := popover287(t)
		before := snap287(view)
		focus, query, regex, index := view.focused, view.searchBar.Query, view.regexMode, view.searchIndex
		res := click287(view, view.searchRect.X+spot.x, view.searchRect.Y+spot.y)
		if !res.Consumed {
			t.Fatalf("%s: click on the search panel under a popover was not consumed: %+v", spot.name, res)
		}
		if view.edit.ModalOpen() {
			t.Fatalf("%s: click did not dismiss the popover", spot.name)
		}
		if view.focused != focus {
			t.Fatalf("%s: click changed the search focus: %v, want %v", spot.name, view.focused, focus)
		}
		if view.searchBar.Query != query || view.regexMode != regex || view.searchIndex != index {
			t.Fatalf("%s: click reached the search bar: query %q mode %v index %d, want %q %v %d",
				spot.name, view.searchBar.Query, view.regexMode, view.searchIndex, query, regex, index)
		}
		requirePopoverGoneSame287(t, spot.name, before, view)
	}
}

// Probe 2: the popover captures every click, so wherever the popover is drawn
// it must be what the user sees. Cells of an active popover that the search
// panel paints over are a draw order bug.
func Test287PopoverIsNotPaintedOverBySearchPanel(t *testing.T) {
	view, withSearch := popover287(t)

	// The popover's own cells: the same frame without search, with and without
	// the popover.
	view.showSearch = false
	withPopover := draw287(view, 100, 24)
	view.edit.ShowPopover = false
	withoutPopover := draw287(view, 100, 24)
	view.edit.ShowPopover = true
	view.showSearch = true
	draw287(view, 100, 24)

	rect := view.searchRect
	popoverCells, overlap, hidden := 0, 0, []string{}
	for y := 0; y < withPopover.Rows(); y++ {
		for x := 0; x < withPopover.Cols(); x++ {
			want := withPopover.Get(x, y)
			if want == withoutPopover.Get(x, y) {
				continue
			}
			popoverCells++
			if !rect.Contains(x, y) {
				continue
			}
			overlap++
			got := withSearch.Get(x, y)
			if want.Style.BG == (loom.Color{}) {
				// The pointer glyph sets no background of its own, so it
				// correctly takes whatever surface lies under it.
				got.Style.BG = want.Style.BG
			}
			if got != want {
				hidden = append(hidden, fmt.Sprintf("(%d,%d) popover %q, search panel paints %q", x, y, want.Text, got.Text))
			}
		}
	}
	if popoverCells == 0 || overlap == 0 {
		t.Fatalf("setup: popover cells %d, inside search panel %d; the probe needs an overlap\n%s",
			popoverCells, overlap, canvasScreenText(withPopover))
	}
	if len(hidden) > 0 {
		t.Fatalf("search panel %+v paints over %d of %d active popover cells inside it (clicks go to the popover):\n%s\n"+
			"popover alone:\n%s\nwith search panel:\n%s",
			rect, len(hidden), overlap, strings.Join(hidden, "\n"),
			rowsText287(withPopover, rect.Y, rect.Y+3), rowsText287(withSearch, rect.Y, rect.Y+3))
	}
}

// rowsText287 renders canvas rows [from, to) as plain text for failure output.
func rowsText287(c *loom.Canvas, from, to int) string {
	var out []string
	for y := from; y < to && y < c.Rows(); y++ {
		var row strings.Builder
		for x := 0; x < c.Cols(); x++ {
			cell := c.Get(x, y)
			if cell.Continuation {
				continue
			}
			if cell.Text == "" {
				row.WriteString(" ")
			} else {
				row.WriteString(cell.Text)
			}
		}
		out = append(out, fmt.Sprintf("%2d|%s", y, row.String()))
	}
	return strings.Join(out, "\n")
}

// Probe 3: mode clicks while the popover is open reach the popover only; the
// search mode stays as it was in both directions.
func Test287PopoverModeClicksDoNotSwitchSearchMode(t *testing.T) {
	defs := loom.SpeccedDefaults.Editor
	for _, regex := range []bool{false, true} {
		view, _ := popover287(t)
		view.regexMode = regex
		label := defs.RegexLabel
		if regex {
			label = defs.NormalLabel
		}
		view.searchIndex = 0
		view.updateSearchMatches()
		// Draw with search above the popover is not needed to locate the label:
		// the mode row is below the popover bar.
		c := draw287(view, 100, 24)
		if !view.edit.ModalOpen() {
			t.Fatalf("regex=%v: setup lost the popover", regex)
		}
		x, y := locate287(t, c, label)
		before := snap287(view)
		focus := view.focused
		res := click287(view, x, y)
		if !res.Consumed {
			t.Fatalf("regex=%v: click on %q was not consumed: %+v", regex, label, res)
		}
		if view.regexMode != regex {
			t.Fatalf("regex=%v: click on %q under a popover switched the search mode to %v", regex, label, view.regexMode)
		}
		if view.focused != focus {
			t.Fatalf("regex=%v: click on %q changed the search focus: %v, want %v", regex, label, view.focused, focus)
		}
		if view.edit.ModalOpen() {
			t.Fatalf("regex=%v: click on %q did not dismiss the popover", regex, label)
		}
		requirePopoverGoneSame287(t, label+" click", before, view)
	}
}
