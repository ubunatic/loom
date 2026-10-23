// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package ptytest

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// Session runs a command on a real PTY and mirrors its output into a VT.
type Session struct {
	t      *testing.T
	master *os.File
	cmd    *exec.Cmd
	done   chan struct{}

	mu        sync.Mutex
	vt        *VT
	raw       []byte
	changed   chan struct{}
	revision  uint64
	afterSend uint64
}

// Start runs the binary on a cols x rows PTY. It skips the test when PTYs are
// unavailable, per docs/Canary.md, and kills the child on cleanup.
func Start(t *testing.T, cols, rows int, name string, args ...string) *Session {
	t.Helper()
	if testing.Short() {
		t.Skip("requires a real PTY; disabled for short tests")
	}
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("open /dev/ptmx: %v", err)
	}
	fd := int(master.Fd())
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		master.Close()
		t.Skipf("unlock pty: %v", err)
	}
	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		master.Close()
		t.Skipf("pty number: %v", err)
	}
	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), os.O_RDWR, 0)
	if err != nil {
		master.Close()
		t.Skipf("open slave: %v", err)
	}
	defer slave.Close()

	s := &Session{t: t, master: master, done: make(chan struct{}), changed: make(chan struct{}), vt: NewVT(cols, rows)}
	s.setSize(cols, rows)
	s.cmd = exec.Command(name, args...)
	s.cmd.Stdin, s.cmd.Stdout, s.cmd.Stderr = slave, slave, slave
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "LOOM_ZWJ=") && !strings.HasPrefix(entry, "LOOMCOLOR=") {
			s.cmd.Env = append(s.cmd.Env, entry)
		}
	}
	s.cmd.Env = append(s.cmd.Env, "LOOM_ZWJ=join")       // PTY tests have no terminal emulator to answer DSR.
	s.cmd.Env = append(s.cmd.Env, "LOOMCOLOR=truecolor") // Keep test child rendering independent of the caller's terminal environment.
	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := s.cmd.Start(); err != nil {
		master.Close()
		t.Fatalf("start %s: %v", name, err)
	}
	go s.pump()
	t.Cleanup(s.Close)
	return s
}

func (s *Session) pump() {
	defer close(s.done)
	buf := make([]byte, 4096)
	for {
		n, err := s.master.Read(buf)
		if n > 0 {
			s.mu.Lock()
			s.raw = append(s.raw, buf[:n]...)
			s.vt.Write(buf[:n]) //nolint:errcheck
			s.revision++
			close(s.changed)
			s.changed = make(chan struct{})
			s.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

func (s *Session) setSize(cols, rows int) {
	ws := &unix.Winsize{Col: uint16(cols), Row: uint16(rows)}
	if err := unix.IoctlSetWinsize(int(s.master.Fd()), unix.TIOCSWINSZ, ws); err != nil {
		s.t.Fatalf("set winsize: %v", err)
	}
}

// Resize changes the PTY size (the kernel raises SIGWINCH in the child) and
// the VT grid.
func (s *Session) Resize(cols, rows int) {
	s.t.Helper()
	s.mu.Lock()
	s.vt.Resize(cols, rows)
	s.mu.Unlock()
	s.setSize(cols, rows)
}

// Send writes keystrokes to the child.
func (s *Session) Send(keys string) {
	s.t.Helper()
	s.markSend()
	if _, err := s.master.WriteString(keys); err != nil {
		s.t.Fatalf("send %q: %v", keys, err)
	}
}

// SendRaw writes terminal input bytes without interpreting them.
func (s *Session) SendRaw(data []byte) {
	s.t.Helper()
	s.markSend()
	if _, err := s.master.Write(data); err != nil {
		s.t.Fatalf("send raw input: %v", err)
	}
}

func (s *Session) markSend() {
	s.mu.Lock()
	s.afterSend = s.revision
	s.mu.Unlock()
}

// VT returns the underlying virtual terminal model.
func (s *Session) VT() *VT {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vt
}

// Signal sends an OS signal to the child command process.
func (s *Session) Signal(sig os.Signal) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd != nil && s.cmd.Process != nil {
		return s.cmd.Process.Signal(sig)
	}
	return nil
}

// Screen returns the current screen rows.
func (s *Session) Screen() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vt.Screen()
}

// Text returns the virtual terminal's visible screen joined by newlines.
func (s *Session) Text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vt.Text()
}

