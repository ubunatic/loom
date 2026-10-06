// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
	"ubunatic.com/loom/internal/ptytest"
)

type firstDrawBoundsProbe struct {
	pane     *Pane
	cancel   context.CancelFunc
	rect     Rect
	startRow int
	draws    int
	winch    bool
}

func (w *firstDrawBoundsProbe) Draw(_ *Canvas, r Rect) {
	w.draws++
	w.rect, w.startRow = r, w.pane.startRow
	if w.winch && w.draws == 1 {
		// A ready resize notification competes with cancellation in select.
		w.pane.winch <- unix.SIGWINCH
	}
	w.cancel()
}
func (*firstDrawBoundsProbe) ConsumeKey(KeyEvent) EventResult     { return Ignored() }
func (*firstDrawBoundsProbe) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (*firstDrawBoundsProbe) PaneRequest() PaneRequest            { return PaneRequest{} }

func TestPaneStartupResize(t *testing.T) {
	for _, mode := range []string{"inline", "full", "alt"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("LOOM_STARTUP_RESIZE_HELPER", mode)
			s := ptytest.Start(t, 80, 24, os.Args[0], "-test.run=^TestPaneStartupResizeHelper$")
			if err := s.Wait(5 * time.Second); err != nil {
				t.Fatalf("startup resize: %v\n%s", err, s.Raw())
			}
		})
	}
}

func TestPaneStartupResizeHelper(t *testing.T) {
	mode := os.Getenv("LOOM_STARTUP_RESIZE_HELPER")
	if mode == "" {
		return
	}
	// Ensure the pre-registration signal is discarded even if Go's signal
	// goroutine would otherwise deliver it after New installs Notify.
	signal.Ignore(unix.SIGWINCH)
	paneBeforeSignalHandler = func(p *Pane) {
		if p.cols != 80 || p.winch != nil {
			t.Fatal("hook must run after the initial size read and before signal registration")
		}
		// The controlling PTY sends SIGWINCH here, before Notify is installed.
		set := &unix.Winsize{Col: 100, Row: 30}
		if err := unix.IoctlSetWinsize(p.fd, unix.TIOCSWINSZ, set); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { paneBeforeSignalHandler = nil }()
	p, err := New(8)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	p.SetScreenMode(ScreenInline)
	p.ResizeConfig.AutoFullscreen = false
	wantStart, wantHeight := 17, 8
	switch mode {
	case "full":
		p.SetScreenMode(ScreenFull)
		wantStart, wantHeight = 1, 29
	case "alt":
		p.SetScreenMode(ScreenAlt)
		wantStart, wantHeight = 1, 30
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	probe := &firstDrawBoundsProbe{pane: p, cancel: cancel}
	if err := p.run(ctx, probe, nil, nil, nil); err != context.Canceled {
		t.Fatalf("run = %v", err)
	}
	if probe.startRow != wantStart || probe.rect != (Rect{W: 100, H: wantHeight}) {
		t.Fatalf("first draw row %d, %+v; want row %d, 100x%d", probe.startRow, probe.rect, wantStart, wantHeight)
	}
}

func TestPaneStartupRefreshInlineBounds(t *testing.T) {
	for _, tc := range []struct {
		name                           string
		cols, termRows, rows, wantRows int
		start, wantStart, wantHeight   int
	}{
		{"unchanged", 80, 24, 8, 8, 17, 17, 8},
		{"grow", 100, 30, 23, 26, 2, 2, 26},
		{"shrink", 60, 12, 8, 8, 17, 5, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			master, slave := openPTY(t)
			if err := unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{
				Col: uint16(tc.cols), Row: uint16(tc.termRows),
			}); err != nil {
				t.Fatal(err)
			}
			p := &Pane{tty: slave, fd: int(slave.Fd()), cols: 80, rows: tc.rows,
				wantRows: tc.wantRows, startRow: tc.start, inlineStart: tc.start,
				ResizeConfig: DefaultResizeConfig()}
			p.ResizeConfig.OutOfBandClear = true
			p.refreshStartupSize()
			if p.cols != tc.cols || p.rows != tc.wantHeight || p.startRow != tc.wantStart || p.inlineStart != tc.start {
				t.Fatalf("startup bounds = %dx%d at %d, inlineStart %d", p.cols, p.rows, p.startRow, p.inlineStart)
			}
			poll := []unix.PollFd{{Fd: int32(master.Fd()), Events: unix.POLLIN}}
			if n, err := unix.Poll(poll, 0); err != nil || n != 0 {
				t.Fatalf("startup refresh wrote terminal output: poll = %d, %v", n, err)
			}
		})
	}
}

