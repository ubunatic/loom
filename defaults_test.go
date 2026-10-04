package loom

import (
	"testing"
	"time"
)

func TestScrollbarSpecGlyphsFitOneCell(t *testing.T) {
	if err := SpeccedDefaults.Scrollbar.validate(); err != nil {
		t.Fatal(err)
	}
	for _, glyph := range []string{"", " ", "AB", "界", "\x1b"} {
		defs := ScrollbarDefaults{Mode: ScrollbarAuto, ForegroundChar: glyph, BackgroundChar: "░"}
		if err := defs.validate(); err == nil {
			t.Fatalf("accepted invalid scrollbar glyph %q", glyph)
		}
	}
}

func TestScrollbarModeDefaultsAndOverrides(t *testing.T) {
	if got := SpeccedDefaults.Scrollbar.Mode; got != ScrollbarAuto {
		t.Fatalf("default mode = %q, want %q", got, ScrollbarAuto)
	}
	if !scrollbarVisible("", true) || scrollbarVisible("", false) {
		t.Fatal("empty widget mode did not inherit auto behavior")
	}
	if !scrollbarVisible(ScrollbarAlways, false) {
		t.Fatal("always mode hid scrollbar without overflow")
	}
	if scrollbarVisible(ScrollbarNever, true) {
		t.Fatal("never mode showed scrollbar during overflow")
	}
	if err := (ScrollbarDefaults{Mode: "invalid", ForegroundChar: "▓", BackgroundChar: "░"}).validate(); err == nil {
		t.Fatal("accepted invalid scrollbar mode")
	}
}

func TestMediaDefaultsLoadFromSpec(t *testing.T) {
	got := SpeccedDefaults.Media
	if got.LoadingLabel != "loading" || got.RenderErrorLabel != "render error" || got.LoadingThreshold != 50*time.Millisecond {
		t.Fatalf("media defaults = %+v, want loading/render error/50ms", got)
	}
}

func TestRichTextEditDefaultsLoadFromSpec(t *testing.T) {
	got := SpeccedDefaults.RichTextEdit
	if got.SelectionBG != 24 || got.ToolbarFG != 15 || got.ToolbarBG != 239 || got.SeparatorGlyph != "│" || got.SeparatorFG != 8 || got.SeparatorBG != 239 || got.PointerUpGlyph != "▲" || got.PointerDownGlyph != "▼" || got.PointerFG != 8 || got.PopoverFocusFG != 15 || got.PopoverFocusBG != 24 {
		t.Fatalf("rich text edit defaults = %+v", got)
	}
	want := []string{"B", "I", "U", "S", "Link", "#FG", "#BG", "Box", "Draw"}
	if len(got.PopoverLabels) != len(want) {
		t.Fatalf("popover labels = %q, want %q", got.PopoverLabels, want)
	}
	for i := range want {
		if got.PopoverLabels[i] != want[i] {
			t.Fatalf("popover label %d = %q, want %q", i, got.PopoverLabels[i], want[i])
		}
	}
	if got.BoxStyleDefault != "plain" || len(got.BoxStyleLabels) != 2 || got.BoxStyleLabels[0] != "Plain" || got.BoxStyleLabels[1] != "Rounded" {
		t.Fatalf("box style defaults = %q %q, want plain and Plain/Rounded", got.BoxStyleDefault, got.BoxStyleLabels)
	}
	if got.LinkFG != 39 || !got.LinkUnderline {
		t.Fatalf("link defaults = %d, %v", got.LinkFG, got.LinkUnderline)
	}
	if err := got.validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPaneDefaultsLoadFromSpec(t *testing.T) {
	got := SpeccedDefaults.Pane
	if got.MaxCols != 50 {
		t.Fatalf("pane.MaxCols = %d, want 50", got.MaxCols)
	}
	if got.EscKeyTimeout != 50*time.Millisecond {
		t.Fatalf("pane.EscKeyTimeout = %v, want 50ms", got.EscKeyTimeout)
	}
	if got.GuardDuration != time.Second {
		t.Fatalf("pane.GuardDuration = %v, want 1s", got.GuardDuration)
	}
	if got.ViewPanStep != 10 {
		t.Fatalf("pane.ViewPanStep = %d, want 10", got.ViewPanStep)
	}
}

func TestBackgroundDefaultsLoadFromSpec(t *testing.T) {
	got := SpeccedBackground
	if got.PeakFloor != 105 {
		t.Fatalf("background.PeakFloor = %d, want 105", got.PeakFloor)
	}
	if got.PeakMargin != 130 {
		t.Fatalf("background.PeakMargin = %d, want 130", got.PeakMargin)
	}
}
