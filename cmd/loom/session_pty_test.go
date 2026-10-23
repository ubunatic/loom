// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
	"ubunatic.com/loom/vt"
)

func TestSessionPassThroughIO(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	// Run bash echoing a prompt and waiting for input
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'READY\\n'; read -r line; printf 'ECHO:%s\\n' \"$line\"")
	session.WaitFor("READY", time.Second*5)

	session.Send("hello-session\n")
	session.WaitFor("ECHO:hello-session", time.Second*5)
}

func TestSessionExitCodePassThrough(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)

	// Exit code 0
	s0 := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "exit 0")
	if err := s0.Wait(time.Second * 5); err != nil {
		t.Fatalf("expected exit 0, got err: %v", err)
	}

	// Exit code 42
	s42 := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "exit 42")
	err := s42.Wait(time.Second * 5)
	if err == nil {
		t.Fatal("expected non-zero exit, got nil")
	}
	if !strings.Contains(err.Error(), "42") {
		t.Fatalf("expected exit code 42, got err: %v", err)
	}
}

func TestSessionWindowResize(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "trap 'printf \"WINCH:%s\\n\" \"$(stty size 2>/dev/null)\"' WINCH; printf 'READY\\n'; while true; do sleep 1; done")
	session.WaitFor("READY", time.Second*5)

	session.Resize(50, 15)
	session.WaitFor("WINCH:14 50", time.Second*5)
	session.SendRaw([]byte{17}) // ctrl-q
	_ = session.Wait(time.Second * 5)
}

func TestSessionDirectHotkeyScreenshot(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	dir := t.TempDir()

	// Run session inside temporary directory so screenshot is created there
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf '\\033[1;31mHELLO SCREENSHOT\\033[0m\\n'; read -r line; exit 0")
	session.WaitFor("HELLO SCREENSHOT", time.Second*5)

	// Send ctrl-p (0x10) to trigger screenshot
	session.SendRaw([]byte{16})

	// Allow disk write
	time.Sleep(200 * time.Millisecond)

	// Finish child
	session.Send("done\n")
	_ = session.Wait(time.Second * 5)

	// Find the created screenshot file
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var shotFile string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "loom-bash-") && strings.HasSuffix(e.Name(), ".ansi") {
			shotFile = filepath.Join(dir, e.Name())
			break
		}
	}
	if shotFile == "" {
		t.Fatalf("screenshot file not found in %s; entries: %v", dir, entries)
	}

	data, err := os.ReadFile(shotFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "HELLO SCREENSHOT") {
		t.Fatalf("screenshot content missing text: %q", content)
	}
	if !strings.Contains(content, "\x1b[0;1;31mHELLO SCREENSHOT\x1b[0m") {
		t.Fatalf("screenshot content missing SGR attributes: %q", content)
	}
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("screenshot line count = %d, want 1 trimmed row; lines: %q", len(lines), lines)
	}

	// Verify loom measure and loom eval accept the screenshot
	if err := execute([]string{"measure", shotFile}, os.Stdout); err != nil {
		t.Fatalf("loom measure failed on screenshot: %v", err)
	}
	if err := execute([]string{"eval", shotFile}, os.Stdout); err != nil {
		t.Fatalf("loom eval failed on screenshot: %v", err)
	}
}

func TestSessionDirectHotkeyQuit(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'RUNNING\\n'; while true; do sleep 1; done")
	session.WaitFor("RUNNING", time.Second*5)

	// Send ctrl-q (0x11)
	session.SendRaw([]byte{17})

	// Wait for session to terminate
	_ = session.Wait(time.Second * 5)
}

func TestSessionPrefixHotkeys(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	dir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Test prefix then 's' (screenshot)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'PREFIX TEST\\n'; read -r line; exit 0")
	session.WaitFor("PREFIX TEST", time.Second*5)

	// Send ctrl-b (2) followed by 's'
	session.SendRaw([]byte{2, 's'})
	time.Sleep(200 * time.Millisecond)

	// Exit child
	session.Send("done\n")
	_ = session.Wait(time.Second * 5)

	entries, _ := os.ReadDir(dir)
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "loom-bash-") && strings.HasSuffix(e.Name(), ".ansi") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("prefix 's' screenshot not found; entries: %v", entries)
	}

	// Test prefix then 'q' (quit)
	sessionQ := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'PREFIX Q\\n'; while true; do sleep 1; done")
	sessionQ.WaitFor("PREFIX Q", time.Second*5)
	// Send ctrl-b followed by 'q'
	sessionQ.SendRaw([]byte{2, 'q'})
	_ = sessionQ.Wait(time.Second * 5)

	// Test prefix twice sends ctrl-b to child
	sessionSend := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "read -r -n 1 char; printf 'RECEIVED:%d\\n' \"$(printf '%d' \"'$char'\")\"")
	// Send ctrl-b twice
	sessionSend.SendRaw([]byte{2, 2})
	sessionSend.WaitFor("RECEIVED:2", time.Second*5)
}

func TestSessionCtrlRightBracketPassesThrough(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "read -r -n 1 char; printf 'RECEIVED:%d\\n' \"$(printf '%d' \"'$char'\")\"")
	// Send ctrl-] (29) once; it must reach the child directly without prefix interception
	session.SendRaw([]byte{29})
	session.WaitFor("RECEIVED:29", time.Second*5)
}

