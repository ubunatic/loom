// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
	"time"
)

// TestPaneKittyKeyboardPushPop runs an opted-in pane on a PTY: the flag is
// pushed on start, popped before a suspended child runs, pushed again on
// resume and popped on close; CSI-u keys reach the widget by name.
func TestPaneKittyKeyboardPushPop(t *testing.T) {
	t.Setenv("LOOM_ZWJ", "join")
	master, slave, cooked := openPTYWithCookedState(t)
	setPTYSize(t, master, 40, 10)
	capture := capturePTY(master, "")

	p := &Pane{tty: slave, fd: int(slave.Fd()), oldState: cooked, rows: 3, wantRows: 3,
		startRow: 2, cols: 40, InlineOnly: true, KittyKeyboard: true}
	w := newSuspendWidget()
	resumed := make(chan struct{})
	w.onSuspend = func() {
		defer close(resumed)
		if err := p.Suspend(func() error {
			_, _ = slave.WriteString(suspendMarker)
			return nil
		}); err != nil {
			t.Errorf("Suspend: %v", err)
		}
	}
	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()
	w.waitDraw(t, 0)
	if _, err := master.Write([]byte("\x1b[13;2u\x1b[13;5u\x1b[105;5u\t\r")); err != nil {
		t.Fatal(err)
	}
	if _, err := master.Write([]byte{0x07}); err != nil { // ctrl-g
		t.Fatal(err)
	}
	select {
	case <-resumed:
	case <-time.After(3 * time.Second):
		t.Fatal("Suspend did not return")
	}
	if _, err := master.Write([]byte("q")); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Run did not return; keys=%v", w.keyNames())
	}
	p.Close()
	before, after := capture.split(t)

	if got, want := strings.Join(w.keyNames(), ","), "shift-enter,ctrl-enter,ctrl-i,tab,enter,ctrl-g,q"; got != want {
		t.Errorf("keys = %s, want %s", got, want)
	}
	if strings.Count(before, kittyKeyboardPush) != 1 || strings.Count(before, kittyKeyboardPop) != 1 ||
		strings.Index(before, kittyKeyboardPush) > strings.Index(before, kittyKeyboardPop) {
		t.Errorf("before child: want one push then one pop, got %q", before)
	}
	if strings.Count(after, kittyKeyboardPush) != 1 || strings.Count(after, kittyKeyboardPop) != 1 ||
		strings.Index(after, kittyKeyboardPush) > strings.Index(after, kittyKeyboardPop) {
		t.Errorf("after child: want push on resume and pop on close, got %q", after)
	}
}

// panicWidget panics on any key.
type panicWidget struct{ drawn chan struct{} }

func (w *panicWidget) Draw(*Canvas, Rect) {
	select {
	case w.drawn <- struct{}{}:
	default:
	}
}
func (w *panicWidget) ConsumeKey(KeyEvent) EventResult     { panic("boom") }
func (w *panicWidget) ConsumeMouse(MouseEvent) EventResult { return Ignored() }

// TestPaneKittyKeyboardPanicPops checks the panic exit restores the keyboard
// mode, and that a pane without the opt-in never pushes it.
func TestPaneKittyKeyboardPanicPops(t *testing.T) {
	for _, opt := range []bool{true, false} {
		t.Setenv("LOOM_ZWJ", "join")
		master, slave, cooked := openPTYWithCookedState(t)
		setPTYSize(t, master, 40, 10)
		capture := capturePTY(master, "")
		p := &Pane{tty: slave, fd: int(slave.Fd()), oldState: cooked, rows: 2, wantRows: 2,
			startRow: 2, cols: 40, InlineOnly: true, KittyKeyboard: opt}
		w := &panicWidget{drawn: make(chan struct{}, 1)}
		recovered := make(chan any, 1)
		go func() {
			defer func() { recovered <- recover() }()
			_ = p.Run(w)
		}()
		select {
		case <-w.drawn:
		case <-time.After(3 * time.Second):
			t.Fatal("no frame")
		}
		_, _ = slave.WriteString(suspendMarker)
		if _, err := master.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
		select {
		case r := <-recovered:
			if r != "boom" {
				t.Fatalf("recovered %v, want boom", r)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("Run did not panic")
		}
		before, after := capture.split(t)
		if !opt {
			if strings.Contains(before+after, kittyKeyboardPush) || strings.Contains(before+after, kittyKeyboardPop) {
				t.Errorf("pane without KittyKeyboard wrote the protocol: %q", before+after)
			}
			continue
		}
		if !strings.Contains(before, kittyKeyboardPush) || !strings.Contains(after, kittyKeyboardPop) {
			t.Errorf("want push at start and pop on panic exit; before=%q after=%q", before, after)
		}
	}
}

// TestPaneKittyKeyboardAltScreenBalanced keeps one push per screen: each
// screen has its own kitty stack, so a switch pops on the old screen and
// pushes on the new one.
func TestPaneKittyKeyboardAltScreenBalanced(t *testing.T) {
	master, slave, _ := openPTYWithCookedState(t)
	capture := capturePTY(master, "")
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 2, startRow: 1, cols: 20}
	p.setKittyKeyboard(true)
	p.switchAltScreen(true)
	_, _ = slave.WriteString(suspendMarker)
	p.switchAltScreen(false)
	p.setKittyKeyboard(false)
	time.Sleep(50 * time.Millisecond)
	before, after := capture.split(t)
	enter := strings.SplitN(before, "\x1b[?1049h", 2)
	if len(enter) != 2 || strings.Count(enter[0], kittyKeyboardPush) != 1 || strings.Count(enter[0], kittyKeyboardPop) != 1 ||
		strings.Count(enter[1], kittyKeyboardPush) != 1 {
		t.Errorf("enter alt: want push, pop before ?1049h and push after: %q", before)
	}
	leave := strings.SplitN(after, "\x1b[?1049l", 2)
	if len(leave) != 2 || !strings.HasSuffix(leave[0], kittyKeyboardPop) ||
		strings.Count(leave[1], kittyKeyboardPush) != 1 || !strings.HasSuffix(leave[1], kittyKeyboardPop) {
		t.Errorf("leave alt: want pop, ?1049l, push, final pop: %q", after)
	}
}
