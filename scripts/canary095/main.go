// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Canary for issue 095 (docs/Canary.md). It probes the mechanisms the PTY test
// view mode depends on, before feature code is built on them:
//
//	A1 can the harness attach the observer terminal (/dev/tty) while the test
//	   child runs in its own session with its own controlling terminal?
//	A2 can it enter raw mode and restore it on SIGINT?
//	A3 does a full-duplex relay work: child output to the observer, observer
//	   keystrokes to the child?
//	A4 can chrome rows (scenario name + prompt) survive, by rendering from the
//	   cell model instead of passing bytes through?
//
// Run it inside a real terminal (tmux or foot). Findings go to stdout and to
// /tmp/canary095.log. Inject a keystroke from another window, or press q.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"

	"ubunatic.com/loom/vt"
)

const (
	cols    = 80
	rows    = 24
	chrome  = 1
	logPath = "/tmp/canary095.log"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-child" {
		runChild()
		return
	}
	if err := runCanary(); err != nil {
		report("FAIL: %v", err)
		os.Exit(1)
	}
}

// report writes a finding to the observer pane and to the canary log.
func report(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	fmt.Println(line)
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(line + "\n")
}

func runCanary() error {
	report("A1 stdout is a terminal: %v", term.IsTerminal(1))
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open /dev/tty: %w", err)
	}
	defer tty.Close()

	fd := int(tty.Fd())
	before, err := term.GetState(fd)
	if err != nil {
		return fmt.Errorf("observer termios: %w", err)
	}
	beforeTermios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return fmt.Errorf("observer termios: %w", err)
	}

	restoredOnSignal := false
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		<-signals
		_ = term.Restore(fd, before)
		restoredOnSignal = true
	}()

	if _, err := term.MakeRaw(fd); err != nil {
		return fmt.Errorf("raw mode: %w", err)
	}
	report("A2 raw mode on /dev/tty: echo off=%v, canonical off=%v",
		flagOff(fd, unix.ECHO), flagOff(fd, unix.ICANON))

	master, slave, err := openPTY()
	if err != nil {
		return err
	}
	defer master.Close()
	defer slave.Close()

	child := exec.Command(os.Args[0], "-child")
	child.Stdin, child.Stdout, child.Stderr = slave, slave, slave
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := child.Start(); err != nil {
		return fmt.Errorf("start child: %w", err)
	}

	observer := vt.NewVT(paneSize(fd))
	model := vt.NewVT(cols, rows)
	_, _ = tty.WriteString("\x1b[?1049h")

	key := relay(tty, master, observer, model)
	if err := child.Wait(); err != nil {
		return fmt.Errorf("child: %w", err)
	}

	_ = term.Restore(fd, before)
	after, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return fmt.Errorf("observer termios after restore: %w", err)
	}
	report("A6 final restore identical: %v (signal path used: %v)", *after == *beforeTermios, restoredOnSignal)
	report("A4 chrome row after render: %q", rowText(observer, 0))
	report("A4 child frame row through render: %q", rowText(observer, chrome))
	report("A3 key seen by relay: %q", key)
	return nil
}

// relay mirrors child output into the observer cell model and forwards observer
// keystrokes to the child. It renders chrome plus the frame from the model.
func relay(tty *os.File, master *os.File, observer, model *vt.VT) string {
	deadline := time.Now().Add(10 * time.Second)
	key := ""
	for time.Now().Before(deadline) {
		fds := []int{int(master.Fd()), int(tty.Fd())}
		ready, err := readable(fds, 200*time.Millisecond)
		if err != nil {
			report("A3 select failed: %v", err)
			return key
		}
		if ready[int(master.Fd())] {
			buf := make([]byte, 4096)
			n, err := master.Read(buf)
			if n > 0 {
				writeVT(model, string(buf[:n]))
				render(tty, observer, model)
			}
			if err != nil {
				return key
			}
		}
		if ready[int(tty.Fd())] {
			buf := make([]byte, 64)
			n, err := tty.Read(buf)
			if n > 0 {
				key += string(buf[:n])
				_, _ = master.Write(buf[:n])
				if strings.Contains(key, "q") {
					return key
				}
			}
			if err != nil {
				return key
			}
		}
	}
	report("A3 no keystroke arrived within 10s")
	return key
}

