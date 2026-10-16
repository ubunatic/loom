// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

const suspendMarker = "<<child-runs>>"

// suspendWidget calls onSuspend for ctrl-g and quits on q.
type suspendWidget struct {
	mu        sync.Mutex
	keys      []string
	sizes     [][2]int
	drawn     chan struct{}
	onSuspend func()
}

func newSuspendWidget() *suspendWidget {
	return &suspendWidget{drawn: make(chan struct{}, 256)}
}

func (w *suspendWidget) Draw(c *Canvas, _ Rect) {
	w.mu.Lock()
	w.sizes = append(w.sizes, [2]int{c.Cols(), c.Rows()})
	n := len(w.sizes)
	w.mu.Unlock()
	c.Write(0, 0, fmt.Sprintf("frame %d", n), Style{})
	select {
	case w.drawn <- struct{}{}:
	default:
	}
}

func (w *suspendWidget) ConsumeMouse(MouseEvent) EventResult { return Ignored() }

func (w *suspendWidget) ConsumeKey(e KeyEvent) EventResult {
	w.mu.Lock()
	w.keys = append(w.keys, e.Name())
	w.mu.Unlock()
	switch {
	case e.Is("ctrl-g"):
		w.onSuspend()
	case e.Is("q"):
		return QuitResult()
	}
	return Consumed()
}

func (w *suspendWidget) draws() [][2]int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([][2]int(nil), w.sizes...)
}

func (w *suspendWidget) keyNames() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]string(nil), w.keys...)
}

// waitDraw waits until the widget has drawn more than n frames.
func (w *suspendWidget) waitDraw(t *testing.T, n int) {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for len(w.draws()) <= n {
		select {
		case <-w.drawn:
		case <-deadline:
			t.Fatalf("no frame after draw %d", n)
		}
	}
}

// ptyCapture records everything the pane writes and can answer the cursor
// position query that follows the suspend marker.
type ptyCapture struct {
	mu    sync.Mutex
	buf   bytes.Buffer
	reply string // written to master after the first DSR following the marker
	done  chan struct{}
}

func capturePTY(master *os.File, reply string) *ptyCapture {
	c := &ptyCapture{reply: reply, done: make(chan struct{})}
	go func() {
		chunk := make([]byte, 4096)
		replied := false
		for {
			select {
			case <-c.done:
				return
			default:
			}
			_ = master.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
			n, _ := master.Read(chunk)
			if n == 0 {
				continue
			}
			c.mu.Lock()
			c.buf.Write(chunk[:n])
			out := c.buf.String()
			c.mu.Unlock()
			if !replied && c.reply != "" {
				if i := strings.Index(out, suspendMarker); i >= 0 && strings.Contains(out[i:], "\x1b[6n") {
					replied = true
					_, _ = master.Write([]byte(c.reply))
				}
			}
		}
	}()
	return c
}

func (c *ptyCapture) split(t *testing.T) (before, after string) {
	t.Helper()
	// Keep reading until the pty stays quiet, so bytes written just before
	// split (resume modes, close teardown) are not lost in the pty buffer.
	last, quiet := -1, time.Now()
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		c.mu.Lock()
		n := c.buf.Len()
		c.mu.Unlock()
		if n != last {
			last, quiet = n, time.Now()
		} else if time.Since(quiet) >= 100*time.Millisecond {
			break
		}
	}
	close(c.done)
	c.mu.Lock()
	defer c.mu.Unlock()
	out := c.buf.String()
	i := strings.Index(out, suspendMarker)
	if i < 0 {
		t.Fatalf("suspend marker missing from output %q", out)
	}
	return out[:i], out[i+len(suspendMarker):]
}

func ttyCanonical(t *testing.T, f *os.File) bool {
	t.Helper()
	tio, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	if err != nil {
		t.Fatalf("get termios: %v", err)
	}
	return tio.Lflag&unix.ICANON != 0
}

