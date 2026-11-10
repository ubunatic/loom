package loom

import (
	"testing"
)

func TestTextEntrySelectionDragCapture(t *testing.T) {
	pane := &Pane{}

	entry := NewTextEntry(TextEntryOptions{})
	entry.SetValue("hello world")
	entry.SetFocus(true)

	other := NewTextEntry(TextEntryOptions{})
	root := NewStack(Horizontal, entry, other)

	canvas := NewCanvas(80, 24)
	root.Draw(canvas, Rect{0, 0, 80, 24})

	pane.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 0})
	pane.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 41, Y: 0})

	if !entry.dragging {
		t.Errorf("entry lost drag capture during move outside, entry.dragging=false")
	}

	pane.dispatchMouse(root, MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 41, Y: 0})

	if entry.dragging {
		t.Errorf("Release outside left entry.dragging = true")
	}

	// Verify later motion over the entry does not extend selection
	entry.dragging = false // Clear dragging if it was left on by bug
	pane.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 2, Y: 0})
	if entry.dragging {
		t.Errorf("motion over entry extended selection after release")
	}

	// Focus change mid-drag
	entry.SetFocus(true)
	entry.lastClick = entry.now().Add(-SpeccedDefaults.Mouse.DoubleClickInterval - 1)
	pane.dispatchMouse(root, MouseEvent{Action: MousePress, Button: MouseLeft, X: 1, Y: 0})
	pane.dispatchMouse(root, MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 2, Y: 0})

	if !entry.dragging {
		t.Errorf("expected entry.dragging = true before SetFocus(false)")
	}

	entry.SetFocus(false)

	if entry.dragging {
		t.Errorf("SetFocus(false) mid-drag left entry.dragging = true")
	}

	t.Log("Note: skip auto-scroll test per instruction")
}
