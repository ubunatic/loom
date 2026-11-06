// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"ubunatic.com/loom/internal/clitest"
	"ubunatic.com/loom/internal/ptytest"
)

func TestServeSessionChildCanHandleHangup(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	bin := clitest.BuildLoomBinary(t)
	name := "child-hangup"
	t.Cleanup(func() { _ = Kill(name) })
	// A shell cannot install a trap for a signal inherited as ignored. Run the
	// real CLI so the daemon's signal handling is isolated from the test runner.
	child := `printf 'HANGUP_READY\n'
read -r release
trap 'printf "HANGUP_HANDLED\n"' HUP
kill -HUP $$
read -r release`
	client := ptytest.Start(t, 80, 24, bin, "session", "-s", name, "--", "/bin/sh", "-c", child)
	client.WaitFor("HANGUP_READY", 5*time.Second)
	status, err := SessionStatus(name)
	if err != nil || status.PID <= 0 || status.PID == os.Getpid() {
		t.Fatalf("expected a separate daemon: %#v, %v", status, err)
	}
	if err := syscall.Kill(status.PID, syscall.SIGHUP); err != nil {
		t.Fatalf("hang up daemon: %v", err)
	}
	client.Send("go\n")
	client.WaitForUpdate("HANGUP_HANDLED", 5*time.Second)
	if after, err := SessionStatus(name); err != nil || after.PID != status.PID {
		t.Fatalf("daemon did not survive SIGHUP: %#v, %v", after, err)
	}
	client.Send("done\n")
	if err := client.Wait(5 * time.Second * ptytest.RaceFactor); err != nil {
		t.Fatalf("child could not handle SIGHUP: %v", err)
	}
}

func TestServeSessionInteractiveShellClose(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	t.Setenv("SHELL", "/bin/sh")
	t.Setenv("ENV", "") // No host interactive-shell startup scripts.
	bin := clitest.BuildLoomBinary(t)
	name := "interactive-close"
	t.Cleanup(func() { _ = Kill(name) })
	primary := `stty -echo
printf 'PRIMARY_READY\n'
read -r line
test "$line" = done`
	client := ptytest.Start(t, 80, 24, bin, "session", "-s", name, "--", "/bin/sh", "-c", primary)
	client.WaitFor("PRIMARY_READY", 5*time.Second)
	status, err := SessionStatus(name)
	if err != nil || status.PID <= 0 || status.PID == os.Getpid() {
		t.Fatalf("expected a separate daemon: %#v, %v", status, err)
	}
	if err := syscall.Kill(status.PID, syscall.SIGHUP); err != nil {
		t.Fatalf("hang up daemon: %v", err)
	}
	client.SendRaw([]byte{14}) // ctrl-n opens $SHELL on its own controlling PTY.
	client.WaitFor("2: ", 5*time.Second)
	// Verify that this is an interactive shell, without trapping or overriding
	// its default HUP/TERM handling. Split the marker to avoid matching input echo.
	client.Send(`stty -echo
if test "${-#*i}" != "$-"
then printf 'INTERACTIVE_%s\n' READY
else exit 42
fi
`)
	client.WaitForUpdate("INTERACTIVE_READY", 5*time.Second)
	pid, err := strconv.Atoi(childReply(t, client, `printf 'SHELL_PID:%s\n' "$$"`, "SHELL_PID:"))
	if err != nil || pid <= 0 || pid == os.Getpid() {
		t.Fatalf("interactive shell PID = %d, %v", pid, err)
	}
	closed := false
	t.Cleanup(func() {
		if !closed {
			// A failed close must not leave the immune shell behind.
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})
	client.SendRaw([]byte{2, 'x'})
	client.WaitForUpdateState(5*time.Second*ptytest.RaceFactor, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ")
	})
	closed = true
	after, err := SessionStatus(name)
	if err != nil || after.PID != status.PID || after.Tabs != 1 || after.Status != "attached" {
		t.Fatalf("daemon after HUP and tab close = %#v, %v", after, err)
	}
	client.Send("done\n")
	if err := client.Wait(5 * time.Second * ptytest.RaceFactor); err != nil {
		t.Fatalf("primary tab did not exit cleanly: %v", err)
	}
}
