// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package measure

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func readProbeCursor(fd int, timeout time.Duration) (int, int, error) {
	deadline := time.Now().Add(timeout)
	var reply strings.Builder
	for reply.Len() < 64 {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, 0, fmt.Errorf("terminal probe timed out")
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		ms := int(remaining.Milliseconds())
		if ms < 1 {
			ms = 1
		}
		n, err := unix.Poll(fds, ms)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return 0, 0, err
		}
		if n == 0 {
			return 0, 0, fmt.Errorf("terminal probe timed out")
		}
		var one [1]byte
		count, err := unix.Read(fd, one[:])
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return 0, 0, err
		}
		if count == 0 {
			continue
		}
		reply.WriteByte(one[0])
		if one[0] == 'R' {
			var row, col int
			if _, err := fmt.Sscanf(reply.String(), "\x1b[%d;%dR", &row, &col); err != nil || row < 1 || col < 1 {
				return 0, 0, fmt.Errorf("invalid cursor report %q", reply.String())
			}
			return row, col, nil
		}
	}
	return 0, 0, fmt.Errorf("invalid cursor report %q", reply.String())
}
