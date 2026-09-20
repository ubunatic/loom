// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiviewer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
	"golang.org/x/sys/unix"
)

// RecordCommand runs command in a private 100x30 PTY, captures output until
// delay (or until the command exits), terminates only its process group, and
// waits for the group to be reaped before returning.
func RecordCommand(ctx context.Context, out io.Writer, delay time.Duration, command string, args ...string) error {
	master, slave, err := openRecorderPTY()
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
	vt := ptytest.NewVT(100, 30)
	if _, err := vt.Write(snapshot); err != nil {
		return fmt.Errorf("ansiviewer: replay recording: %w", err)
	}
	_, err = io.WriteString(out, vt.Text()+"\n")
	return err
}

func openRecorderPTY() (*os.File, *os.File, error) {
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
	if err = unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ, &unix.Winsize{Col: 100, Row: 30}); err != nil {
		slave.Close()
		master.Close()
		return nil, nil, fmt.Errorf("ansiviewer: size PTY: %w", err)
	}
	return master, slave, nil
}