func TestPaneFirstDrawUsesScreenBounds(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mode       ScreenMode
		auto       bool
		wantStart  int
		wantHeight int
	}{
		{"inline", ScreenInline, false, 7, 24},
		{"full", ScreenFull, false, 1, 29},
		{"alt", ScreenAlt, false, 1, 30},
		{"wrapped-request-auto-alt", ScreenInline, true, 1, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			master, slave := openPTY(t)
			setPTYSize(t, master, 100, 30)
			p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 24, wantRows: 24, cols: 100, startRow: 7, MaxCols: DefaultMaxCols, ResizeConfig: DefaultResizeConfig()}
			state, err := term.GetState(p.fd)
			if err != nil {
				t.Fatal(err)
			}
			p.oldState = state
			p.SetScreenMode(tc.mode)
			p.ResizeConfig.AutoFullscreen = tc.auto
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			probe := &firstDrawBoundsProbe{pane: p, cancel: cancel}
			done := make(chan struct{})
			drainPTY(master, done)
			defer close(done)
			defer p.Close()
			if err := p.run(ctx, &tickerWrapper{child: probe}, nil, nil, nil); err != context.Canceled {
				t.Fatalf("run = %v, want canceled after first draw", err)
			}
			if probe.startRow != tc.wantStart || probe.rect != (Rect{W: 100, H: tc.wantHeight}) {
				t.Fatalf("first draw at row %d with %+v, want row %d and 100x%d", probe.startRow, probe.rect, tc.wantStart, tc.wantHeight)
			}
			if p.altActive && (p.savedStart != 7 || p.savedRows != 24) {
				t.Fatalf("saved inline bounds = row %d, height %d, want row 7, height 24", p.savedStart, p.savedRows)
			}
		})
	}
}

func TestPaneCanceledDrawDoesNotReflow(t *testing.T) {
	master, slave := openPTY(t)
	if err := unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Col: 100, Row: 30}); err != nil {
		t.Fatal(err)
	}
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 24, wantRows: 24,
		cols: 100, startRow: 7, ResizeConfig: DefaultResizeConfig()}
	state, err := term.GetState(p.fd)
	if err != nil {
		t.Fatal(err)
	}
	p.oldState = state
	p.SetScreenMode(ScreenInline)
	p.ResizeConfig.AutoFullscreen = false
	// Register before drawing, with a larger queue so an external WINCH
	// cannot block the test's explicit notification in Draw.
	p.installSignalHandler()
	signal.Stop(p.winch)
	p.winch = make(chan os.Signal, 2)
	defer p.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	probe := &firstDrawBoundsProbe{pane: p, cancel: cancel, winch: true}
	done := make(chan struct{})
	drainPTY(master, done)
	defer close(done)
	if err := p.run(ctx, probe, nil, nil, nil); err != context.Canceled {
		t.Fatalf("run = %v", err)
	}
	if probe.draws != 1 || probe.rect.W != 100 {
		t.Fatalf("canceled first draw was overwritten: draws %d, bounds %+v", probe.draws, probe.rect)
	}
	// An already canceled context must also leave the last draw untouched.
	if err := p.run(ctx, probe, nil, nil, nil); err != context.Canceled || probe.draws != 1 {
		t.Fatalf("already canceled run = %v, draws = %d; want canceled without drawing", err, probe.draws)
	}
}

func TestPaneAlternateScreenModeRestoresPrimaryScreen(t *testing.T) {
	master, slave := openPTY(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 4, cols: 40, startRow: 3}
	state, err := term.GetState(p.fd)
	if err != nil {
		t.Fatal(err)
	}
	p.oldState = state
	p.switchAltScreen(true)
	p.Close()
	_ = master.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 4096)
	var output strings.Builder
	for !strings.Contains(output.String(), "\x1b[?1049l") {
		n, err := master.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		if err != nil {
			t.Fatalf("read terminal output before alternate-screen exit: %v", err)
		}
	}
	raw := output.String()
	enter := strings.Index(raw, "\x1b[?1049h")
	leave := strings.Index(raw, "\x1b[?1049l")
	if enter < 0 || leave < enter {
		t.Fatalf("alternate screen output = %q, want enter then leave sequences", raw)
	}
}

// openPTY opens a real Linux pseudo-terminal pair via /dev/ptmx, returning
// the master (test-controlled) and slave (what Pane.run reads/writes) ends.
// It skips the test rather than failing when ptys are unavailable (e.g. some
// sandboxes), matching this repo's "probe external mechanisms" canary-first
// convention (docs/Canary.md).
func openPTY(t *testing.T) (master, slave *os.File) {
	t.Helper()
	t.Setenv("LOOM_ZWJ", "join") // This test PTY has no terminal emulator to answer DSR.
	return openPTYForProbe(t)
}

func openPTYForProbe(t *testing.T) (master, slave *os.File) {
	t.Helper()
	m, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("open /dev/ptmx: %v", err)
	}
	fd := int(m.Fd())
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		m.Close()
		t.Skipf("unlock pty: %v", err)
	}
	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		m.Close()
		t.Skipf("get pty number: %v", err)
	}
	path := fmt.Sprintf("/dev/pts/%d", n)
	s, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		m.Close()
		t.Skipf("open %s: %v", path, err)
	}
	// A freshly opened pty slave starts in canonical (line-buffered, echoing)
	// mode, so a master Write without a trailing newline would sit unread
	// until one arrived — exactly what Pane.New's term.MakeRaw normally
	// avoids on the real /dev/tty. Put the slave in raw mode here so these
	// tests see bytes as soon as they are written, like the real pane does.
	if _, err := term.MakeRaw(int(s.Fd())); err != nil {
		s.Close()
		m.Close()
		t.Skipf("raw mode: %v", err)
	}
	t.Cleanup(func() {
		s.Close()
		m.Close()
	})
	return m, s
}

