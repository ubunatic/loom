// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
	"ubunatic.com/loom/terminal"
)

type clientRead struct {
	message ipcMessage
	err     error
}

type clientInput struct {
	data []byte
	err  error
}

// attachSession connects a foreground client to a running session.
func attachSession(ctx context.Context, name string, in io.Reader, out io.Writer) error {
	path, err := sessionSocketPath(name)
	if err != nil {
		return err
	}
	conn, err := net.Dial("unix", path)
	if err != nil {
		return fmt.Errorf("connect to session %q: %w", name, err)
	}
	defer conn.Close()
	clientCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	cols, rows := 80, 24
	var state *term.State
	fd := -1
	if file, ok := in.(*os.File); ok {
		fd = int(file.Fd())
		if term.IsTerminal(fd) {
			if width, height, sizeErr := term.GetSize(fd); sizeErr == nil && width > 0 && height > 0 {
				cols, rows = width, height
			}
			state, _ = term.MakeRaw(fd)
		}
	} else if file, ok := out.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		if width, height, sizeErr := term.GetSize(int(file.Fd())); sizeErr == nil && width > 0 && height > 0 {
			cols, rows = width, height
		}
	}
	restorer := terminal.NewRestorer(fd, state, out)
	defer restorer.Restore()
	restorer.OnRestore(func() {
		_, _ = io.WriteString(out, "\x1b[r\x1b[2K")
	})
	if err := writeMessage(conn, ipcMessage{Type: msgAttach, Name: name, Cols: cols, Rows: rows}); err != nil {
		return err
	}

	readCh := make(chan clientRead, 1)
	go func() {
		for {
			message, readErr := readMessage(conn)
			select {
			case readCh <- clientRead{message: message, err: readErr}:
			case <-clientCtx.Done():
				return
			}
			if readErr != nil {
				return
			}
		}
	}()
	inputCh := make(chan clientInput, 1)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := in.Read(buf)
			if n > 0 {
				data := append([]byte(nil), buf[:n]...)
				select {
				case inputCh <- clientInput{data: data}:
				case <-clientCtx.Done():
					return
				}
			}
			if readErr != nil {
				select {
				case inputCh <- clientInput{err: readErr}:
				case <-clientCtx.Done():
				}
				return
			}
		}
	}()

	sigWinch := make(chan os.Signal, 1)
	sigHangup := make(chan os.Signal, 1)
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)
	signal.Notify(sigHangup, syscall.SIGHUP)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigWinch)
	defer signal.Stop(sigHangup)
	defer signal.Stop(sigTerm)
	for {
		select {
		case <-ctx.Done():
			_ = writeMessage(conn, ipcMessage{Type: msgDetach})
			return ctx.Err()
		case <-sigHangup:
			_ = writeMessage(conn, ipcMessage{Type: msgDetach})
			return nil
		case sig := <-sigTerm:
			if unixSignal, ok := sig.(syscall.Signal); ok {
				_ = writeMessage(conn, ipcMessage{Type: msgKill, Code: 128 + int(unixSignal)})
			}
		case <-sigWinch:
			if fd >= 0 && term.IsTerminal(fd) {
				if width, height, sizeErr := term.GetSize(fd); sizeErr == nil && width > 0 && height > 0 {
					cols, rows = width, height
					if err := writeMessage(conn, ipcMessage{Type: msgResize, Cols: cols, Rows: rows}); err != nil {
						return err
					}
				}
			}
		case input := <-inputCh:
			if len(input.data) > 0 {
				if err := writeMessage(conn, ipcMessage{Type: msgInput, Data: input.data}); err != nil {
					return err
				}
			}
			if input.err != nil {
				if !errors.Is(input.err, io.EOF) {
					return input.err
				}
				_ = writeMessage(conn, ipcMessage{Type: msgDetach})
				return nil
			}
		case result := <-readCh:
			if result.err != nil {
				if errors.Is(result.err, net.ErrClosed) || errors.Is(result.err, io.EOF) {
					return nil
				}
				return result.err
			}
			switch result.message.Type {
			case msgOutput:
				if _, err := out.Write(result.message.Data); err != nil {
					return err
				}
			case msgStderr:
				if _, err := out.Write(result.message.Data); err != nil {
					return err
				}
			case msgStatus:
				if result.message.Status == "modes" {
					restorer.SetAltScreen(result.message.AltScreen)
					restorer.SetKittyKeyboard(result.message.KittyKeyboard)
				}
			case msgDetached:
				return nil
			case msgExit:
				if result.message.Code != 0 {
					return ExitError{code: result.message.Code}
				}
				return nil
			case msgError:
				return errors.New(result.message.Error)
			default:
				return fmt.Errorf("unexpected server message %q", result.message.Type)
			}
		}
	}
}
