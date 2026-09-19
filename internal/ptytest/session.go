// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

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

	mu  sync.Mutex
	vt  *VT
	raw []byte
}

// Start runs the binary on a cols x rows PTY. It skips the test when PTYs are
// unavailable, per docs/Canary.md, and kills the child on cleanup.
func Start(t *testing.T, cols, rows int, name string, args ...string) *Session {
	t.Helper()
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

	s := &Session{t: t, master: master, done: make(chan struct{}), vt: NewVT(cols, rows)}
	s.setSize(cols, rows)
	s.cmd = exec.Command(name, args...)
	s.cmd.Stdin, s.cmd.Stdout, s.cmd.Stderr = slave, slave, slave
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
	if _, err := s.master.WriteString(keys); err != nil {
		s.t.Fatalf("send %q: %v", keys, err)
	}
}

// Screen returns the current screen rows.
func (s *Session) Screen() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vt.Screen()
}

// Frames returns a copy of the snapshots taken at each ?2026 frame end.
func (s *Session) Frames() [][]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([][]string(nil), s.vt.Frames...)
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
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(s.dump(), want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	s.t.Fatalf("timed out waiting for %q; screen:\n%s", want, s.dump())
}

func (s *Session) dump() string {
	return strings.Join(s.Screen(), "\n")
}

// Wait waits for the child to exit and returns its error (nil on status 0).
func (s *Session) Wait(timeout time.Duration) error {
	s.t.Helper()
	errc := make(chan error, 1)
	go func() { errc <- s.cmd.Wait() }()
	select {
	case err := <-errc:
		return err
	case <-time.After(timeout):
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
