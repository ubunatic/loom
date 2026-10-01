// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testAnimatedBackground struct{ interval time.Duration }

func (testAnimatedBackground) DrawBackground(*Canvas, Rect)                {}
func (b testAnimatedBackground) DrawBackgroundAt(*Canvas, Rect, time.Time) {}
func (b testAnimatedBackground) BackgroundInterval() time.Duration         { return b.interval }

func TestBackgroundTickerLifecycle(t *testing.T) {
	background := testAnimatedBackground{interval: 20 * time.Millisecond}
	var clock backgroundTicker
	if !clock.reconcile(background, false, false) || clock.ticker == nil {
		t.Fatal("always mode did not start the animated background ticker")
	}
	first := clock.ticker
	if clock.reconcile(background, false, false) || clock.ticker != first {
		t.Fatal("unchanged ticker configuration replaced the timer")
	}
	select {
	case <-clock.frames:
	case <-time.After(time.Second):
		t.Fatal("always mode ticker did not fire")
	}
	if !clock.reconcile(background, false, true) || clock.ticker != nil || clock.frames != nil {
		t.Fatal("on-redraw mode did not stop the background ticker")
	}
	if !clock.reconcile(background, false, false) || clock.ticker == nil {
		t.Fatal("returning to always mode did not restart the background ticker")
	}
	if !clock.reconcile(background, true, false) || clock.ticker != nil {
		t.Fatal("ReduceMotion did not suppress the background ticker")
	}
	if !clock.reconcile(background, false, false) || clock.ticker == nil {
		t.Fatal("disabling ReduceMotion did not restart the background ticker")
	}
	if !clock.reconcile(nil, false, false) || clock.ticker != nil || clock.frames != nil {
		t.Fatal("switching to a plain background did not stop the ticker")
	}
	clock.stop()
}

type paneSelectionWidget struct {
	selection    int
	consumeCalls int
}

func (*paneSelectionWidget) Draw(*Canvas, Rect)                  {}
func (*paneSelectionWidget) ConsumeMouse(MouseEvent) EventResult { return Ignored() }

func (w *paneSelectionWidget) ConsumeKey(e KeyEvent) EventResult {
	w.consumeCalls++
	if e.Key == "down" {
		w.selection++
		return Handled()
	}
	return Ignored()
}

func TestPaneDispatchKeyConsumesSelectionKeyOnce(t *testing.T) {
	p := &Pane{}
	w := &paneSelectionWidget{}

	if p.dispatchKey(w, KeyEvent{Key: "down"}) {
		t.Fatal("down key unexpectedly requested quit")
	}
	if w.selection != 1 {
		t.Fatalf("selection after one down key = %d, want 1", w.selection)
	}
	if w.consumeCalls != 1 {
		t.Fatalf("ConsumeKey calls = %d, want 1", w.consumeCalls)
	}
}

func TestPaneClickTrackingIsDisabledOnTeardown(t *testing.T) {
	out, err := os.Create(filepath.Join(t.TempDir(), "mouse-sequences"))
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	p := &Pane{tty: out}
	p.EnableMouseClicks()
	p.disableMouse()
	p.disableMouse()
	if p.mouse || p.mouseMode != 0 {
		t.Fatal("mouse tracking remained enabled")
	}
	if _, err := out.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(out)
	if err != nil {
		t.Fatal(err)
	}
	want := "\x1b[?1000h\x1b[?1006h\x1b[?1000l\x1b[?1006l"
	if string(got) != want {
		t.Fatalf("mouse sequences = %q, want %q", got, want)
	}
}

func TestPaneOwnershipRejectsNestedPane(t *testing.T) {
	if !claimPaneOwnership() {
		t.Fatal("initial pane ownership claim failed")
	}
	defer releasePaneOwnership()
	if claimPaneOwnership() {
		t.Fatal("nested pane ownership claim succeeded")
	}
}

func TestPaneOwnershipCanBeReclaimedAfterRelease(t *testing.T) {
	if !claimPaneOwnership() {
		t.Fatal("pane ownership claim failed")
	}
	releasePaneOwnership()
	if !claimPaneOwnership() {
		t.Fatal("pane ownership was not released")
	}
	releasePaneOwnership()
}

