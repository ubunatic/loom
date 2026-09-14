// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"os"

	"codeberg.org/ubunatic/loom"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// This raw-terminal adapter keeps the live demo responsive to Loom's quit
// keys without adding input mechanics to the graph rendering example.
var quitKeys = func() map[string]bool {
	m := make(map[string]bool, len(loom.SpeccedDefaults.FallbackQuitKeys))
	for _, k := range loom.SpeccedDefaults.FallbackQuitKeys {
		m[k] = true
	}
	return m
}()

// watchForQuitKey returns a no-op cleanup when no controlling terminal is
// available, so --watch degrades to signal-only quitting.
func watchForQuitKey(cancel context.CancelFunc) (cleanup func()) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return func() {}
	}
	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		_ = tty.Close()
		return func() {}
	}
	done := make(chan struct{})
	go func() {
		pfd := []unix.PollFd{{Fd: int32(tty.Fd()), Events: unix.POLLIN}}
		buf := make([]byte, 64)
		for {
			select {
			case <-done:
				return
			default:
			}
			n, perr := unix.Poll(pfd, 200)
			if perr != nil {
				if perr == unix.EINTR {
					continue
				}
				return
			}
			if n == 0 {
				continue
			}
			m, rerr := tty.Read(buf)
			if m > 0 {
				ke := loom.DecodeKey(buf[:m])
				key := ke.Key
				if key == "" {
					key = ke.Text
				}
				if quitKeys[key] {
					cancel()
					return
				}
			}
			if rerr != nil {
				return
			}
		}
	}()
	return func() {
		close(done)
		_ = term.Restore(int(tty.Fd()), oldState)
		_ = tty.Close()
	}
}
