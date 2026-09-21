package loom

import "testing"

func TestViewPagerNavigation(t *testing.T) {
	v := NewView(make([]string, 30))
	v.Draw(NewCanvas(20, 6), Rect{W: 20, H: 6})
	for _, tc := range []struct {
		key  KeyEvent
		want int
	}{
		{KeyEvent{Key: "pgdown"}, 6},
		{KeyEvent{Key: "ctrl-f"}, 12},
		{KeyEvent{Text: " "}, 18},
		{KeyEvent{Key: "ctrl-d"}, 21},
		{KeyEvent{Key: "end"}, 24},
		{KeyEvent{Text: "G"}, 24},
		{KeyEvent{Key: "ctrl-u"}, 21},
		{KeyEvent{Key: "pgup"}, 15},
		{KeyEvent{Key: "ctrl-b"}, 9},
		{KeyEvent{Text: "b"}, 3},
		{KeyEvent{Key: "home"}, 0},
		{KeyEvent{Text: "g"}, 0},
	} {
		if v.HandleKey(tc.key) || v.Scroll != tc.want {
			t.Fatalf("key %+v: scroll %d, want %d", tc.key, v.Scroll, tc.want)
		}
	}
}

func TestViewScrollbarTrackClick(t *testing.T) {
	v := NewView(make([]string, 30))
	v.Draw(NewCanvas(12, 5), Rect{X: 2, Y: 1, W: 10, H: 4})
	for _, tc := range []struct {
		x, y int
		want int
	}{
		{11, 4, 26}, // bottom of the track
		{11, 2, 8},  // second track row
		{10, 5, 8},  // content column: no jump
		{11, 1, 0},  // top of the track
	} {
		v.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: tc.x, Y: tc.y})
		if v.Scroll != tc.want {
			t.Fatalf("click (%d,%d): scroll=%d, want %d", tc.x, tc.y, v.Scroll, tc.want)
		}
	}
	v.HandleMouse(MouseEvent{Action: MousePress, Button: MouseRight, X: 11, Y: 4})
	if v.Scroll != 0 {
		t.Fatal("right click moved scrollbar")
	}
}

func TestViewDrawsScrollbarTrack(t *testing.T) {
	thumb := SpeccedDefaults.Scrollbar.ForegroundChar
	track := SpeccedDefaults.Scrollbar.BackgroundChar
	v := NewView(make([]string, 20))
	v.Scrollbar = ScrollbarStyle{
		Track: Style{FG: ColorIndex(33), BG: ColorIndex(27)},
		Thumb: Style{FG: ColorIndex(51), BG: ColorIndex(27)},
	}
	canvas := NewCanvas(10, 4)
	v.Draw(canvas, Rect{W: 10, H: 4})
	for y := 0; y < 4; y++ {
		want := track
		if y == 0 {
			want = thumb
		}
		if got := canvas.Get(9, y).Text; got != want {
			t.Fatalf("track row %d = %q, want %q", y, got, want)
		}
		wantStyle := v.Scrollbar.Track
		if y == 0 {
			wantStyle = v.Scrollbar.Thumb
		}
		if got := canvas.Get(9, y).Style; got != wantStyle {
			t.Fatalf("track row %d style = %+v, want %+v", y, got, wantStyle)
		}
	}
	v.Scroll = 16
	v.Draw(canvas, Rect{W: 10, H: 4})
	if canvas.Get(9, 0).Text != track || canvas.Get(9, 3).Text != thumb {
		t.Fatal("thumb did not move over the visible track")
	}
}

func TestViewFocusable(t *testing.T) {
	v := NewView([]string{"item 1", "item 2"})
	v.Style = Style{Dim: true}
	v.FocusStyle = Style{Bold: true}

	if v.Focused() {
		t.Fatal("expected initially not focused")
	}

	canvas := NewCanvas(10, 2)
	v.Draw(canvas, canvas.Bounds())
	if !canvas.Get(0, 0).Style.Dim {
		t.Fatal("expected dim style when unfocused")
	}

	v.SetFocus(true)
	if !v.Focused() {
		t.Fatal("expected focused after SetFocus(true)")
	}

	canvas.Clear()
	v.Draw(canvas, canvas.Bounds())
	if !canvas.Get(0, 0).Style.Bold {
		t.Fatal("expected bold style when focused")
	}

	v.SetFocus(false)
	if v.Focused() {
		t.Fatal("expected not focused after SetFocus(false)")
	}
}

func TestScrollbarMapping(t *testing.T) {
	for _, tc := range []struct {
		name                             string
		track, content, viewport, offset int
		wantThumb, wantStart             int
	}{
		{"normal", 10, 100, 20, 40, 2, 4},
		{"minimum", 3, 1000, 1, 0, 1, 0},
		{"equal", 8, 10, 10, 0, 0, 0},
		{"short-content", 8, 4, 10, 0, 0, 0},
		{"length-one", 1, 2, 1, 1, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			thumb := scrollbarThumbLength(tc.track, tc.content, tc.viewport)
			if thumb != tc.wantThumb {
				t.Fatalf("thumb=%d, want %d", thumb, tc.wantThumb)
			}
			if got := scrollbarThumbStart(tc.track, thumb, tc.offset, tc.content-tc.viewport); got != tc.wantStart {
				t.Fatalf("start=%d, want %d", got, tc.wantStart)
			}
		})
	}
}

func TestScrollbarDragSequenceAndHorizontalMapping(t *testing.T) {
	var drag scrollbarDrag
	drag.press(2, 1, 3)
	if got := scrollbarOffset(10, 3, 8, drag.grab, 7); got != 7 {
		t.Fatalf("drag offset=%d, want 7", got)
	}
	drag.cancel()
	if drag.active {
		t.Fatal("cancel left drag active")
	}
	if got := scrollbarOffset(12, 2, 99, 0, 10); got != 10 {
		t.Fatalf("clamped offset=%d, want 10", got)
	}
	// Horizontal scrollbars use the same integer mapping with an independent axis.
	if got := scrollbarThumbStart(20, scrollbarThumbLength(20, 80, 10), 35, 70); got != 9 {
		t.Fatalf("horizontal thumb start=%d, want 9", got)
	}
}

func TestViewScrollbarDrag(t *testing.T) {
	v := NewView(make([]string, 100))
	v.Draw(NewCanvas(12, 10), Rect{W: 12, H: 10})
	v.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 11, Y: 1})
	v.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 11, Y: 9})
	if v.Scroll <= 0 {
		t.Fatal("drag did not scroll")
	}
	v.HandleMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 11, Y: 9})
	if v.drag.active {
		t.Fatal("release left drag active")
	}
}

func TestViewScrollbarDragCancelAndCapture(t *testing.T) {
	v := NewView(make([]string, 100))
	v.Draw(NewCanvas(12, 10), Rect{W: 12, H: 10})
	v.Scroll = 20
	v.Draw(NewCanvas(12, 10), Rect{W: 12, H: 10})
	start := v.Scroll
	v.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 11, Y: 2})
	v.HandleMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 11, Y: 99})
	if v.Scroll == start {
		t.Fatal("captured drag did not update outside the widget")
	}
	v.HandleKey(KeyEvent{Key: "esc"})
	if v.Scroll != start || v.drag.active {
		t.Fatalf("escape cancel: scroll=%d active=%v, want %d false", v.Scroll, v.drag.active, start)
	}
	v.HandleMouse(MouseEvent{Action: MousePress, Button: MouseRight, X: 11, Y: 2})
	if v.drag.active {
		t.Fatal("non-primary press started drag")
	}
}
