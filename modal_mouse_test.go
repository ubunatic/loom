package loom

import (
	"fmt"
	"testing"
)

type modalDocumentState struct {
	cursor, from, to, anchor RichPosition
	selected                 bool
	scrollX, scrollY         int
	modified                 bool
	document                 string
}

func modalEditorState(e *RichTextEdit) modalDocumentState {
	return modalDocumentState{e.Cursor, e.SelectionFrom, e.SelectionTo, e.selectionAnchor,
		e.HasSelection, e.ScrollX, e.ScrollY, e.IsModified(), e.Document.ToANSI()}
}

func modalTestEditor(t *testing.T) *RichTextEdit {
	t.Helper()
	e := NewRichTextEdit(&RichDocument{Lines: []RichLine{
		{Spans: []RichSpan{{Text: "first word"}}},
		{Spans: []RichSpan{{Text: "second line"}}},
	}})
	e.Cursor.Offset = 2
	e.Draw(NewCanvas(80, 24), Rect{X: 3, Y: 2, W: 70, H: 20})
	if result := e.ConsumeKey(KeyEvent{Key: "ctrl-space"}); result != Handled() {
		t.Fatalf("open popover: %+v", result)
	}
	// Capture is based on state, including before the next frame is rendered.
	if target, _ := ActiveModalMouse(e); target != e {
		t.Fatal("popover did not acquire capture immediately")
	}
	e.Draw(NewCanvas(80, 24), Rect{X: 3, Y: 2, W: 70, H: 20})
	return e
}

func Test304PopoverMouseIsolation(t *testing.T) {
	for _, action := range []MouseAction{MousePress, MouseRelease, MouseDrag, MouseHover, MouseScrollUp, MouseScrollDown} {
		for _, button := range []MouseButton{MouseLeft, MouseRight, MouseMiddle} {
			t.Run(fmt.Sprintf("%v/%v", action, button), func(t *testing.T) {
				e := modalTestEditor(t)
				before := modalEditorState(e)
				if got := e.ConsumeMouse(MouseEvent{X: 25, Y: 0, Action: action, Button: button}); got != Handled() {
					t.Fatalf("backdrop result = %+v", got)
				}
				if after := modalEditorState(e); after != before || e.dragSelecting {
					t.Fatalf("backdrop changed editor: before=%+v after=%+v", before, after)
				}
				wantOpen := action != MousePress || button != MouseLeft
				if e.ModalOpen() != wantOpen {
					t.Fatalf("modal open = %v, want %v", e.ModalOpen(), wantOpen)
				}
				if !wantOpen {
					for _, trailing := range []MouseAction{MouseDrag, MouseRelease} {
						e.Draw(NewCanvas(80, 24), Rect{X: 3, Y: 2, W: 70, H: 20})
						if got := e.ConsumeMouse(MouseEvent{X: 8, Y: 1, Action: trailing, Button: MouseLeft}); got != Handled() {
							t.Fatalf("trailing event = %+v", got)
						}
					}
					if modalEditorState(e) != before {
						t.Fatal("dismissal gesture changed cursor or selection")
					}
					// A new gesture must work normally after dismissal.
					e.ConsumeMouse(MouseEvent{X: 3, Y: 1, Action: MousePress, Button: MouseLeft})
					e.ConsumeMouse(MouseEvent{X: 6, Y: 1, Action: MouseDrag, Button: MouseLeft})
					e.ConsumeMouse(MouseEvent{X: 6, Y: 1, Action: MouseRelease, Button: MouseLeft})
					if e.Cursor != (RichPosition{Line: 1, Offset: 6}) || !e.HasSelection || e.dragSelecting {
						t.Fatalf("normal selection did not resume: %+v", modalEditorState(e))
					}
				}
			})
		}
	}
}

func Test304PopoverCancelsOldDragAndKeyboardDismisses(t *testing.T) {
	e := modalTestEditor(t)
	e.ConsumeKey(KeyEvent{Key: "esc"})
	e.ConsumeMouse(MouseEvent{X: 1, Y: 0, Action: MousePress, Button: MouseLeft})
	e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	if e.dragSelecting {
		t.Fatal("keyboard modal retained document drag")
	}
	before := modalEditorState(e)
	if got := e.ConsumeKey(KeyEvent{Key: "esc"}); got != Handled() || e.ModalOpen() {
		t.Fatalf("keyboard dismissal: %+v open=%v", got, e.ModalOpen())
	}
	if modalEditorState(e) != before {
		t.Fatal("keyboard dismissal changed document selection")
	}
}

type modalMouseProbe struct {
	open    bool
	height  int
	events  []MouseEvent
	focused bool
	result  EventResult
}

