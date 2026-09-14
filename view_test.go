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
