// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func readCursorReport(fd int, timeout time.Duration) (int, int, error) {
	deadline := time.Now().Add(timeout)
	var reply strings.Builder
	for reply.Len() < 64 {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, 0, fmt.Errorf("no DSR reply within %s", timeout)
		}
		pollFD := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		waitMS := int(remaining.Milliseconds())
		if waitMS < 1 {
			waitMS = 1
		}
		ready, err := unix.Poll(pollFD, waitMS)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return 0, 0, fmt.Errorf("wait for DSR reply: %w", err)
		}
		if ready == 0 {
			return 0, 0, fmt.Errorf("no DSR reply within %s", timeout)
		}
		var one [1]byte
		n, err := unix.Read(fd, one[:])
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return 0, 0, fmt.Errorf("read DSR reply: %w", err)
		}
		if n != 1 {
			continue
		}
		reply.WriteByte(one[0])
		if one[0] == 'R' {
			return parseCursorReport(reply.String())
		}
	}
	return 0, 0, fmt.Errorf("invalid DSR reply %q", reply.String())
}

// keyGap is how long readChunk waits for the rest of a multi-byte key.
const keyGap = 20 * time.Millisecond

// readChunk waits up to timeout (forever if zero) for input, then returns the
// bytes that arrive without a gap longer than keyGap: one key, usually.
func readChunk(fd int, timeout time.Duration) ([]byte, error) {
	wait := -1
	if timeout > 0 {
		wait = max(int(timeout.Milliseconds()), 1)
	}
	var chunk []byte
	buf := make([]byte, 64)
	for {
		ready, err := unix.Poll([]unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}, wait)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return chunk, err
		}
		if ready == 0 {
			return chunk, nil
		}
		n, err := unix.Read(fd, buf)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return chunk, err
		}
		if n == 0 {
			// Hangup: poll keeps reporting the fd ready, so stop instead of spinning.
			return chunk, io.EOF
		}
		chunk = append(chunk, buf[:n]...)
		wait = int(keyGap.Milliseconds())
	}
}