func drainPTY(master *os.File, done <-chan struct{}) {
	go func() {
		buf := make([]byte, 4096)
		for {
			select {
			case <-done:
				return
			default:
				_ = master.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
				_, err := master.Read(buf)
				if err != nil {
					time.Sleep(10 * time.Millisecond)
				}
			}
		}
	}()
}

// keyRecorder is a minimal Widget that records every KeyEvent it receives
// and quits once it has seen want of them, so tests can drive Pane.run
// against a real pty without a full widget tree.
type keyRecorder struct {
	mu   sync.Mutex
	keys []KeyEvent
	want int
	done chan struct{}
}

func newKeyRecorder(want int) *keyRecorder {
	return &keyRecorder{want: want, done: make(chan struct{})}
}

func (r *keyRecorder) Draw(*Canvas, Rect)                  {}
func (r *keyRecorder) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (r *keyRecorder) ConsumeKey(e KeyEvent) EventResult {
	r.mu.Lock()
	r.keys = append(r.keys, e)
	n := len(r.keys)
	r.mu.Unlock()
	if n >= r.want {
		select {
		case <-r.done:
		default:
			close(r.done)
		}
		return QuitResult()
	}
	return Ignored()
}

func (r *keyRecorder) snapshot() []KeyEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]KeyEvent, len(r.keys))
	copy(out, r.keys)
	return out
}

type cursorEffectsPTYWidget struct {
	hintUpdates chan CursorHint
	mouseEvents chan MouseEvent
	keyEvents   chan KeyEvent
}

func newCursorEffectsPTYWidget() *cursorEffectsPTYWidget {
	return &cursorEffectsPTYWidget{
		hintUpdates: make(chan CursorHint, 8),
		mouseEvents: make(chan MouseEvent, 8),
		keyEvents:   make(chan KeyEvent, 8),
	}
}

func (w *cursorEffectsPTYWidget) Draw(c *Canvas, r Rect) {
	base := Style{BG: ColorRGB(20, 40, 60)}
	c.PaintSurface(r, base)
	c.Set(1, 1, Cell{Text: "T", Style: Style{FG: ColorIndex(15), BG: base.BG}})
	if hint, ok := c.CursorHintAt(2, 1); ok {
		select {
		case w.hintUpdates <- hint:
		default:
		}
	}
}

func (w *cursorEffectsPTYWidget) ConsumeMouse(e MouseEvent) EventResult {
	w.mouseEvents <- e
	return Ignored()
}

func TestFramePTYMouseDragOutsidePane(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 24, 10)
	w := newCursorEffectsPTYWidget()
	frame := &Frame{Boxes: []Box{{ID: "content", Width: 24, Height: 8, Child: w}}}
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 10, cols: 24, startRow: 1, DisableDefaultQuit: true}
	p.EnableMouse()
	done := make(chan error, 1)
	go func() { done <- p.Run(frame) }()
	t.Cleanup(func() {
		_, _ = master.WriteString("q")
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("Pane.Run did not stop")
		}
	})

	outer := frame.Layout(24, 10)[0]
	inner := Rect{X: outer.X + 1, Y: outer.Y + 1, W: outer.W - 2, H: outer.H - 2}
	pressX, pressY := inner.X+1, inner.Y+1
	if _, err := master.WriteString(fmt.Sprintf("\x1b[<0;%d;%dM", pressX+1, pressY+2)); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-w.mouseEvents:
		if e.Action != MousePress {
			t.Fatalf("press delivered as %+v", e)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for press")
	}
	if _, err := master.WriteString(fmt.Sprintf("\x1b[<32;40;%dM", pressY+2)); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-w.mouseEvents:
		if e.Action != MouseDrag || e.X != 38 || e.Y != pressY-inner.Y+1 {
			t.Fatalf("outside drag delivered as %+v", e)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for outside drag")
	}
	if _, err := master.WriteString(fmt.Sprintf("\x1b[<0;40;%dm", pressY+3)); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-w.mouseEvents:
		if e.Action != MouseRelease || e.X != 38 || e.Y != pressY-inner.Y+2 {
			t.Fatalf("outside release delivered as %+v", e)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for outside release")
	}
}

func (w *cursorEffectsPTYWidget) ConsumeKey(e KeyEvent) EventResult {
	w.keyEvents <- e
	return EventResult{Consumed: e.Is("q"), Quit: e.Is("q")}
}

