// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestScanKeySeparatesPrintableKeys(t *testing.T) {
	for _, input := range []string{"r ", "ä🙂r ", "r\x1b[A "} {
		t.Run(input, func(t *testing.T) {
			var got []KeyEvent
			for raw := []byte(input); len(raw) > 0; {
				event, used, ok := scanKey(raw)
				if !ok || used <= 0 {
					t.Fatalf("scanKey(%q) made no progress", raw)
				}
				got = append(got, event)
				raw = raw[used:]
			}
			var want []KeyEvent
			switch input {
			case "r ":
				want = []KeyEvent{{Text: "r"}, {Text: " "}}
			case "ä🙂r ":
				want = []KeyEvent{{Text: "ä"}, {Text: "🙂"}, {Text: "r"}, {Text: " "}}
			default:
				want = []KeyEvent{{Text: "r"}, {Key: "up"}, {Text: " "}}
			}
			if len(got) != len(want) {
				t.Fatalf("events = %+v, want %+v", got, want)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("event %d = %+v, want %+v", i, got[i], want[i])
				}
			}
		})
	}
}

func TestScanKeyWaitsForCompleteUTF8Rune(t *testing.T) {
	raw := []byte("🙂")
	for end := 1; end < len(raw); end++ {
		if event, used, ok := scanKey(raw[:end]); ok || used != 0 || event != (KeyEvent{}) {
			t.Fatalf("partial rune decoded as (%+v, %d, %v)", event, used, ok)
		}
	}
}