func TestSessionHelpText(t *testing.T) {
	cmd := sessionCommand()
	if !strings.Contains(cmd.Long, "Prefix ctrl-b followed by s saves a screenshot, q quits, and ctrl-b forwards ctrl-b to the child.") {
		t.Fatalf("unexpected Long help text: %q", cmd.Long)
	}
}

func TestSessionMultiByteInputNotSplit(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)

	// Child reads 5 bytes and prints hex
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'PROMPT\\n'; IFS= read -r -d '' -n 5 input; printf 'HEX:%s\\n' \"$(printf '%s' \"$input\" | od -An -tx1 | tr -d ' \\n')\"")
	session.WaitFor("PROMPT", time.Second*5)

	// Send an escape sequence containing byte 16 (ctrl-p): \x1b[16~
	// It should NOT trigger a screenshot; it should pass to child
	session.SendRaw([]byte("\x1b[16~"))
	session.WaitFor("HEX:1b5b31367e", time.Second*5)

	// Test bracketed paste containing ctrl-b (2), ctrl-p (16), ctrl-q (17), ctrl-] (29)
	sessionPaste := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "stty -ixon 2>/dev/null; printf 'PASTE_PROMPT\\n'; IFS= read -r line; printf 'PASTE:%s\\n' \"$(printf '%s' \"$line\" | tr '\\002\\020\\021\\035' 'BPQD')\"")
	sessionPaste.WaitFor("PASTE_PROMPT", time.Second*5)
	sessionPaste.SendRaw([]byte("\x1b[200~a\x02b\x10c\x11d\x1de\x1b[201~\n"))
	sessionPaste.WaitFor("PASTE:aBbPcQdDe", time.Second*5)
}

func TestSessionAlternateScreenScreenshot(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	dir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Child switches to alternate screen (\x1b[?1049h), prints text, waits
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf '\\033[?1049h\\033[1;1HALTSCREEN_HOTKEY\\033[0m\\n'; read -r line; exit 0")
	session.WaitFor("ALTSCREEN_HOTKEY", time.Second*5)

	// Send ctrl-p
	session.SendRaw([]byte{16})
	time.Sleep(200 * time.Millisecond)

	session.Send("done\n")
	_ = session.Wait(time.Second * 5)

	entries, _ := os.ReadDir(dir)
	var shotFile string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "loom-bash-") && strings.HasSuffix(e.Name(), ".ansi") {
			shotFile = filepath.Join(dir, e.Name())
			break
		}
	}
	if shotFile == "" {
		t.Fatalf("alt screen screenshot not found; entries: %v", entries)
	}

	data, err := os.ReadFile(shotFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "ALTSCREEN_HOTKEY") {
		t.Fatalf("screenshot missing alternate screen text: %q", string(data))
	}
}

func TestSessionLoneEscReachesChildAtOnce(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	// Child reads a single byte and prints its ASCII code
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'READY\\n'; IFS= read -r -d '' -n 1 key; printf 'KEY:%d\\n' \"$(printf '%d' \"'$key'\")\"")
	session.WaitFor("READY", time.Second*5)

	// Send lone Esc (27)
	session.SendRaw([]byte{27})

	// Must reach child immediately without waiting for a second byte
	session.WaitFor("KEY:27", time.Second*2)
}

func TestSessionFailedScreenshotReportedOnStderr(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	dir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Make directory read-only so screenshot creation fails
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'PROMPT\\n'; read -r line; exit 0")
	session.WaitFor("PROMPT", time.Second*5)

	// Send ctrl-p (trigger screenshot which should fail)
	session.SendRaw([]byte{16})
	time.Sleep(200 * time.Millisecond)

	session.Send("done\n")
	_ = session.Wait(time.Second * 5)

	// The stderr output must report the failure
	output := string(session.Raw())
	if !strings.Contains(output, "Error saving screenshot:") {
		t.Fatalf("expected error reported on stderr, raw output: %q", output)
	}
}

func TestSessionStderrCleanLineStart(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	dir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'PROMPT\\n'; read -r line; exit 0")
	session.WaitFor("PROMPT", time.Second*5)

	session.SendRaw([]byte{16}) // ctrl-p
	time.Sleep(200 * time.Millisecond)

	session.Send("done\n")
	_ = session.Wait(time.Second * 5)

	raw := string(session.Raw())
	if !strings.Contains(raw, "Saved screenshot: ") {
		t.Fatalf("expected screenshot notification, got raw: %q", raw)
	}
	// Verify that because raw mode was restored before printing, the notification line
	// follows a standard cooked newline \r\n (or starts at column 0).
	if !strings.Contains(raw, "\r\nSaved screenshot: ") && !strings.HasPrefix(raw, "Saved screenshot: ") {
		t.Fatalf("notification does not start at clean line start: %q", raw)
	}
}