func (*modalMouseProbe) Draw(*Canvas, Rect)   {}
func (p *modalMouseProbe) ContentHeight() int { return max(6, p.height) }
func (p *modalMouseProbe) SetFocus(f bool)    { p.focused = f }
func (p *modalMouseProbe) Focused() bool      { return p.focused }
func (p *modalMouseProbe) ConsumeKey(e KeyEvent) EventResult {
	if p.open && e.Is("esc") {
		p.open = false
		return Handled()
	}
	return Ignored()
}
func (p *modalMouseProbe) ConsumeMouse(e MouseEvent) EventResult {
	p.events = append(p.events, e)
	if e.Action == MousePress && e.Button == MouseLeft {
		p.open = false
	}
	return p.result
}
func (p *modalMouseProbe) ModalMouseTarget() (Widget, Rect) {
	if p.open {
		return p, Rect{}
	}
	return nil, Rect{}
}

func Test304ContainersCaptureBeforeFocusAndHitTesting(t *testing.T) {
	for _, test := range []struct {
		name  string
		build func(Widget, Widget) Widget
	}{
		{"split", func(a, b Widget) Widget { return NewSplit(a, b) }},
		{"stack", func(a, b Widget) Widget { return NewStack(Horizontal, a, b) }},
		{"grid", func(a, b Widget) Widget { return NewGrid(2, a, b) }},
		{"tabs", func(a, b Widget) Widget {
			return NewTabs(Tab{Title: "First", Widget: a}, Tab{Title: "Other", Widget: b})
		}},
		{"viewport", func(a, _ Widget) Widget { return NewViewport(a) }},
		{"form", func(a, b Widget) Widget {
			return NewForm([]FormField{{Label: "First", Widget: a}, {Label: "Other", Widget: b}})
		}},
		{"frame", func(a, b Widget) Widget {
			return &Frame{Boxes: []Box{{Child: a, Width: 40, Height: 22}, {Child: b, Width: 40, Height: 22}}}
		}},
		{"box", func(a, _ Widget) Widget { return &Box{Child: a, Padding: 1} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			modal, sibling := &modalMouseProbe{open: true}, &modalMouseProbe{}
			root := test.build(modal, sibling)
			root.Draw(NewCanvas(100, 30), Rect{X: 7, Y: 3, W: 80, H: 24})
			focus := sibling.focused
			for _, action := range []MouseAction{MouseHover, MouseScrollDown, MousePress, MouseDrag, MouseRelease} {
				if got := root.ConsumeMouse(MouseEvent{X: 79, Y: 23, Action: action, Button: MouseLeft}); got != Handled() {
					t.Fatalf("%v = %+v", action, got)
				}
				root.Draw(NewCanvas(100, 30), Rect{X: 7, Y: 3, W: 80, H: 24})
			}
			if modal.open || len(modal.events) != 3 || len(sibling.events) != 0 || sibling.focused != focus {
				t.Fatalf("modal=%+v sibling=%+v", modal, sibling)
			}
			if tabs, ok := root.(*Tabs); ok && tabs.Focus() != 0 {
				t.Fatal("modal click switched tab")
			}
		})
	}
}

func Test304NestedCoordinatesAndTopmostModal(t *testing.T) {
	a, b := &modalMouseProbe{open: true}, &modalMouseProbe{open: true}
	inner := NewStack(Horizontal, a, b)
	tabs := NewTabs(Tab{Title: "Active", Widget: inner}, Tab{Title: "Hidden", Widget: &modalMouseProbe{open: true}})
	split := NewSplit(NewView(nil), tabs)
	split.Draw(NewCanvas(100, 30), Rect{X: 7, Y: 3, W: 80, H: 24})
	// Split's second child starts at x=41, tab content at y=2, and the
	// stack's second child at x=19: the nested modal origin is (60,2).
	target, origin := ActiveModalMouse(split)
	if target != b || origin.X != 60 || origin.Y != 2 {
		t.Fatalf("nested target/origin = %T %+v, want last modal at (60,2)", target, origin)
	}
	if got := split.ConsumeMouse(MouseEvent{X: 5, Y: 4, Action: MousePress, Button: MouseLeft}); got != Handled() {
		t.Fatal(got)
	}
	if len(a.events) != 0 || len(b.events) != 1 || b.events[0].X != -55 || b.events[0].Y != 2 {
		t.Fatalf("bad modal routing: a=%+v b=%+v", a.events, b.events)
	}
	// Release belongs to the dismissed topmost overlay, not the one below it.
	split.ConsumeMouse(MouseEvent{X: 5, Y: 4, Action: MouseRelease, Button: MouseLeft})
	if len(a.events) != 0 {
		t.Fatal("release leaked to lower overlay")
	}
	if target, _ := ActiveModalMouse(split); target != a {
		t.Fatal("dismissing top modal lost lower modal")
	}
}

func Test304PaneModalCapturePreservesResult(t *testing.T) {
	for _, result := range []EventResult{Ignored(), DoneResult(), QuitResult()} {
		modal := &modalMouseProbe{open: true, result: result}
		p := &Pane{}
		got := p.dispatchMouse(modal, MouseEvent{Action: MousePress, Button: MouseLeft})
		result.Consumed = true
		if got != result {
			t.Fatalf("result = %+v, want %+v", got, result)
		}
		if got := p.dispatchMouse(modal, MouseEvent{Action: MouseRelease, Button: MouseLeft}); got != Handled() || len(modal.events) != 1 {
			t.Fatal("pane leaked release after dismissal")
		}
	}
}

