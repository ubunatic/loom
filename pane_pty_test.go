// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// openPTY opens a real Linux pseudo-terminal pair via /dev/ptmx, returning
// the master (test-controlled) and slave (what Pane.run reads/writes) ends.
// It skips the test rather than failing when ptys are unavailable (e.g. some
// sandboxes), matching this repo's "probe external mechanisms" canary-first
// convention (docs/Canary.md).
func openPTY(t *testing.T) (master, slave *os.File) {
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

func (r *keyRecorder) Draw(*Canvas, Rect)          {}
func (r *keyRecorder) HandleMouse(MouseEvent) bool { return false }
func (r *keyRecorder) HandleKey(e KeyEvent) bool {
	r.mu.Lock()
	r.keys = append(r.keys, e)
	n := len(r.keys)
	r.mu.Unlock()
	if n >= r.want {
		close(r.done)
		return true
	}
	return false
}

func (r *keyRecorder) snapshot() []KeyEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]KeyEvent, len(r.keys))
	copy(out, r.keys)
	return out
}

// TestPaneRunDecodesMultipleKeysFromOneRead is the ticket-053 acceptance
// case for bug 1: two complete key sequences written in a single Write (so
// they are very likely to land in a single tty.Read) must both be decoded
// and dispatched via HandleKey, in order. Before the fix, Pane.run decoded
// only the first sequence in raw and silently dropped the rest.
func TestPaneRunDecodesMultipleKeysFromOneRead(t *testing.T) {
	master, slave := openPTY(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, cols: 20}
	rec := newKeyRecorder(2)

	errc := make(chan error, 1)
	go func() { errc <- p.Run(rec) }()

	if _, err := master.Write([]byte("\x1b[A\x1b[B")); err != nil {
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
	if len(got) != 2 || got[0].Key != "up" || got[1].Key != "down" {
		t.Fatalf("got %+v, want [{Key:up} {Key:down}]", got)
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

func (w *ptyResizeWidget) HandleMouse(MouseEvent) bool { return false }

func (w *ptyResizeWidget) HandleKey(e KeyEvent) bool {
	if e.Key == "q" || e.Text == "q" {
		close(w.done)
		return true
	}
	return false
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
