// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiviewer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// RecordCommand runs command in a private PTY sized like the hosting terminal,
// captures output until
// delay (or until the command exits), terminates only its process group, and
// waits for the group to be reaped before returning.
func RecordCommand(ctx context.Context, out io.Writer, delay time.Duration, command string, args ...string) error {
	cols, rows := recordingSize()
	master, slave, err := openRecorderPTY(cols, rows)
	if err != nil {
		return err
	}
	defer master.Close()
	cmd := exec.Command(command, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = slave, slave, slave
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := cmd.Start(); err != nil {
		slave.Close()
		return fmt.Errorf("ansiviewer: start recording command: %w", err)
	}
	if err := slave.Close(); err != nil {
		return fmt.Errorf("ansiviewer: close recording slave: %w", err)
	}
	var dataMu sync.Mutex
	var data []byte
	dataDone := make(chan struct{})
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := master.Read(buf)
			if n > 0 {
				dataMu.Lock()
				data = append(data, buf[:n]...)
				dataMu.Unlock()
			}
			if err != nil {
				break
			}
		}
		close(dataDone)
	}()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	timer := time.NewTimer(delay)
	defer timer.Stop()
	var snapshot []byte
	select {
	case <-ctx.Done():
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		<-done
		return ctx.Err()
	case <-done:
		<-dataDone
		dataMu.Lock()
		snapshot = append([]byte(nil), data...)
		dataMu.Unlock()
	case <-timer.C:
		time.Sleep(30 * time.Millisecond)
		dataMu.Lock()
		snapshot = append([]byte(nil), data...)
		dataMu.Unlock()
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
			return fmt.Errorf("ansiviewer: stop recording command: %w", err)
		}
		<-done
	}
	_ = master.Close()
	<-dataDone
	if len(snapshot) == 0 {
		dataMu.Lock()
		snapshot = append([]byte(nil), data...)
		dataMu.Unlock()
	}
	if len(snapshot) == 0 {
		return fmt.Errorf("ansiviewer: recording command produced no terminal output")
	}
	// The command's output is already a terminal recording. Replaying it
	// through a text-only VT would discard SGR colors and alter line wrapping.
	_, err = out.Write(stripRecordingState(snapshot))
	return err
}

var recordingOSC = regexp.MustCompile("\\x1b\\](?:0|7);[^\\x07\\x1b]*(?:\\x07|\\x1b\\\\)")

func stripRecordingState(data []byte) []byte {
	clean := recordingOSC.ReplaceAll(data, nil)
	for _, sequence := range []string{
		"\x1b[?1h", "\x1b[?1l", "\x1b=", "\x1b>",
		"\x1b[?1001s", "\x1b[?1001r", "\x1b[?1002h", "\x1b[?1002l",
		"\x1b[?1003h", "\x1b[?1003l", "\x1b[?1006h", "\x1b[?1006l",
		"\x1b[?1015h", "\x1b[?1015l", "\x1b[?1049h", "\x1b[?1049l",
		"\x1b[?2004h", "\x1b[?2004l", "\x1b[22;0;0t", "\x1b[23;0;0t", "\x1b[4l",
	} {
		clean = bytes.ReplaceAll(clean, []byte(sequence), nil)
	}
	return clean
}

func recordingSize() (int, int) {
	for _, fd := range []uintptr{os.Stdout.Fd(), os.Stdin.Fd()} {
		ws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
		if err == nil && ws.Col > 0 && ws.Row > 0 {
			return int(ws.Col), int(ws.Row)
		}
	}
	return 100, 30
}

func openRecorderPTY(cols, rows int) (*os.File, *os.File, error) {
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("ansiviewer: open PTY: %w", err)
	}
	fd := int(master.Fd())
	if err = unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("ansiviewer: unlock PTY: %w", err)
	}
	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("ansiviewer: get PTY number: %w", err)
	}
	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), os.O_RDWR, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("ansiviewer: open PTY slave: %w", err)
	}
	term, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
	if err != nil {
		slave.Close()
		master.Close()
		return nil, nil, fmt.Errorf("ansiviewer: get PTY termios: %w", err)
	}
	term.Oflag &^= unix.OPOST
	if err = unix.IoctlSetTermios(int(slave.Fd()), unix.TCSETS, term); err != nil {
		slave.Close()
		master.Close()
		return nil, nil, fmt.Errorf("ansiviewer: set PTY termios: %w", err)
	}
	if err = unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ, &unix.Winsize{Col: uint16(cols), Row: uint16(rows)}); err != nil {
		slave.Close()
		master.Close()
		return nil, nil, fmt.Errorf("ansiviewer: size PTY: %w", err)
	}
	return master, slave, nil
}
