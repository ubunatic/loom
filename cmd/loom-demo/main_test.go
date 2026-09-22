// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestHostedModeSetup(t *testing.T) {
	host, err := newHostedTabs()
	if err != nil {
		t.Fatalf("newHostedTabs failed: %v", err)
	}

	// Verify the hosted setup
	if host.Focus() != 0 {
		t.Errorf("initial focus should be 0, got %d", host.Focus())
	}

	// Render initial state
	canvas := loom.NewCanvas(80, 24)
	host.Draw(canvas, canvas.Bounds())

	// At least the tab bar should be rendered
	row0 := canvas.Row(0)
	if row0 == "" {
		t.Error("first row is empty, expected tab bar")
	}
}

func TestNestedTabsMeta(t *testing.T) {
	// Test that with the hosted tabs example focused, left/right/ctrl-t keys
	// reach the inner Tabs and don't switch the outer host tabs.
	// Host must have at least TWO tabs (split and tabs) to be meaningful.

	host, err := newHostedTabs()
	if err != nil {
		t.Fatalf("newHostedTabs failed: %v", err)
	}

	if len(host.Tabs) < 2 {
		t.Fatalf("host must have at least 2 tabs for nested test, got %d", len(host.Tabs))
	}

	// Find the tabs example tab index
	tabsIdx := -1
	for i, tab := range host.Tabs {
		if tab.Title == "tabs" {
			tabsIdx = i
			break
		}
	}
	if tabsIdx < 0 {
		t.Skip("tabs example not in hosted tabs")
	}

	// Ensure we're NOT on the tabs tab initially
	initialIdx := host.Focus()
	if initialIdx == tabsIdx {
		// Switch to split if tabs was default
		host.SetFocusIndex(0)
		initialIdx = 0
	}

	// Switch focus to the hosted tabs example
	host.SetFocusIndex(tabsIdx)
	if host.Focus() != tabsIdx {
		t.Fatalf("failed to switch to tabs tab, focus is %d", host.Focus())
	}

	// Render before and after keys to verify visual change (the inner tabs should switch)
	framesBefore := loom.Render(host, 80, 24)
	beforeStr := strings.Join(framesBefore, "\n")

	// Test right arrow key: should reach the inner tabs
	_ = host.HandleKey(loom.KeyEvent{Key: "right"})
	if host.Focus() != tabsIdx {
		t.Errorf("after right key, host focus changed to %d, expected %d (still on tabs tab)", host.Focus(), tabsIdx)
	}

	framesAfterRight := loom.Render(host, 80, 24)
	afterRightStr := strings.Join(framesAfterRight, "\n")
	if beforeStr == afterRightStr {
		t.Error("after right key, render output did not change (expected inner tabs to switch)")
	}

	// Test left arrow key: should go back
	_ = host.HandleKey(loom.KeyEvent{Key: "left"})
	if host.Focus() != tabsIdx {
		t.Errorf("after left key, host focus changed to %d, expected %d (still on tabs tab)", host.Focus(), tabsIdx)
	}

	framesAfterLeft := loom.Render(host, 80, 24)
	afterLeftStr := strings.Join(framesAfterLeft, "\n")
	if afterLeftStr != beforeStr {
		t.Error("after left key, render output did not return to initial state")
	}

	// Test ctrl-t (cycle): inner tabs should advance
	_ = host.HandleKey(loom.KeyEvent{Key: "ctrl-t"})
	if host.Focus() != tabsIdx {
		t.Errorf("after ctrl-t, host focus changed to %d, expected %d (still on tabs tab)", host.Focus(), tabsIdx)
	}

	framesAfterCycle := loom.Render(host, 80, 24)
	afterCycleStr := strings.Join(framesAfterCycle, "\n")
	if afterCycleStr == beforeStr {
		t.Error("after ctrl-t, render output did not change (expected inner tabs to cycle)")
	}
}

func TestQuitContainment(t *testing.T) {
	// Test that a hosted app's quit key does not propagate quit to the host.

	host, err := newHostedTabs()
	if err != nil {
		t.Fatalf("newHostedTabs failed: %v", err)
	}

	// The OnChildQuit callback should return false (stay in hosted mode)
	// when a child quits. Verify the callback is set.
	if host.OnChildQuit == nil {
		t.Error("OnChildQuit callback is nil")
		return
	}

	// Call the callback and verify it returns false
	shouldQuit := host.OnChildQuit(0)
	if shouldQuit {
		t.Error("OnChildQuit returned true, expected false to stay in hosted mode")
	}
}

func TestGenerateM3HostedEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	host, err := newHostedTabs()
	if err != nil {
		t.Fatalf("newHostedTabs failed: %v", err)
	}

	// Render initial state (M3-hosted-split.ansi)
	frames := loom.Render(host, 80, 24)
	var buf strings.Builder
	for _, line := range frames {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	repoRoot := findM3RepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "062")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M3-hosted-split.ansi")
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}
	t.Logf("M3 initial evidence saved to %s (%d bytes)", outPath, len(buf.String()))

	// Now test the nested Tabs meta-test: switch to tabs example and send right arrow
	// Find the tabs tab
	tabsIdx := -1
	for i, tab := range host.Tabs {
		if tab.Title == "tabs" {
			tabsIdx = i
			break
		}
	}
	if tabsIdx < 0 {
		t.Skip("tabs example not in hosted tabs")
	}

	// Switch focus to tabs tab
	host.SetFocusIndex(tabsIdx)

	// Send right arrow to the host, which should delegate to the tabs widget
	_ = host.HandleKey(loom.KeyEvent{Key: "right"})

	// Render after the key (M3-hosted-tabs-inner-switched.ansi)
	frames2 := loom.Render(host, 80, 24)
	var buf2 strings.Builder
	for _, line := range frames2 {
		buf2.WriteString(line)
		buf2.WriteString("\n")
	}

	outPath2 := filepath.Join(progressDir, "M3-hosted-tabs-inner-switched.ansi")
	if err := os.WriteFile(outPath2, []byte(buf2.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}
	t.Logf("M3 switched evidence saved to %s (%d bytes)", outPath2, len(buf2.String()))

	// Verify the switched frame is different from the initial frame
	if buf.String() == buf2.String() {
		t.Error("M3-hosted-tabs-inner-switched.ansi should differ from the initial frame")
	}
}

// findM3RepoRoot walks up from the current directory to find the repo root (where go.mod is)
func findM3RepoRoot(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting current directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			t.Fatalf("go.mod not found in any parent directory")
		}
		cwd = parent
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"monitor", "monitor"},
		{"examples/monitor", "monitor"},
		{"examples/split/", "split"},
		{"  SPLIT  ", "split"},
	}
	for _, tc := range tests {
		if got := normalizeName(tc.input); got != tc.want {
			t.Errorf("normalizeName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestRunByNameAndExecute(t *testing.T) {
	// Running with -h/--help should succeed quickly without blocking or requiring TTY.
	if err := runByName("examples/split/", "--help"); err != nil {
		t.Errorf("runByName with --help failed: %v", err)
	}
	if err := execute([]string{"examples/monitor", "--help"}); err != nil {
		t.Errorf("execute with --help failed: %v", err)
	}
	if err := runByName("nonexistent"); err == nil {
		t.Error("expected error for nonexistent example, got nil")
	}
}
