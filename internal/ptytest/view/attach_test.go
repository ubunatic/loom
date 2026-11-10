// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build linux

package view

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestObserverAttachSkip(t *testing.T) {
	_, err := attachFiles("/dev/does-not-exist", os.Stdin)
	if err != ErrNoTTY {
		t.Fatalf("expected ErrNoTTY, got: %v", err)
	}
}

func TestObserverAttachWithPTY(t *testing.T) {
	// Create a PTY pair
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("cannot open /dev/ptmx: %v", err)
	}
	defer master.Close()

	fd := int(master.Fd())
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		t.Skipf("cannot unlock pty: %v", err)
	}

	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		t.Skipf("cannot get pty number: %v", err)
	}

	slavePath := fmt.Sprintf("/dev/pts/%d", n)

	obs, err := attachFiles(slavePath, os.Stdin)
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	// Verify raw mode flags (ICANON and ECHO cleared, ISIG retained)
	termios, err := unix.IoctlGetTermios(int(obs.Stdin.Fd()), unix.TCGETS)
	if err != nil {
		t.Fatalf("IoctlGetTermios failed: %v", err)
	}
	if termios.Lflag&unix.ISIG == 0 {
		t.Error("expected ISIG to be set in raw mode")
	}
	if termios.Lflag&unix.ICANON != 0 {
		t.Error("expected ICANON to be cleared in raw mode")
	}
	if termios.Lflag&unix.ECHO != 0 {
		t.Error("expected ECHO to be cleared in raw mode")
	}

	initialTermios, err := unix.IoctlGetTermios(int(obs.Stdin.Fd()), unix.TCGETS)
	if err != nil {
		t.Fatalf("Initial IoctlGetTermios failed: %v", err)
	}

	// Normal exit restore check
	obs.Restore()

	restoredTermios, err := unix.IoctlGetTermios(int(obs.Stdin.Fd()), unix.TCGETS)
	if err == nil && restoredTermios != nil && initialTermios != nil {
		if initialTermios.Lflag == restoredTermios.Lflag {
			t.Error("Expected termios to be restored on exit")
		}
	}

	// Idempotency
	obs.Restore()
}
