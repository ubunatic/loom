package loom

import (
	"strings"
	"testing"
)

func TestEditorStatusBarSpecAndNarrowWidths(t *testing.T) {
	icons := SpeccedDefaults.Editor.StatusIcons
	for _, enabled := range []bool{false, true} {
		cfg := EditorConfig{Theme: "plain", MouseGrab: enabled, AltScreen: enabled}
		bar := NewEditorStatusBar(cfg)
		mouse, alt := icons.MouseOff, icons.AltOff
		if enabled {
			mouse, alt = icons.MouseOn, icons.AltOn
		}
		if bar.Entries[0].Key != icons.Theme || bar.Entries[1].Key != mouse || bar.Entries[2].Key != alt || bar.Style.Cap.FG != ColorIndex(uint8(icons.FG)) || bar.Style.Cap.BG != ColorIndex(uint8(icons.BG)) {
			t.Fatalf("status bar does not reflect spec/config: %+v", bar)
		}
		if !strings.Contains(bar.Text(bar.ContentWidth()), cfg.Theme) {
			t.Fatal("full width lost theme detail")
		}
		for width := 1; width <= bar.ContentWidth(); width++ {
			c := NewCanvas(width+2, 1)
			c.Set(0, 0, Cell{Text: "│"})
			c.Set(width+1, 0, Cell{Text: "│"})
			bar.Draw(c, Rect{X: 1, W: width, H: 1})
			if c.Get(0, 0).Text != "│" || c.Get(width+1, 0).Text != "│" {
				t.Fatalf("width %d overwrote border", width)
			}
		}
	}
}