func TestCursorEffectsThroughPTY(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 12, 4)
	w := newCursorEffectsPTYWidget()
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 4, cols: 12, startRow: 1, DisableDefaultQuit: true}
	p.EnableCursorEffects()

	screenMu := sync.Mutex{}
	screen := ptytest.NewVT(12, 4)
	screenDone := make(chan struct{})
	screenStop := make(chan struct{})
	var stopScreen sync.Once
	stopScreenReader := func() { stopScreen.Do(func() { close(screenStop) }) }
	go func() {
		defer close(screenDone)
		buf := make([]byte, 4096)
		for {
			select {
			case <-screenStop:
				return
			default:
			}
			fds := []unix.PollFd{{Fd: int32(master.Fd()), Events: unix.POLLIN}}
			if _, err := unix.Poll(fds, 50); err != nil || fds[0].Revents&unix.POLLIN == 0 {
				continue
			}
			n, err := master.Read(buf)
			if n > 0 {
				screenMu.Lock()
				_, _ = screen.Write(buf[:n])
				screenMu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	errC := make(chan error, 1)
	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		errC <- p.Run(w)
	}()
	t.Cleanup(func() {
		stopScreenReader()
		_ = master.Close()
		select {
		case <-screenDone:
		case <-time.After(2 * time.Second):
			t.Error("PTY screen reader did not stop")
		}
		select {
		case <-runDone:
		case <-time.After(2 * time.Second):
			t.Error("Pane.Run did not stop")
		}
	})

	cellAt := func(x, y int) ptytest.Cell {
		screenMu.Lock()
		defer screenMu.Unlock()
		return screen.Cell(x, y)
	}
	waitCell := func(x, y int, match func(ptytest.Cell) bool, what string) ptytest.Cell {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			cell := cellAt(x, y)
			if match(cell) {
				return cell
			}
			time.Sleep(5 * time.Millisecond)
		}
		cell := cellAt(x, y)
		t.Fatalf("timed out waiting for %s at (%d,%d): %+v", what, x, y, cell)
		return cell
	}
	waitCell(1, 1, func(cell ptytest.Cell) bool { return cell.Rune == 'T' }, "initial widget frame")
	if len(w.hintUpdates) != 0 {
		t.Fatal("cursor hint appeared before motion")
	}

	if _, err := master.WriteString("\x1b[<35;3;2M"); err != nil {
		t.Fatalf("send motion: %v", err)
	}
	select {
	case event := <-w.mouseEvents:
		if event.Action != MouseHover || event.X != 2 || event.Y != 1 {
			t.Fatalf("motion delivered as %+v; want hover at (2,1)", event)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for motion event")
	}
	select {
	case hint := <-w.hintUpdates:
		if hint.DX != 0 || hint.DY != 0 {
			t.Fatalf("cursor cell hint = %+v; want zero delta", hint)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for draw-time cursor hint")
	}
	waitCell(1, 1, func(cell ptytest.Cell) bool {
		return cell.Style.BG == ptytest.ColorRGB(143, 153, 162)
	}, "brightened text background")

	if _, err := master.WriteString("\x1b[<0;4;2M"); err != nil {
		t.Fatalf("send click: %v", err)
	}
	select {
	case event := <-w.mouseEvents:
		if event.Action != MousePress || event.X != 3 || event.Y != 1 {
			t.Fatalf("click delivered as %+v; want press at (3,1)", event)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for click event")
	}
	waitCell(2, 1, func(cell ptytest.Cell) bool { return cell.Rune == '✦' }, "trailing star")
	waitCell(3, 1, func(cell ptytest.Cell) bool {
		return cell.Style.BG == ptytest.ColorRGB(255, 255, 255)
	}, "button pulse origin")

	time.Sleep(SpeccedCursorPressPulse.Lifetime + 30*time.Millisecond)
	waitCell(2, 1, func(cell ptytest.Cell) bool { return cell.Rune != '✦' }, "expired trail")
	waitCell(3, 1, func(cell ptytest.Cell) bool {
		return cell.Style.BG == ptytest.ColorRGB(185, 191, 197)
	}, "expired button pulse")

	if _, err := master.WriteString("k"); err != nil {
		t.Fatalf("send key: %v", err)
	}
	select {
	case key := <-w.keyEvents:
		if key.Text != "k" {
			t.Fatalf("key delivered as %+v; want text k", key)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for key event")
	}
	waitCell(3, 1, func(cell ptytest.Cell) bool {
		return cell.Style.BG == ptytest.ColorRGB(255, 255, 255)
	}, "key pulse origin")
	time.Sleep(SpeccedCursorPressPulse.Lifetime + 30*time.Millisecond)
	waitCell(3, 1, func(cell ptytest.Cell) bool {
		return cell.Style.BG == ptytest.ColorRGB(185, 191, 197)
	}, "expired key pulse")

	if _, err := master.WriteString("q"); err != nil {
		t.Fatalf("send quit: %v", err)
	}
	select {
	case key := <-w.keyEvents:
		if !key.Is("q") {
			t.Fatalf("quit key delivered as %+v", key)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for quit key")
	}
	select {
	case err := <-errC:
		if err != nil {
			t.Fatalf("Pane.Run: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Pane.Run did not exit after q")
	}
	stopScreenReader()
	select {
	case <-screenDone:
	case <-time.After(2 * time.Second):
		t.Fatal("PTY screen reader did not stop after q")
	}
}

// Complete key sequences written in one Write must be dispatched in order,
// including printable shortcuts such as Stopwatch's reset and resume.
func TestPaneRunDecodesMultipleKeysFromOneRead(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		want        []KeyEvent
	}{
		{"arrows", "\x1b[A\x1b[B", []KeyEvent{{Key: "up"}, {Key: "down"}}},
		{"reset-resume", "r ", []KeyEvent{{Text: "r"}, {Text: " "}}},
		{"unicode", "ä🙂", []KeyEvent{{Text: "ä"}, {Text: "🙂"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			master, slave := openPTY(t)
			p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, cols: 20}
			rec := newKeyRecorder(len(tc.want))

			errc := make(chan error, 1)
			go func() { errc <- p.Run(rec) }()

			if _, err := master.Write([]byte(tc.input)); err != nil {
				t.Fatalf("write: %v", err)
			}

			select {
			case <-rec.done:
			case <-time.After(3 * time.Second):
				t.Fatal("timed out waiting for both coalesced keys to be decoded")
			}
			if err := <-errc; err != nil {
				t.Fatalf("Pane.Run: %v", err)
			}

			got := rec.snapshot()
			if len(got) != len(tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("event %d = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

type pastePTYWidget struct{ pasted chan PasteEvent }

func (w *pastePTYWidget) Draw(*Canvas, Rect)                  {}
func (w *pastePTYWidget) ConsumeKey(KeyEvent) EventResult     { return Ignored() }
func (w *pastePTYWidget) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (w *pastePTYWidget) ConsumePaste(event PasteEvent) EventResult {
	w.pasted <- event
	return QuitResult()
}

func TestPaneRunDispatchesBracketedPaste(t *testing.T) {
	master, slave := openPTY(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, cols: 20}
	w := &pastePTYWidget{pasted: make(chan PasteEvent, 1)}
	errC := make(chan error, 1)
	go func() { errC <- p.Run(w) }()
	if _, err := master.Write([]byte("\x1b[200~first\nsecond\x1b[201~")); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-w.pasted:
		if got.Text != "first\nsecond" {
			t.Fatalf("paste text = %q", got.Text)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for paste event")
	}
	if err := <-errC; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
	if p.pasteMode {
		t.Fatal("bracketed paste mode remained enabled after Pane.Run")
	}
}

func TestPaneFlushesOversizedUnterminatedPaste(t *testing.T) {
	master, slave := openPTY(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, cols: 20}
	w := &pastePTYWidget{pasted: make(chan PasteEvent, 1)}
	errC := make(chan error, 1)
	go func() { errC <- p.Run(w) }()
	payload := strings.Repeat("x", maxPendingPasteSize+1)
	if _, err := master.Write([]byte("\x1b[200~" + payload)); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-w.pasted:
		if len(got.Text) != len(payload) {
			t.Fatalf("flushed paste length = %d, want %d", len(got.Text), len(payload))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for oversized paste flush")
	}
	if err := <-errC; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
}

func TestPaneProbesZWJAfterEnteringAltScreen(t *testing.T) {
	t.Setenv("LOOM_ZWJ", "")
	master, slave := openPTYForProbe(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 4, cols: 20, startRow: 1, ResizeConfig: DefaultResizeConfig()}
	p.ResizeConfig.AltScreen = true
	rec := newKeyRecorder(1)
	output := make(chan string, 1)
	probed := make(chan struct{})
	stop := make(chan struct{})
	go func() {
		var captured strings.Builder
		replied := false
		buf := make([]byte, 256)
		for {
			select {
			case <-stop:
				output <- captured.String()
				return
			default:
			}
			fds := []unix.PollFd{{Fd: int32(master.Fd()), Events: unix.POLLIN}}
			if n, err := unix.Poll(fds, 20); err != nil || n == 0 {
				continue
			}
			n, err := master.Read(buf)
			if err != nil {
				continue
			}
			captured.Write(buf[:n])
			if !replied && strings.Contains(captured.String(), "\x1b[6n") {
				_, _ = master.Write([]byte("\x1b[1;3R"))
				replied = true
				select {
				case <-probed:
				default:
					close(probed)
				}
			}
		}
	}()
	errC := make(chan error, 1)
	go func() { errC <- p.Run(rec) }()
	select {
	case <-probed:
	case <-time.After(time.Second):
		t.Fatal("pane did not issue ZWJ probe")
	}
	if _, err := master.Write([]byte("q")); err != nil {
		t.Fatal(err)
	}
	if err := <-errC; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
	close(stop)
	select {
	case raw := <-output:
		alt := strings.Index(raw, "\x1b[?1049h")
		probe := strings.Index(raw, "\x1b7\x1b[1;1H👨‍👩‍👧‍👦\x1b[6n")
		erase := strings.Index(raw, "\r\x1b[2K\x1b8")
		if alt < 0 || probe < alt || erase < probe {
			t.Fatalf("startup output does not enter alt screen before probing and erasing glyph: %q", raw)
		}
	case <-time.After(time.Second):
		t.Fatal("pane probe output capture did not finish")
	}
}

func TestKey088PTYCapture(t *testing.T) {
	master, slave := openPTY(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, cols: 40}
	rec := newKeyRecorder(4)
	errC := make(chan error, 1)
	go func() { errC <- p.Run(rec) }()
	for _, raw := range [][]byte{[]byte("ä"), {27, 'a'}, {1}, []byte("\x1b[A")} {
		if _, err := master.Write(raw); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case <-rec.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for PTY key capture")
	}
	if err := <-errC; err != nil {
		t.Fatal(err)
	}
	got := rec.snapshot()
	want := []string{"ä", "alt-a", "ctrl-a", "up"}
	if len(got) != len(want) {
		t.Fatalf("captured %d keys, want %d: %+v", len(got), len(want), got)
	}
	for i, event := range got {
		if event.Name() != want[i] {
			t.Errorf("key %d = %q, want %q", i, event.Name(), want[i])
		}
	}
}

// TestPaneRunReassemblesSplitEscapeSequence is the ticket-053 acceptance
// case for bug 2: an escape sequence split across two Write calls (with a
// short delay, simulating two separate tty reads) must decode as the single
// intended key event, not as a lone "esc" plus garbage. Before the fix,
// Pane.run had no carry-over buffer, so the lone leading ESC from the first
// read was decoded as an "esc" keypress on its own.
func TestPaneRunReassemblesSplitEscapeSequence(t *testing.T) {
	master, slave := openPTY(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, cols: 20}
	rec := newKeyRecorder(1)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(rec) }()

	if _, err := master.Write([]byte{0x1b}); err != nil {
		t.Fatalf("write esc: %v", err)
	}
	// Well under the pane's escape-completion timeout (50ms), so the second
	// write should still be joined onto the pending lone ESC.
	time.Sleep(10 * time.Millisecond)
	if _, err := master.Write([]byte("[A")); err != nil {
		t.Fatalf("write rest: %v", err)
	}

	select {
	case <-rec.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for the split escape sequence to be decoded")
	}
	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}

	got := rec.snapshot()
	if len(got) != 1 || got[0].Key != "up" {
		t.Fatalf("got %+v, want a single {Key:up} event (not esc)", got)
	}
}

// TestPaneRunStandaloneEscTimesOut confirms a genuinely standalone ESC
// keypress (no following bytes) still resolves to a plain "esc" event
// promptly instead of hanging forever waiting for a sequence that never
// arrives.
func TestPaneRunStandaloneEscTimesOut(t *testing.T) {
	master, slave := openPTY(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, cols: 20}
	rec := newKeyRecorder(1)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(rec) }()

	if _, err := master.Write([]byte{0x1b}); err != nil {
		t.Fatalf("write esc: %v", err)
	}

	select {
	case <-rec.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for standalone ESC to resolve")
	}
	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}

	got := rec.snapshot()
	if len(got) != 1 || got[0].Key != "esc" {
		t.Fatalf("got %+v, want a single {Key:esc} event", got)
	}
}

func setPTYSize(t *testing.T, master *os.File, cols, rows int) {
	t.Helper()
	ws := &unix.Winsize{Row: uint16(rows), Col: uint16(cols)}
	if err := unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, ws); err != nil {
		t.Fatalf("set pty size: %v", err)
	}
	_ = unix.Kill(unix.Getpid(), unix.SIGWINCH)
}

type ptyResizeWidget struct {
	mu      sync.Mutex
	draws   int
	bounds  []Rect
	lastCol int
	lastRow int
	done    chan struct{}
}

func newPTYResizeWidget() *ptyResizeWidget {
	return &ptyResizeWidget{done: make(chan struct{})}
}

func (w *ptyResizeWidget) Draw(c *Canvas, r Rect) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.draws++
	w.bounds = append(w.bounds, r)
	w.lastCol = c.Cols()
	w.lastRow = c.Rows()
	c.Write(0, 0, fmt.Sprintf("draw %d: %dx%d", w.draws, c.Cols(), c.Rows()), Style{})
}

func (w *ptyResizeWidget) ConsumeMouse(MouseEvent) EventResult { return Ignored() }

func (w *ptyResizeWidget) ConsumeKey(e KeyEvent) EventResult {
	if e.Key == "q" || e.Text == "q" {
		close(w.done)
		return QuitResult()
	}
	return Ignored()
}

func (w *ptyResizeWidget) lastDimensions() (int, int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastCol, w.lastRow
}

func (w *ptyResizeWidget) history() []Rect {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]Rect, len(w.bounds))
	copy(out, w.bounds)
	return out
}

func TestPTYResizeBurstWideNarrowWide(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)

	p := &Pane{
		tty:                slave,
		fd:                 int(slave.Fd()),
		rows:               20,
		wantRows:           20,
		cols:               80,
		MaxCols:            0,
		Resizeable:         true,
		ResizeConfig:       DefaultResizeConfig(),
		WidthGuardDuration: 30 * time.Millisecond,
	}
	w := newPTYResizeWidget()
	drainPTY(master, w.done)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()

	// Let the initial frame render
	time.Sleep(20 * time.Millisecond)

	// Simulate rapid resize bursts: 80x24 -> 40x20 -> 25x10 -> 60x18 -> 80x24
	burst := []struct{ cols, rows int }{
		{40, 20},
		{25, 10},
		{60, 18},
		{80, 24},
	}
	for _, sz := range burst {
		setPTYSize(t, master, sz.cols, sz.rows)
		time.Sleep(15 * time.Millisecond)
	}

	// Wait for coalesced resize handling and width guard restoration
	time.Sleep(80 * time.Millisecond)

	// Send 'q' to exit
	if _, err := master.Write([]byte("q")); err != nil {
		t.Fatalf("write q: %v", err)
	}

	select {
	case <-w.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for pty resize test to finish")
	}

	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}

	lastCols, lastRows := w.lastDimensions()
	if lastCols != 80 || lastRows < 20 {
		t.Fatalf("final dimensions %dx%d, want 80x>=20", lastCols, lastRows)
	}
}

func TestPTYResizeModeToggles(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 60, 20)

	cfg := DefaultResizeConfig()
	p := &Pane{
		tty:          slave,
		fd:           int(slave.Fd()),
		rows:         10,
		wantRows:     10,
		cols:         60,
		MaxCols:      0,
		Resizeable:   true,
		ResizeConfig: cfg,
	}
	w := newPTYResizeWidget()

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()

	time.Sleep(30 * time.Millisecond)

	// Drain output from master
	buf := make([]byte, 4096)
	n, _ := master.Read(buf)
	initialOutput := string(buf[:n])

	if !strings.Contains(initialOutput, "\x1b[?2026h") {
		t.Errorf("expected synchronized output sequence in initial frame: %q", initialOutput)
	}
	if !strings.Contains(initialOutput, "\x1b[?7l") {
		t.Errorf("expected disable auto-wrap sequence in initial frame: %q", initialOutput)
	}

	// Exit
	master.Write([]byte("q")) //nolint:errcheck
	select {
	case <-w.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for exit")
	}
	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
}