// render draws chrome on row 0 and the child frame below it from the cell model.
func render(tty *os.File, observer, model *vt.VT) {
	observer.ResetParser()
	writeVT(observer, "\x1b[H\x1b[2J")
	writeVT(observer, chromeLine())
	for y, row := range model.Cells() {
		writeVT(observer, fmt.Sprintf("\x1b[%d;1H", y+chrome+1))
		for _, cell := range row {
			writeVT(observer, cell.Style.SGR()+glyph(cell))
		}
	}
	_, _ = tty.WriteString(observer.Repaint())
}

func writeVT(v *vt.VT, s string) {
	_, _ = v.Write([]byte(s))
}

func chromeLine() string {
	return "\x1b[1;1H\x1b[1;34m" + "095 CANARY chrome row (must survive the child frame)"
}

func glyph(cell vt.Cell) string {
	if cell.Text != "" {
		return cell.Text
	}
	if cell.Rune == 0 {
		return " "
	}
	return string(cell.Rune)
}

func readable(fds []int, timeout time.Duration) (map[int]bool, error) {
	set := &unix.FdSet{}
	highest := 0
	for _, fd := range fds {
		set.Set(fd)
		if fd > highest {
			highest = fd
		}
	}
	tv := unix.NsecToTimeval(timeout.Nanoseconds())
	if _, err := unix.Select(highest+1, set, nil, nil, &tv); err != nil {
		if errors.Is(err, syscall.EINTR) {
			// Go's runtime can interrupt Select with a signal; a spurious EINTR must
			// not end the relay.
			return nil, nil
		}
		return nil, err
	}
	ready := make(map[int]bool)
	for _, fd := range fds {
		ready[fd] = set.IsSet(fd)
	}
	return ready, nil
}

func flagOff(fd int, flag uint32) bool {
	state, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return false
	}
	return state.Lflag&flag == 0
}

func openPTY() (*os.File, *os.File, error) {
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
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
	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), os.O_RDWR, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("open slave: %w", err)
	}
	ws := &unix.Winsize{Col: uint16(cols), Row: uint16(rows)}
	if err := unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ, ws); err != nil {
		master.Close()
		slave.Close()
		return nil, nil, fmt.Errorf("winsize: %w", err)
	}
	return master, slave, nil
}

// runChild runs in its own session with its own controlling terminal, reports
// what it sees, echoes keystrokes, and quits on q.
func runChild() {
	own := resolve("/proc/self/fd/0")
	ownTTY := childTTY()
	report("A1 child fd0 -> %s, child /dev/tty -> %s (same as fd0: %v)", own, ownTTY, own == ownTTY)
	report("A2 child raw mode: %v", flagOff(0, unix.ECHO))
	report("A3 child ready, waiting for a keystroke")
	buf := make([]byte, 64)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			key := string(buf[:n])
			report("A3 child got key %q", key)
			if strings.Contains(key, "q") {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func resolve(path string) string {
	target, err := os.Readlink(path)
	if err != nil {
		return "unresolved"
	}
	return target
}

// childTTY resolves the child's controlling terminal by opening /dev/tty.
func childTTY() string {
	file, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "no controlling terminal: " + err.Error()
	}
	defer file.Close()
	return resolve(fmt.Sprintf("/proc/self/fd/%d", int(file.Fd())))
}

func paneSize(fd int) (int, int) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return cols, rows + chrome
	}
	return int(ws.Col), int(ws.Row)
}

func rowText(v *vt.VT, y int) string {
	rows := v.Screen()
	if y >= len(rows) {
		return ""
	}
	return strings.TrimSpace(rows[y])
}