// TestPaneSuspendHandsTTYToChild runs a child that reads one line while the
// pane is suspended: the child gets the line, the pane sees no stray key, the
// modes are toggled around the child and the frame is repainted afterwards.
func TestPaneSuspendHandsTTYToChild(t *testing.T) {
	head, err := exec.LookPath("head")
	if err != nil {
		t.Skip("head not available")
	}
	t.Setenv("LOOM_ZWJ", "join")
	master, slave, cooked := openPTYWithCookedState(t)
	setPTYSize(t, master, 40, 10)
	// Answer the resume cursor query with typeahead ahead of the report: the
	// "x" must reach the pane, the report must not.
	capture := capturePTY(master, "x\x1b[7;1R")

	p := &Pane{tty: slave, fd: int(slave.Fd()), oldState: cooked, rows: 3, wantRows: 3,
		startRow: 2, cols: 40, InlineOnly: true}
	p.EnableMouseClicks()
	w := newSuspendWidget()
	childStarted := make(chan struct{})
	var childOut bytes.Buffer
	var suspendErr error
	var canonical bool
	drawsBefore, drawsDuring := -1, -2
	resumed := make(chan struct{})
	w.onSuspend = func() {
		defer close(resumed)
		suspendErr = p.Suspend(func() error {
			canonical = ttyCanonical(t, slave)
			drawsBefore = len(w.draws())
			_, _ = slave.WriteString(suspendMarker)
			cmd := exec.Command(head, "-n", "1")
			cmd.Stdin, cmd.Stdout = slave, &childOut
			close(childStarted)
			err := cmd.Run()
			drawsDuring = len(w.draws())
			return err
		})
	}

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()
	w.waitDraw(t, 0)
	if _, err := master.Write([]byte{0x07}); err != nil { // ctrl-g
		t.Fatal(err)
	}
	select {
	case <-childStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("child never started")
	}
	if _, err := master.Write([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-resumed:
	case <-time.After(3 * time.Second):
		t.Fatal("Suspend did not return after the child read its line")
	}
	w.waitDraw(t, drawsBefore) // the resume repaint, then quit
	if _, err := master.Write([]byte("q")); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(3 * time.Second):
		_, aft := capture.split(t)
		t.Fatalf("Run did not return after q; keys=%v after=%q", w.keyNames(), aft)
	}
	p.Close()
	before, after := capture.split(t)

	if suspendErr != nil {
		t.Fatalf("Suspend: %v", suspendErr)
	}
	if !canonical {
		t.Error("tty was not in cooked mode while the child ran")
	}
	if got := childOut.String(); got != "hello\n" {
		t.Errorf("child read %q, want %q", got, "hello\n")
	}
	if drawsDuring != drawsBefore {
		t.Errorf("pane drew %d frames while suspended", drawsDuring-drawsBefore)
	}
	if got, want := strings.Join(w.keyNames(), ","), "ctrl-g,x,q"; got != want {
		t.Errorf("pane keys = %s, want %s (no stray key from the child's line)", got, want)
	}
	for _, seq := range []string{"\x1b[?1000l", "\x1b[?2004l", "\x1b[?25h"} {
		if !strings.Contains(before, seq) {
			t.Errorf("output before child lacks %q", seq)
		}
	}
	for _, seq := range []string{"\x1b[?1000h", "\x1b[?2004h", "\x1b[7;1H"} {
		if !strings.Contains(after, seq) {
			t.Errorf("output after child lacks %q: %q", seq, after)
		}
	}
	if !strings.Contains(after, fmt.Sprintf("frame %d", drawsBefore+1)) {
		t.Errorf("no repainted frame after the child: %q", after)
	}
}

