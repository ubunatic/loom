// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

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
		{"home", []byte{27, '[', '1', '~'}, KeyEvent{Key: "home"}},
		{"delete", []byte{27, '[', '3', '~'}, KeyEvent{Key: "delete"}},
		{"end", []byte{27, '[', '4', '~'}, KeyEvent{Key: "end"}},
		{"pgup", []byte{27, '[', '5', '~'}, KeyEvent{Key: "pgup"}},
		{"pgdown", []byte{27, '[', '6', '~'}, KeyEvent{Key: "pgdown"}},
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

func (d *dummyEventConsumer) Draw(*Canvas, Rect)          {}
func (d *dummyEventConsumer) HandleKey(KeyEvent) bool     { return false }
func (d *dummyEventConsumer) HandleMouse(MouseEvent) bool { return false }
func (d *dummyEventConsumer) ConsumeKey(e KeyEvent) EventResult {
	if d.onKey != nil {
		return d.onKey(e)
	}
	return Ignored()
}

type dummyMouseConsumer struct {
	onMouse func(MouseEvent) EventResult
}

func (d *dummyMouseConsumer) Draw(*Canvas, Rect)          {}
func (d *dummyMouseConsumer) HandleKey(KeyEvent) bool     { return false }
func (d *dummyMouseConsumer) HandleMouse(MouseEvent) bool { return false }
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
