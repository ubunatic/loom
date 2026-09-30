// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestKeyHelpRendersLabeledBindingsInStableOrder(t *testing.T) {
	km := loom.NewKeyMapWithLabels(map[string][]string{
		"save":    {"ctrl-s", "s"},
		"quit":    {"q"},
		"ignored": {"i"},
	}, map[string]string{"save": "Save", "quit": "Quit"})

	canvas := loom.NewCanvas(40, 1)
	loom.NewKeyHelp(km).Draw(canvas, canvas.Bounds())
	got := strings.TrimRight(visibleKeyHelpRow(canvas.Row(0)), " ")
	if want := "q Quit · ctrl-s Save"; got != want {
		t.Fatalf("key help = %q, want %q", got, want)
	}
}

func TestKeyHelpTruncatesAtDisplayWidth(t *testing.T) {
	km := loom.NewKeyMapWithLabels(map[string][]string{"save": {"s"}}, map[string]string{"save": "保存 changes"})
	canvas := loom.NewCanvas(8, 1)
	loom.NewKeyHelp(km).Draw(canvas, canvas.Bounds())

	row := visibleKeyHelpRow(canvas.Row(0))
	if got := loom.StringWidth(row); got != 8 {
		t.Fatalf("rendered width = %d, want 8: %q", got, row)
	}
	if want := "s 保存 …"; row != want {
		t.Fatalf("truncated help = %q, want %q", row, want)
	}
}

func visibleKeyHelpRow(row string) string {
	var text strings.Builder
	for _, cell := range loom.ParseANSI(row) {
		if !cell.Continuation {
			text.WriteString(cell.Text)
		}
	}
	return text.String()
}

func TestKeyHelpAndViewportThemeSurfaces(t *testing.T) {
	theme := loom.Theme("mc-dark")
	for _, entry := range []struct {
		name   string
		widget loom.Widget
	}{
		{"KeyHelp", loom.NewKeyHelp(loom.NewKeyMapWithLabels(map[string][]string{"save": {"s"}}, map[string]string{"save": "保存"}))},
		{"Viewport", loom.NewViewport(loom.NewView([]string{"first", "second", "third", "fourth", "fifth"}))},
		{"EmptyViewport", loom.NewViewport(nil)},
	} {
		t.Run(entry.name, func(t *testing.T) {
			themeable, ok := entry.widget.(loom.Themeable)
			if !ok {
				t.Fatal("widget has no theme support")
			}
			for _, theme := range []loom.ThemeColors{theme, loom.Theme("mc")} {
				themeable.ApplyTheme(theme)
				c := loom.NewCanvas(14, 6)
				r := loom.Rect{X: 2, Y: 1, W: 10, H: 4}
				entry.widget.Draw(c, r)
				for y := r.Y; y < r.Y+r.H; y++ {
					for x := r.X; x < r.X+r.W; x++ {
						if got := c.Get(x, y).Style.BG; got != theme.NormalBG.Color() {
							t.Fatalf("background at (%d,%d) = %v, want %v", x, y, got, theme.NormalBG.Color())
						}
					}
				}
				if c.Get(0, 0).Style.BG != loom.ColorReset() {
					t.Fatal("theme surface escaped widget bounds")
				}
			}
		})
	}
}