func TestSessionNewTabCtrlN(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'TAB1_READY\\n'; read -r line; exit 0")
	session.WaitFor("TAB1_READY", time.Second*5)

	// Send ctrl-n (14) to open new tab running $SHELL
	session.SendRaw([]byte{14})

	// Wait for tab bar to appear
	session.WaitFor("2: ", time.Second*5)

	// In new tab, print TAB2_RUNNING
	session.Send("printf 'TAB2_RUNNING\\n'\n")
	session.WaitFor("TAB2_RUNNING", time.Second*5)

	// Close tab 2 with ctrl-b x (2, 'x')
	session.SendRaw([]byte{2, 'x'})

	// Wait for tab bar to disappear (only tab 1 remains)
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ")
	})

	// Finish tab 1
	session.Send("done\n")
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session did not exit cleanly: %v", err)
	}
}

func TestSessionTabSwitchPrefixNAndP(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'SCREEN1_CONTENT\\n'; read -r line; exit 0")
	session.WaitFor("SCREEN1_CONTENT", time.Second*5)

	// Open tab 2 with ctrl-n
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	session.Send("printf 'SCREEN2_CONTENT\\n'\n")
	session.WaitFor("SCREEN2_CONTENT", time.Second*5)

	// Switch to previous tab with ctrl-b p (2, 'p')
	session.SendRaw([]byte{2, 'p'})
	session.WaitForUpdate("SCREEN1_CONTENT", time.Second*5)

	// Switch to next tab with ctrl-b n (2, 'n')
	session.SendRaw([]byte{2, 'n'})
	session.WaitForUpdate("SCREEN2_CONTENT", time.Second*5)

	// Close tab 2 and finish tab 1
	session.SendRaw([]byte{2, 'x'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ")
	})
	session.Send("done\n")
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session did not exit cleanly: %v", err)
	}
}

func TestSessionTabSwitchPrefixSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'TAB_ONE_MARKER\\n'; read -r line; exit 0")
	session.WaitFor("TAB_ONE_MARKER", time.Second*5)

	// Open tab 2 with ctrl-n
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	session.Send("printf 'TAB_TWO_MARKER\\n'\n")
	session.WaitFor("TAB_TWO_MARKER", time.Second*5)

	// Select tab 1 with ctrl-b 1 (2, '1')
	session.SendRaw([]byte{2, '1'})
	session.WaitForUpdate("TAB_ONE_MARKER", time.Second*5)

	// Select tab 2 with ctrl-b 2 (2, '2')
	session.SendRaw([]byte{2, '2'})
	session.WaitForUpdate("TAB_TWO_MARKER", time.Second*5)

	// Close tab 2 and finish tab 1
	session.SendRaw([]byte{2, 'x'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ")
	})
	session.Send("done\n")
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session did not exit cleanly: %v", err)
	}
}

func TestSessionPrefixNewTabAndCloseTab(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'TAB_KEEP\\n'; read -r line; exit 0")
	session.WaitFor("TAB_KEEP", time.Second*5)

	// Open tab 2 with prefix c (2, 'c')
	session.SendRaw([]byte{2, 'c'})
	session.WaitFor("2: ", time.Second*5)

	session.Send("printf 'TAB_TO_CLOSE\\n'\n")
	session.WaitFor("TAB_TO_CLOSE", time.Second*5)

	// Close tab 2 with prefix x (2, 'x')
	session.SendRaw([]byte{2, 'x'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ")
	})

	// Finish tab 1
	session.Send("done\n")
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session did not exit cleanly: %v", err)
	}
}

func TestSessionLastTabExitCode(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'TAB1_READY\\n'; read -r line; exit 42")
	session.WaitFor("TAB1_READY", time.Second*5)

	// Open tab 2 with ctrl-n
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// Exit tab 2 with code 0
	session.Send("exit 0\n")
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ")
	})

	// Finish tab 1 which exits 42
	session.Send("done\n")
	err := session.Wait(time.Second * 5)
	if err == nil {
		t.Fatal("expected exit code 42, got nil")
	}
	if !strings.Contains(err.Error(), "42") {
		t.Fatalf("expected error containing 42, got %v", err)
	}
}

func TestSessionAltScreenTabSwitch(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'MAIN_SCREEN_TAB\\n'; read -r line; exit 0")
	session.WaitFor("MAIN_SCREEN_TAB", time.Second*5)

	// Open tab 2 with ctrl-n
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// In tab 2, enter alt screen and print text
	session.Send("printf '\\033[?1049h\\033[1;1HALT_SCREEN_TAB\\033[0m\\n'\n")
	session.WaitFor("ALT_SCREEN_TAB", time.Second*5)

	// Switch to tab 1 (main screen) with ctrl-b p
	session.SendRaw([]byte{2, 'p'})
	session.WaitForUpdate("MAIN_SCREEN_TAB", time.Second*5)

	// Switch back to tab 2 (alt screen) with ctrl-b n
	session.SendRaw([]byte{2, 'n'})
	session.WaitForUpdate("ALT_SCREEN_TAB", time.Second*5)

	// Close tab 2 with ctrl-b x
	session.SendRaw([]byte{2, 'x'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ")
	})

	// Finish tab 1
	session.Send("done\n")
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session did not exit cleanly: %v", err)
	}
}

