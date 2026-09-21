// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestSplitLayoutRatioAndOrientation(t *testing.T) {
	tests := []struct {
		name        string
		orientation Orientation
		bounds      Rect
		ratio       float64
		gap         int
		wantFirst   Rect
		wantSecond  Rect
	}{
		{"horizontal", Horizontal, Rect{X: 2, Y: 3, W: 21, H: 8}, 0.25, 1, Rect{X: 2, Y: 3, W: 5, H: 8}, Rect{X: 8, Y: 3, W: 15, H: 8}},
		{"vertical", Vertical, Rect{X: 2, Y: 3, W: 8, H: 21}, 0.75, 1, Rect{X: 2, Y: 3, W: 8, H: 15}, Rect{X: 2, Y: 19, W: 8, H: 5}},
		{"ratio-clamped-low", Horizontal, Rect{W: 10, H: 2}, -2, 0, Rect{W: 0, H: 2}, Rect{W: 10, H: 2}},
		{"ratio-clamped-high", Horizontal, Rect{W: 10, H: 2}, 2, 0, Rect{W: 10, H: 2}, Rect{X: 10, W: 0, H: 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			split := &Split{Orientation: tc.orientation, Ratio: tc.ratio, Gap: tc.gap}
			first, second := split.Layout(tc.bounds)
			if first != tc.wantFirst || second != tc.wantSecond {
				t.Fatalf("Layout() = %+v, %+v; want %+v, %+v", first, second, tc.wantFirst, tc.wantSecond)
			}
		})
	}
}

func TestSplitLayoutMinimumSizes(t *testing.T) {
	tests := []struct {
		name                  string
		ratio                 float64
		minFirst, minSecond   int
		wantFirst, wantSecond int
	}{
		{"first minimum", 0.1, 6, 2, 6, 13},
		{"second minimum", 0.9, 2, 7, 12, 7},
		{"both exact", 0.5, 8, 11, 8, 11},
		{"unsatisfied favors first", 0.5, 15, 10, 15, 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			split := &Split{Orientation: Horizontal, Ratio: tc.ratio, Gap: 1, MinFirst: tc.minFirst, MinSecond: tc.minSecond}
			first, second := split.Layout(Rect{W: 20, H: 4})
			if first.W != tc.wantFirst || second.W != tc.wantSecond {
				t.Fatalf("widths = %d,%d; want %d,%d", first.W, second.W, tc.wantFirst, tc.wantSecond)
			}
		})
	}
}

func TestSplitSetRatioAndDivider(t *testing.T) {
	split := NewSplit(nil, nil)
	split.SetRatio(4)
	if split.Ratio != 1 {
		t.Fatalf("SetRatio(4) = %v, want 1", split.Ratio)
	}
	split.HandleKey(KeyEvent{Text: "["})
	if split.Ratio != 0.95 {
		t.Fatalf("ratio after [ = %v, want .95", split.Ratio)
	}
	split.Ratio = 0.5
	split.Divider = DividerStyle{Glyph: "#", Style: Style{Bold: true}}
	canvas := NewCanvas(9, 3)
	split.Draw(canvas, canvas.Bounds())
	for y := 0; y < 3; y++ {
		cell := canvas.Get(4, y)
		if cell.Text != "#" || !cell.Style.Bold {
			t.Fatalf("divider cell at y=%d = %+v", y, cell)
		}
	}
}

func TestNestedSplitFocusTraversal(t *testing.T) {
	a, b, c := &focusProbe{}, &focusProbe{}, &focusProbe{}
	nested := NewSplit(a, b)
	root := NewSplit(nested, c)
	root.SetFocus(true)
	if !a.Focused() || b.Focused() || c.Focused() {
		t.Fatal("initial nested focus should select first leaf")
	}
	if !root.FocusNext() || a.Focused() || !b.Focused() || c.Focused() {
		t.Fatal("first traversal should select nested second leaf")
	}
	if !root.FocusNext() || a.Focused() || b.Focused() || !c.Focused() {
		t.Fatal("second traversal should select outer second leaf")
	}
	if root.FocusNext() {
		t.Fatal("FocusNext should report the outer boundary")
	}
	if !root.FocusPrevious() || a.Focused() || !b.Focused() || c.Focused() {
		t.Fatal("reverse traversal should enter nested split at its last leaf")
	}
	if !root.FocusPrevious() || !a.Focused() || b.Focused() || c.Focused() {
		t.Fatal("second reverse traversal should select first leaf")
	}
	if root.FocusPrevious() {
		t.Fatal("FocusPrevious should report the outer boundary")
	}
}

