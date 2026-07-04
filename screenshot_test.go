// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// multiViewYAML is a minimal three-view layout used to exercise routing.
const multiViewYAML = `
app:
  root: main
  height: 6
views:
  - name: main
    grid: |
      +---+
      |L  |
      +---+
    elements:
      L:
        type: choice
        static: ["list", "settings"]
  - name: list
    grid: |
      +---+
      |L  |
      +---+
    elements:
      L:
        type: view
        static: |
          alpha
          beta
  - name: settings
    grid: |
      +---+
      |S  |
      +---+
    elements:
      S:
        type: view
        static: |
          a setting
`

func newRouter(t *testing.T) *loom.Router {
	t.Helper()
	widget, _, err := loom.BuildWidget(strings.NewReader(multiViewYAML))
	if err != nil {
		t.Fatalf("BuildWidget: %v", err)
	}
	router, ok := widget.(*loom.Router)
	if !ok {
		t.Fatalf("expected *loom.Router, got %T", widget)
	}
	return router
}

func TestDeeplinkSingleView(t *testing.T) {
	r := newRouter(t)
	if err := r.Deeplink("settings"); err != nil {
		t.Fatalf("Deeplink: %v", err)
	}
	if r.Current() != "settings" {
		t.Errorf("current = %q, want settings", r.Current())
	}
	// One step from root: Esc/back returns to main, then stops.
	if !r.GoBack() || r.Current() != "main" {
		t.Errorf("GoBack did not return to main, current = %q", r.Current())
	}
}

func TestDeeplinkPathBuildsHistory(t *testing.T) {
	r := newRouter(t)
	if err := r.Deeplink("list:settings"); err != nil {
		t.Fatalf("Deeplink: %v", err)
	}
	if r.Current() != "settings" {
		t.Fatalf("current = %q, want settings", r.Current())
	}
	// History walks back settings → list → main, mirroring the colon path.
	if !r.GoBack() || r.Current() != "list" {
		t.Errorf("first GoBack: current = %q, want list", r.Current())
	}
	if !r.GoBack() || r.Current() != "main" {
		t.Errorf("second GoBack: current = %q, want main", r.Current())
	}
}

func TestDeeplinkEmptyAndWhitespaceSegments(t *testing.T) {
	r := newRouter(t)
	// Leading/trailing colons and blanks are skipped, not errors.
	if err := r.Deeplink(":settings:"); err != nil {
		t.Fatalf("Deeplink: %v", err)
	}
	if r.Current() != "settings" {
		t.Errorf("current = %q, want settings", r.Current())
	}
}

func TestDeeplinkUnknownView(t *testing.T) {
	r := newRouter(t)
	err := r.Deeplink("settings:nope")
	if err == nil {
		t.Fatal("Deeplink with unknown view should error")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error should name the bad segment: %v", err)
	}
	// The valid prefix still routed before the failure was reported.
	if r.Current() != "settings" {
		t.Errorf("current = %q, want settings (valid prefix applied)", r.Current())
	}
}

func TestRenderProducesRows(t *testing.T) {
	v := loom.NewView([]string{"hello", "world"})
	rows := loom.Render(v, 10, 3)
	if len(rows) != 3 {
		t.Fatalf("Render returned %d rows, want 3", len(rows))
	}
	if !strings.Contains(rows[0], "hello") || !strings.Contains(rows[1], "world") {
		t.Errorf("rendered rows missing content: %q", rows)
	}
}

func TestScreenshotScriptRoundTrip(t *testing.T) {
	rows := []string{
		"\x1b[1mbold\x1b[0m",
		"plain it's fine", // embedded single quote must survive
	}
	script := loom.ScreenshotScript(rows, "test\nshot")

	if !strings.HasPrefix(script, "#!/usr/bin/env bash\n") {
		t.Errorf("script missing shebang: %q", script)
	}
	// The header comment collapses newlines so it stays a single comment line.
	if !strings.Contains(script, "# test shot\n") {
		t.Errorf("script missing collapsed comment: %q", script)
	}
	// Raw ESC bytes must not appear; they are encoded as the \033 escape.
	if strings.Contains(script, "\x1b") {
		t.Errorf("script must not contain raw ESC bytes: %q", script)
	}
	if !strings.Contains(script, `\033[1mbold`) {
		t.Errorf("script should encode ESC as \\033: %q", script)
	}
	// Embedded single quote is closed-escaped-reopened.
	if !strings.Contains(script, `it'\''s fine`) {
		t.Errorf("script should escape single quotes: %q", script)
	}
}