func TestDecodeKey(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected KeyEvent
	}{
		{"empty", nil, KeyEvent{}},
		{"ctrl-c", []byte{3}, KeyEvent{Key: "ctrl-c"}},
		{"ctrl-d", []byte{4}, KeyEvent{Key: "ctrl-d"}},
		{"ctrl-b", []byte{2}, KeyEvent{Key: "ctrl-b"}},
		{"ctrl-f", []byte{6}, KeyEvent{Key: "ctrl-f"}},
		{"ctrl-u", []byte{21}, KeyEvent{Key: "ctrl-u"}},
		{"shift-tab", []byte{27, '[', 'Z'}, KeyEvent{Key: "shift-tab"}},
		{"enter", []byte{13}, KeyEvent{Key: "enter"}},
		{"esc", []byte{27}, KeyEvent{Key: "esc"}},
		{"up", []byte{27, '[', 'A'}, KeyEvent{Key: "up"}},
		{"down", []byte{27, '[', 'B'}, KeyEvent{Key: "down"}},
		{"left", []byte{27, '[', 'D'}, KeyEvent{Key: "left"}},
		{"right", []byte{27, '[', 'C'}, KeyEvent{Key: "right"}},
		{"shift-up xterm", []byte{27, '[', '1', ';', '2', 'A'}, KeyEvent{Key: "shift-up"}},
		{"shift-down xterm", []byte{27, '[', '1', ';', '2', 'B'}, KeyEvent{Key: "shift-down"}},
		{"ctrl-up xterm", []byte{27, '[', '1', ';', '5', 'A'}, KeyEvent{Key: "ctrl-up"}},
		{"ctrl-down xterm", []byte{27, '[', '1', ';', '5', 'B'}, KeyEvent{Key: "ctrl-down"}},
		{"alt-up xterm", []byte{27, '[', '1', ';', '3', 'A'}, KeyEvent{Key: "alt-up"}},
		{"alt-down xterm", []byte{27, '[', '1', ';', '3', 'B'}, KeyEvent{Key: "alt-down"}},
		{"shift-up rxvt", []byte{27, '[', 'a'}, KeyEvent{Key: "shift-up"}},
		{"shift-down rxvt", []byte{27, '[', 'b'}, KeyEvent{Key: "shift-down"}},
		{"f1 SS3", []byte{27, 'O', 'P'}, KeyEvent{Key: "f1"}},
		{"f2 SS3", []byte{27, 'O', 'Q'}, KeyEvent{Key: "f2"}},
		{"f3 SS3", []byte{27, 'O', 'R'}, KeyEvent{Key: "f3"}},
		{"f4 SS3", []byte{27, 'O', 'S'}, KeyEvent{Key: "f4"}},
		{"f5 tilde", []byte{27, '[', '1', '5', '~'}, KeyEvent{Key: "f5"}},
		{"f10 tilde", []byte{27, '[', '2', '1', '~'}, KeyEvent{Key: "f10"}},
		{"f12 tilde", []byte{27, '[', '2', '4', '~'}, KeyEvent{Key: "f12"}},
		{"shift-f12 tilde", []byte{27, '[', '2', '4', ';', '2', '~'}, KeyEvent{Key: "shift-f12"}},
		{"ctrl-f12 tilde", []byte{27, '[', '2', '4', ';', '5', '~'}, KeyEvent{Key: "ctrl-f12"}},
		{"shift-f1 csi 1", []byte{27, '[', '1', ';', '2', 'P'}, KeyEvent{Key: "shift-f1"}},
		{"home", []byte{27, '[', '1', '~'}, KeyEvent{Key: "home"}},
		{"delete", []byte{27, '[', '3', '~'}, KeyEvent{Key: "delete"}},
		{"end", []byte{27, '[', '4', '~'}, KeyEvent{Key: "end"}},
		{"pgup", []byte{27, '[', '5', '~'}, KeyEvent{Key: "pgup"}},
		{"pgdown", []byte{27, '[', '6', '~'}, KeyEvent{Key: "pgdown"}},
		{"legacy tab stays tab", []byte{9}, KeyEvent{Key: "tab"}},
		{"csi-u ctrl-i", []byte("\x1b[105;5u"), KeyEvent{Key: "ctrl-i"}},
		{"csi-u plain tab", []byte("\x1b[9u"), KeyEvent{Key: "tab"}},
		{"ctrl-space nul", []byte{0}, KeyEvent{Key: "ctrl-space"}},
		{"csi-u ctrl-space", []byte("\x1b[32;5u"), KeyEvent{Key: "ctrl-space"}},
		{"shift-insert", []byte("\x1b[2;2~"), KeyEvent{Key: "shift-insert"}},
		{"ctrl-insert", []byte("\x1b[2;5~"), KeyEvent{Key: "ctrl-insert"}},
		{"shift-delete", []byte("\x1b[3;2~"), KeyEvent{Key: "shift-delete"}},
		{"plain insert", []byte("\x1b[2~"), KeyEvent{Key: "insert"}},
		{"csi-u ctrl-shift-z", []byte("\x1b[122;6u"), KeyEvent{Key: "ctrl-shift-z"}},
		{"csi-u ctrl-shift-b", []byte("\x1b[98;6u"), KeyEvent{Key: "ctrl-shift-b"}},
		{"csi-u ctrl-shift-y", []byte("\x1b[121;6u"), KeyEvent{Key: "ctrl-shift-y"}},
		{"legacy home csi H", []byte("\x1b[H"), KeyEvent{Key: "home"}},
		{"legacy end csi F", []byte("\x1b[F"), KeyEvent{Key: "end"}},
		{"ss3 home", []byte("\x1bOH"), KeyEvent{Key: "home"}},
		{"ss3 end", []byte("\x1bOF"), KeyEvent{Key: "end"}},
		{"tilde 1", []byte("\x1b[1~"), KeyEvent{Key: "home"}},
		{"tilde 4", []byte("\x1b[4~"), KeyEvent{Key: "end"}},
		{"tilde 7", []byte("\x1b[7~"), KeyEvent{Key: "home"}},
		{"tilde 8", []byte("\x1b[8~"), KeyEvent{Key: "end"}},
		{"shift-home H", []byte("\x1b[1;2H"), KeyEvent{Key: "shift-home"}},
		{"shift-end F", []byte("\x1b[1;2F"), KeyEvent{Key: "shift-end"}},
		{"ctrl-home H", []byte("\x1b[1;5H"), KeyEvent{Key: "ctrl-home"}},
		{"ctrl-end F", []byte("\x1b[1;5F"), KeyEvent{Key: "ctrl-end"}},
		{"alt-home H", []byte("\x1b[1;3H"), KeyEvent{Key: "alt-home"}},
		{"ctrl-shift-end F", []byte("\x1b[1;6F"), KeyEvent{Key: "ctrl-shift-end"}},
		{"shift-home tilde 1", []byte("\x1b[1;2~"), KeyEvent{Key: "shift-home"}},
		{"shift-end tilde 4", []byte("\x1b[4;2~"), KeyEvent{Key: "shift-end"}},
		{"ctrl-home tilde 7", []byte("\x1b[7;5~"), KeyEvent{Key: "ctrl-home"}},
		{"shift-end tilde 8", []byte("\x1b[8;2~"), KeyEvent{Key: "shift-end"}},
		{"csi-u ctrl-shift-a", []byte("\x1b[97;6u"), KeyEvent{Key: "ctrl-shift-a"}},
		{"csi-u ctrl-shift-e", []byte("\x1b[101;6u"), KeyEvent{Key: "ctrl-shift-e"}},
		{"legacy ctrl-a", []byte{1}, KeyEvent{Key: "ctrl-a"}},
		{"legacy ctrl-e", []byte{5}, KeyEvent{Key: "ctrl-e"}},
		{"legacy ctrl-z", []byte{26}, KeyEvent{Key: "ctrl-z"}},
		{"printable t", []byte{'t'}, KeyEvent{Text: "t"}},
		{"printable ?", []byte{'?'}, KeyEvent{Text: "?"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeKey(tt.input)
			if got != tt.expected {
				t.Errorf("DecodeKey(%v) = %+v, want %+v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDecodePaste(t *testing.T) {
	got, used, ok := DecodePaste([]byte("\x1b[200~one\ntwo\x1b[201~x"))
	if !ok || used != len("\x1b[200~one\ntwo\x1b[201~") || got.Text != "one\ntwo" {
		t.Fatalf("DecodePaste() = %+v, %d, %t", got, used, ok)
	}
	if _, _, ok := DecodePaste([]byte("\x1b[200~unfinished")); ok {
		t.Fatal("DecodePaste accepted an unterminated paste")
	}
}

func TestKeyEventName(t *testing.T) {
	tests := []struct {
		name  string
		event KeyEvent
		want  string
	}{
		{"up", KeyEvent{Key: "up"}, "up"},
		{"pgup", KeyEvent{Key: "pgup"}, "pgup"},
		{"esc", KeyEvent{Key: "esc"}, "esc"},
		{"j", KeyEvent{Text: "j"}, "j"},
		{"q", KeyEvent{Text: "q"}, "q"},
		{"question mark", KeyEvent{Text: "?"}, "?"},
		{"key takes precedence", KeyEvent{Key: "up", Text: "j"}, "up"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.Name(); got != tt.want {
				t.Errorf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeyEventIs(t *testing.T) {
	if !((KeyEvent{Key: "up"}).Is("down", "up", "left")) {
		t.Error("Is() = false, want true for a matching candidate")
	}
	if (KeyEvent{Text: "j"}).Is("q", "?") {
		t.Error("Is() = true, want false when no candidate matches")
	}
}

func TestKeyMap(t *testing.T) {
	km := NewKeyMapWithLabels(map[string][]string{
		"move-up": {"up", "w", "k"},
		"quit":    {"q", "ctrl-c", "esc"},
	}, map[string]string{"move-up": "↑", "quit": "Quit"})

	tests := []struct {
		name   string
		event  KeyEvent
		action string
	}{
		{"single special key", KeyEvent{Key: "up"}, "move-up"},
		{"alias lowercase", KeyEvent{Text: "k"}, "move-up"},
		{"alias uppercase", KeyEvent{Text: "W"}, "move-up"},
		{"modifier", KeyEvent{Key: "CTRL-C"}, "quit"},
		{"special alias", KeyEvent{Key: "ESC"}, "quit"},
		{"unmatched", KeyEvent{Text: "x"}, ""},
		{"empty event", KeyEvent{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := km.Action(tt.event); got != tt.action {
				t.Errorf("Action(%+v) = %q, want %q", tt.event, got, tt.action)
			}
			if got := km.Matches(tt.event, tt.action); got != (tt.action != "") {
				t.Errorf("Matches(%+v, %q) = %t", tt.event, tt.action, got)
			}
		})
	}
	if got := km.Label("move-up"); got != "↑" {
		t.Errorf("Label(move-up) = %q, want %q", got, "↑")
	}
	if got := km.Label("unknown"); got != "" {
		t.Errorf("Label(unknown) = %q, want empty", got)
	}

	// Case folding is limited to ASCII; visually similar non-ASCII letters stay distinct.
	unicodeMap := NewKeyMap(map[string][]string{"unicode": {"é"}})
	if got := unicodeMap.Action(KeyEvent{Text: "É"}); got != "" {
		t.Errorf("Action(non-ASCII case variant) = %q, want empty", got)
	}
}

func TestKeyEventRune(t *testing.T) {
	tests := []struct {
		name  string
		event KeyEvent
		want  rune
	}{
		{"printable", KeyEvent{Text: "j"}, 'j'},
		{"unicode printable", KeyEvent{Text: "🙂"}, '🙂'},
		{"empty", KeyEvent{}, 0},
		{"non-printable", KeyEvent{Text: "\n"}, 0},
		{"special key", KeyEvent{Key: "up"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.Rune(); got != tt.want {
				t.Errorf("Rune() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEventResultConstructors(t *testing.T) {
	if h := Handled(); !h.Consumed || h.Quit {
		t.Errorf("Handled() = %+v, want Consumed:true, Quit:false", h)
	}
	if c := Consumed(); !c.Consumed || c.Quit {
		t.Errorf("Consumed() = %+v, want Consumed:true, Quit:false", c)
	}
	if ig := Ignored(); ig.Consumed || ig.Quit {
		t.Errorf("Ignored() = %+v, want Consumed:false, Quit:false", ig)
	}
	if un := Unhandled(); un.Consumed || un.Quit {
		t.Errorf("Unhandled() = %+v, want Consumed:false, Quit:false", un)
	}
	if q := Quit(); !q.Consumed || !q.Quit {
		t.Errorf("Quit() = %+v, want Consumed:true, Quit:true", q)
	}
	if qr := QuitResult(); !qr.Consumed || !qr.Quit {
		t.Errorf("QuitResult() = %+v, want Consumed:true, Quit:true", qr)
	}
}

type dummyEventConsumer struct {
	onKey func(KeyEvent) EventResult
}

func (d *dummyEventConsumer) Draw(*Canvas, Rect)                  {}
func (d *dummyEventConsumer) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (d *dummyEventConsumer) ConsumeKey(e KeyEvent) EventResult {
	if d.onKey != nil {
		return d.onKey(e)
	}
	return Ignored()
}

type dummyMouseConsumer struct {
	onMouse func(MouseEvent) EventResult
}

func (d *dummyMouseConsumer) Draw(*Canvas, Rect)              {}
func (d *dummyMouseConsumer) ConsumeKey(KeyEvent) EventResult { return Ignored() }
func (d *dummyMouseConsumer) ConsumeMouse(e MouseEvent) EventResult {
	if d.onMouse != nil {
		return d.onMouse(e)
	}
	return Ignored()
}

func TestDispatchKeyAndMouseEvent(t *testing.T) {
	// 1. Nil root returns Ignored
	if res := DispatchKeyEvent(nil, KeyEvent{Key: "up"}); res != Ignored() {
		t.Errorf("DispatchKeyEvent(nil) = %+v, want Ignored", res)
	}
	if res := DispatchMouseEvent(nil, MouseEvent{}); res != Ignored() {
		t.Errorf("DispatchMouseEvent(nil) = %+v, want Ignored", res)
	}

	// 2. EventConsumer takes precedence
	ec := &dummyEventConsumer{
		onKey: func(e KeyEvent) EventResult {
			if e.Key == "up" {
				return Handled()
			}
			if e.Key == "f10" {
				return QuitResult()
			}
			return Ignored()
		},
	}
	if res := DispatchKeyEvent(ec, KeyEvent{Key: "up"}); res != Handled() {
		t.Errorf("DispatchKeyEvent(ec, up) = %+v, want Handled", res)
	}
	if res := DispatchKeyEvent(ec, KeyEvent{Key: "f10"}); res != QuitResult() {
		t.Errorf("DispatchKeyEvent(ec, f10) = %+v, want QuitResult", res)
	}
	if res := DispatchKeyEvent(ec, KeyEvent{Key: "x"}); res != Ignored() {
		t.Errorf("DispatchKeyEvent(ec, x) = %+v, want Ignored", res)
	}

	// 3. MouseConsumer takes precedence
	mc := &dummyMouseConsumer{
		onMouse: func(e MouseEvent) EventResult {
			if e.Action == MousePress {
				return Handled()
			}
			return Ignored()
		},
	}
	if res := DispatchMouseEvent(mc, MouseEvent{Action: MousePress}); res != Handled() {
		t.Errorf("DispatchMouseEvent(mc, press) = %+v, want Handled", res)
	}
	if res := DispatchMouseEvent(mc, MouseEvent{Action: MouseRelease}); res != Ignored() {
		t.Errorf("DispatchMouseEvent(mc, release) = %+v, want Ignored", res)
	}
}