// TestPaneSuspendResizeAppliesOnResume resizes the terminal while suspended on
// the alternate screen; nothing is drawn meanwhile and the first resumed frame
// has the new size.
func TestPaneSuspendResizeAppliesOnResume(t *testing.T) {
	t.Setenv("LOOM_ZWJ", "join")
	master, slave, cooked := openPTYWithCookedState(t)
	setPTYSize(t, master, 40, 10)
	capture := capturePTY(master, "")

	p := &Pane{tty: slave, fd: int(slave.Fd()), oldState: cooked, rows: 3, wantRows: 3,
		startRow: 2, cols: 40, ResizeConfig: DefaultResizeConfig()}
	// A late SIGWINCH from setPTYSize under load would narrow frames by the
	// width guard; this test is about the size, not the guard.
	p.ResizeConfig.WidthGuard = false
	p.SetScreenMode(ScreenAlt)
	w := newSuspendWidget()
	var mu sync.Mutex
	drawsBefore, drawsDuring := -1, -1
	resumed := make(chan struct{})
	w.onSuspend = func() {
		_ = p.Suspend(func() error {
			_, _ = slave.WriteString(suspendMarker)
			mu.Lock()
			drawsBefore = len(w.draws())
			mu.Unlock()
			setPTYSize(t, master, 30, 8)
			time.Sleep(100 * time.Millisecond) // let SIGWINCH arrive while suspended
			mu.Lock()
			drawsDuring = len(w.draws())
			mu.Unlock()
			return nil
		})
		close(resumed)
	}

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()
	w.waitDraw(t, 0)
	if got := w.draws()[len(w.draws())-1]; got != [2]int{40, 10} {
		t.Fatalf("initial alt frame = %v, want 40x10", got)
	}
	_, _ = master.Write([]byte{0x07})
	select {
	case <-resumed:
	case <-time.After(3 * time.Second):
		t.Fatal("Suspend did not return")
	}
	mu.Lock()
	before := drawsBefore
	mu.Unlock()
	w.waitDraw(t, before)
	_, _ = master.Write([]byte("q"))
	if err := <-errc; err != nil {
		t.Fatalf("Run: %v", err)
	}
	p.Close()
	out, after := capture.split(t)

	if drawsDuring != drawsBefore {
		t.Errorf("pane drew %d frames while suspended", drawsDuring-drawsBefore)
	}
	draws := w.draws()
	for i, size := range draws[before:] {
		if size != [2]int{30, 8} {
			t.Errorf("resumed frame %d = %v, want 30x8", i, size)
		}
	}
	if !strings.Contains(out, "\x1b[?1049l") || !strings.Contains(after, "\x1b[?1049h") {
		t.Errorf("alternate screen not left before and re-entered after the child")
	}
}

// TestPaneSuspendRestoresAfterPanicAndError checks the terminal returns to raw
// mode when fn fails, and that a panic is re-raised after the restore.
func TestPaneSuspendRestoresAfterPanicAndError(t *testing.T) {
	master, slave, cooked := openPTYWithCookedState(t)
	drainDone := make(chan struct{})
	defer close(drainDone)
	drainPTY(master, drainDone)
	p := &Pane{tty: slave, fd: int(slave.Fd()), oldState: cooked, rows: 2, wantRows: 2,
		startRow: 1, cols: 20, fullActive: true}

	boom := errors.New("boom")
	if err := p.Suspend(func() error {
		if !ttyCanonical(t, slave) {
			t.Error("not cooked inside Suspend")
		}
		return boom
	}); !errors.Is(err, boom) {
		t.Fatalf("Suspend error = %v, want %v", err, boom)
	}
	if ttyCanonical(t, slave) {
		t.Fatal("raw mode not restored after an error")
	}

	func() {
		defer func() {
			if r := recover(); r != "panic in child" {
				t.Fatalf("recovered %v, want the child's panic", r)
			}
		}()
		_ = p.Suspend(func() error { panic("panic in child") })
	}()
	if ttyCanonical(t, slave) {
		t.Fatal("raw mode not restored after a panic")
	}
	if p.suspended {
		t.Fatal("pane still marked suspended")
	}
}

// TestPaneSuspendCloseInsideFn closes the pane from fn: Suspend must not
// resume onto the closed tty, and Run must return cleanly.
func TestPaneSuspendCloseInsideFn(t *testing.T) {
	t.Setenv("LOOM_ZWJ", "join")
	master, slave, cooked := openPTYWithCookedState(t)
	setPTYSize(t, master, 40, 10)
	capture := capturePTY(master, "")

	p := &Pane{tty: slave, fd: int(slave.Fd()), oldState: cooked, rows: 3, wantRows: 3,
		startRow: 2, cols: 40, InlineOnly: true}
	p.EnableMouseClicks()
	w := newSuspendWidget()
	suspendErr := errors.New("Suspend not called")
	w.onSuspend = func() {
		suspendErr = p.Suspend(func() error {
			_, _ = slave.WriteString(suspendMarker)
			p.Close()
			return nil
		})
	}

	errc := make(chan error, 1)
	go func() { errc <- p.Run(w) }()
	w.waitDraw(t, 0)
	_, _ = master.Write([]byte{0x07})
	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after Close inside Suspend")
	}
	time.Sleep(50 * time.Millisecond) // let the capture see the last writes
	_, after := capture.split(t)
	if suspendErr != nil {
		t.Fatalf("Suspend: %v", suspendErr)
	}
	for _, seq := range []string{"\x1b[?1000h", "\x1b[?2004h", "\x1b[6n"} {
		if strings.Contains(after, seq) {
			t.Errorf("Suspend resumed onto the closed tty: output after fn has %q", seq)
		}
	}
	if p.suspended {
		t.Error("pane still marked suspended")
	}
}