func TestSessionQuitErasesTabBarOnOuterScreen(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'RUNNING\\n'; while true; do sleep 1; done")
	session.WaitFor("RUNNING", time.Second*5)

	// Open tab 2 with ctrl-n so tab bar appears
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// Send ctrl-q (17) to quit session while bar is shown
	session.SendRaw([]byte{17})
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session wait: %v", err)
	}

	// The outer VT screen after exit must hold no tab bar text
	for y, row := range session.Screen() {
		if strings.Contains(row, "1: bash") || strings.Contains(row, "2: ") {
			t.Fatalf("stale tab bar found at row %d after exit: %q", y, row)
		}
	}
}

func TestSessionQuitInAltScreenErasesTabBarOnOuterScreen(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'MAIN\\n'; while true; do sleep 1; done")
	session.WaitFor("MAIN", time.Second*5)

	// Open tab 2 with ctrl-n
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// Enter alt screen in tab 2
	session.Send("printf '\\033[?1049h\\033[1;1HALT\\033[0m\\n'\n")
	session.WaitFor("ALT", time.Second*5)

	// Quit while in alt screen with ctrl-q (17)
	session.SendRaw([]byte{17})
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session wait: %v", err)
	}

	// The outer VT screen after exit must hold no tab bar text
	for y, row := range session.Screen() {
		if strings.Contains(row, "1: bash") || strings.Contains(row, "2: ") {
			t.Fatalf("stale tab bar found at row %d after exit: %q", y, row)
		}
	}
}

func TestSessionHideBarErasesBottomRow(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'PERSIST\\n'; while true; do sleep 1; done")
	session.WaitFor("PERSIST", time.Second*5)

	// Open tab 2 with ctrl-n so tab bar appears
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// Close tab 2 with ctrl-b x; the configured bar remains with one tab.
	session.SendRaw([]byte{2, 'x'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ") && strings.Contains(strings.Join(rows, "\n"), "1: bash")
	})

	if !strings.Contains(session.Screen()[len(session.Screen())-1], "1: bash") {
		t.Fatalf("single-tab bar missing after closing tab: %q", session.Screen()[len(session.Screen())-1])
	}

	session.SendRaw([]byte{17}) // ctrl-q
	_ = session.Wait(time.Second * 5)
}

func TestSessionSafeRestoreOnCtrlQ(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	// Fixture child enables mouse 1000+1006, focus 1004, bracketed paste 2004, and waits
	fixture := "printf '\\033[?1000h\\033[?1006h\\033[?1004h\\033[?2004hREADY\\n'; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("READY", time.Second*5)

	// Child turned tracking on; verify it was active
	if session.VT().MouseMode != 1000 || !session.VT().MouseSGR {
		t.Fatalf("expected child to enable mouse, got mode=%d sgr=%v", session.VT().MouseMode, session.VT().MouseSGR)
	}

	// Quit with ctrl-q (17)
	session.SendRaw([]byte{17})
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session wait: %v", err)
	}

	// The outer VT must have all modes off
	vt := session.VT()
	if vt.MouseMode != 0 || vt.MouseSGR {
		t.Fatalf("outer VT mouse mode left on after ctrl-q: mode=%d sgr=%v", vt.MouseMode, vt.MouseSGR)
	}
	if vt.FocusEvents {
		t.Fatalf("outer VT focus events left on after ctrl-q")
	}
	if vt.BracketedPaste {
		t.Fatalf("outer VT bracketed paste left on after ctrl-q")
	}
}

func TestSessionSafeRestoreOnChildExit(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	// Fixture child enables mouse 1000+1006, focus 1004, bracketed paste 2004, then exits directly
	fixture := "printf '\\033[?1000h\\033[?1006h\\033[?1004h\\033[?2004hEXITING\\n'; exit 0"
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", fixture)
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session wait: %v", err)
	}

	vt := session.VT()
	if vt.MouseMode != 0 || vt.MouseSGR {
		t.Fatalf("outer VT mouse mode left on after child exit: mode=%d sgr=%v", vt.MouseMode, vt.MouseSGR)
	}
	if vt.FocusEvents {
		t.Fatalf("outer VT focus events left on after child exit")
	}
	if vt.BracketedPaste {
		t.Fatalf("outer VT bracketed paste left on after child exit")
	}
}

func TestSessionSafeRestoreOnLastTabClosed(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	fixture := "printf '\\033[?1000h\\033[?1006h\\033[?1004h\\033[?2004hREADY\\n'; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("READY", time.Second*5)

	// Close the only tab with prefix x (ctrl-b x)
	session.SendRaw([]byte{2, 'x'})
	_ = session.Wait(time.Second * 5)

	vt := session.VT()
	if vt.MouseMode != 0 || vt.MouseSGR {
		t.Fatalf("outer VT mouse mode left on after prefix x: mode=%d sgr=%v", vt.MouseMode, vt.MouseSGR)
	}
	if vt.FocusEvents {
		t.Fatalf("outer VT focus events left on after prefix x")
	}
	if vt.BracketedPaste {
		t.Fatalf("outer VT bracketed paste left on after prefix x")
	}
}

