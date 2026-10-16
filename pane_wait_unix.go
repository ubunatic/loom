// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package loom

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func waitTTY(fd int, timeout time.Duration) (bool, error) {
	milliseconds := int(timeout.Milliseconds())
	if milliseconds < 1 {
		milliseconds = 1
	}
	pollFD := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	ready, err := unix.Poll(pollFD, milliseconds)
	return ready > 0, err
}

func isInterrupted(err error) bool {
	return errors.Is(err, unix.EINTR)
}

func drainTerminalInput(fd int) {
	const tcflsh = 0x540B // Linux TCFLSH; argument 0 selects TCIFLUSH (input queue).
	_, _, _ = unix.Syscall(unix.SYS_IOCTL, uintptr(fd), tcflsh, 0)
}

func notifyPaneSignals(interrupts, resizes chan<- os.Signal) {
	signal.Notify(interrupts, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	signal.Notify(resizes, syscall.SIGWINCH)
}

// newWakePipe creates the self-pipe that wakes a blocked input reader. ok is
// false when no pipe could be created; the reader then polls with a timeout.
func newWakePipe() (r, w int, ok bool) {
	var fds [2]int
	if unix.Pipe(fds[:]) != nil {
		return -1, -1, false
	}
	unix.CloseOnExec(fds[0])
	unix.CloseOnExec(fds[1])
	return fds[0], fds[1], true
}

// wakeAndClose writes one byte to the pipe's write end and closes it. A write
// after the read end closed fails with EPIPE, which is harmless (Go ignores
// SIGPIPE for fds other than 1 and 2).
func wakeAndClose(w int) {
	_, _ = unix.Write(w, []byte{0})
	_ = unix.Close(w)
}

func closeWakeEnd(fd int) {
	_ = unix.Close(fd)
}

// waitInput blocks until fd is readable (ttyReady) or wake, when >= 0, becomes
// readable. A negative timeout waits indefinitely.
func waitInput(fd, wake int, timeout time.Duration) (ttyReady bool, err error) {
	pfd := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	if wake >= 0 {
		pfd = append(pfd, unix.PollFd{Fd: int32(wake), Events: unix.POLLIN})
	}
	ms := -1
	if timeout >= 0 {
		ms = max(1, int(timeout.Milliseconds()))
	}
	n, err := unix.Poll(pfd, ms)
	if err != nil {
		return false, err
	}
	return n > 0 && pfd[0].Revents != 0, nil
}
