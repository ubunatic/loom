// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build linux

package view

import (
	"errors"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
	"ubunatic.com/loom/terminal"
)

var ErrNoTTY = errors.New("no controlling terminal available")

// Observer represents the attached observer terminal.
type Observer struct {
	Stdin    *os.File
	Stdout   *os.File
	restorer *terminal.Restorer
	termios  *term.State
	closed   bool
}

// Attach connects to the observer's terminal (writing to /dev/tty, reading from
// os.Stdin or /dev/tty) and sets it to raw mode while keeping ISIG enabled.
func Attach() (*Observer, error) {
	return attachFiles("/dev/tty", os.Stdin)
}

func attachFiles(ttyPath string, fallbackStdin *os.File) (*Observer, error) {
	// Write side is always /dev/tty (or ttyPath for testing)
	stdout, err := os.OpenFile(ttyPath, os.O_WRONLY, 0)
	if err != nil {
		return nil, ErrNoTTY
	}

	// Read side is os.Stdin if it's a terminal, else /dev/tty
	var stdin *os.File
	if term.IsTerminal(int(fallbackStdin.Fd())) {
		stdin = fallbackStdin
	} else {
		stdin, err = os.OpenFile(ttyPath, os.O_RDONLY, 0)
		if err != nil {
			stdout.Close()
			return nil, ErrNoTTY
		}
	}

	fd := int(stdin.Fd())

	// Per Canary A5: Clear only ICANON and ECHO, keep ISIG.
	rawTermios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		if stdin != fallbackStdin {
			stdin.Close()
		}
		stdout.Close()
		return nil, err
	}

	// Create term.State to use with restorer, but we just manually get state from x/term
	state, err := term.GetState(fd)
	if err != nil {
		if stdin != fallbackStdin {
			stdin.Close()
		}
		stdout.Close()
		return nil, err
	}

	newTermios := *rawTermios
	newTermios.Lflag &^= unix.ICANON | unix.ECHO

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &newTermios); err != nil {
		if stdin != fallbackStdin {
			stdin.Close()
		}
		stdout.Close()
		return nil, err
	}

	restorer := terminal.NewRestorer(fd, state, stdout)
	restorer.CatchSignals(nil, syscall.SIGINT, syscall.SIGTERM)

	return &Observer{
		Stdin:    stdin,
		Stdout:   stdout,
		restorer: restorer,
		termios:  state,
	}, nil
}

// Restore restores the observer's terminal state and closes the file descriptors.
// It is idempotent.
func (o *Observer) Restore() {
	if o.closed {
		return
	}
	o.closed = true
	o.restorer.Restore()
	if o.Stdin != nil && o.Stdin != os.Stdin {
		o.Stdin.Close()
	}
	if o.Stdout != nil {
		o.Stdout.Close()
	}
}
