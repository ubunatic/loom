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

	// A separate fd on the slave reads termios after Restore closes the observer's fds.
	probe, err := os.OpenFile(slavePath, os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Fatalf("open slave probe: %v", err)
	}
	defer probe.Close()
	lflag := func() uint32 {
		t.Helper()
		termios, err := unix.IoctlGetTermios(int(probe.Fd()), unix.TCGETS)
		if err != nil {
			t.Fatalf("IoctlGetTermios: %v", err)
		}
		return termios.Lflag
	}
	before := lflag()

	// A non-terminal fallback stdin makes attachFiles read from the slave,
	// never from the terminal that runs the tests.
	devnull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer devnull.Close()

	obs, err := attachFiles(slavePath, devnull)
	if err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	// Raw mode: ICANON and ECHO cleared, ISIG kept.
	raw := lflag()
	if raw&unix.ISIG == 0 {
		t.Error("expected ISIG to be set in raw mode")
	}
	if raw&unix.ICANON != 0 {
		t.Error("expected ICANON to be cleared in raw mode")
	}
	if raw&unix.ECHO != 0 {
		t.Error("expected ECHO to be cleared in raw mode")
	}

	obs.Restore()
	if got := lflag(); got != before {
		t.Errorf("Lflag after Restore = %#x, want the original %#x", got, before)
	}

	// Idempotency: a second Restore is a no-op.
	obs.Restore()
	if got := lflag(); got != before {
		t.Errorf("Lflag after second Restore = %#x, want %#x", got, before)
	}
}
