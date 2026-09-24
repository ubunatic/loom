// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"strings"
	"testing"
)

func TestRowLabel(t *testing.T) {
	cases := []struct {
		idx  int
		want string
	}{
		{0, "a"},
		{1, "b"},
		{25, "z"},
		{26, "aa"},
		{27, "ab"},
		{51, "az"},
		{52, "ba"},
	}
	for _, tc := range cases {
		if got := rowLabel(tc.idx); got != tc.want {
			t.Errorf("rowLabel(%d) = %q, want %q", tc.idx, got, tc.want)
		}
	}
}

func TestRenderEmojiGrid(t *testing.T) {
	glyphs := uniqueGlyphs()
	if len(glyphs) == 0 {
		t.Fatal("no glyphs found")
	}

	// Test all sections (no width filter)
	all, err := RenderEmojiGrid(glyphs, "")
	if err != nil {
		t.Fatalf("RenderEmojiGrid() err: %v", err)
	}
	for _, want := range []string{"VTE Width 2:", "VTE Width 3:", "VTE Width 4:", "VTE Width 1:", "a |", "|"} {
		if !strings.Contains(all, want) {
			t.Errorf("RenderEmojiGrid missing %q:\n%s", want, all)
		}
	}

	// Test specific width filters
	for _, w := range []string{"1", "2", "3", "4"} {
		out, err := RenderEmojiGrid(glyphs, w)
		if err != nil {
			t.Fatalf("RenderEmojiGrid(w=%s) err: %v", w, err)
		}
		if !strings.Contains(out, "VTE Width "+w+":") {
			t.Errorf("RenderEmojiGrid(w=%s) missing section header:\n%s", w, out)
		}
	}

	// Test invalid width filter
	if _, err := RenderEmojiGrid(glyphs, "99"); err == nil {
		t.Error("RenderEmojiGrid(w=99) expected error, got nil")
	}

	// Test RenderEmojiGridWithPath for standard / non-vte path
	stdOut, err := RenderEmojiGridWithPath(glyphs, "2", "standard")
	if err != nil {
		t.Fatalf("RenderEmojiGridWithPath(standard) err: %v", err)
	}
	if !strings.Contains(stdOut, "STANDARD Width 2:") {
		t.Errorf("RenderEmojiGridWithPath(standard) missing STANDARD header:\n%s", stdOut)
	}
	// For standard path, ☠️ is not padded with trailing space
	if !strings.Contains(stdOut, "☠️|") {
		t.Errorf("RenderEmojiGridWithPath(standard) expected unpadded '☠️|':\n%s", stdOut)
	}
}
