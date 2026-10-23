// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
	"ubunatic.com/loom"
	"ubunatic.com/loom/vt"
)

type sessionExitError struct {
	code int
	err  error
}

func (e sessionExitError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e sessionExitError) Unwrap() error { return e.err }

func sessionCommand() *cobra.Command {
	sessionDefs := loom.SpeccedDefaults.Session
	shotKey := strings.Join(sessionDefs.Bindings.Screenshot, ", ")
	quitKey := strings.Join(sessionDefs.Bindings.Quit, ", ")
	prefixShot := strings.Join(sessionDefs.PrefixBindings.Screenshot, ", ")
	prefixQuit := strings.Join(sessionDefs.PrefixBindings.Quit, ", ")
	sendPrefix := strings.Join(sessionDefs.PrefixBindings.SendPrefix, ", ")
	longHelp := fmt.Sprintf(
		"Run the command in a pseudo-terminal, passing all input and output through unchanged. Hotkeys: %s saves an ANSI screenshot, %s quits. Prefix %s followed by %s saves a screenshot, %s quits, and %s forwards %s to the child.",
		shotKey, quitKey, sessionDefs.Prefix, prefixShot, prefixQuit, sendPrefix, sessionDefs.Prefix,
	)
	cmd := &cobra.Command{
		Use:   "session -- <cmd> [args...]",
		Short: "Run a command in a PTY wrapper with screenshot and quit hotkeys",
		Long:  longHelp,
		Args:  cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSession(cmd.Context(), os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], args[1:]...)
		},
	}
	return cmd
}

func openSessionPTY(cols, rows int) (master, slave *os.File, err error) {
	master, err = os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("open /dev/ptmx: %w", err)
	}
	fd := int(master.Fd())
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("unlock pty: %w", err)
	}
	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("pty number: %w", err)
	}
	slave, err = os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), os.O_RDWR, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("open slave: %w", err)
	}
	setPTYSize(master, cols, rows)
	return master, slave, nil
}

func setPTYSize(f *os.File, cols, rows int) {
	ws := &unix.Winsize{Col: uint16(cols), Row: uint16(rows)}
	_ = unix.IoctlSetWinsize(int(f.Fd()), unix.TIOCSWINSZ, ws)
}