func TestSessionSafeRestoreOnSIGTERM(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	fixture := "printf '\\033[?1000h\\033[?1006h\\033[?1004h\\033[?2004hREADY\\n'; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("READY", time.Second*5)

	// Send SIGTERM to loom session
	_ = session.Signal(os.Interrupt) // or syscall.SIGTERM
	_ = session.Wait(time.Second * 5)

	vt := session.VT()
	if vt.MouseMode != 0 || vt.MouseSGR {
		t.Fatalf("outer VT mouse mode left on after SIGTERM: mode=%d sgr=%v", vt.MouseMode, vt.MouseSGR)
	}
	if vt.FocusEvents {
		t.Fatalf("outer VT focus events left on after SIGTERM")
	}
	if vt.BracketedPaste {
		t.Fatalf("outer VT bracketed paste left on after SIGTERM")
	}
}

func TestSessionPerTabModes(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	// Tab 1 enables SGR mouse mode and stays alive
	fixture := "printf '\\033[?1000h\\033[?1006hTAB1_READY\\n'; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("TAB1_READY", time.Second*5)

	// Tab 1 turned mouse on in outer VT
	if session.VT().MouseMode != 1000 || !session.VT().MouseSGR {
		t.Fatalf("Tab 1 expected mouse on, got mode=%d sgr=%v", session.VT().MouseMode, session.VT().MouseSGR)
	}

	// Open Tab 2 with ctrl-n (14)
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// In Tab 2, mouse tracking in outer VT must be OFF
	if session.VT().MouseMode != 0 || session.VT().MouseSGR {
		t.Fatalf("Tab 2 expected mouse off, got mode=%d sgr=%v", session.VT().MouseMode, session.VT().MouseSGR)
	}

	// Switch back to Tab 1 using prefix 1 (ctrl-b 1) or alt-1 (\x1b1)
	session.SendRaw([]byte{2, '1'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return session.VT().MouseMode == 1000
	})

	if session.VT().MouseMode != 1000 || !session.VT().MouseSGR {
		t.Fatalf("Tab 1 switch-back expected mouse on, got mode=%d sgr=%v", session.VT().MouseMode, session.VT().MouseSGR)
	}

	// Quit with ctrl-q
	session.SendRaw([]byte{17})
	_ = session.Wait(time.Second * 5)
}

func TestSessionDirectTabSelectAltDigits(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'ORIGINAL_TAB1\\n'; while true; do sleep 1; done")
	session.WaitFor("ORIGINAL_TAB1", time.Second*5)

	// Open Tab 2 with ctrl-n
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// Send alt-1 (\x1b1) to switch to Tab 1
	session.SendRaw([]byte{27, '1'})
	session.WaitFor("ORIGINAL_TAB1", time.Second*5)

	// Send alt-2 (\x1b2) to switch back to Tab 2
	session.SendRaw([]byte{27, '2'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "ORIGINAL_TAB1")
	})

	// Close Tab 2 with prefix x (ctrl-b x) so only Tab 1 remains
	session.SendRaw([]byte{2, 'x'})
	session.WaitFor("ORIGINAL_TAB1", time.Second*5)

	// With 1 tab, send alt-2 (\x1b2): must be swallowed and NOT printed to shell
	session.SendRaw([]byte{27, '2'})
	time.Sleep(200 * time.Millisecond)

	for _, row := range session.Screen() {
		if strings.Contains(row, "2") && !strings.Contains(row, "1: bash") {
			t.Fatalf("alt-2 was not swallowed: screen row = %q", row)
		}
	}

	session.SendRaw([]byte{17}) // ctrl-q
	_ = session.Wait(time.Second * 5)
}

func TestSessionCleanOuterScreenAfterQuitWithTabs(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'INITIAL_SHELL_TEXT\\n'; while true; do sleep 1; done")
	session.WaitFor("INITIAL_SHELL_TEXT", time.Second*5)

	// Open tab 2 with ctrl-n
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// Print unique text in Tab 2
	session.Send("printf 'TAB2_ISOLATED\\n'\n")
	session.WaitFor("TAB2_ISOLATED", time.Second*5)

	// Switch back to Tab 1 with alt-1 (\x1b1)
	session.SendRaw([]byte{27, '1'})
	session.WaitFor("INITIAL_SHELL_TEXT", time.Second*5)

	// Quit session with ctrl-q (17)
	session.SendRaw([]byte{17})
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session wait: %v", err)
	}

	// Verify that the outer terminal after exit does NOT contain TAB2_ISOLATED
	// or the tab bar, and does not have corrupted screen text
	for y, row := range session.Screen() {
		if strings.Contains(row, "TAB2_ISOLATED") {
			t.Fatalf("stale Tab 2 text found on outer screen at row %d: %q", y, row)
		}
		if strings.Contains(row, "1: bash") || strings.Contains(row, "2: ") {
			t.Fatalf("stale tab bar found on outer screen at row %d: %q", y, row)
		}
	}
}

