// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"strings"
	"testing"
)

func hintBarRow(c *Canvas, y int) string {
	var text strings.Builder
	for x := 0; x < c.Cols(); x++ {
		cell := c.Get(x, y)
		if cell.Continuation {
			continue
		}
		if cell.Text == "" {
			text.WriteByte(' ')
		} else {
			text.WriteString(cell.Text)
		}
	}
	return text.String()
}

func TestHintBarFitsWholePairsAndNames(t *testing.T) {
	bar := NewHintBar(
		HintEntry{Key: "F8", Binding: "f8", Label: "BG", Detail: "plain"},
		HintEntry{Key: "F9", Binding: "f9", Label: "Theme", Detail: "julia256"},
		HintEntry{Key: "F10", Binding: "f10", Label: "Quit"},
	)
	for width := 1; width <= 120; width++ {
		c := NewCanvas(width+4, 3)
		bar.Draw(c, Rect{X: 2, Y: 1, W: width, H: 1})
		for _, hit := range bar.hits {
			if hit.rect.X < 0 || hit.rect.X+hit.rect.W > width {
				t.Fatalf("width %d: clipped pair %+v", width, hit)
			}
		}
		if strings.TrimSpace(c.Get(1, 1).Text) != "" || strings.TrimSpace(c.Get(width+2, 1).Text) != "" {
			t.Fatalf("width %d: escaped bounds", width)
		}
		text := hintBarRow(c, 1)
		if width == 80 && !strings.Contains(text, "Theme julia256") {
			t.Fatal(text)
		}
		if width == 40 && strings.TrimSpace(text) != "F8  BG  F9  Theme  F10  Quit" {
			t.Fatal(text)
		}
	}
}

func TestHintBarKeyAndMouseRunSameAction(t *testing.T) {
	for _, key := range []string{"F1", "^S", "^Shift+S", "F7", "F8", "F9", "F10"} {
		t.Run(key, func(t *testing.T) {
			calls := 0
			bar := NewHintBar(HintEntry{Key: key, Binding: menuKeyName(key), Label: "界面", Action: func() EventResult { calls++; return Handled() }})
			c := NewCanvas(40, 5)
			bar.Draw(c, Rect{X: 3, Y: 2, W: 30, H: 1})
			if r := bar.ConsumeKey(KeyEvent{Key: menuKeyName(key)}); !r.Consumed || calls != 1 {
				t.Fatalf("key result/calls: %+v/%d", r, calls)
			}
			for x := 1; x < bar.hits[0].rect.X+bar.hits[0].rect.W; x++ {
				before := calls
				bar.ConsumeMouse(MouseEvent{X: x, Y: 0, Action: MouseHover})
				bar.ConsumeMouse(MouseEvent{X: x, Y: 0, Action: MouseRelease, Button: MouseLeft})
				bar.ConsumeMouse(MouseEvent{X: x, Y: 0, Action: MousePress, Button: MouseRight})
				if calls != before {
					t.Fatal("non-click ran action")
				}
				if r := bar.ConsumeMouse(MouseEvent{X: x, Y: 0, Action: MousePress, Button: MouseLeft}); !r.Consumed || calls != before+1 {
					t.Fatalf("local column %d: %+v calls=%d", x, r, calls)
				}
			}
			bar.Draw(c, Rect{W: 1, H: 1})
			if r := bar.ConsumeMouse(MouseEvent{X: 1, Action: MousePress, Button: MouseLeft}); r.Consumed {
				t.Fatal("stale hit after resize")
			}
		})
	}
}

func TestRichTextEditHintBarMatchesDesign001(t *testing.T) {
	data, err := os.ReadFile("docs/data/richtext-toolbar-design-001-keycaps.ansi")
	if err != nil {
		t.Fatal(err)
	}
	fixture := &RichDocument{}
	fixture.FromANSI(string(data))
	rows := strings.Split(fixture.ToPlainText(), "\n")
	e := NewRichTextEdit(nil)
	for _, tc := range []struct {
		width int
		want  string
	}{
		{80, "F1  Help  ^S  Save  ^Shift+S  Save as  F7  View"},
		{40, "F1  Help  ^S  Save  F7  View"},
	} {
		c := NewCanvas(tc.width, 1)
		bar := e.HotkeyBar()
		bar.ApplyTheme(Theme("julia256"))
		bar.Draw(c, c.Bounds())
		if got := strings.TrimSpace(hintBarRow(c, 0)); got != tc.want {
			t.Fatalf("%d columns: %q, want %q", tc.width, got, tc.want)
		}
		fixtureRow := 5
		if tc.width == 40 {
			fixtureRow = 13
		}
		if got, want := strings.TrimSpace(hintBarRow(c, 0)), strings.TrimSpace(rows[fixtureRow]); got != want {
			t.Fatalf("design 001 at %d: %q, want %q", tc.width, got, want)
		}
		if got := c.Get(2, 0).Style; !got.Bold || got.BG != Theme("julia256").KeyCapBG.Color() {
			t.Fatalf("cap style: %+v", got)
		}
		if c.Get(6, 0).Style.Bold {
			t.Fatal("label is bold")
		}
	}
}

func TestRichTextEditHintBarAtomicAtEveryWidth(t *testing.T) {
	bar := NewRichTextEdit(nil).HotkeyBar()
	for width := 1; width <= 120; width++ {
		c := NewCanvas(width, 1)
		bar.Draw(c, c.Bounds())
		for _, hit := range bar.hits {
			entry := bar.Entries[hit.index]
			if hit.rect.X+hit.rect.W > width {
				t.Fatalf("width %d clipped %+v", width, hit)
			}
			var text strings.Builder
			for x := hit.rect.X; x < hit.rect.X+hit.rect.W; x++ {
				text.WriteString(c.Get(x, 0).Text)
			}
			if got, want := text.String(), " "+entry.Key+"  "+entry.Label; got != want {
				t.Fatalf("width %d: partial pair %q, want %q", width, got, want)
			}
		}
	}
}

func TestRichTextEditHintBarActions(t *testing.T) {
	for _, binding := range []string{"f1", "ctrl-s", "ctrl-shift-s", "f7"} {
		for _, labelClick := range []bool{false, true} {
			e := NewRichTextEdit(nil)
			e.ShowFileBar = true
			bar := e.HotkeyBar()
			bar.Draw(NewCanvas(80, 1), Rect{W: 80, H: 1})
			var x int
			for _, hit := range bar.hits {
				if bar.Entries[hit.index].Binding == binding {
					x = hit.rect.X + 1
					if labelClick {
						x = hit.rect.X + StringWidth(bar.Entries[hit.index].Key) + 3
					}
				}
			}
			bar.ConsumeMouse(MouseEvent{X: x, Action: MouseHover})
			if e.ViewMode || e.helpPopup != nil || e.savePopup != nil {
				t.Fatal("hover changed editor")
			}
			if r := bar.ConsumeMouse(MouseEvent{X: x, Action: MousePress, Button: MouseLeft}); !r.Consumed {
				t.Fatalf("%s click: %+v", binding, r)
			}
			switch binding {
			case "f1":
				if e.helpPopup == nil {
					t.Fatal("help did not open")
				}
			case "ctrl-s", "ctrl-shift-s":
				if e.savePopup == nil {
					t.Fatal("save picker did not open")
				}
			case "f7":
				if !e.ViewMode {
					t.Fatal("view did not toggle")
				}
			}
		}
	}
}