func runSession(ctx context.Context, in io.Reader, out io.Writer, errOut io.Writer, command string, args ...string) error {
	cols, rows := 80, 24
	var oldState *term.State
	if inFile, ok := in.(*os.File); ok && term.IsTerminal(int(inFile.Fd())) {
		if w, h, err := term.GetSize(int(inFile.Fd())); err == nil && w > 0 && h > 0 {
			cols, rows = w, h
		}
		state, err := term.MakeRaw(int(inFile.Fd()))
		if err == nil {
			oldState = state
			defer func() {
				if oldState != nil {
					_ = term.Restore(int(inFile.Fd()), oldState)
				}
			}()
		}
	} else if outFile, ok := out.(*os.File); ok && term.IsTerminal(int(outFile.Fd())) {
		if w, h, err := term.GetSize(int(outFile.Fd())); err == nil && w > 0 && h > 0 {
			cols, rows = w, h
		}
	}

	master, slave, err := openSessionPTY(cols, rows)
	if err != nil {
		return err
	}
	defer master.Close()

	child := exec.CommandContext(ctx, command, args...)
	child.Stdin = slave
	child.Stdout = slave
	child.Stderr = slave
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := child.Start(); err != nil {
		slave.Close()
		return fmt.Errorf("session start %s: %w", command, err)
	}
	slave.Close()

	var vtMu sync.Mutex
	screenModel := vt.NewVT(cols, rows)

	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)
	defer signal.Stop(sigWinch)
	go func() {
		for range sigWinch {
			if inFile, ok := in.(*os.File); ok && term.IsTerminal(int(inFile.Fd())) {
				if w, h, err := term.GetSize(int(inFile.Fd())); err == nil && w > 0 && h > 0 {
					setPTYSize(master, w, h)
					vtMu.Lock()
					screenModel.Resize(w, h)
					vtMu.Unlock()
				}
			}
		}
	}()

	var savedShots []string
	var shotErrors []error
	var shotMu sync.Mutex
	saveScreenshot := func() error {
		vtMu.Lock()
		content := screenModel.ANSI()
		vtMu.Unlock()

		baseCmd := filepath.Base(command)
		safeCmd := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, baseCmd)
		if safeCmd == "" || safeCmd == "-" {
			safeCmd = "session"
		}

		timestamp := time.Now().Format("20060102-150405")
		path := fmt.Sprintf("./loom-%s-%s.ansi", safeCmd, timestamp)
		for i := 1; ; i++ {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err == nil {
				_, writeErr := f.WriteString(content)
				closeErr := f.Close()
				if writeErr != nil {
					return writeErr
				}
				if closeErr != nil {
					return closeErr
				}
				shotMu.Lock()
				savedShots = append(savedShots, path)
				shotMu.Unlock()
				return nil
			}
			if os.IsExist(err) {
				path = fmt.Sprintf("./loom-%s-%s-%d.ansi", safeCmd, timestamp, i)
				continue
			}
			return err
		}
	}

	stdoutDone := make(chan struct{})
	go func() {
		defer close(stdoutDone)
		buf := make([]byte, 4096)
		for {
			n, err := master.Read(buf)
			if n > 0 {
				out.Write(buf[:n]) //nolint:errcheck
				vtMu.Lock()
				screenModel.Write(buf[:n]) //nolint:errcheck
				vtMu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()

	stdinDone := make(chan struct{})
	go func() {
		defer close(stdinDone)
		buf := make([]byte, 2048)
		var pending []byte
		prefixMode := false

		sessionDefs := loom.SpeccedDefaults.Session

		for {
			n, err := in.Read(buf)
			if n > 0 {
				pending = append(pending, buf[:n]...)

				for len(pending) > 0 {
					// A lone Esc (0x1b with no following byte in the same read) must reach
					// the child at once rather than waiting for more input.
					if len(pending) == 1 && pending[0] == 0x1b {
						prefixMode = false
						master.Write(pending) //nolint:errcheck
						pending = nil
						break
					}

					// Check for bracketed paste first so pasted control characters are never split
					if paste, used, ok := loom.DecodePaste(pending); ok {
						_ = paste
						master.Write(pending[:used]) //nolint:errcheck
						pending = pending[used:]
						prefixMode = false
						continue
					}

					key, used, ok := loom.ScanKey(pending)
					if !ok {
						// Incomplete sequence; wait for more data
						break
					}

					if prefixMode {
						switch {
						case key.Is(sessionDefs.PrefixBindings.Screenshot...):
							if err := saveScreenshot(); err != nil {
								shotMu.Lock()
								shotErrors = append(shotErrors, err)
								shotMu.Unlock()
							}
							prefixMode = false
						case key.Is(sessionDefs.PrefixBindings.Quit...):
							killProcessGroup(child.Process.Pid)
							return
						case key.Is(sessionDefs.PrefixBindings.SendPrefix...):
							master.Write(pending[:used]) //nolint:errcheck
							prefixMode = false
						default:
							// Any other key cancels prefix mode and forwards the key
							prefixMode = false
							master.Write(pending[:used]) //nolint:errcheck
						}
						pending = pending[used:]
						continue
					}

					// Direct hotkey matching
					switch {
					case key.Is(sessionDefs.Bindings.Screenshot...):
						if err := saveScreenshot(); err != nil {
							shotMu.Lock()
							shotErrors = append(shotErrors, err)
							shotMu.Unlock()
						}
					case key.Is(sessionDefs.Bindings.Quit...):
						killProcessGroup(child.Process.Pid)
						return
					case key.Is(sessionDefs.Prefix):
						prefixMode = true
					default:
						master.Write(pending[:used]) //nolint:errcheck
					}
					pending = pending[used:]
				}
			}
			if err != nil {
				return
			}
		}
	}()

	waitErr := child.Wait()

	// Restore outer terminal to cooked mode before printing status to errOut so lines
	// start cleanly at column 0 and do not staircase.
	if inFile, ok := in.(*os.File); ok && oldState != nil {
		_ = term.Restore(int(inFile.Fd()), oldState)
		oldState = nil
	}

	shotMu.Lock()
	for _, p := range savedShots {
		fmt.Fprintf(errOut, "Saved screenshot: %s\n", p)
	}
	for _, err := range shotErrors {
		fmt.Fprintf(errOut, "Error saving screenshot: %v\n", err)
	}
	shotMu.Unlock()

	var exitCode int
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return waitErr
		}
	}

	if exitCode != 0 {
		return sessionExitError{code: exitCode}
	}
	return nil
}

func killProcessGroup(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	_ = syscall.Kill(pid, syscall.SIGTERM)
}
