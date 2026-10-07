package loom

import "testing"

type nestedRichTextEditWrapper struct{ edit *RichTextEdit }

func (w *nestedRichTextEditWrapper) Draw(c *Canvas, r Rect) { w.edit.Draw(c, r) }
func (w *nestedRichTextEditWrapper) ConsumeKey(key KeyEvent) EventResult {
	return w.edit.ConsumeKey(key)
}
func (*nestedRichTextEditWrapper) ConsumeMouse(MouseEvent) EventResult { return Ignored() }

func newSubmenuTestEditor() (*RichTextEdit, *Canvas) {
	edit := richTestEditor(RichSpan{Text: "abc"})
	edit.SetSelection(RichPosition{Offset: 0}, RichPosition{Offset: 3})
	canvas := NewCanvas(60, 8)
	edit.Draw(canvas, canvas.Bounds())
	return edit, canvas
}

func popoverFocusAction(t *testing.T, edit *RichTextEdit, action string) {
	t.Helper()
	if got := edit.ConsumeKey(KeyEvent{Key: "tab"}); !got.Consumed {
		t.Fatalf("initial popover Tab = %+v", got)
	}
	for range len(richPopoverActions) {
		if edit.popoverFocus >= 0 && edit.popoverFocus < len(richPopoverActions) && richPopoverActions[edit.popoverFocus] == action {
			return
		}
		if got := edit.ConsumeKey(KeyEvent{Key: "tab"}); !got.Consumed {
			t.Fatalf("Tab toward %s = %+v", action, got)
		}
	}
	t.Fatalf("could not focus popover action %q, focus=%d", action, edit.popoverFocus)
}

func assertMenuHandled(t *testing.T, got EventResult) {
	t.Helper()
	if !got.Consumed || got.Done || got.Quit {
		t.Fatalf("popover key result = %+v, want consumed without Done/Quit", got)
	}
}

func TestRichTextEditKeyboardColorSubmenusWrapApplyAndUndo(t *testing.T) {
	for _, palette := range []string{"#FG", "#BG"} {
		t.Run(palette, func(t *testing.T) {
			edit, canvas := newSubmenuTestEditor()
			popoverFocusAction(t, edit, palette)
			assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
			if edit.popoverPalette != palette || !edit.popoverSubmenuFocusSet || edit.popoverSubmenuFocus != 0 {
				t.Fatalf("opened %s submenu state = palette:%q focus:%d set:%v", palette, edit.popoverPalette, edit.popoverSubmenuFocus, edit.popoverSubmenuFocusSet)
			}
			for _, key := range []string{"right", "left", "left", "down", "tab"} {
				assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: key}))
			}
			if edit.popoverSubmenuFocus != 1 {
				t.Fatalf("submenu wrap sequence focus = %d, want 1", edit.popoverSubmenuFocus)
			}
			edit.Draw(canvas, canvas.Bounds())
			if len(edit.popoverSwatches) != 16 || !canvas.Get(edit.popoverSwatches[1].rect.X, edit.popoverSwatches[1].rect.Y).Style.Invert {
				t.Fatal("focused palette choice was not visibly highlighted")
			}
			for i := 0; i < 5; i++ {
				assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "tab"}))
			}
			assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "enter"}))
			if edit.popoverPalette != "" || edit.popoverSubmenu != "" || edit.popoverFocusSet {
				t.Fatalf("palette apply left menu open: palette=%q submenu=%q focusSet=%v", edit.popoverPalette, edit.popoverSubmenu, edit.popoverFocusSet)
			}
			want := ColorIndex(6)
			if got := edit.Document.Lines[0].Spans[0].Style; palette == "#FG" && got.FG != want || palette == "#BG" && got.BG != want {
				t.Fatalf("applied %s style = %+v, want color %v", palette, got, want)
			}
			edit.undo()
			if got := edit.Document.Lines[0].Spans[0].Style; got != (Style{}) {
				t.Fatalf("undo %s color = %+v, want original style", palette, got)
			}
		})
	}
}

