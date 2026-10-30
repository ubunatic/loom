package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

// Issue 287: the search overlay owns its input and never edits the document.
// The cursor-isolation tests live in edit_287_cursor_test.go.

// new287View opens content in an edit view backed by a temp file.
func new287View(t *testing.T, content string) (*editView, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatal(err)
	}
	return view, path
}

func draw287(v *editView, w, h int) *loom.Canvas {
	c := loom.NewCanvas(w, h)
	v.Draw(c, c.Bounds())
	return c
}

// state287 captures everything a search interaction must leave alone.
type state287 struct {
	cursor, from, to loom.RichPosition
	hasSelection     bool
	scrollY          int
	ansi             string
	modified         bool
}

func snap287(v *editView) state287 {
	e := v.edit
	return state287{e.Cursor, e.SelectionFrom, e.SelectionTo, e.HasSelection, e.ScrollY, e.Document.ToANSI(), e.IsModified()}
}

// requireSame287 fails when the editing state differs. The cursor, selection
// and scroll offset can be excluded for tests whose subject is the document.
func requireSame287(t *testing.T, what string, want state287, v *editView, cursor, scroll bool) {
	t.Helper()
	got := snap287(v)
	if cursor && got.cursor != want.cursor {
		t.Fatalf("%s moved the editing cursor: %+v, want %+v", what, got.cursor, want.cursor)
	}
	// Without a selection, From/To are leftover anchors that a click rewrites.
	if got.hasSelection != want.hasSelection || (want.hasSelection && (got.from != want.from || got.to != want.to)) {
		t.Fatalf("%s changed the editing selection: %+v %+v %v, want %+v %+v %v",
			what, got.from, got.to, got.hasSelection, want.from, want.to, want.hasSelection)
	}
	if scroll && got.scrollY != want.scrollY {
		t.Fatalf("%s changed ScrollY: %d, want %d", what, got.scrollY, want.scrollY)
	}
	if got.ansi != want.ansi {
		t.Fatalf("%s changed the document: %q, want %q", what, got.ansi, want.ansi)
	}
	if got.modified != want.modified {
		t.Fatalf("%s changed IsModified: %v, want %v", what, got.modified, want.modified)
	}
}

func type287(v *editView, text string) {
	for _, r := range text {
		v.ConsumeKey(loom.KeyEvent{Text: string(r)})
	}
}

func click287(v *editView, x, y int) loom.EventResult {
	res := v.ConsumeMouse(loom.MouseEvent{X: x, Y: y, Action: loom.MousePress, Button: loom.MouseLeft})
	v.ConsumeMouse(loom.MouseEvent{X: x, Y: y, Action: loom.MouseRelease, Button: loom.MouseLeft})
	return res
}

// locate287 returns the terminal cell of the first occurrence of label.
func locate287(t *testing.T, c *loom.Canvas, label string) (x, y int) {
	t.Helper()
	for row := 0; row < c.Rows(); row++ {
		// canvasPlainRow collapses blanks, so keep each glyph's cell column.
		var text strings.Builder
		var columns []int
		for col := 0; col < c.Cols(); col++ {
			cell := c.Get(col, row)
			if cell.Continuation {
				continue
			}
			glyph := cell.Text
			if glyph == "" {
				glyph = " "
			}
			for range []byte(glyph) { // strings.Index reports byte offsets
				columns = append(columns, col)
			}
			text.WriteString(glyph)
		}
		if i := strings.Index(text.String(), label); i >= 0 {
			return columns[i], row
		}
	}
	t.Fatalf("missing %q on screen:\n%s", label, canvasScreenText(c))
	return 0, 0
}

// matches287 is an independent oracle for the matches of query in the current
// document, case-insensitive like the editor's default.
func matches287(v *editView, query string) []searchMatch {
	var out []searchMatch
	q := []rune(strings.ToLower(query))
	for lineIdx, line := range v.edit.Document.Lines {
		text := []rune(strings.ToLower(richLinePlainText(line)))
		for i := 0; i+len(q) <= len(text); i++ {
			if string(text[i:i+len(q)]) == string(q) {
				out = append(out, searchMatch{line: lineIdx, start: i, length: len(q)})
			}
		}
	}
	return out
}