func TestPTYResizeStaleRowClearing(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)

	p := &Pane{
		tty:          slave,
		fd:           int(slave.Fd()),
		rows:         20,
		wantRows:     20,
		cols:         80,
		MaxCols:      0,
		Resizeable:   true,
		ResizeConfig: DefaultResizeConfig(),
	}
	w := newPTYResizeWidget()
	drainPTY(master, w.done)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()

	time.Sleep(30 * time.Millisecond)

	// Shrink height significantly: 24 -> 12 rows
	setPTYSize(t, master, 80, 12)
	time.Sleep(50 * time.Millisecond)

	// Exit
	master.Write([]byte("q")) //nolint:errcheck
	select {
	case <-w.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for exit")
	}
	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
}

func TestPTYResizeReduceMotionAndTheme(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)

	p := &Pane{
		tty:          slave,
		fd:           int(slave.Fd()),
		rows:         15,
		wantRows:     15,
		cols:         80,
		MaxCols:      0,
		ReduceMotion: true,
		Background:   NewAstraBackground(),
		Resizeable:   true,
		ResizeConfig: DefaultResizeConfig(),
	}
	w := newPTYResizeWidget()
	drainPTY(master, w.done)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()

	time.Sleep(30 * time.Millisecond)
	setPTYSize(t, master, 50, 18)
	time.Sleep(30 * time.Millisecond)

	master.Write([]byte("q")) //nolint:errcheck
	select {
	case <-w.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for exit")
	}
	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
}