func TestRichTextEditKeyboardBoxSubmenuWrapAndApply(t *testing.T) {
	edit, canvas := newSubmenuTestEditor()
	popoverFocusAction(t, edit, "Box")
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
	if edit.popoverSubmenu != "Box" {
		t.Fatalf("Box submenu = %q", edit.popoverSubmenu)
	}
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "left"}))
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "left"}))
	if edit.popoverSubmenuFocus != 1 {
		t.Fatalf("reverse arrow wrap focus = %d, want 1", edit.popoverSubmenuFocus)
	}
	edit.Draw(canvas, canvas.Bounds())
	if len(edit.popoverBoxChoices) != 3 {
		t.Fatalf("box choices = %d, want three", len(edit.popoverBoxChoices))
	}
	choice := edit.popoverBoxChoices[1]
	if got := canvas.Get(choice.rect.X+1, choice.rect.Y).Style.BG; got != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.PopoverFocusBG)) {
		t.Fatalf("focused box choice background = %v", got)
	}
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "enter"}))
	if got, want := richTestText(edit), "╭───╮\n│abc│\n╰───╯"; got != want {
		t.Fatalf("rounded submenu result = %q, want %q", got, want)
	}
	if edit.popoverSubmenu != "" || edit.popoverFocusSet {
		t.Fatalf("box apply left keyboard menu active: submenu=%q focusSet=%v", edit.popoverSubmenu, edit.popoverFocusSet)
	}
	edit.undo()
	if got := richTestText(edit); got != "abc" {
		t.Fatalf("undo box style = %q, want abc", got)
	}
}

func TestRichTextEditSubmenuEscapeReturnsToBarThenText(t *testing.T) {
	edit, _ := newSubmenuTestEditor()
	popoverFocusAction(t, edit, "Box")
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "esc"}))
	if edit.popoverSubmenu != "" || !edit.popoverFocusSet || edit.popoverFocus != 7 || !edit.HasSelection {
		t.Fatalf("first Esc did not restore Box bar focus: submenu=%q focusSet=%v focus=%d selection=%v", edit.popoverSubmenu, edit.popoverFocusSet, edit.popoverFocus, edit.HasSelection)
	}
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "esc"}))
	if edit.popoverFocusSet || edit.popoverAtCursor || !edit.popoverSuppressed || richTestText(edit) != "abc" {
		t.Fatalf("second Esc did not dismiss without editing: focusSet=%v atCursor=%v suppressed=%v text=%q", edit.popoverFocusSet, edit.popoverAtCursor, edit.popoverSuppressed, richTestText(edit))
	}
}

func TestRichTextEditSubmenuKeysStayConsumedByFrameAndSplit(t *testing.T) {
	frameEdit, _ := newSubmenuTestEditor()
	frame := &Frame{Boxes: []Box{{ID: "editor", Child: frameEdit}}}
	frame.Draw(NewCanvas(60, 8), Rect{W: 60, H: 8})
	popoverFocusAction(t, frameEdit, "#FG")
	assertMenuHandled(t, frame.ConsumeKey(KeyEvent{Key: "space"}))
	focus := frame.focused
	assertMenuHandled(t, frame.ConsumeKey(KeyEvent{Key: "tab"}))
	if frame.focused != focus {
		t.Fatalf("Frame traversed focus while submenu was open: %d -> %d", focus, frame.focused)
	}

	splitEdit, _ := newSubmenuTestEditor()
	split := NewSplit(splitEdit, richEditorLines("other"))
	split.Orientation = Vertical
	split.Ratio = 0.7
	split.Draw(NewCanvas(60, 8), Rect{W: 60, H: 8})
	popoverFocusAction(t, splitEdit, "#BG")
	assertMenuHandled(t, split.ConsumeKey(KeyEvent{Key: "space"}))
	focus = split.focus
	assertMenuHandled(t, split.ConsumeKey(KeyEvent{Key: "tab"}))
	if split.focus != focus {
		t.Fatalf("Split traversed focus while submenu was open: %d -> %d", focus, split.focus)
	}
}

