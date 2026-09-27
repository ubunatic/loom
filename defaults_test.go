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