type runtimeBackgroundProbe struct {
	pane     *Pane
	frames   chan bool
	interval time.Duration
}

func (b *runtimeBackgroundProbe) DrawBackground(*Canvas, Rect) {}
func (b *runtimeBackgroundProbe) DrawBackgroundAt(*Canvas, Rect, time.Time) {
	select {
	case b.frames <- b.pane.BackgroundOnRedraw:
	default:
	}
}
func (b *runtimeBackgroundProbe) BackgroundInterval() time.Duration { return b.interval }

type runtimeBackgroundWidget struct {
	pane  *Pane
	bg    Background
	draws chan bool
}

func (w *runtimeBackgroundWidget) Draw(*Canvas, Rect) {
	select {
	case w.draws <- w.pane.BackgroundOnRedraw:
	default:
	}
}
func (*runtimeBackgroundWidget) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (w *runtimeBackgroundWidget) ConsumeKey(e KeyEvent) EventResult {
	switch {
	case e.Is("a"):
		w.pane.Background = w.bg
	case e.Is("o"):
		w.pane.Background = w.bg
		w.pane.BackgroundOnRedraw = true
	case e.Is("p"):
		w.pane.Background = nil
		w.pane.BackgroundOnRedraw = false
	case e.Is("q"):
		return QuitResult()
	}
	return Handled()
}

