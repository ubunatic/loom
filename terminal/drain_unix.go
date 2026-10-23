// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package terminal

import "golang.org/x/sys/unix"

// DrainInput flushes the terminal input queue (TCIFLUSH) so raw-mode leftovers
// do not bleed into the outer shell or subsequent terminal readers.
func DrainInput(fd int) {
	const tcflsh = 0x540B // Linux TCFLSH; argument 0 selects TCIFLUSH (input queue).
	_, _, _ = unix.Syscall(unix.SYS_IOCTL, uintptr(fd), tcflsh, 0)
}