// TestWinchBounds covers the pane-placement math used when the terminal window
// is resized: height is clamped to leave the prompt line, and the top row is
// lifted (never below 1) when the pane would overflow the new bottom.
func TestWinchBounds(t *testing.T) {
	cases := []struct {
		name                     string
		startRow, rows, termRows int
		wantStartRow, wantRows   int
	}{
		{"fits unchanged", 5, 8, 24, 5, 8},
		{"height clamped then lifted to fit", 5, 30, 10, 2, 9},
		{"overflow lifts top row", 20, 6, 22, 17, 6},
		{"shrink clamps and lifts", 5, 8, 6, 2, 5},
		{"overflow with clamp lifts to fit", 3, 10, 5, 2, 4},
		{"single-row floor", 1, 2, 2, 1, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotStart, gotRows := winchBounds(c.startRow, c.rows, c.termRows)
			if gotStart != c.wantStartRow || gotRows != c.wantRows {
				t.Fatalf("winchBounds(%d,%d,%d) = (%d,%d), want (%d,%d)",
					c.startRow, c.rows, c.termRows, gotStart, gotRows, c.wantStartRow, c.wantRows)
			}
			// Invariants: stays on screen, at least one row, top row >= 1.
			if gotStart < 1 {
				t.Errorf("startRow %d < 1", gotStart)
			}
			if gotRows < 1 {
				t.Errorf("rows %d < 1", gotRows)
			}
			if gotStart+gotRows-1 > c.termRows {
				t.Errorf("bottom %d overflows termRows %d", gotStart+gotRows-1, c.termRows)
			}
		})
	}
}

func TestPaneConsumeKeyFallback(t *testing.T) {
	p := &Pane{}

	// Default fallback exits on Esc, Ctrl-C, Ctrl-Q, Ctrl-D, and 'q'
	exitKeys := []KeyEvent{
		{Key: "esc"},
		{Key: "ctrl-c"},
		{Key: "ctrl-q"},
		{Key: "ctrl-d"},
		{Text: "q"},
	}

	for _, ke := range exitKeys {
		if !p.handleKeyFallback(ke) {
			t.Errorf("expected handleKeyFallback(%+v) = true", ke)
		}
	}

	// Non-exit keys do not quit
	nonExitKeys := []KeyEvent{
		{Key: "enter"},
		{Key: "up"},
		{Key: "down"},
		{Text: "a"},
		{Text: "x"},
	}

	for _, ke := range nonExitKeys {
		if p.handleKeyFallback(ke) {
			t.Errorf("expected handleKeyFallback(%+v) = false", ke)
		}
	}

	// When DisableDefaultQuit is set, no fallback exit occurs
	p.DisableDefaultQuit = true
	for _, ke := range exitKeys {
		if p.handleKeyFallback(ke) {
			t.Errorf("expected handleKeyFallback(%+v) = false when DisableDefaultQuit is true", ke)
		}
	}
}

func TestPaneHelpOverlayUsesRootCanvasAndCapturesInput(t *testing.T) {
	p := &Pane{}
	previous := paneHelpRequest
	defer func() { paneHelpRequest = previous }()
	paneHelpRequest = func(cmds []Cmd) {
		p.help = NewPopup("Help", newHelpWidget(cmds))
	}

	bar := newCmdBar()
	bar.active = true
	bar.query = "help"
	if bar.execute(bar.match()) != cmdNone {
		t.Fatal("help command unexpectedly changed navigation")
	}
	if p.help == nil {
		t.Fatal("help did not request the pane overlay")
	}

	canvas := NewCanvas(40, 12)
	canvas.Clear()
	p.help.Draw(canvas, canvas.Bounds())
	if got := canvas.Get(10, 3).Text; got != "┌" {
		t.Fatalf("overlay left border at %d,3 = %q, want popup corner", 10, got)
	}
	quit, handled := p.handleHelpKey(KeyEvent{Text: "x"})
	if quit {
		t.Fatal("help dismissal unexpectedly requested application quit")
	}
	if !handled {
		t.Fatal("help key was not captured")
	}
	if p.help != nil {
		t.Fatal("help overlay remained open after dismissal key")
	}
}

type paneEventConsumerProbe struct {
	handledKeys []KeyEvent
	quitKey     string
}

func (*paneEventConsumerProbe) Draw(*Canvas, Rect)                  {}
func (*paneEventConsumerProbe) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (p *paneEventConsumerProbe) ConsumeKey(e KeyEvent) EventResult {
	if e.Key == p.quitKey {
		return QuitResult()
	}
	p.handledKeys = append(p.handledKeys, e)
	return Handled()
}

