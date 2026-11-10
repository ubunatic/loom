// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build linux

package view

import (
	"errors"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// Relay configures the duplex relay between a child PTY and the observer.
type Relay struct {
	Master *os.File
	Stdin  *os.File
	Stdout *os.File
	DSR    func() []byte // Called to get DSR replies to drain to the child
	Render func()        // Called when child writes
}

// Run starts the event-driven duplex loop.
// It returns nil on normal exit or EOF.
func (r *Relay) Run() error {
	masterFd := int(r.Master.Fd())
	stdinFd := int(r.Stdin.Fd())

	pollFds := []unix.PollFd{
		{Fd: int32(masterFd), Events: unix.POLLIN},
		{Fd: int32(stdinFd), Events: unix.POLLIN},
	}

	buf := make([]byte, 4096)

	for {
		// Drain DSR replies to the child before polling
		if r.DSR != nil {
			if replies := r.DSR(); len(replies) > 0 {
				_, _ = r.Master.Write(replies)
			}
		}

		n, err := unix.Poll(pollFds, -1)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue // Retry on EINTR
			}
			return err
		}

		if n > 0 {
			// Observer input -> Child
			if pollFds[1].Revents&unix.POLLIN != 0 {
				nr, err := r.Stdin.Read(buf)
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
				if nr > 0 {
					// Intercept Ctrl-C (0x03) as stop request
					for i := 0; i < nr; i++ {
						if buf[i] == 0x03 {
							return nil
						}
					}
					_, _ = r.Master.Write(buf[:nr])
				}
			}

			// Child output -> Observer
			if pollFds[0].Revents&unix.POLLIN != 0 {
				nr, err := r.Master.Read(buf)
				if err != nil {
					if err == io.EOF {
						return nil
					}
					// EIO is returned on Linux when the PTY slave is closed
					if errors.Is(err, unix.EIO) {
						return nil
					}
					return err
				}
				if nr > 0 {
					if r.Render != nil {
						r.Render()
					} else if r.Stdout != nil {
						_, _ = r.Stdout.Write(buf[:nr])
					}
				}
			}

			if pollFds[0].Revents&(unix.POLLERR|unix.POLLHUP) != 0 {
				return nil
			}
			if pollFds[1].Revents&(unix.POLLERR|unix.POLLHUP) != 0 {
				return nil
			}
		}
	}
}
