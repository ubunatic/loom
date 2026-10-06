// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestGalleryHintBarsClickAndKeyPaths(t *testing.T) {
	for _, key := range []string{"f1", "ctrl-s", "ctrl-shift-s", "f7", "f8", "f9", "f10"} {
		for _, label := range []bool{false, true} {
			t.Run(key, func(t *testing.T) {
				e := loom.NewRichTextEdit(nil)
				e.ShowFileBar = true
				g := newThemedGallery(e, "julia256")
				c := loom.NewCanvas(86, 15)
				g.Draw(c, loom.Rect{X: 3, Y: 2, W: 80, H: 12})
				bar, y := g.hintBar, 10
				if strings.HasPrefix(key, "f") && key != "f1" && key != "f7" {
					bar, y = g.controls, 11
				}
				x := 1
				for _, entry := range bar.Entries {
					if entry.Binding == key {
						if label {
							x += loom.StringWidth(entry.Key) + 3
						} else {
							x++
						}
						break
					}
					x += loom.StringWidth(entry.Key) + 3 + loom.StringWidth(entry.Label) + 1
					if entry.Detail != "" {
						x += 1 + loom.StringWidth(entry.Detail)
					}
				}
				beforeTheme := g.themeName
				g.ConsumeMouse(loom.MouseEvent{X: x, Y: y, Action: loom.MouseHover})
				if e.ViewMode || g.bgIndex != 0 || g.themeName != beforeTheme {
					t.Fatal("hover changed state")
				}
				result := g.ConsumeMouse(loom.MouseEvent{X: x, Y: y, Action: loom.MousePress, Button: loom.MouseLeft})
				if !result.Consumed {
					t.Fatalf("%s cap/label click ignored at (%d,%d)", key, x, y)
				}
				switch key {
				case "f1":
					g.Draw(c, loom.Rect{W: 80, H: 12})
					if !strings.Contains(canvasScreenText(c), "RichTextEdit Help") {
						t.Fatal("click did not open help")
					}
				case "ctrl-s", "ctrl-shift-s":
					g.Draw(c, loom.Rect{W: 80, H: 12})
					if !strings.Contains(canvasScreenText(c), "Save as") {
						t.Fatal("click did not open save picker")
					}
				case "f7":
					if !e.ViewMode {
						t.Fatal("click did not toggle mode")
					}
					g.ConsumeKey(loom.KeyEvent{Key: key})
					if e.ViewMode {
						t.Fatal("key did not toggle back")
					}
				case "f8":
					if g.bgIndex != 1 {
						t.Fatal("click did not cycle BG")
					}
					g.ConsumeKey(loom.KeyEvent{Key: key})
					if g.bgIndex != 2 {
						t.Fatal("key did not cycle BG")
					}
				case "f9":
					if g.themeName == beforeTheme {
						t.Fatal("click did not cycle theme")
					}
					name := g.themeName
					g.ConsumeKey(loom.KeyEvent{Key: key})
					if g.themeName == name {
						t.Fatal("key did not cycle theme")
					}
				case "f10":
					if !result.Quit || !g.ConsumeKey(loom.KeyEvent{Key: key}).Quit {
						t.Fatal("click/key did not quit")
					}
				}
			})
		}
	}
}

func TestGalleryControlsMatchDesignAtNormalAndNarrowWidth(t *testing.T) {
	for _, tc := range []struct {
		width int
		want  string
	}{
		{80, "F8 BG plain F9 Theme julia256 F10 Quit"},
		{40, "F8 BG F9 Theme F10 Quit"},
	} {
		g := newThemedGallery(loom.NewRichTextEdit(nil), "julia256")
		c := loom.NewCanvas(tc.width, 12)
		g.Draw(c, c.Bounds())
		if got := canvasPlainRow(c, 11); got != tc.want {
			t.Fatalf("%d columns: %q, want %q", tc.width, got, tc.want)
		}
		if style := c.Get(2, 11).Style; style != loom.Theme("julia256").HintBarStyle().Cap {
			t.Fatalf("cap style: %+v", style)
		}
	}
}
