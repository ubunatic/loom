package main

import (
	"os"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestEditBottomBarClicks(t *testing.T) {
	for _, label := range []string{"Files", "Search", "Save", "Help", "Quit", "Box", "Screenshot"} {
		for _, capClick := range []bool{true, false} {
			t.Run(label, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				v, path := newDirtyEditView(t)
				quit := false
				v.quit = func() { quit = true }
				if label == "Save" {
					v.edit.ConsumeKey(loom.KeyEvent{Text: "x"})
				}
				c := loom.NewCanvas(160, 24)
				v.Draw(c, c.Bounds())
				// Find by display cells, including styled cap padding.
				x := -1
				for col := 0; col < c.Bounds().W; col++ {
					var text strings.Builder
					for j := col; j < c.Bounds().W; j++ {
						text.WriteString(c.Get(j, 22).Text)
					}
					if strings.HasPrefix(text.String(), label) {
						x = col
						break
					}
				}
				if x < 0 {
					t.Fatalf("missing %s", label)
				}
				if capClick {
					x -= 3
				}
				for _, action := range []loom.MouseAction{loom.MouseHover, loom.MouseRelease} {
					if got := v.ConsumeMouse(loom.MouseEvent{X: x, Y: 22, Action: action, Button: loom.MouseLeft}); got.Consumed {
						t.Fatalf("inert mouse action consumed: %+v", got)
					}
				}
				got := v.ConsumeMouse(loom.MouseEvent{X: x, Y: 22, Action: loom.MousePress, Button: loom.MouseLeft})
				if !got.Consumed {
					t.Fatalf("%s click ignored", label)
				}
				switch label {
				case "Files":
					if !v.showSidePanel {
						t.Fatal("browser not opened")
					}
				case "Search":
					if !v.showSearch {
						t.Fatal("search not opened")
					}
				case "Save":
					data, _ := os.ReadFile(path)
					if !strings.Contains(string(data), "x") || v.edit.IsModified() {
						t.Fatal("save not performed")
					}
				case "Help":
					v.Draw(c, c.Bounds())
					if !strings.Contains(canvasScreenText(c), "RichTextEdit Help") {
						t.Fatal("help not opened")
					}
				case "Quit":
					if !quit {
						t.Fatal("quit not requested")
					}
				case "Box":
					if !v.edit.BoxMode {
						t.Fatal("draw mode not enabled")
					}
				case "Screenshot":
					if v.statusMessage == "" {
						t.Fatal("screenshot not attempted")
					}
				}
			})
		}
	}
}

func TestEditStatusClicksAndClippedBars(t *testing.T) {
	for index := 0; index < 3; index++ {
		v, _ := newDirtyEditView(t)
		c := loom.NewCanvas(100, 24)
		v.Draw(c, c.Bounds())
		before := v.config
		x := v.indicatorRect.X + 2
		for i := 0; i < index; i++ {
			e := v.indicators.Entries[i]
			x += loom.StringWidth(e.Key) + 4 + loom.StringWidth(e.Label)
			if e.Detail != "" {
				x += 1 + loom.StringWidth(e.Detail)
			}
		}
		if got := v.ConsumeMouse(loom.MouseEvent{X: x, Y: 21, Action: loom.MousePress, Button: loom.MouseLeft}); !got.Consumed {
			t.Fatalf("indicator %d ignored", index)
		}
		if index == 0 && v.config.Theme == before.Theme || index == 1 && v.config.MouseGrab == before.MouseGrab || index == 2 && v.config.AltScreen == before.AltScreen {
			t.Fatalf("indicator %d did not toggle", index)
		}
		v.Draw(c, loom.Rect{W: 10, H: 4})
		if got := v.ConsumeMouse(loom.MouseEvent{X: x, Y: 21, Action: loom.MousePress, Button: loom.MouseLeft}); got.Consumed {
			t.Fatal("stale status targets survived short draw")
		}
	}
}

func TestEditPrimaryAndSecondaryToggles(t *testing.T) {
	for _, key := range []string{"ctrl-o", "f2", "ctrl-f", "f3"} {
		v, _ := newDirtyEditView(t)
		for _, open := range []bool{true, false} {
			if got := v.ConsumeKey(loom.KeyEvent{Key: key}); !got.Consumed {
				t.Fatalf("%s ignored", key)
			}
			state := v.showSidePanel
			if key == "ctrl-f" || key == "f3" {
				state = v.showSearch
			}
			if state != open {
				t.Fatalf("%s state %v want %v", key, state, open)
			}
		}
	}
}