func TestSessionTabBarClickSwitchesTabWhenChildMouseOn(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	// Tab 1 enables mouse SGR tracking
	fixture := "printf '\\033[?1000h\\033[?1006hTAB1_READY\\n'; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("TAB1_READY", time.Second*5)

	// Open Tab 2 with ctrl-n
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// Switch back to Tab 1 so active tab has mouse tracking on
	session.SendRaw([]byte{27, '1'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return session.VT().MouseMode == 1000
	})

	// Click on row 24 (the tab bar) in the region of Tab 2: col 14, row 24
	// SGR mouse left press: \x1b[<0;14;24M
	session.SendRaw([]byte("\x1b[<0;14;24M"))
	session.SendRaw([]byte("\x1b[<0;14;24m"))

	// Verify that active tab switched to Tab 2 (where mouse mode is 0)
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return session.VT().MouseMode == 0
	})

	if session.VT().MouseMode != 0 {
		t.Fatalf("expected active tab to switch to Tab 2 (MouseMode=0), got %d", session.VT().MouseMode)
	}

	session.SendRaw([]byte{17}) // ctrl-q
	_ = session.Wait(time.Second * 5)
}

func TestSessionTabBarClicksSafeModePlainShellDoesNotEnableMouse(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	t.Setenv("SHELL", "/bin/sh")
	bin := buildLoomBinary(t)
	// Both tabs are plain shells (no child enables mouse)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'PLAIN_SHELL\\n'; while true; do sleep 1; done")
	session.WaitFor("PLAIN_SHELL", time.Second*5)

	// Open Tab 2 with ctrl-n so tab bar appears
	session.SendRaw([]byte{14})
	session.WaitFor("2: ", time.Second*5)

	// In safe mode, plain shells must NOT cause mouse tracking to be enabled on the outer VT
	if session.VT().MouseMode != 0 || session.VT().MouseSGR {
		t.Fatalf("expected outer VT mouse tracking to stay OFF in safe mode for plain shells, got mode=%d sgr=%v", session.VT().MouseMode, session.VT().MouseSGR)
	}

	session.SendRaw([]byte{17}) // ctrl-q
	_ = session.Wait(time.Second * 5)
}

func TestSessionBarAlwaysAndNarrowDoesNotWrap(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 40, 12, bin, "session", "--", "bash", "-c", "while true; do sleep 1; done")
	session.WaitFor("Tools", time.Second*5)
	rows := session.Screen()
	bar := rows[len(rows)-1]
	for _, want := range []string{"Tools", "🖱"} {
		if !strings.Contains(bar, want) {
			t.Fatalf("single-tab bar missing %q: %q; screen=%q", want, bar, rows)
		}
	}
	if strings.Contains(rows[len(rows)-2], "🖱") {
		t.Fatalf("mouse icon wrapped: %q", rows[len(rows)-2])
	}
	barY := len(rows) - 1
	assertSessionBarCellStyle(t, session, 0, barY, loom.DefaultTabsStyle().Active)
	capCol := displayColumnOf(bar, "⌃B…")
	labelCol := displayColumnOf(bar, "Tools")
	iconCol := displayColumnOf(bar, loom.SpeccedDefaults.Session.BarLabels.MouseIcon)
	searchCapCol := displayColumnOf(bar, "⌃B F")
	searchLabelCol := displayColumnOf(bar, loom.SpeccedDefaults.Session.BarLabels.Search)
	if capCol < 0 || labelCol <= capCol || iconCol < 0 || searchCapCol < 0 || searchLabelCol <= searchCapCol {
		t.Fatalf("bar boundaries missing: %q", bar)
	}
	hintStyle := loom.DefaultTheme().HintBarStyle()
	assertSessionBarCellStyle(t, session, capCol, barY, hintStyle.Cap)
	assertSessionBarCellStyle(t, session, capCol+1, barY, hintStyle.Cap)
	assertSessionBarCellStyle(t, session, labelCol, barY, hintStyle.Label)
	assertSessionBarCellStyle(t, session, searchCapCol, barY, hintStyle.Cap)
	assertSessionBarCellStyle(t, session, searchLabelCol, barY, hintStyle.Label)
	assertSessionBarCellStyle(t, session, iconCol, barY, hintStyle.Label)
	session.SendRaw([]byte{17})
	_ = session.Wait(time.Second * 5)
}

func displayColumnOf(line, text string) int {
	runes := []rune(line)
	for i := 0; i+len([]rune(text)) <= len(runes); i++ {
		if string(runes[i:i+len([]rune(text))]) == text {
			return loom.StringWidth(string(runes[:i]))
		}
	}
	return -1
}

func assertSessionBarCellStyle(t *testing.T, session *ptytest.Session, x, y int, want loom.Style) {
	t.Helper()
	got := session.Cell(x, y).Style
	wantFR, wantFG, wantFB, wantFGOK := want.FG.RGB()
	gotFR, gotFG, gotFB, gotFGOK := got.FG.RGB()
	wantBR, wantBG, wantBB, wantBGOK := want.BG.RGB()
	gotBR, gotBG, gotBB, gotBGOK := got.BG.RGB()
	if wantFGOK != gotFGOK || wantBGOK != gotBGOK || (wantFGOK && (wantFR != gotFR || wantFG != gotFG || wantFB != gotFB)) || (wantBGOK && (wantBR != gotBR || wantBG != gotBG || wantBB != gotBB)) {
		t.Fatalf("cell (%d,%d) FG/BG = %v/%v, want %v/%v", x, y, got.FG, got.BG, want.FG, want.BG)
	}
}