func TestRichTextEditPopoverSpaceAndArrowsStayInsideNestedGalleryWrapper(t *testing.T) {
	edit := richTestEditor(RichSpan{Text: "abc"})
	edit.SetSelection(RichPosition{Offset: 0}, RichPosition{Offset: 3})
	edit.ShowPopover = true
	wrapper := &nestedRichTextEditWrapper{edit: edit}
	other := richEditorLines("other")
	tabs := NewTabs(Tab{Title: "RichTextEdit", Widget: wrapper}, Tab{Title: "Other", Widget: other})
	canvas := NewCanvas(80, 10)
	tabs.Draw(canvas, canvas.Bounds())

	popoverFocusAction(t, edit, "#FG")
	cursor, selectionFrom, selectionTo, selection := edit.Cursor, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection
	focus := tabs.Focus()

	// DecodeKey represents a physical Space key as printable text, not Key:"space".
	space := DecodeKey([]byte{' '})
	if space.Key != "" || space.Text != " " {
		t.Fatalf("decoded Space = %+v, want printable space text", space)
	}
	assertMenuHandled(t, tabs.ConsumeKey(space))
	if edit.popoverPalette != "#FG" {
		t.Fatalf("Space did not open focused FG submenu: %q", edit.popoverPalette)
	}
	for _, key := range []string{"up", "down", "left", "right"} {
		assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: key}))
	}
	if tabs.Focus() != focus || edit.Cursor != cursor || edit.SelectionFrom != selectionFrom || edit.SelectionTo != selectionTo || edit.HasSelection != selection {
		t.Fatalf("submenu arrows escaped wrapper: tabFocus=%d cursor=%+v selection=%+v..%+v active=%v", tabs.Focus(), edit.Cursor, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "esc"}))
	if edit.popoverPalette != "" || !edit.popoverFocusSet {
		t.Fatal("Escape did not return from submenu to the focused bar")
	}
	// Horizontal arrows move the visible bar focus, while vertical arrows are
	// consumed without moving the document or Tabs focus.
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "right"}))
	barFocus := edit.popoverFocus
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "left"}))
	if edit.popoverFocus == barFocus {
		t.Fatal("left arrow did not move focus along the bar")
	}
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "up"}))
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "down"}))
	if tabs.Focus() != focus || edit.Cursor != cursor || edit.SelectionFrom != selectionFrom || edit.SelectionTo != selectionTo || edit.HasSelection != selection {
		t.Fatalf("bar arrows escaped wrapper: tabFocus=%d cursor=%+v selection=%+v..%+v active=%v", tabs.Focus(), edit.Cursor, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
	popoverFocusAction(t, edit, "#FG")
	assertMenuHandled(t, tabs.ConsumeKey(space))
	assertMenuHandled(t, tabs.ConsumeKey(space))
	if edit.popoverPalette != "" || edit.HasSelection != selection || edit.Cursor != cursor || edit.SelectionFrom != selectionFrom || edit.SelectionTo != selectionTo {
		t.Fatalf("submenu Space did not apply in place: palette=%q cursor=%+v selection=%+v..%+v active=%v", edit.popoverPalette, edit.Cursor, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
}

func TestRichTextEditKeyboardSubmenuDismissesWhenItCannotFit(t *testing.T) {
	edit := richTestEditor(RichSpan{Text: "abc"})
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 3})
	edit.lastRect = Rect{W: 18, H: 4}
	edit.popoverFocusSet = true
	edit.popoverFocus = 5
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
	if edit.popoverPalette != "" || edit.popoverFocusSet || !edit.popoverSuppressed {
		t.Fatalf("unrenderable palette retained focus: palette=%q focusSet=%v suppressed=%v", edit.popoverPalette, edit.popoverFocusSet, edit.popoverSuppressed)
	}
}

func TestRichTextEditDrawIsNestedAndKeyboardAccessible(t *testing.T) {
	for _, selection := range []bool{false, true} {
		edit := richTestEditor(RichSpan{Text: "abc "})
		if selection {
			edit.SetSelection(RichPosition{}, RichPosition{Offset: 3})
		} else {
			edit.Cursor.Offset = 3
			edit.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		}
		canvas := NewCanvas(60, 8)
		edit.Draw(canvas, canvas.Bounds())
		for _, button := range edit.popoverButtons {
			if button.action == "Draw" {
				t.Fatal("top-level Draw")
			}
		}
		popoverFocusAction(t, edit, "Box")
		assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "enter"}))
		if selection {
			edit.ConsumeKey(KeyEvent{Key: "left"})
		}
		if edit.popoverSubmenuFocus != 2 {
			t.Fatalf("Draw focus=%d", edit.popoverSubmenuFocus)
		}
		assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "enter"}))
		if !edit.BoxMode || edit.HasSelection {
			t.Fatal("nested Draw failed")
		}
	}
}