func TestPTYBackgroundTickerRuntimeLifecycle(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 15, wantRows: 15, cols: 80, ResizeConfig: DefaultResizeConfig()}
	background := &runtimeBackgroundProbe{pane: p, frames: make(chan bool, 16), interval: 15 * time.Millisecond}
	w := &runtimeBackgroundWidget{pane: p, bg: background, draws: make(chan bool, 16)}
	done := make(chan struct{})
	drainPTY(master, done)
	errC := make(chan error, 1)
	go func() { errC <- p.Run(w) }()
	defer func() {
		master.Write([]byte("q")) //nolint:errcheck
		select {
		case <-errC:
		case <-time.After(3 * time.Second):
			t.Error("pane did not stop")
		}
		close(done)
	}()

	writeKey := func(key byte) {
		t.Helper()
		if _, err := master.Write([]byte{key}); err != nil {
			t.Fatalf("write key %q: %v", key, err)
		}
	}
	waitDraw := func(want bool) {
		t.Helper()
		deadline := time.NewTimer(time.Second)
		defer deadline.Stop()
		for {
			select {
			case got := <-w.draws:
				if got == want {
					return
				}
			case <-deadline.C:
				t.Fatalf("pane did not redraw with BackgroundOnRedraw=%v", want)
			}
		}
	}
	waitFrame := func(want bool) {
		t.Helper()
		deadline := time.NewTimer(time.Second)
		defer deadline.Stop()
		for {
			select {
			case got := <-background.frames:
				if got == want {
					return
				}
			case <-deadline.C:
				t.Fatalf("animated background did not draw in mode %v", want)
			}
		}
	}
	waitQuiet := func() {
		t.Helper()
		select {
		case mode := <-background.frames:
			t.Fatalf("background drew without a foreground redraw (on-redraw=%v)", mode)
		case <-time.After(5 * background.interval):
		}
	}

	waitDraw(false) // Initial plain frame.
	writeKey('a')
	waitDraw(false)
	waitFrame(false) // Runtime switch redraw.
	waitFrame(false) // Independent ticker redraw.
	writeKey('o')
	waitDraw(true)
	waitFrame(true) // Runtime on-redraw switch redraw.
	waitQuiet()
	writeKey('p')
	waitDraw(false) // Switching to plain redraws the foreground.
	waitQuiet()
}

