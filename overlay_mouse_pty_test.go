package loom

import (
	"os"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

type overlayMousePTYWidget struct{ pane *Pane }

func (*overlayMousePTYWidget) Draw(*Canvas, Rect) {}
func (*overlayMousePTYWidget) ConsumeMouse(MouseEvent) EventResult {
	return Ignored()
}
func (w *overlayMousePTYWidget) ConsumeKey(e KeyEvent) EventResult {
	if e.Is("f1") {
		w.pane.help = NewPopup("Pane mouse help", NewView(nil))
		return Handled()
	}
	if e.Is("q") {
		return EventResult{Quit: true, Consumed: true}
	}
	return Ignored()
}

func TestOverlayMousePaneHelpPTY(t *testing.T) {
	t.Setenv("LOOM_OVERLAY_MOUSE_HELPER", "1")
	s := ptytest.Start(t, 80, 24, os.Args[0], "-test.run=^TestOverlayMousePaneHelpPTYHelper$")
	wait := func(predicate func() bool) {
		t.Helper()
		until := time.Now().Add(3 * time.Second)
		for time.Now().Before(until) {
			if predicate() {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("mouse lifecycle did not arrive: %q", s.Raw())
	}
	disable := "\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l"
	wait(func() bool { return strings.Contains(string(s.Raw()), disable) })
	s.Send("\x1bOP")
	s.WaitFor("Pane mouse help", 3*time.Second)
	if !strings.Contains(string(s.Raw()), "\x1b[?1000h\x1b[?1006h") {
		t.Fatal("pane help did not enable temporary mouse tracking")
	}
	from := len(s.Raw())
	s.Send("\x1b[<0;1;24M\x1b[<0;1;24m")
	wait(func() bool { return strings.Contains(string(s.Raw()[from:]), disable) })
	wait(func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "Pane mouse help") })
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestOverlayMousePaneHelpPTYHelper(t *testing.T) {
	if os.Getenv("LOOM_OVERLAY_MOUSE_HELPER") == "" {
		return
	}
	p, err := New(24)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	p.DisableDefaultQuit = true
	p.SetScreenMode(ScreenAlt)
	p.DisableMouse()
	if err := p.Run(&overlayMousePTYWidget{pane: p}); err != nil {
		t.Fatal(err)
	}
}
