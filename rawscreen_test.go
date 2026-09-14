// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom/measure"
)

// TestClipRowNeverExceedsWidth: no matter what a renderer produced, a row
// RawScreen writes to the terminal must never exceed the real, live
// terminal width -- ClipRow is what enforces that.
func TestClipRowNeverExceedsWidth(t *testing.T) {
	plain := strings.Repeat("x", 50)
	colored := "\x1b[41mleft\x1b[0m middle \x1b[97mright\x1b[0m tail overflow"
	for _, row := range []string{plain, colored, "", "short"} {
		for _, width := range []int{0, 1, 5, 10, 20, 200} {
			clipped := ClipRow(row, width)
			if w := measure.StringWidth(clipped); w > width {
				t.Errorf("ClipRow(%q, %d) visible width = %d, want <= %d", row, width, w, width)
			}
		}
	}
}

// TestClipRowPreservesColorOfKeptPrefix: dropping color is not acceptable --
// inline ANSI is the one thing a drawing component is allowed to emit; only
// trim what doesn't fit.
func TestClipRowPreservesColorOfKeptPrefix(t *testing.T) {
	row := "\x1b[41mAB\x1b[0mCD"
	got := ClipRow(row, 3)
	want := "\x1b[41mAB\x1b[0mC\x1b[0m"
	if got != want {
		t.Errorf("ClipRow(%q, 3) = %q, want %q", row, got, want)
	}
	if w := measure.StringWidth(got); w != 3 {
		t.Errorf("ClipRow(%q, 3) visible width = %d, want 3", row, w)
	}
}

// TestClipRowNeverBleedsStyleIntoNextRow: a row clipped mid-style must still
// end in a reset, or its color would leak into whatever is drawn next.
func TestClipRowNeverBleedsStyleIntoNextRow(t *testing.T) {
	got := ClipRow("\x1b[41mABCDEFG", 3)
	if !strings.HasSuffix(got, "\x1b[0m") {
		t.Errorf("ClipRow with an open style must end in a reset, got %q", got)
	}
}

// TestClipRowLeavesUnstyledRowsUnstyled: ClipRow must not invent a reset
// code for plain text that never opened any styling.
func TestClipRowLeavesUnstyledRowsUnstyled(t *testing.T) {
	got := ClipRow("hello world", 5)
	if strings.Contains(got, "\x1b[") {
		t.Errorf("ClipRow of an unstyled row must not introduce escape codes, got %q", got)
	}
	if got != "hello" {
		t.Errorf("ClipRow(%q, 5) = %q, want %q", "hello world", got, "hello")
	}
}

// TestRawScreenNonTerminalPassesRowsThroughUncapped exercises Draw's
// non-terminal path (os.Pipe ends are never terminals) with an os.Pipe: a
// piped consumer gets plain, uncapped, newline-separated rows -- there's no
// PTY to auto-wrap against, and clipping to an assumed width would corrupt
// legitimate piped output.
func TestRawScreenNonTerminalPassesRowsThroughUncapped(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	s := OpenRawScreen(w)
	if s.isTerm {
		t.Fatal("expected isTerm=false for an os.Pipe write end")
	}
	longRow := strings.Repeat("y", 500)
	if err := s.Draw([]string{longRow, "second row"}); err != nil {
		t.Fatalf("Draw: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	w.Close()

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	got := string(buf[:n])
	want := longRow + "\nsecond row\n"
	if got != want {
		t.Errorf("non-terminal Draw output = %q, want %q (uncapped, no escape codes)", got, want)
	}
}
