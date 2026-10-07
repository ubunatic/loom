// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestRichTextEditStatusColorsFollowTheme(t *testing.T) {
	e := NewRichTextEdit(nil)
	e.ShowFileBar = true
	theme := Theme("julia256")
	theme.ModifiedFG, theme.SavedFG = ThemeColorRGB(1, 2, 3), ThemeColorRGB(4, 5, 6)
	theme.MediaErrorFG, theme.PlaceholderFG = ThemeColorRGB(7, 8, 9), ThemeColorRGB(10, 11, 12)
	e.ApplyTheme(theme)
	check := func(state string, fg Color, dim bool) {
		t.Helper()
		c := NewCanvas(80, 4)
		e.Draw(c, c.Bounds())
		row := hintBarRow(c, 3)
		// Locate by rune column so a Unicode filename cannot shift the test.
		runes, target := []rune(row), []rune(state)
		for x := 0; x+len(target) <= len(runes); x++ {
			if string(runes[x:x+len(target)]) == state {
				style := c.Get(StringWidth(string(runes[:x])), 3).Style
				if style.FG != fg || style.Dim != dim {
					t.Fatalf("%s style: %+v, want fg=%+v dim=%v", state, style, fg, dim)
				}
				return
			}
		}
		t.Fatalf("status %s missing: %q", state, row)
	}
	check("Unsaved", theme.PlaceholderFG.Color(), true)
	e.ConsumeKey(KeyEvent{Text: "changed"})
	check("Modified", theme.ModifiedFG.Color(), false)
	if err := e.SaveAs(filepath.Join(t.TempDir(), "界面.ansi")); err != nil {
		t.Fatal(err)
	}
	check("Saved", theme.SavedFG.Color(), false)
	e.LastSaveError = errors.New("failed")
	check("Error", theme.MediaErrorFG.Color(), false)
	menu := e.ensureFileBar().menu
	for i, want := range []string{"⌃O", "⌃W", "⌃S", ""} {
		if got := menu.Menus[0].Items[i].Shortcut; got != want {
			t.Fatalf("shortcut: %q, want %q", got, want)
		}
	}
	if help := strings.Join(newRichTextEditHelp(nil).plainLines(72), "\n"); !strings.Contains(help, "⌃S ") || strings.Contains(help, "⌃⌥S") {
		t.Fatal(help)
	}
}
