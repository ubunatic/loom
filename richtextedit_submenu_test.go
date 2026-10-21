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
	_, _, _, pointer, open := edit.popoverGeometry(edit.lastRect, edit.documentLines())
	if !open {
		t.Fatal("popover unavailable")
	}
	key := "down"
	if pointer == SpeccedDefaults.RichTextEdit.PointerDownGlyph {
		key = "up"
	}
	if got := edit.ConsumeKey(KeyEvent{Key: key}); !got.Consumed || !edit.popoverFocusSet {
		t.Fatalf("initial popover focus = %+v", got)
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

func TestRichTextEditPopoverSubmenusDoNotCoverSelection(t *testing.T) {
	for _, tc := range []struct {
		name     string
		from, to int
	}{
		{name: "toolbar above selection", from: 5, to: 5},
		{name: "toolbar below selection", from: 0, to: 0},
		{name: "multiline toolbar below selection", from: 0, to: 1},
	} {
		for _, submenu := range []string{"ColorPicker", "Box"} {
			t.Run(tc.name+"/"+submenu, func(t *testing.T) {
				lines := make([]RichLine, 7)
				for i := range lines {
					lines[i] = RichLine{Spans: []RichSpan{{Text: "row"}}}
				}
				for line := tc.from; line <= tc.to; line++ {
					lines[line] = RichLine{Spans: []RichSpan{{Text: "selected"}}}
				}
				edit := NewRichTextEdit(&RichDocument{Lines: lines})
				edit.SetSelection(RichPosition{Line: tc.from}, RichPosition{Line: tc.to, Offset: 8})
				canvas := NewCanvas(60, 8)
				edit.Draw(canvas, canvas.Bounds())
				popoverFocusAction(t, edit, submenu)
				assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
				edit.Draw(canvas, canvas.Bounds())
				for line := tc.from; line <= tc.to; line++ {
					for x, ch := range "selected" {
						if got := canvas.Get(x, line).Text; got != string(ch) {
							t.Fatalf("selected cell (%d,%d) = %q, want %q", x, line, got, string(ch))
						}
					}
				}
			})
		}
	}
}

func TestRichTextEditColorPickerAppliesBothChannels(t *testing.T) {
	edit, canvas := newSubmenuTestEditor()
	popoverFocusAction(t, edit, "ColorPicker")
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
	picker := edit.popoverColorPicker
	if picker == nil {
		t.Fatal("Space did not open embedded ColorPicker")
	}
	before := edit.Document.Lines[0].Spans[0].Style
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "tab"}))
	if picker.Target != ColorPickerBG || edit.Document.Lines[0].Spans[0].Style != before {
		t.Fatalf("Tab changed mode/style unexpectedly: target=%v style=%+v", picker.Target, edit.Document.Lines[0].Spans[0].Style)
	}
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "right"}))
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
	if got := edit.Document.Lines[0].Spans[0].Style.BG; got != picker.Palette[1] {
		t.Fatalf("selected BG = %v, want %v", got, picker.Palette[1])
	}
	edit.Draw(canvas, canvas.Bounds())
	before = edit.Document.Lines[0].Spans[0].Style
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "up"}))
	if picker.Target != ColorPickerFG || edit.Document.Lines[0].Spans[0].Style != before {
		t.Fatalf("Up changed mode/style unexpectedly: target=%v style=%+v", picker.Target, edit.Document.Lines[0].Spans[0].Style)
	}
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
	if got := edit.Document.Lines[0].Spans[0].Style.FG; got != picker.Palette[1] {
		t.Fatalf("selected FG = %v, want %v", got, picker.Palette[1])
	}
	for _, key := range []KeyEvent{{Text: "f"}, {Text: "x"}, {Key: "shift-space"}} {
		if got := edit.ConsumeKey(key); got.Consumed {
			t.Errorf("removed picker key %+v was consumed", key)
		}
	}
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "esc"}))
	if edit.popoverColorPicker != nil {
		t.Fatal("Escape did not close ColorPicker")
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
	if edit.popoverSubmenu != "" || !edit.popoverFocusSet || edit.popoverFocus != len(richPopoverActions)-1 || !edit.HasSelection {
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
	popoverFocusAction(t, frameEdit, "ColorPicker")
	assertMenuHandled(t, frame.ConsumeKey(KeyEvent{Key: "space"}))
	focus := frame.focused
	assertMenuHandled(t, frame.ConsumeKey(KeyEvent{Key: "tab"})) // The picker owns Tab while open.
	if frame.focused != focus {
		t.Fatalf("Frame traversed focus while submenu was open: %d -> %d", focus, frame.focused)
	}

	splitEdit, _ := newSubmenuTestEditor()
	split := NewSplit(splitEdit, richEditorLines("other"))
	split.Orientation = Vertical
	split.Ratio = 0.7
	split.Draw(NewCanvas(60, 8), Rect{W: 60, H: 8})
	popoverFocusAction(t, splitEdit, "ColorPicker")
	assertMenuHandled(t, split.ConsumeKey(KeyEvent{Key: "space"}))
	focus = split.focus
	assertMenuHandled(t, split.ConsumeKey(KeyEvent{Key: "tab"}))
	if split.focus != focus || splitEdit.popoverColorPicker.Target != ColorPickerBG {
		t.Fatalf("active picker Tab escaped to Split: focus=%d target=%v", split.focus, splitEdit.popoverColorPicker.Target)
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

	popoverFocusAction(t, edit, "ColorPicker")
	cursor, selectionFrom, selectionTo, selection := edit.Cursor, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection
	focus := tabs.Focus()

	// DecodeKey represents a physical Space key as printable text, not Key:"space".
	space := DecodeKey([]byte{' '})
	if space.Key != "" || space.Text != " " {
		t.Fatalf("decoded Space = %+v, want printable space text", space)
	}
	assertMenuHandled(t, tabs.ConsumeKey(space))
	if edit.popoverColorPicker == nil {
		t.Fatal("Space did not open embedded ColorPicker")
	}
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "tab"}))
	if edit.popoverColorPicker.Target != ColorPickerBG || tabs.Focus() != focus {
		t.Fatalf("active picker Tab escaped to Tabs: target=%v focus=%d", edit.popoverColorPicker.Target, tabs.Focus())
	}
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "tab"}))
	if edit.popoverColorPicker.Target != ColorPickerFG || tabs.Focus() != focus {
		t.Fatalf("second picker Tab escaped to Tabs or failed to toggle back: target=%v focus=%d", edit.popoverColorPicker.Target, tabs.Focus())
	}
	for _, key := range []string{"left", "right"} {
		assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: key}))
	}
	if tabs.Focus() != focus || edit.Cursor != cursor || edit.SelectionFrom != selectionFrom || edit.SelectionTo != selectionTo || edit.HasSelection != selection {
		t.Fatalf("submenu arrows escaped wrapper: tabFocus=%d cursor=%+v selection=%+v..%+v active=%v", tabs.Focus(), edit.Cursor, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "esc"}))
	if edit.popoverColorPicker != nil {
		t.Fatal("Escape did not close ColorPicker")
	}
	// Horizontal arrows move bar focus. The below-anchor toolbar consumes Up
	// as dismissal without moving the document or Tabs focus.
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "right"}))
	barFocus := edit.popoverFocus
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "left"}))
	if edit.popoverFocus == barFocus {
		t.Fatal("left arrow did not move focus along the bar")
	}
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "up"}))
	if !edit.popoverSuppressed || edit.popoverFocusSet {
		t.Fatal("arrow away did not dismiss the focused toolbar")
	}
	if tabs.Focus() != focus || edit.Cursor != cursor || edit.SelectionFrom != selectionFrom || edit.SelectionTo != selectionTo || edit.HasSelection != selection {
		t.Fatalf("bar arrows escaped wrapper: tabFocus=%d cursor=%+v selection=%+v..%+v active=%v", tabs.Focus(), edit.Cursor, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "ctrl-space"}))
	popoverFocusAction(t, edit, "ColorPicker")
	assertMenuHandled(t, tabs.ConsumeKey(space))
	assertMenuHandled(t, tabs.ConsumeKey(KeyEvent{Key: "esc"}))
	if edit.popoverColorPicker != nil || edit.HasSelection != selection || edit.Cursor != cursor || edit.SelectionFrom != selectionFrom || edit.SelectionTo != selectionTo {
		t.Fatalf("picker session changed editor selection: cursor=%+v selection=%+v..%+v active=%v", edit.Cursor, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
}

func TestRichTextEditKeyboardSubmenuDismissesWhenItCannotFit(t *testing.T) {
	edit := richTestEditor(RichSpan{Text: "abc"})
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 3})
	edit.lastRect = Rect{W: 18, H: 4}
	edit.popoverFocusSet = true
	edit.popoverFocus = 5
	assertMenuHandled(t, edit.ConsumeKey(KeyEvent{Key: "space"}))
	if edit.popoverColorPicker != nil || edit.popoverFocusSet || !edit.popoverSuppressed {
		t.Fatalf("unrenderable toolbar retained input focus: picker=%v focusSet=%v suppressed=%v", edit.popoverColorPicker != nil, edit.popoverFocusSet, edit.popoverSuppressed)
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