func TestPTYResizeWidthGuardBurstAndRestore(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)

	cfg := DefaultResizeConfig()
	cfg.WidthGuard = true
	cfg.WidthGuardN = 1

	p := &Pane{
		tty:                slave,
		fd:                 int(slave.Fd()),
		rows:               20,
		wantRows:           20,
		cols:               80,
		MaxCols:            0,
		Resizeable:         true,
		ResizeConfig:       cfg,
		WidthGuardDuration: 80 * time.Millisecond,
	}
	w := newPTYResizeWidget()
	drainPTY(master, w.done)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()

	// Let the initial frame render (width 80)
	time.Sleep(30 * time.Millisecond)

	// Resize to 60 columns. During active burst, width guard should render at 60 - 1 = 59.
	setPTYSize(t, master, 60, 20)

	// Wait briefly for the WINCH event to process
	time.Sleep(30 * time.Millisecond)

	burstCols, _ := w.lastDimensions()
	if burstCols != 59 {
		t.Fatalf("during active burst: cols = %d, want 59 (60-1)", burstCols)
	}

	// Wait for the 80ms guard timer to expire and restore full width 60
	time.Sleep(120 * time.Millisecond)

	restoredCols, _ := w.lastDimensions()
	if restoredCols != 60 {
		t.Fatalf("after burst expiry: cols = %d, want 60", restoredCols)
	}

	// Exit cleanly
	master.Write([]byte("q")) //nolint:errcheck
	select {
	case <-w.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for exit")
	}
	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
}

func TestPTYResizeWidthGuardConfigurableN(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)

	cfg := DefaultResizeConfig()
	cfg.WidthGuard = true
	cfg.WidthGuardN = 3

	p := &Pane{
		tty:                slave,
		fd:                 int(slave.Fd()),
		rows:               20,
		wantRows:           20,
		cols:               80,
		MaxCols:            0,
		Resizeable:         true,
		ResizeConfig:       cfg,
		WidthGuardDuration: 80 * time.Millisecond,
	}
	w := newPTYResizeWidget()
	drainPTY(master, w.done)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()

	time.Sleep(30 * time.Millisecond)

	// Resize to 70 columns. With n=3, should render at 70 - 3 = 67.
	setPTYSize(t, master, 70, 20)
	time.Sleep(30 * time.Millisecond)

	burstCols, _ := w.lastDimensions()
	if burstCols != 67 {
		t.Fatalf("during active burst (n=3): cols = %d, want 67 (70-3)", burstCols)
	}

	// Wait for restore
	time.Sleep(120 * time.Millisecond)
	restoredCols, _ := w.lastDimensions()
	if restoredCols != 70 {
		t.Fatalf("after burst expiry: cols = %d, want 70", restoredCols)
	}

	master.Write([]byte("q")) //nolint:errcheck
	select {
	case <-w.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for exit")
	}
	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
}

func TestPTYResizeWidthGuardDisabled(t *testing.T) {
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)

	cfg := DefaultResizeConfig()
	cfg.WidthGuard = false

	p := &Pane{
		tty:          slave,
		fd:           int(slave.Fd()),
		rows:         20,
		wantRows:     20,
		cols:         80,
		MaxCols:      0,
		Resizeable:   true,
		ResizeConfig: cfg,
	}
	w := newPTYResizeWidget()
	drainPTY(master, w.done)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()

	time.Sleep(25 * time.Millisecond)

	// Resize to 60 columns. With WidthGuard=false, should render at full 60 columns immediately.
	setPTYSize(t, master, 60, 20)
	time.Sleep(25 * time.Millisecond)

	cols, _ := w.lastDimensions()
	if cols != 60 {
		t.Fatalf("with width guard disabled: cols = %d, want 60", cols)
	}

	master.Write([]byte("q")) //nolint:errcheck
	select {
	case <-w.done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for exit")
	}
	if err := <-errc; err != nil {
		t.Fatalf("Pane.Run: %v", err)
	}
}
