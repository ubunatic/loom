// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
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
	session.WaitFor("WINCH:15 50", time.Second*5)
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
	session.WaitFor("1: bash", time.Second*5)

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
	session.WaitFor("1: bash", time.Second*5)

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
	session.WaitFor("1: bash", time.Second*5)

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
	session.WaitFor("1: bash", time.Second*5)

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
	session.WaitFor("1: bash", time.Second*5)

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
	session.WaitFor("1: bash", time.Second*5)

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
	session.WaitFor("1: bash", time.Second*5)

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
	session.WaitFor("1: bash", time.Second*5)

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
	session.WaitFor("1: bash", time.Second*5)

	// Close tab 2 with ctrl-b x so bar hides
	session.SendRaw([]byte{2, 'x'})
	session.WaitForUpdateState(time.Second*5, func(rows []string) bool {
		return !strings.Contains(strings.Join(rows, "\n"), "2: ")
	})

	// The outer VT bottom row must have no bar text
	for y, row := range session.Screen() {
		if strings.Contains(row, "1: bash") || strings.Contains(row, "2: ") {
			t.Fatalf("stale tab bar found at row %d after hiding bar: %q", y, row)
		}
	}

	session.SendRaw([]byte{17}) // ctrl-q
	_ = session.Wait(time.Second * 5)
}


