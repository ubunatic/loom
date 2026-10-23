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
	newTabKey := strings.Join(sessionDefs.Bindings.NewTab, ", ")
	prefixShot := strings.Join(sessionDefs.PrefixBindings.Screenshot, ", ")
	prefixQuit := strings.Join(sessionDefs.PrefixBindings.Quit, ", ")
	sendPrefix := strings.Join(sessionDefs.PrefixBindings.SendPrefix, ", ")
	prefixNewTab := strings.Join(sessionDefs.PrefixBindings.NewTab, ", ")
	nextTab := strings.Join(sessionDefs.PrefixBindings.NextTab, ", ")
	prevTab := strings.Join(sessionDefs.PrefixBindings.PreviousTab, ", ")
	closeTab := strings.Join(sessionDefs.PrefixBindings.CloseTab, ", ")
	longHelp := fmt.Sprintf(
		"Run the command in a pseudo-terminal, passing all input and output through unchanged. Hotkeys: %s saves an ANSI screenshot, %s quits, %s opens a new tab. Prefix %s followed by %s saves a screenshot, %s quits, and %s forwards %s to the child. Prefix %s opens a new tab, %s/%s switches tabs, %s closes a tab.",
		shotKey, quitKey, newTabKey, sessionDefs.Prefix, prefixShot, prefixQuit, sendPrefix, sessionDefs.Prefix, prefixNewTab, nextTab, prevTab, closeTab,
	)
	cmd := &cobra.Command{
		Use:          "session [-- <cmd> [args...]]",
		Short:        "Run a command in a PTY wrapper with screenshot and quit hotkeys",
		Long:         longHelp,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdName := ""
			var cmdArgs []string
			if len(args) > 0 {
				cmdName = args[0]
				cmdArgs = args[1:]
			}
			return runSession(cmd.Context(), os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr(), cmdName, cmdArgs...)
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

type sessionTab struct {
	id       int
	title    string
	command  string
	args     []string
	master   *os.File
	child    *exec.Cmd
	vt       *vt.VT
	exitCode int
	waitErr  error
	done     chan struct{}
}

func drawSessionTabBar(out io.Writer, tabs []*sessionTab, activeIndex, cols, rows int) {
	if len(tabs) <= 1 {
		return
	}
	var b strings.Builder
	b.WriteString("\x1b7") // DEC save cursor & attributes
	fmt.Fprintf(&b, "\x1b[%d;1H", rows)

	style := loom.DefaultTabsStyle()
	usedCols := 0
	for i, t := range tabs {
		tabStyle := style.Inactive
		if i == activeIndex {
			tabStyle = style.Active
		}
		title := fmt.Sprintf(" %d: %s ", i+1, t.title)
		b.WriteString(tabStyle.ANSI())
		b.WriteString(title)
		usedCols += loom.StringWidth(title)
	}
	if usedCols < cols {
		b.WriteString(style.Inactive.ANSI())
		b.WriteString(strings.Repeat(" ", cols-usedCols))
	}
	b.WriteString("\x1b[0m\x1b8") // reset style and restore cursor
	_, _ = out.Write([]byte(b.String()))
}

func repaintActiveTab(out io.Writer, tabs []*sessionTab, activeIndex, cols, rows int) {
	if activeIndex < 0 || activeIndex >= len(tabs) {
		return
	}
	tab := tabs[activeIndex]
	repaint := tab.vt.Repaint()
	_, _ = out.Write([]byte(repaint))

	if len(tabs) > 1 {
		fmt.Fprintf(out, "\x1b[1;%dr", rows-1)
		drawSessionTabBar(out, tabs, activeIndex, cols, rows)
		x, y := tab.vt.Cursor()
		fmt.Fprintf(out, "\x1b[%d;%dH", y+1, x+1)
		if tab.vt.CursorVisible {
			_, _ = out.Write([]byte("\x1b[?25h"))
		} else {
			_, _ = out.Write([]byte("\x1b[?25l"))
		}
	} else {
		_, _ = out.Write([]byte("\x1b[r"))
	}
}

func runSession(ctx context.Context, in io.Reader, out io.Writer, errOut io.Writer, command string, args ...string) error {
	if command == "" {
		command = os.Getenv("SHELL")
		if command == "" {
			command = "/bin/sh"
		}
	}

	cols, rows := 80, 24
	var oldState *term.State
	if inFile, ok := in.(*os.File); ok && term.IsTerminal(int(inFile.Fd())) {
		if w, h, err := term.GetSize(int(inFile.Fd())); err == nil && w > 0 && h > 0 {
			cols, rows = w, h
		}
		state, err := term.MakeRaw(int(inFile.Fd()))
		if err == nil {
			oldState = state
		}
	} else if outFile, ok := out.(*os.File); ok && term.IsTerminal(int(outFile.Fd())) {
		if w, h, err := term.GetSize(int(outFile.Fd())); err == nil && w > 0 && h > 0 {
			cols, rows = w, h
		}
	}

	defer func() {
		_, _ = out.Write([]byte("\x1b[r\x1b[?1049l\x1b[?25h"))
		if inFile, ok := in.(*os.File); ok && oldState != nil {
			_ = term.Restore(int(inFile.Fd()), oldState)
			oldState = nil
		}
	}()

	var sessionMu sync.Mutex
	var tabs []*sessionTab
	activeIndex := 0
	nextTabID := 1

	tabExitChan := make(chan *sessionTab, 16)
	sessionQuitChan := make(chan struct{})

	spawnTab := func(cmdName string, cmdArgs ...string) (*sessionTab, error) {
		childRows := rows
		if len(tabs) >= 1 {
			childRows = rows - 1
		}
		master, slave, err := openSessionPTY(cols, childRows)
		if err != nil {
			return nil, err
		}
		child := exec.CommandContext(ctx, cmdName, cmdArgs...)
		child.Stdin = slave
		child.Stdout = slave
		child.Stderr = slave
		child.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
		if err := child.Start(); err != nil {
			master.Close()
			slave.Close()
			return nil, fmt.Errorf("session start %s: %w", cmdName, err)
		}
		slave.Close()

		baseCmd := filepath.Base(cmdName)
		tab := &sessionTab{
			id:      nextTabID,
			title:   baseCmd,
			command: cmdName,
			args:    cmdArgs,
			master:  master,
			child:   child,
			vt:      vt.NewVT(cols, childRows),
			done:    make(chan struct{}),
		}
		nextTabID++

		go func(t *sessionTab) {
			defer close(t.done)
			buf := make([]byte, 4096)
			for {
				n, readErr := t.master.Read(buf)
				if n > 0 {
					sessionMu.Lock()
					_, _ = t.vt.Write(buf[:n])
					if len(tabs) > 0 && activeIndex < len(tabs) && tabs[activeIndex] == t {
						_, _ = out.Write(buf[:n])
						if len(tabs) > 1 {
							drawSessionTabBar(out, tabs, activeIndex, cols, rows)
						}
					}
					sessionMu.Unlock()
				}
				if readErr != nil {
					break
				}
			}
			waitErr := t.child.Wait()
			var exitCode int
			if waitErr != nil {
				var exitErr *exec.ExitError
				if errors.As(waitErr, &exitErr) {
					exitCode = exitErr.ExitCode()
				}
			}
			t.exitCode = exitCode
			t.waitErr = waitErr
			tabExitChan <- t
		}(tab)

		return tab, nil
	}

	initialTab, err := spawnTab(command, args...)
	if err != nil {
		return err
	}
	tabs = append(tabs, initialTab)

	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)
	defer signal.Stop(sigWinch)
	go func() {
		for range sigWinch {
			sessionMu.Lock()
			if inFile, ok := in.(*os.File); ok && term.IsTerminal(int(inFile.Fd())) {
				if w, h, err := term.GetSize(int(inFile.Fd())); err == nil && w > 0 && h > 0 {
					cols, rows = w, h
					childRows := rows
					if len(tabs) > 1 {
						childRows = rows - 1
					}
					for _, t := range tabs {
						setPTYSize(t.master, cols, childRows)
						t.vt.Resize(cols, childRows)
					}
					repaintActiveTab(out, tabs, activeIndex, cols, rows)
				}
			}
			sessionMu.Unlock()
		}
	}()

	var savedShots []string
	var shotErrors []error
	var shotMu sync.Mutex
	saveScreenshot := func() error {
		sessionMu.Lock()
		if len(tabs) == 0 || activeIndex >= len(tabs) {
			sessionMu.Unlock()
			return errors.New("no active tab for screenshot")
		}
		curTab := tabs[activeIndex]
		content := curTab.vt.ANSI()
		cmdName := curTab.command
		sessionMu.Unlock()

		baseCmd := filepath.Base(cmdName)
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
					sessionMu.Lock()
					if len(tabs) == 0 {
						sessionMu.Unlock()
						return
					}
					activeMaster := tabs[activeIndex].master

					// Lone Esc
					if len(pending) == 1 && pending[0] == 0x1b {
						prefixMode = false
						_, _ = activeMaster.Write(pending)
						pending = nil
						sessionMu.Unlock()
						break
					}

					// Bracketed paste
					if paste, used, ok := loom.DecodePaste(pending); ok {
						_ = paste
						_, _ = activeMaster.Write(pending[:used])
						pending = pending[used:]
						prefixMode = false
						sessionMu.Unlock()
						continue
					}

					key, used, ok := loom.ScanKey(pending)
					if !ok {
						sessionMu.Unlock()
						break
					}

					if prefixMode {
						switch {
						case key.Is(sessionDefs.PrefixBindings.Screenshot...):
							prefixMode = false
							sessionMu.Unlock()
							if err := saveScreenshot(); err != nil {
								shotMu.Lock()
								shotErrors = append(shotErrors, err)
								shotMu.Unlock()
							}
						case key.Is(sessionDefs.PrefixBindings.Quit...):
							sessionMu.Unlock()
							close(sessionQuitChan)
							return
						case key.Is(sessionDefs.PrefixBindings.SendPrefix...):
							_, _ = activeMaster.Write(pending[:used])
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.NewTab...):
							shell := os.Getenv("SHELL")
							if shell == "" {
								shell = "/bin/sh"
							}
							if len(tabs) == 1 {
								setPTYSize(tabs[0].master, cols, rows-1)
								tabs[0].vt.Resize(cols, rows-1)
							}
							newTab, err := spawnTab(shell)
							if err == nil {
								tabs = append(tabs, newTab)
								activeIndex = len(tabs) - 1
								repaintActiveTab(out, tabs, activeIndex, cols, rows)
							}
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.NextTab...):
							if len(tabs) > 1 {
								activeIndex = (activeIndex + 1) % len(tabs)
								repaintActiveTab(out, tabs, activeIndex, cols, rows)
							}
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.PreviousTab...):
							if len(tabs) > 1 {
								activeIndex = (activeIndex - 1 + len(tabs)) % len(tabs)
								repaintActiveTab(out, tabs, activeIndex, cols, rows)
							}
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.CloseTab...):
							if len(tabs) > 0 {
								t := tabs[activeIndex]
								killProcessGroup(t.child.Process.Pid)
								_ = t.master.Close()
							}
							prefixMode = false
							sessionMu.Unlock()
						default:
							// Check select_tab bindings by index
							selected := -1
							for idx, binding := range sessionDefs.PrefixBindings.SelectTab {
								if key.Is(binding) {
									selected = idx
									break
								}
							}
							if selected >= 0 {
								if selected < len(tabs) {
									activeIndex = selected
									repaintActiveTab(out, tabs, activeIndex, cols, rows)
								}
							} else {
								_, _ = activeMaster.Write(pending[:used])
							}
							prefixMode = false
							sessionMu.Unlock()
						}
						pending = pending[used:]
						continue
					}

					// Direct hotkey matching
					switch {
					case key.Is(sessionDefs.Bindings.Screenshot...):
						sessionMu.Unlock()
						if err := saveScreenshot(); err != nil {
							shotMu.Lock()
							shotErrors = append(shotErrors, err)
							shotMu.Unlock()
						}
					case key.Is(sessionDefs.Bindings.Quit...):
						sessionMu.Unlock()
						close(sessionQuitChan)
						return
					case key.Is(sessionDefs.Bindings.NewTab...):
						shell := os.Getenv("SHELL")
						if shell == "" {
							shell = "/bin/sh"
						}
						if len(tabs) == 1 {
							setPTYSize(tabs[0].master, cols, rows-1)
							tabs[0].vt.Resize(cols, rows-1)
						}
						newTab, err := spawnTab(shell)
						if err == nil {
							tabs = append(tabs, newTab)
							activeIndex = len(tabs) - 1
							repaintActiveTab(out, tabs, activeIndex, cols, rows)
						}
						sessionMu.Unlock()
					case key.Is(sessionDefs.Prefix):
						prefixMode = true
						sessionMu.Unlock()
					default:
						_, _ = activeMaster.Write(pending[:used])
						sessionMu.Unlock()
					}
					pending = pending[used:]
				}
			}
			if err != nil {
				return
			}
		}
	}()

	var finalExitCode int
	var finalWaitErr error

	for {
		select {
		case <-ctx.Done():
			sessionMu.Lock()
			for _, t := range tabs {
				killProcessGroup(t.child.Process.Pid)
			}
			sessionMu.Unlock()
			return ctx.Err()
		case <-sessionQuitChan:
			sessionMu.Lock()
			for _, t := range tabs {
				killProcessGroup(t.child.Process.Pid)
			}
			sessionMu.Unlock()
			goto done
		case exitedTab := <-tabExitChan:
			sessionMu.Lock()
			idx := -1
			for i, t := range tabs {
				if t == exitedTab {
					idx = i
					break
				}
			}
			if idx >= 0 {
				exitedTab.master.Close()
				tabs = append(tabs[:idx], tabs[idx+1:]...)
				if len(tabs) == 0 {
					finalExitCode = exitedTab.exitCode
					finalWaitErr = exitedTab.waitErr
					sessionMu.Unlock()
					goto done
				}
				if activeIndex >= len(tabs) {
					activeIndex = len(tabs) - 1
				}
				if len(tabs) == 1 {
					setPTYSize(tabs[0].master, cols, rows)
					tabs[0].vt.Resize(cols, rows)
					_, _ = out.Write([]byte("\x1b[r"))
				}
				repaintActiveTab(out, tabs, activeIndex, cols, rows)
			}
			sessionMu.Unlock()
		}
	}

done:
	// Restore outer terminal cooked mode before printing status messages to errOut
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

	if finalWaitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(finalWaitErr, &exitErr) {
			finalExitCode = exitErr.ExitCode()
		} else {
			return finalWaitErr
		}
	}

	if finalExitCode != 0 {
		return sessionExitError{code: finalExitCode}
	}
	return nil
}

func killProcessGroup(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGHUP)
	_ = syscall.Kill(pid, syscall.SIGHUP)
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	_ = syscall.Kill(pid, syscall.SIGTERM)
}