func TestFrameTraversesNestedSplit(t *testing.T) {
	a, b, c := &focusProbe{}, &focusProbe{}, &focusProbe{}
	nested := NewSplit(a, b)
	frame := &Frame{Boxes: []Box{{ID: "split", Child: nested}, {ID: "tail", Child: c}}}
	frame.HandleKey(KeyEvent{Key: "tab"})
	if !b.Focused() || c.Focused() {
		t.Fatal("frame tab did not traverse within split")
	}
	frame.HandleKey(KeyEvent{Key: "tab"})
	if b.Focused() || !c.Focused() {
		t.Fatal("frame tab did not continue after split")
	}
	frame.HandleKey(KeyEvent{Key: "shift-tab"})
	if !b.Focused() || c.Focused() {
		t.Fatal("frame reverse tab did not enter split at last leaf")
	}
}

func TestSplitAndBoxMouseRouting(t *testing.T) {
	left, right := &focusProbe{}, &focusProbe{}
	split := NewSplit(left, right)
	canvas := NewCanvas(20, 6)
	split.Draw(canvas, Rect{X: 2, Y: 1, W: 15, H: 4})
	click := MouseEvent{Action: MousePress, Button: MouseLeft, X: 14, Y: 3}
	if split.HandleMouse(click) || len(left.mice) != 0 || len(right.mice) != 1 || !right.Focused() {
		t.Fatalf("split mouse route: left=%d right=%d focused=%v", len(left.mice), len(right.mice), right.Focused())
	}
	if right.mice[0].X != 4 || right.mice[0].Y != 2 {
		t.Fatalf("translated split event = %+v, want X=4 Y=2", right.mice[0])
	}

	child := &focusProbe{}
	box := &Box{Padding: 1, Child: child}
	box.Draw(canvas, Rect{X: 3, Y: 0, W: 10, H: 6})
	box.HandleMouse(MouseEvent{Action: MouseHover, X: 6, Y: 4})
	if len(child.mice) != 1 || child.mice[0].X != 1 || child.mice[0].Y != 2 {
		t.Fatalf("box child event = %+v", child.mice)
	}
	box.HandleMouse(MouseEvent{Action: MouseHover, X: 4, Y: 2})
	if len(child.mice) != 1 {
		t.Fatal("box border/padding event reached child")
	}
}

func TestSplitDividerMouseDragChangesRatio(t *testing.T) {
	split := NewSplit(&focusProbe{}, &focusProbe{})
	canvas := NewCanvas(21, 4)
	split.Draw(canvas, canvas.Bounds())
	if split.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 10, Y: 2}) {
		t.Fatal("divider press requested quit")
	}
	split.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 4, Y: 2})
	if split.Ratio != 0.2 {
		t.Fatalf("dragged ratio = %v, want .2", split.Ratio)
	}
	split.HandleMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 4, Y: 2})
}

func TestSplitCapturesDragAndReleaseOutsideChild(t *testing.T) {
	left, right := &focusProbe{}, &focusProbe{}
	split := NewSplit(left, right)
	canvas := NewCanvas(21, 4)
	split.Draw(canvas, canvas.Bounds())
	split.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 15, Y: 1})
	split.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 3, Y: 9})
	split.HandleMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 3, Y: 9})
	split.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 3, Y: 9})
	if len(left.mice) != 0 || len(right.mice) != 3 {
		t.Fatalf("left=%d right=%d events, want 0 and 3 (press, drag, release)", len(left.mice), len(right.mice))
	}
	if got := right.mice[1]; got.X != -8 || got.Y != 9 {
		t.Fatalf("captured drag = %+v, want child-relative X=-8 Y=9", got)
	}
}
