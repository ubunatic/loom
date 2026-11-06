// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Probe for issue 437 / 095: which descriptor actually sees observer keystrokes.
// The relay selects on /dev/tty and never saw a key, while the pane's own stdin
// is the same terminal. This mode reports both paths.
//
// Run inside a real terminal (tmux pane) and send keys from another window.
package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func probe() error {
	fd0 := 0
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open /dev/tty: %w", err)
	}
	defer tty.Close()
	ttyfd := int(tty.Fd())

	if _, err := term.MakeRaw(fd0); err != nil {
		return fmt.Errorf("raw fd0: %w", err)
	}
	if _, err := term.MakeRaw(ttyfd); err != nil {
		return fmt.Errorf("raw tty: %w", err)
	}

	fmt.Printf("P1 fd0 is terminal: %v, /dev/tty is terminal: %v\n",
		term.IsTerminal(fd0), term.IsTerminal(ttyfd))
	fmt.Printf("P1 fd0 -> %s, /dev/tty fd -> %s\n", resolve("/proc/self/fd/0"), resolve(fmt.Sprintf("/proc/self/fd/%d", ttyfd)))
	fmt.Println("P2 waiting 15s for a keystroke on either path")

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		fds := []int{fd0, ttyfd}
		ready, err := readable(fds, 500*time.Millisecond)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			return fmt.Errorf("select: %w", err)
		}
		for _, fd := range fds {
			if !ready[fd] {
				continue
			}
			buf := make([]byte, 64)
			n, err := unix.Read(fd, buf)
			if n > 0 {
				fmt.Printf("P2 fd %d saw %q\n", fd, string(buf[:n]))
				if n > 0 && buf[0] == 3 {
					return nil
				}
			}
			if err != nil {
				fmt.Printf("P2 fd %d read error: %v\n", fd, err)
			}
		}
	}
	fmt.Println("P2 no keystroke seen on either path")
	return nil
}