func TestSessionBarFitAndHitBoundaries(t *testing.T) {
	tabs := []*sessionTab{{title: "bash"}, {title: "zsh"}, {title: "fish"}}
	for _, cols := range []int{10, 40, 100} {
		bar, mouseStart, hits := sessionBarContent(tabs, 1, cols, true)
		if width := loom.StringWidth(bar); width > cols {
			t.Fatalf("bar width at %d columns = %d", cols, width)
		}
		if mouseStart < 0 || mouseStart >= cols {
			t.Fatalf("mouse icon start at %d columns = %d", cols, mouseStart)
		}
		for _, hit := range hits {
			if hit.start < 0 || hit.end > cols || tabIndexAtCol(tabs, hit.start, cols+1) != hit.index || tabIndexAtCol(tabs, hit.end-1, cols+1) != hit.index {
				t.Fatalf("tab hit boundary at %d columns: %+v", cols, hit)
			}
		}
	}
}

func TestSessionBarGoldens(t *testing.T) {
	for _, fixture := range []struct {
		name   string
		titles []string
	}{
		{"session-bar-one.ansi", []string{"bash"}},
		{"session-bar-three.ansi", []string{"bash", "vim", "shell"}},
	} {
		tabs := make([]*sessionTab, len(fixture.titles))
		for i, title := range fixture.titles {
			tabs[i] = &sessionTab{title: title, vt: vt.NewVT(100, 23)}
		}
		var output bytes.Buffer
		drawSessionTabBar(&output, tabs, 0, 100, 24)
		path := filepath.Join("..", "..", "docs", "progress", "373", fixture.name)
		if os.Getenv("UPDATE_SESSION_BAR_GOLDENS") == "1" {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, output.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		golden, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(golden, output.Bytes()) {
			t.Fatalf("%s differs from rendered session bar", fixture.name)
		}
	}
}

func TestSessionPrefixHelpAndMouseToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", "printf 'HELP_READY\\n'; while true; do sleep 1; done")
	session.WaitFor("HELP_READY", time.Second*5)
	session.SendRaw([]byte{2, '?'})
	session.WaitFor("Session tools", time.Second*5)
	session.SendRaw([]byte{'z'})
	session.WaitFor("HELP_READY", time.Second*5)
	session.SendRaw([]byte{2, 'm'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool { return session.VT().MouseMode == 1000 })
	if session.VT().MouseMode != 1000 || !session.VT().MouseSGR {
		t.Fatalf("mouse mode = %d SGR=%v, want enabled", session.VT().MouseMode, session.VT().MouseSGR)
	}
	session.SendRaw([]byte{2, 'm'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool { return session.VT().MouseMode == 0 })
	if session.VT().MouseMode != 0 {
		t.Fatalf("mouse mode = %d, want disabled", session.VT().MouseMode)
	}
	session.SendRaw([]byte{2, 'm'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool { return session.VT().MouseMode == 1000 })
	session.SendRaw([]byte("\x1b[<0;78;24M"))
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool { return session.VT().MouseMode == 0 })
	if session.VT().MouseMode != 0 {
		t.Fatalf("mouse click left mode %d enabled", session.VT().MouseMode)
	}
	session.SendRaw([]byte{17})
	_ = session.Wait(time.Second * 5)
}

func TestSessionPrefixHelpFitsShortTerminal(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 100, 12, bin, "session", "--", "bash", "-c", "while true; do sleep 1; done")
	session.SendRaw([]byte{2, '?'})
	session.WaitFor("Session tools", time.Second*5)
	session.WaitFor("return", time.Second*5)
	rows := session.Screen()
	text := strings.Join(rows, "\n")
	for _, want := range []string{"Search", "Close tab", "Press any key to return"} {
		if !strings.Contains(text, want) {
			t.Fatalf("short-terminal help missing %q: %q", want, rows)
		}
	}
	session.SendRaw([]byte{'z'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "Session tools")
	})
	session.SendRaw([]byte{17})
	_ = session.Wait(time.Second * 5)
}

func TestSessionSingleTabAltScreenChildQuitCtrlQ(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	// Single-tab child enters alternate screen (?1049h)
	fixture := "printf '\\033[?1049h\\033[1;1HALT_RUNNING\\n'; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("ALT_RUNNING", time.Second*5)

	if !session.VT().AltScreen() {
		t.Fatalf("expected child to enter alternate screen")
	}

	// Quit with ctrl-q
	session.SendRaw([]byte{17})
	if err := session.Wait(time.Second * 5); err != nil {
		t.Fatalf("session wait: %v", err)
	}

	// The outer VT must NOT be left in alternate screen
	if session.VT().AltScreen() {
		t.Fatalf("outer VT left in alternate screen after ctrl-q with alt-screen child")
	}
}

func TestSessionSingleTabAltScreenChildSIGTERM(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	bin := buildLoomBinary(t)
	// Single-tab child enters alternate screen (?1049h)
	fixture := "printf '\\033[?1049h\\033[1;1HALT_RUNNING\\n'; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 24, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("ALT_RUNNING", time.Second*5)

	if !session.VT().AltScreen() {
		t.Fatalf("expected child to enter alternate screen")
	}

	// Send SIGTERM to loom session
	if err := session.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("send SIGTERM: %v", err)
	}

	waitErr := session.Wait(time.Second * 5)
	if waitErr == nil {
		t.Fatalf("expected non-zero exit code on SIGTERM, got nil")
	}
	if !strings.Contains(waitErr.Error(), "143") {
		t.Fatalf("expected exit code 143 (128+SIGTERM), got %v", waitErr)
	}

	// The outer VT must NOT be left in alternate screen
	if session.VT().AltScreen() {
		t.Fatalf("outer VT left in alternate screen after SIGTERM with alt-screen child")
	}
}

