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
		{12, 5, 26}, // bottom of the track
		{12, 3, 8},  // second track row
		{11, 5, 8},  // content column: no jump
		{12, 2, 0},  // top of the track
	} {
		v.HandleMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: tc.x, Y: tc.y})
		if v.Scroll != tc.want {
			t.Fatalf("click (%d,%d): scroll=%d, want %d", tc.x, tc.y, v.Scroll, tc.want)
		}
	}
	v.HandleMouse(MouseEvent{Action: MousePress, Button: MouseRight, X: 12, Y: 5})
	if v.Scroll != 0 {
		t.Fatal("right click moved scrollbar")
	}
}
