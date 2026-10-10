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