func TestSessionSearchFindNavigateNoMatchAndClose(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	bin := buildLoomBinary(t)
	fixture := "printf 'old needle\\n'; for i in $(seq 1 20); do printf 'filler-%s\\n' \"$i\"; done; printf 'new needle\\n'; read -r line; printf 'CHILD:%s\\n' \"$line\"; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 12, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("new needle", time.Second*5)

	session.SendRaw([]byte{2, 'f'})
	session.WaitFor("Find:", time.Second*5)
	session.Send("needle")
	session.WaitFor("2/2", time.Second*5)
	newY := sessionRowContaining(session.Screen(), "new needle")
	if newY < 0 || !session.Cell(4, newY).Style.Reverse {
		t.Fatalf("newest match cell was not highlighted at row %d: %+v screen=%q", newY, session.Cell(4, newY), session.Text())
	}
	if strings.Contains(session.Text(), "old needle") {
		t.Fatalf("old scrollback match visible before navigation: %q", session.Text())
	}

	session.Send("\n")
	session.WaitForUpdate("1/2", time.Second*5)
	oldY := sessionRowContaining(session.Screen(), "old needle")
	if oldY < 0 || strings.Contains(session.Text(), "new needle") || !session.Cell(4, oldY).Style.Reverse {
		t.Fatalf("enter did not select older match at row %d: %+v", oldY, session.Cell(4, oldY))
	}
	session.SendRaw([]byte("\x1b[A"))
	session.WaitForUpdate("2/2", time.Second*5)
	newY = sessionRowContaining(session.Screen(), "new needle")
	if newY < 0 || strings.Contains(session.Text(), "old needle") || !session.Cell(4, newY).Style.Reverse {
		t.Fatalf("up did not select newer match at row %d: %+v", newY, session.Cell(4, newY))
	}

	session.SendRaw([]byte{14, 16, 2, 'f'})
	session.SendRaw([]byte("\x1b1"))
	session.SendRaw([]byte("\x1b[<0;2;1M\n"))
	session.Send("missing")
	session.WaitFor("No matches", time.Second*5)
	for _, row := range session.Cells() {
		for _, cell := range row {
			if cell.Style.Reverse {
				t.Fatal("no-match query left a match highlighted")
			}
		}
	}

	session.SendRaw([]byte{0x1b})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		for _, row := range rows {
			if strings.Contains(row, "Find:") {
				return false
			}
		}
		return true
	})
	session.Send("restored-input\n")
	session.WaitFor("CHILD:restored-input", time.Second*5)
	session.SendRaw([]byte{17})
	_ = session.Wait(time.Second * 5)
}

func sessionRowContaining(rows []string, text string) int {
	for y, row := range rows {
		if strings.Contains(row, text) {
			return y
		}
	}
	return -1
}

func TestSessionSearchAltScreenOnlySearchesVisibleScreen(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	bin := buildLoomBinary(t)
	fixture := "printf 'HISTORY_NEEDLE\\n'; printf '\\033[?1049hALT_NEEDLE'; while true; do sleep 1; done"
	session := ptytest.Start(t, 80, 12, bin, "session", "--", "bash", "-c", fixture)
	session.WaitFor("ALT_NEEDLE", time.Second*5)

	session.SendRaw([]byte{2, 'f'})
	session.WaitFor("Find:", time.Second*5)
	session.Send("HISTORY_NEEDLE")
	session.WaitFor("No matches", time.Second*5)
	session.SendRaw([]byte{0x1b})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		for _, row := range rows {
			if strings.Contains(row, "Find:") {
				return false
			}
		}
		return true
	})
	session.SendRaw([]byte{2, 'f'})
	session.WaitFor("Find:", time.Second*5)
	session.Send("ALT_NEEDLE")
	session.WaitFor("1/1", time.Second*5)
	if !strings.Contains(session.Text(), "ALT_NEEDLE") {
		t.Fatalf("visible alternate-screen text was not searchable: %q", session.Text())
	}
	session.SendRaw([]byte{0x1b})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		for _, row := range rows {
			if strings.Contains(row, "Find:") {
				return false
			}
		}
		return true
	})
	if !session.VT().AltScreen() {
		t.Fatal("closing search left the child alternate screen")
	}
	session.SendRaw([]byte{17})
	_ = session.Wait(time.Second * 5)
}

func TestSessionSearchKeepsGlobalQuitBinding(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	bin := buildLoomBinary(t)
	session := ptytest.Start(t, 80, 12, bin, "session", "--", "bash", "-c", "while true; do sleep 1; done")
	session.WaitFor("Search", time.Second*5)
	session.SendRaw([]byte{2, 'f'})
	session.WaitFor("Find:", time.Second*5)
	session.SendRaw([]byte{17})
	_ = session.Wait(time.Second * 5)
}