// Cells returns a snapshot of the current cell grid (runes and styles).
func (s *Session) Cells() [][]Cell {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vt.Cells()
}

// Cell returns the cell at (x, y), or a zero Cell if out of bounds.
func (s *Session) Cell(x, y int) Cell {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vt.Cell(x, y)
}

// Frames returns a copy of the snapshots taken at each ?2026 frame end.
func (s *Session) Frames() [][]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([][]string(nil), s.vt.Frames...)
}

// CellFrames returns snapshots of the cell grid taken at each ?2026 frame end.
func (s *Session) CellFrames() [][][]Cell {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vt.FrameCells()
}

// Raw returns every byte the child has written so far.
func (s *Session) Raw() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.raw...)
}

// WaitFor polls until the screen text contains want, failing with the screen
// dump on timeout.
func (s *Session) WaitFor(want string, timeout time.Duration) {
	s.t.Helper()
	s.WaitForState(timeout*RaceFactor, func(rows []string) bool {
		return strings.Contains(strings.Join(rows, "\n"), want)
	})
}

// WaitForUpdate waits for want to appear in screen output produced after the
// most recent Send or SendRaw call.
func (s *Session) WaitForUpdate(want string, timeout time.Duration) {
	s.t.Helper()
	s.WaitForUpdateState(timeout*RaceFactor, func(rows []string) bool {
		return strings.Contains(strings.Join(rows, "\n"), want)
	})
}

// WaitForState waits for a newly observed screen state. It wakes on PTY output
// rather than polling, so race-instrumented tests do not depend on scheduler
// timing or arbitrary sleeps.
func (s *Session) WaitForState(timeout time.Duration, predicate func([]string) bool) {
	s.waitForState(timeout, false, predicate)
}

// WaitForUpdateState waits for a screen predicate after the latest input.
func (s *Session) WaitForUpdateState(timeout time.Duration, predicate func([]string) bool) {
	s.waitForState(timeout, true, predicate)
}

func (s *Session) waitForState(timeout time.Duration, afterSend bool, predicate func([]string) bool) {
	s.t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		s.mu.Lock()
		rows := s.vt.Screen()
		changed := s.changed
		ready := !afterSend || s.revision > s.afterSend
		s.mu.Unlock()
		if ready && predicate(rows) {
			return
		}
		select {
		case <-changed:
		case <-timer.C:
			s.t.Fatalf("timed out waiting for screen state; screen:\n%s", strings.Join(s.Screen(), "\n"))
			return
		}
	}
}

func (s *Session) dump() string {
	return strings.Join(s.Screen(), "\n")
}

// Wait waits for the child to exit and returns its error (nil on status 0).
func (s *Session) Wait(timeout time.Duration) error {
	s.t.Helper()
	errc := make(chan error, 1)
	go func() { errc <- s.cmd.Wait() }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-errc:
		select {
		case <-s.done:
			return err
		case <-timer.C:
			s.t.Fatalf("PTY output did not reach EOF within %v; screen:\n%s", timeout, s.dump())
			return err
		}
	case <-timer.C:
		s.t.Fatalf("child did not exit within %v; screen:\n%s", timeout, s.dump())
		return nil
	}
}

// Close kills the child if it is still running and releases the PTY.
func (s *Session) Close() {
	if s.cmd.ProcessState == nil {
		_ = s.cmd.Process.Kill()
	}
	s.master.Close()
}
