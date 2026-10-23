// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
	"ubunatic.com/loom/vt"
)

func TestSessionSearchControlsUseKeycaps(t *testing.T) {
	defs := loom.SpeccedDefaults.Session
	got := sessionSearchControls(defs, "needle", 1, 2)
	want := "2/2  Enter/↑ older  ↓ newer  Esc close"
	if got != want {
		t.Fatalf("sessionSearchControls() = %q, want %q", got, want)
	}
	for _, raw := range []string{"enter", "down", "up", "esc"} {
		if strings.Contains(got, raw) {
			t.Fatalf("search controls expose raw binding %q: %q", raw, got)
		}
	}
}

func TestFindSessionMatchesSmartCaseAndDisplayColumns(t *testing.T) {
	term := vt.NewVT(20, 2)
	term.Write([]byte("FOO 中 foo")) //nolint:errcheck
	rows := term.ScrollbackCells()

	matches := findSessionMatches(rows, "foo")
	if len(matches) != 2 {
		t.Fatalf("case-insensitive matches = %d, want 2", len(matches))
	}
	if got := matches[0].start; got != 0 {
		t.Fatalf("first match column = %d, want 0", got)
	}
	if got := matches[1].start; got != 7 {
		t.Fatalf("second match display column = %d, want 7", got)
	}

	matches = findSessionMatches(rows, "Foo")
	if len(matches) != 0 {
		t.Fatalf("smart-case matches = %d, want 0", len(matches))
	}
	matches = findSessionMatches(rows, "中")
	if len(matches) != 1 || matches[0].start != 4 {
		t.Fatalf("wide-rune match = %+v, want display column 4", matches)
	}
}