func Test304OverlayKindsConsumeBackdrop(t *testing.T) {
	for _, w := range []Widget{
		NewPopup("Popup", &modalMouseProbe{}), NewDialog("Dialog", "Body", "Cancel"),
		&Menu{Items: []MenuItem{{Label: "Item"}}},
		&MenuBar{Open: true, Menus: []Menu{{Title: "File", Items: []MenuItem{{Label: "Item"}}}}},
	} {
		t.Run(fmt.Sprintf("%T", w), func(t *testing.T) {
			w.Draw(NewCanvas(100, 30), Rect{X: 7, Y: 3, W: 80, H: 24})
			for _, action := range []MouseAction{MouseHover, MouseScrollDown, MouseDrag, MouseRelease, MousePress} {
				if got := w.ConsumeMouse(MouseEvent{X: -2, Y: -2, Action: action, Button: MouseLeft}); got != Handled() {
					t.Fatalf("%v = %+v", action, got)
				}
			}
			if target, _ := ActiveModalMouse(w); target != nil {
				t.Fatal("backdrop did not dismiss overlay")
			}
		})
	}
}

func Test304ViewportModalPreemptsScrolling(t *testing.T) {
	modal := &modalMouseProbe{open: true, height: 60}
	v := NewViewport(modal)
	v.ScrollY = 4
	v.Draw(NewCanvas(80, 24), Rect{X: 7, Y: 3, W: 60, H: 10})
	v.ConsumeMouse(MouseEvent{X: 59, Y: 9, Action: MouseScrollDown})
	v.ConsumeMouse(MouseEvent{X: 59, Y: 9, Action: MousePress, Button: MouseLeft})
	if v.ScrollY != 4 || v.dragging || len(modal.events) != 2 || modal.events[1].Y != 13 {
		t.Fatalf("viewport intercepted modal input or mistranslated coordinates: scroll=%d events=%+v", v.ScrollY, modal.events)
	}
}

func Test304NestedPopupOwnsOuterBackdrop(t *testing.T) {
	inner := NewPopup("Inner", &modalMouseProbe{})
	outer := NewPopup("Outer", inner)
	outer.Width, outer.Height = 60, 18
	outer.Draw(NewCanvas(80, 24), Rect{W: 80, H: 24})
	for _, action := range []MouseAction{MousePress, MouseDrag, MouseRelease} {
		if got := outer.ConsumeMouse(MouseEvent{X: 0, Y: 0, Action: action, Button: MouseLeft}); got != Handled() {
			t.Fatal(got)
		}
	}
	if inner.Open || !outer.Open {
		t.Fatal("inner backdrop dismissed the outer popup")
	}
	outer.ConsumeMouse(MouseEvent{X: 0, Y: 0, Action: MousePress, Button: MouseLeft})
	if outer.Open {
		t.Fatal("next press did not dismiss outer popup")
	}
}

func Test304CommandHelpUsesModalContract(t *testing.T) {
	choice := NewChoice([]Item{{Name: "One"}})
	choice.cmd.help = NewPopup("Help", NewView(nil))
	choice.Draw(NewCanvas(80, 24), Rect{W: 80, H: 24})
	p := &Pane{}
	if target, _ := ActiveModalMouse(choice); target != choice {
		t.Fatal("command help did not expose capture")
	}
	p.dispatchMouse(choice, MouseEvent{X: 79, Y: 23, Action: MousePress, Button: MouseLeft})
	p.dispatchMouse(choice, MouseEvent{X: 0, Y: 0, Action: MouseRelease, Button: MouseLeft})
	if choice.cmd.help != nil {
		t.Fatal("dismissed command help retained capture")
	}
}

func Test304ModelessContentLeavesSiblingsInteractive(t *testing.T) {
	popup := NewPopup("Embedded", NewView(nil))
	popup.Modeless = true
	dialog := NewDialog("Embedded", "Body", "OK")
	dialog.Modeless = true
	for _, w := range []Widget{
		popup, dialog, &Menu{Modeless: true, Items: []MenuItem{{Label: "Item"}}},
		&MenuBar{Modeless: true, Open: true, Menus: []Menu{{Title: "File", Items: []MenuItem{{Label: "Item"}}}}},
	} {
		t.Run(fmt.Sprintf("%T", w), func(t *testing.T) {
			sibling := &modalMouseProbe{result: Handled()}
			grid := NewGrid(2, w, sibling)
			grid.Draw(NewCanvas(80, 24), Rect{W: 80, H: 24})
			if target, _ := ActiveModalMouse(grid); target != nil {
				t.Fatal("modeless content captured the grid")
			}
			if got := grid.ConsumeMouse(MouseEvent{X: 79, Y: 23, Action: MousePress, Button: MouseLeft}); got != Handled() ||
				grid.Focus() != 1 || len(sibling.events) != 1 {
				t.Fatal("modeless content blocked a sibling")
			}
		})
	}
}