func TestPaneDispatchKeyEventConsumer(t *testing.T) {
	p := &Pane{}
	probe := &paneEventConsumerProbe{quitKey: "f10"}

	// Arrow keys navigate and do not quit
	navKeys := []KeyEvent{
		{Key: "up"},
		{Key: "down"},
		{Key: "left"},
		{Key: "right"},
		{Key: "home"},
		{Key: "end"},
	}

	for _, k := range navKeys {
		if p.dispatchKey(probe, k) {
			t.Fatalf("dispatchKey(%+v) requested quit, want false", k)
		}
	}

	if len(probe.handledKeys) != len(navKeys) {
		t.Fatalf("handledKeys len = %d, want %d", len(probe.handledKeys), len(navKeys))
	}

	// F10 requests quit
	if !p.dispatchKey(probe, KeyEvent{Key: "f10"}) {
		t.Fatal("dispatchKey(f10) = false, want true")
	}
}

type f10SwallowProbe struct{ calls int }

func (*f10SwallowProbe) Draw(*Canvas, Rect)                  {}
func (*f10SwallowProbe) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (p *f10SwallowProbe) ConsumeKey(KeyEvent) EventResult {
	p.calls++
	return Handled()
}

func TestPaneGlobalF10PrecedesFocusedWidgetDispatch(t *testing.T) {
	for _, tc := range []struct {
		name string
		root func(*f10SwallowProbe) Widget
	}{
		{name: "root consumer", root: func(p *f10SwallowProbe) Widget { return p }},
		{name: "focused frame child", root: func(p *f10SwallowProbe) Widget {
			frame := &Frame{Boxes: []Box{{ID: "input", Child: p}}}
			frame.focusFirst()
			return frame
		}},
		{name: "active tab child", root: func(p *f10SwallowProbe) Widget {
			return NewTabs(Tab{Title: "Input", Widget: p})
		}},
		{name: "nested tab and frame child", root: func(p *f10SwallowProbe) Widget {
			frame := &Frame{Boxes: []Box{{ID: "input", Child: p}}}
			frame.focusFirst()
			return NewTabs(Tab{Title: "Editor", Widget: frame})
		}},
		{name: "popup child", root: func(p *f10SwallowProbe) Widget {
			return &Popup{Open: true, Inner: p}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &f10SwallowProbe{}
			pane := &Pane{}
			if !pane.dispatchKey(tc.root(p), KeyEvent{Key: "f10"}) {
				t.Fatal("F10 did not request global quit")
			}
			if p.calls != 0 {
				t.Fatalf("focused child saw F10 %d times, want 0", p.calls)
			}
		})
	}
}

func TestPaneGlobalF10OptOut(t *testing.T) {
	probe := &paneEventConsumerProbe{quitKey: "never"}
	pane := &Pane{DisableGlobalF10Quit: true}
	if pane.dispatchKey(probe, KeyEvent{Text: "F10"}) {
		t.Fatal("opted-out F10 requested global quit")
	}
	if len(probe.handledKeys) != 1 {
		t.Fatalf("opted-out child handled %d keys, want 1", len(probe.handledKeys))
	}
}

func TestPaneGlobalF10BypassesHelpOverlay(t *testing.T) {
	probe := &f10SwallowProbe{}
	pane := &Pane{help: NewPopup("Help", probe)}
	quit, handled := pane.handleHelpKey(KeyEvent{Key: "f10"})
	if !quit || !handled {
		t.Fatalf("handleHelpKey(f10) = quit:%v handled:%v, want true,true", quit, handled)
	}
	if probe.calls != 0 {
		t.Fatalf("help child saw F10 %d times, want 0", probe.calls)
	}
}

func TestPaneNavigationFallbackNeverQuits(t *testing.T) {
	p := &Pane{}
	// Passive widget returning false on everything
	w := &focusProbe{}

	navKeys := []KeyEvent{
		{Key: "up"},
		{Key: "down"},
		{Key: "left"},
		{Key: "right"},
		{Key: "home"},
		{Key: "end"},
		{Key: "pgup"},
		{Key: "pgdn"},
		{Key: "delete"},
		{Key: "insert"},
		{Key: "backspace"},
		{Key: "tab"},
		{Key: "shift-tab"},
		{Key: "ctrl-left"},
		{Key: "ctrl-right"},
	}

	for _, k := range navKeys {
		if p.dispatchKey(w, k) {
			t.Fatalf("dispatchKey(%+v) on passive widget requested quit", k)
		}
	}
}
