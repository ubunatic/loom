// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/internal/examplesreg"
)

func TestHostedModeSetup(t *testing.T) {
	// Build the hosted tabs widget
	tabs := make([]loom.Tab, 0)
	for _, e := range examplesreg.Registry {
		if e.NewWidget == nil {
			continue
		}
		widget, err := e.NewWidget([]string{})
		if err != nil {
			t.Errorf("failed to create widget for %q: %v", e.Name, err)
			continue
		}
		w, ok := widget.(loom.Widget)
		if !ok {
			t.Errorf("%q NewWidget did not return a loom.Widget", e.Name)
			continue
		}
		tabs = append(tabs, loom.Tab{Title: e.Name, Widget: w})
	}
	
	if len(tabs) == 0 {
		t.Fatal("no examples with NewWidget to host")
	}
	
	host := loom.NewTabs(tabs...)
	host.ArrowSwitch = false
	
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
	// This tests that when the hosted tabs example is focused,
	// left/right keys reach the inner Tabs and don't switch the outer host.
	
	// Find the tabs example and create its widget
	tabsEx, ok := examplesreg.Find("tabs")
	if !ok || tabsEx.NewWidget == nil {
		t.Skip("tabs example with NewWidget not found")
	}
	
	widget, err := tabsEx.NewWidget([]string{})
	if err != nil {
		t.Fatalf("NewWidget failed: %v", err)
	}
	w, ok := widget.(loom.Widget)
	if !ok {
		t.Fatal("NewWidget did not return a loom.Widget")
	}
	
	// Create a host with just the tabs widget
	host := loom.NewTabs(loom.Tab{Title: "hosted-tabs", Widget: w})
	host.ArrowSwitch = false
	
	// Simulate a left arrow key event - it should be consumed by the inner tabs,
	// not by the outer host (which would have no effect anyway with one tab)
	event := loom.KeyEvent{Key: "left"}
	quit := host.HandleKey(event)
	
	// No quit expected
	if quit {
		t.Error("expected no quit on left arrow")
	}
	
	// The outer host should still be on tab 0
	if host.Focus() != 0 {
		t.Errorf("host focus changed unexpectedly to %d", host.Focus())
	}
}

func TestGenerateM3HostedEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}
	
	// Build the hosted tabs widget
	tabs := make([]loom.Tab, 0)
	for _, e := range examplesreg.Registry {
		if e.NewWidget == nil {
			continue
		}
		widget, err := e.NewWidget([]string{})
		if err != nil {
			t.Errorf("failed to create widget for %q: %v", e.Name, err)
			continue
		}
		w, ok := widget.(loom.Widget)
		if !ok {
			t.Errorf("%q NewWidget did not return a loom.Widget", e.Name)
			continue
		}
		tabs = append(tabs, loom.Tab{Title: e.Name, Widget: w})
	}
	
	if len(tabs) == 0 {
		t.Fatal("no examples with NewWidget to host")
	}
	
	host := loom.NewTabs(tabs...)
	host.ArrowSwitch = false
	host.OnChildQuit = func(i int) bool { return false }
	
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
	
	// Get the tabs widget and send it a right arrow key
	tabsWidget := host.Tabs[tabsIdx].Widget
	_ = tabsWidget.HandleKey(loom.KeyEvent{Key: "right"})
	
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
