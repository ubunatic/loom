// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"os"
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
