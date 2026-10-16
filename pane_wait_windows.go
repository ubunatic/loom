// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package loom

import (
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/windows"
)

func waitTTY(fd int, timeout time.Duration) (bool, error) {
	milliseconds := uint32(timeout.Milliseconds())
	if milliseconds < 1 {
		milliseconds = 1
	}
	result, err := windows.WaitForSingleObject(windows.Handle(fd), milliseconds)
	if err != nil {
		return false, err
	}
	return result == windows.WAIT_OBJECT_0, nil
}

func isInterrupted(error) bool {
	return false
}

func drainTerminalInput(int) {}

func notifyPaneSignals(interrupts, _ chan<- os.Signal) {
	signal.Notify(interrupts, os.Interrupt)
}

// newWakePipe reports no self-pipe on Windows; the input reader polls with a
// timeout instead.
func newWakePipe() (r, w int, ok bool) { return -1, -1, false }

func wakeAndClose(int) {}

func closeWakeEnd(int) {}

// waitInput waits for fd; wake is unused. A negative timeout falls back to a
// short poll so the caller can re-check its stop signal.
func waitInput(fd, _ int, timeout time.Duration) (bool, error) {
	if timeout < 0 {
		timeout = readerFallbackPoll * time.Millisecond
	}
	return waitTTY(fd, timeout)
}