func requireMatches287(t *testing.T, v *editView, query string) []searchMatch {
	t.Helper()
	want := matches287(v, query)
	if len(v.searchMatches) != len(want) {
		t.Fatalf("searchMatches = %+v, want %+v", v.searchMatches, want)
	}
	for i := range want {
		if v.searchMatches[i] != want[i] {
			t.Fatalf("searchMatches = %+v, want %+v", v.searchMatches, want)
		}
	}
	return want
}

func requireHighlight287(t *testing.T, v *editView, m searchMatch) {
	t.Helper()
	want := loom.RichTextHighlight{
		From:  loom.RichPosition{Line: m.line, Offset: m.start},
		To:    loom.RichPosition{Line: m.line, Offset: m.start + m.length},
		Style: v.edit.SelectionStyle,
	}
	if len(v.edit.Highlights) != 1 || v.edit.Highlights[0].From != want.From || v.edit.Highlights[0].To != want.To {
		t.Fatalf("Highlights = %+v, want one highlight %+v..%+v", v.edit.Highlights, want.From, want.To)
	}
}

// editWhileClosed287 types a query, closes search, appends " alpha" to the
// first line and leaves the document with three matches and search closed.
func editWhileClosed287(t *testing.T) *editView {
	t.Helper()
	view, _ := new287View(t, "alpha beta alpha\n")
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	type287(view, "alpha")
	if len(view.searchMatches) != 2 {
		t.Fatalf("setup: %d matches, want 2", len(view.searchMatches))
	}
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	if view.showSearch || view.focused != focusEditor {
		t.Fatalf("setup: search not closed (show=%v focus=%v)", view.showSearch, view.focused)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "end"})
	type287(view, " alpha")
	if got := len(matches287(view, "alpha")); got != 3 {
		t.Fatalf("setup: document has %d matches, want 3", got)
	}
	return view
}

// Expected red before M2: reopening shows the match list of the old text.
func Test287ReopenRefreshesStaleMatches(t *testing.T) {
	view := editWhileClosed287(t)
	before := snap287(view)

	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	c := draw287(view, 100, 24)

	want := requireMatches287(t, view, "alpha")
	requireHighlight287(t, view, want[0])
	if text := canvasScreenText(c); !strings.Contains(text, "1 / 3 matches") {
		t.Fatalf("counter after reopen is stale:\n%s", text)
	}
	// The cursor is covered by Test287SearchHighlightKeepsEditingCursor.
	requireSame287(t, "reopening search", before, view, false, true)
}

// Expected red before M2: next/previous walk offsets of the old text.
func Test287SearchNextOffsetsFollowEditsWhileClosed(t *testing.T) {
	view := editWhileClosed287(t)
	before := snap287(view)

	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	want := requireMatches287(t, view, "alpha")
	for step := 1; step <= len(want); step++ {
		view.ConsumeKey(loom.KeyEvent{Key: "enter"})
		m := want[step%len(want)]
		if view.searchIndex != step%len(want) {
			t.Fatalf("after %d x next, searchIndex = %d, want %d", step, view.searchIndex, step%len(want))
		}
		requireHighlight287(t, view, m)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "shift-enter"})
	requireHighlight287(t, view, want[len(want)-1])
	// The cursor is covered by Test287SearchHighlightKeepsEditingCursor.
	requireSame287(t, "navigating matches", before, view, false, true)
}

