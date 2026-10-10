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
	want := []string{"B", "I", "U", "S", "Link", "#FG", "#BG", "Box"}
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
	if got.GhostCursorEnabled {
		t.Fatal("ghost cursor default = enabled, want disabled")
	}
	if got.LinkFG != 39 || !got.LinkUnderline {
		t.Fatalf("link defaults = %d, %v", got.LinkFG, got.LinkUnderline)
	}
	if got.SavePopupMaxWidth != 72 || got.SavePopupMaxHeight != 18 {
		t.Fatalf("save popup max dimensions = %dx%d, want 72x18", got.SavePopupMaxWidth, got.SavePopupMaxHeight)
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

func TestQuitDefaultsLoadFromSpec(t *testing.T) {
	got := SpeccedDefaults.QuitKeys
	if len(got) != 2 || got[0] != "ctrl-q" || got[1] != "f10" {
		t.Fatalf("QuitKeys = %v, want [ctrl-q f10]", got)
	}
	if SpeccedDefaults.EscapeQuits {
		t.Fatal("EscapeQuits = true, want false")
	}
}

func TestEditorRefinedBindings(t *testing.T) {
	d := SpeccedDefaults.Editor
	if d.HotkeyFilesKey != "^O" || d.HotkeyFilesBinding != "ctrl-o" || d.HotkeyFilesSecondaryBinding != "f2" || d.HotkeySearchKey != "^F" || d.HotkeySearchBinding != "ctrl-f" || d.HotkeySearchSecondaryBinding != "f3" || d.HotkeyBoxKey != "^D" || d.HotkeyBoxBinding != "ctrl-d" {
		t.Fatalf("editor bindings: %+v", d)
	}
	if SpeccedDefaults.RichTextEdit.BoxDrawLabel != "Draw" {
		t.Fatal("missing nested Draw label")
	}
}

func TestWidgetDefaultsLoadAndValidate(t *testing.T) {
	ni := SpeccedDefaults.NumberInput
	if ni.Step != 1.0 || ni.LeftGlyph != "◂" || ni.RightGlyph != "▸" || ni.InvalidError != "invalid number" {
		t.Fatalf("unexpected NumberInput defaults: %+v", ni)
	}
	if err := ni.validate(); err != nil {
		t.Fatalf("NumberInput validation failed: %v", err)
	}

	tog := SpeccedDefaults.Toggle
	if tog.OnMark != "[✓]" || tog.OffMark != "[ ]" {
		t.Fatalf("unexpected Toggle defaults: %+v", tog)
	}
	if err := tog.validate(); err != nil {
		t.Fatalf("Toggle validation failed: %v", err)
	}

	pag := SpeccedDefaults.Paginator
	if pag.ActiveDot != "●" || pag.InactiveDot != "○" || pag.NumericFormat != "%d/%d" {
		t.Fatalf("unexpected Paginator defaults: %+v", pag)
	}
	if err := pag.validate(); err != nil {
		t.Fatalf("Paginator validation failed: %v", err)
	}

	tr := SpeccedDefaults.Tree
	if tr.CollapsedMarker != "▶ " || tr.ExpandedMarker != "▼ " || tr.LeafMarker != "  " {
		t.Fatalf("unexpected Tree defaults: %+v", tr)
	}
	if err := tr.validate(); err != nil {
		t.Fatalf("Tree validation failed: %v", err)
	}

	sp := SpeccedDefaults.Split
	if sp.VerticalDivider != "─" || sp.HorizontalDivider != "│" || sp.DefaultRatio != 0.5 || sp.DefaultGap != 1 {
		t.Fatalf("unexpected Split defaults: %+v", sp)
	}
	if err := sp.validate(); err != nil {
		t.Fatalf("Split validation failed: %v", err)
	}

	ch := SpeccedDefaults.Choice
	if ch.SelectionMarker != "▶ " || ch.UnselectedMarker != "  " || ch.CheckedMarker != "[✓] " || ch.UncheckedMarker != "[ ] " {
		t.Fatalf("unexpected Choice defaults: %+v", ch)
	}
	if err := ch.validate(); err != nil {
		t.Fatalf("Choice validation failed: %v", err)
	}
}