// Expected red before M2: a retained query repaints highlights on a document
// that was opened while search is closed.
func Test287ClosedSearchDoesNotLeakHighlights(t *testing.T) {
	view, _ := new287View(t, "alpha beta alpha\n")
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	type287(view, "alpha")
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	if view.showSearch || len(view.edit.Highlights) != 0 {
		t.Fatalf("setup: show=%v highlights=%+v", view.showSearch, view.edit.Highlights)
	}

	other := filepath.Join(t.TempDir(), "other.txt")
	if err := os.WriteFile(other, []byte("alpha alpha alpha\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := view.openFile(other); err != nil {
		t.Fatal(err)
	}
	if len(view.edit.Highlights) != 0 {
		t.Fatalf("openFile painted highlights while search is closed: %+v", view.edit.Highlights)
	}
	if view.edit.IsModified() || view.edit.HasSelection {
		t.Fatal("openFile left the new document modified or selected")
	}

	view.closeDocument()
	if len(view.edit.Highlights) != 0 {
		t.Fatalf("closeDocument painted highlights while search is closed: %+v", view.edit.Highlights)
	}

	// Reopening search brings the highlight for the current text back.
	view.edit.Document.Lines = []loom.RichLine{{Spans: []loom.RichSpan{{Text: "alpha"}}}}
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	requireMatches287(t, view, "alpha")
	if len(view.edit.Highlights) != 1 {
		t.Fatalf("reopen did not paint the current match: %+v", view.edit.Highlights)
	}
}

// Expected red before M2: below the draw guard the stale rect swallows clicks
// and the stale mode bar from an earlier, larger draw is still hit-tested.
func Test287InvisibleSearchPanelDoesNotSwallowClicks(t *testing.T) {
	tested := 0
	for _, size := range [][2]int{{4, 7}, {5, 8}, {6, 9}, {8, 8}} {
		view, _ := new287View(t, "alpha beta alpha\nsecond line\nthird line\n")
		view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
		draw287(view, 100, 24) // builds the mode bar and a large search rect
		view.ConsumeKey(loom.KeyEvent{Key: "tab"})

		small := draw287(view, size[0], size[1])
		editor := view.editorRect
		text := canvasScreenText(small)
		if editor.W <= 0 || editor.H < 3 || strings.Contains(text, "Sear") || strings.Contains(text, "Find") || strings.Contains(text, "Regex") {
			continue // the panel is drawn (even truncated) or the editor is too small to probe
		}
		tested++
		for _, row := range []int{0, 2} {
			if row == 2 && !strings.Contains(canvasPlainRow(small, editor.Y+row), "thi") {
				continue // this size does not paint the third document line
			}
			view.edit.Cursor = loom.RichPosition{}
			view.focused = focusEditor
			before := snap287(view)
			res := click287(view, editor.X+1, editor.Y+row)
			if !res.Consumed {
				t.Fatalf("%v row %d: click not consumed", size, row)
			}
			if view.regexMode {
				t.Fatalf("%v row %d: click toggled regex mode through an invisible mode bar", size, row)
			}
			if view.focused != focusEditor {
				t.Fatalf("%v row %d: click in the editor focused %v", size, row, view.focused)
			}
			if view.edit.Cursor.Line != row || view.edit.Cursor.Offset == 0 {
				t.Fatalf("%v row %d: click did not reach the editor: cursor %+v", size, row, view.edit.Cursor)
			}
			requireSame287(t, "click below the search guard", before, view, false, true)
		}
	}
	if tested == 0 {
		t.Fatal("no probed size hides the search panel while keeping an editor")
	}
}

// Expected red before M2: mouse edits (popover Frame) never refresh matches.
func Test287MouseEditRefreshesMatches(t *testing.T) {
	view, _ := new287View(t, "alpha beta alpha\n")
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	type287(view, "alpha")
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	view.edit.Cursor = loom.RichPosition{Offset: 2} // inside the first word
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-space"})
	c := draw287(view, 100, 24)
	if !view.edit.ModalOpen() || !view.edit.HasSelection {
		t.Fatal("setup: the formatting popover did not open on a selected word")
	}
	ansi := view.edit.Document.ToANSI()

	x, y := locate287(t, c, "Frame")
	click287(view, x, y)
	c = draw287(view, 100, 24)
	x, y = locate287(t, c, loom.SpeccedDefaults.RichTextEdit.BoxStyleLabels[0])
	click287(view, x, y)
	draw287(view, 100, 24)

	if view.edit.Document.ToANSI() == ansi {
		t.Fatal("setup: the mouse did not change the document")
	}
	if !view.edit.IsModified() {
		t.Fatal("mouse edit did not mark the document modified")
	}
	want := requireMatches287(t, view, "alpha")
	if len(want) != 2 {
		t.Fatalf("oracle found %d matches, want the 2 words to survive the frame", len(want))
	}
	requireHighlight287(t, view, want[0])
}

// Regression guard: the reported repro. Click the query field, type, save.
func Test287SearchQueryClickTypeLeavesDocumentUnchanged(t *testing.T) {
	content := "first word\nsecond line\n"
	view, path := new287View(t, content)
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	c := draw287(view, 100, 24)
	if view.focused != focusEditor {
		t.Fatalf("setup: focus = %v, want the editor", view.focused)
	}
	before := snap287(view)

	if res := click287(view, view.searchRect.X+5, view.searchRect.Y+1); !res.Consumed {
		t.Fatalf("query click result = %+v, want consumed", res)
	}
	if view.focused != focusSearch {
		t.Fatalf("focus = %v, want the search bar", view.focused)
	}
	type287(view, "X")
	if view.searchQuery != "X" {
		t.Fatalf("searchQuery = %q, want X", view.searchQuery)
	}
	c = draw287(view, 100, 24)
	if row := canvasPlainRow(c, view.searchRect.Y+1); !strings.Contains(row, "X") {
		t.Fatalf("typed query not shown in the query row: %q", row)
	}
	view.ConsumeKey(loom.KeyEvent{Key: loom.SpeccedDefaults.RichTextEdit.HotkeySaveBinding})
	requireSame287(t, "typing into the clicked query field", before, view, true, true)

	var expected loom.RichDocument
	expected.FromANSI(content)
	if data, err := os.ReadFile(path); err != nil || string(data) != expected.ToANSI()+"\n" {
		t.Fatalf("saved document changed: %q, %v", data, err)
	}
}

// Regression guard: Normal/Regex clicks run their action and stop there.
func Test287SearchModeClicksSwitchModeWithoutFallThrough(t *testing.T) {
	view, _ := new287View(t, "alpha beta alpha\n")
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	view.edit.ShowPopover = false // keep a real selection without a modal popover
	view.edit.Cursor = loom.RichPosition{Offset: 10}
	view.edit.SetSelection(loom.RichPosition{Offset: 6}, loom.RichPosition{Offset: 10})
	c := draw287(view, 100, 24)
	if view.edit.ModalOpen() {
		t.Fatal("setup: the popover is open")
	}
	before := snap287(view)

	regexX, row := locate287(t, c, loom.SpeccedDefaults.Editor.RegexLabel)
	normalX, normalRow := locate287(t, c, loom.SpeccedDefaults.Editor.NormalLabel)
	if row != normalRow || row != view.searchRect.Y+2 {
		t.Fatalf("mode bar at rows %d/%d, want %d", row, normalRow, view.searchRect.Y+2)
	}
	if res := click287(view, regexX, row); !res.Consumed || !view.regexMode {
		t.Fatalf("Regex click: result %+v regexMode=%v", res, view.regexMode)
	}
	if view.focused != focusSearch {
		t.Fatalf("focus after mode click = %v, want the search bar", view.focused)
	}
	c = draw287(view, 100, 24)
	if !strings.Contains(canvasPlainRow(c, row), loom.SpeccedDefaults.Editor.RegexLabel+loom.SpeccedDefaults.Editor.ModeMarker) {
		t.Fatalf("mode marker did not move to Regex: %q", canvasPlainRow(c, row))
	}
	requireSame287(t, "Regex click", before, view, true, true)

	if res := click287(view, normalX, row); !res.Consumed || view.regexMode {
		t.Fatalf("Normal click: result %+v regexMode=%v", res, view.regexMode)
	}
	requireSame287(t, "Normal click", before, view, true, true)
}

// Regression guard: no mouse action inside the panel reaches the editor.
func Test287SearchPanelClickNeverMovesEditorState(t *testing.T) {
	view, _ := new287View(t, "alpha beta alpha\n")
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	view.edit.ShowPopover = false // keep a real selection without a modal popover
	view.edit.Cursor = loom.RichPosition{Offset: 10}
	view.edit.SetSelection(loom.RichPosition{Offset: 6}, loom.RichPosition{Offset: 10})
	draw287(view, 100, 24)
	if view.edit.ModalOpen() {
		t.Fatal("setup: the popover is open")
	}
	before := snap287(view)
	panel := view.searchRect
	if panel.W == 0 {
		t.Fatal("setup: search panel not drawn")
	}

	actions := []loom.MouseAction{loom.MousePress, loom.MouseDrag, loom.MouseRelease, loom.MouseHover, loom.MouseScrollUp, loom.MouseScrollDown}
	for y := panel.Y; y < panel.Y+panel.H; y++ {
		for x := panel.X; x < panel.X+panel.W; x++ {
			for _, action := range actions {
				res := view.ConsumeMouse(loom.MouseEvent{X: x, Y: y, Action: action, Button: loom.MouseLeft})
				if !res.Consumed {
					t.Fatalf("action %v at (%d,%d) fell through: %+v", action, x, y, res)
				}
				requireSame287(t, "panel mouse event", before, view, true, false)
			}
		}
	}
	draw287(view, 100, 24)
	requireSame287(t, "panel mouse events", before, view, true, true)
}

// Regression guard: only a press changes focus, never hover or wheel.
func Test287HoverAndScrollDoNotChangeFocus(t *testing.T) {
	view, _ := new287View(t, "alpha beta alpha\n")
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	draw287(view, 100, 24)
	panel, editor := view.searchRect, view.editorRect
	before := snap287(view)

	moves := []loom.MouseAction{loom.MouseHover, loom.MouseScrollUp, loom.MouseScrollDown}
	for _, action := range moves {
		view.ConsumeMouse(loom.MouseEvent{X: panel.X + 5, Y: panel.Y + 1, Action: action})
		view.ConsumeMouse(loom.MouseEvent{X: panel.X + 5, Y: panel.Y + 2, Action: action})
		draw287(view, 100, 24)
		if view.focused != focusEditor {
			t.Fatalf("%v over the search panel moved focus to %v", action, view.focused)
		}
	}
	requireSame287(t, "hover and wheel over the panel", before, view, true, true)

	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if view.focused != focusSearch {
		t.Fatalf("setup: focus = %v, want the search bar", view.focused)
	}
	for _, action := range moves {
		view.ConsumeMouse(loom.MouseEvent{X: editor.X + 2, Y: editor.Y + 2, Action: action})
		draw287(view, 100, 24)
		if view.focused != focusSearch {
			t.Fatalf("%v over the editor moved focus to %v", action, view.focused)
		}
	}
	requireSame287(t, "hover and wheel over the editor", before, view, true, true)

	// Control: a press is routed and does move focus and the cursor.
	click287(view, editor.X+2, editor.Y)
	if view.focused != focusEditor || view.edit.Cursor.Offset != 2 {
		t.Fatalf("control press: focus=%v cursor=%+v", view.focused, view.edit.Cursor)
	}
}

// Regression guard: the formatting popover owns the backdrop. A click on the
// search panel only dismisses it; the next click reaches the panel.
func Test287PopoverOpenSearchClickIsBackdrop(t *testing.T) {
	view, _ := new287View(t, "alpha beta alpha\n")
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-f"})
	type287(view, "alpha")
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	view.edit.Cursor = loom.RichPosition{Offset: 2} // inside the first word
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-space"})
	draw287(view, 100, 24)
	if !view.edit.ModalOpen() || !view.edit.HasSelection {
		t.Fatal("setup: the formatting popover did not open on a selected word")
	}
	before := snap287(view)

	click287(view, view.searchRect.X+5, view.searchRect.Y+1)
	draw287(view, 100, 24)
	if view.edit.ModalOpen() {
		t.Fatal("backdrop click did not dismiss the popover")
	}
	if view.focused != focusEditor || view.searchQuery != "alpha" || !view.showSearch {
		t.Fatalf("backdrop click changed host state: focus=%v query=%q show=%v", view.focused, view.searchQuery, view.showSearch)
	}
	// The popover closing clears the selection by contract; the rest is fixed.
	got := snap287(view)
	if got.cursor != before.cursor || got.ansi != before.ansi || got.modified != before.modified {
		t.Fatalf("backdrop click changed cursor or document: %+v, want %+v", got, before)
	}

	click287(view, view.searchRect.X+5, view.searchRect.Y+1)
	if view.focused != focusSearch {
		t.Fatalf("second click: focus = %v, want the search bar", view.focused)
	}
	type287(view, "!")
	if view.searchQuery != "alpha!" || view.edit.Document.ToANSI() != before.ansi {
		t.Fatalf("typing after the backdrop click went to the document: query=%q", view.searchQuery)
	}
}
