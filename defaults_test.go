package loom

import (
	"reflect"
	"slices"
	"testing"
	"time"

	"ubunatic.com/loom/spec"
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

func TestGraphDefaultsLoadFromSpec(t *testing.T) {
	want, err := spec.LoadGraphDefaults()
	if err != nil {
		t.Fatal(err)
	}
	if got := SpeccedDefaults.Graph; got != want {
		t.Fatalf("graph defaults = %+v, spec has %+v", got, want)
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

func TestMediaDefaultsValidation(t *testing.T) {
	valid := SpeccedDefaults.Media
	if err := valid.validate(); err != nil {
		t.Fatalf("valid defaults failed validation: %v", err)
	}

	invalidThreshold := valid
	invalidThreshold.LoadingThreshold = 0
	if err := invalidThreshold.validate(); err == nil {
		t.Fatal("accepted zero loading threshold")
	}

	invalidWidth := valid
	invalidWidth.LoadingBarWidth = 0
	if err := invalidWidth.validate(); err == nil {
		t.Fatal("accepted zero loading bar width")
	}

	invalidZoom := valid
	invalidZoom.ZoomFactor = 1.0
	if err := invalidZoom.validate(); err == nil {
		t.Fatal("accepted zoom factor <= 1.0")
	}

	invalidPan := valid
	invalidPan.PanStep = 0
	if err := invalidPan.validate(); err == nil {
		t.Fatal("accepted pan step <= 0")
	}

	missingBinding := valid
	missingBinding.Bindings = make(map[string][]string)
	for k, v := range valid.Bindings {
		missingBinding.Bindings[k] = v
	}
	delete(missingBinding.Bindings, "play")
	if err := missingBinding.validate(); err == nil {
		t.Fatal("accepted missing binding")
	}

	missingLabel := valid
	missingLabel.Labels = make(map[string]string)
	for k, v := range valid.Labels {
		missingLabel.Labels[k] = v
	}
	delete(missingLabel.Labels, "play")
	if err := missingLabel.validate(); err == nil {
		t.Fatal("accepted missing label")
	}
}

func TestMediaDefaultsLoadFromSpec(t *testing.T) {
	got := SpeccedDefaults.Media
	if got.LoadingLabel != "loading" || got.RenderErrorLabel != "render error" || got.LoadingThreshold != 50*time.Millisecond {
		t.Fatalf("media defaults = %+v, want loading/render error/50ms", got)
	}
	if got.LoadingBarWidth != 8 || got.ZoomFactor != 1.25 || got.PanStep != 0.12 {
		t.Fatalf("media numeric defaults = width:%d zoom:%f pan:%f, want 8/1.25/0.12", got.LoadingBarWidth, got.ZoomFactor, got.PanStep)
	}
	wantBindings := map[string][]string{
		"play":     {"space", " "},
		"zoom_in":  {"+", "="},
		"zoom_out": {"-", "_"},
		"pan":      {"←/→/↑/↓", "left", "right", "up", "down"},
		"reset":    {"0"},
	}
	for action, wantKeys := range wantBindings {
		keys := got.Bindings[action]
		if len(keys) != len(wantKeys) {
			t.Fatalf("media.bindings.%s = %v, want %v", action, keys, wantKeys)
		}
		for i, k := range wantKeys {
			if keys[i] != k {
				t.Fatalf("media.bindings.%s[%d] = %q, want %q", action, i, keys[i], k)
			}
		}
	}
	wantLabels := map[string]string{
		"play":     "Play/Pause",
		"zoom_in":  "Zoom in",
		"zoom_out": "Zoom out",
		"pan":      "Pan",
		"reset":    "Reset",
	}
	for action, wantLabel := range wantLabels {
		if got.Labels[action] != wantLabel {
			t.Fatalf("media.labels.%s = %q, want %q", action, got.Labels[action], wantLabel)
		}
	}
}

func TestRichTextEditDefaultsLoadFromSpec(t *testing.T) {
	got := SpeccedDefaults.RichTextEdit
	if got.MenuFocusKey != "" || got.PopupOpenKey != "ctrl-space" {
		t.Fatalf("popover key defaults = %q/%q", got.MenuFocusKey, got.PopupOpenKey)
	}
	if got.ToolbarFG != 15 || got.ToolbarBG != 239 || got.SeparatorGlyph != "▏" || got.SeparatorFG != 8 || got.SeparatorBG != 239 || got.PointerUpGlyph != "▲" || got.PointerDownGlyph != "▼" || got.PointerFG != 8 || got.PopoverFocusFG != 15 {
		t.Fatalf("rich text edit defaults = %+v", got)
	}
	want := []string{"B", "I", "U", "S", "🎨", "Frame", "Link"}
	if len(got.PopoverLabels) != len(want) {
		t.Fatalf("popover labels = %q, want %q", got.PopoverLabels, want)
	}
	for i := range want {
		if got.PopoverLabels[i] != want[i] {
			t.Fatalf("popover label %d = %q, want %q", i, got.PopoverLabels[i], want[i])
		}
	}
	if want := []string{"b", "i", "u", "s", "", "f", "l", ""}; !reflect.DeepEqual(got.PopoverHotkeys, want) {
		t.Fatalf("popover hotkeys = %q, want %q", got.PopoverHotkeys, want)
	}
	if want := []int{0, 0, 0, 0, -1, 0, 0, -1}; !reflect.DeepEqual(got.PopoverHotkeyIndices, want) {
		t.Fatalf("popover hotkey indices = %v, want %v", got.PopoverHotkeyIndices, want)
	}
	if got.PopoverHamburgerLabel != "≡" || len(got.PopoverHamburgerItems) != 1 || got.PopoverHamburgerItems[0] != "Select all" || len(got.PopoverHamburgerActions) != 1 || got.PopoverHamburgerActions[0] != "select_all" {
		t.Fatalf("hamburger defaults = %q/%q/%q", got.PopoverHamburgerLabel, got.PopoverHamburgerItems, got.PopoverHamburgerActions)
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

func TestColorPickerDefaultsLoadAndValidate(t *testing.T) {
	got := SpeccedDefaults.ColorPicker
	if err := got.validate(); err != nil {
		t.Fatal(err)
	}
	if len(got.Palette) != 16 || got.Palette[0] != 0 || got.Palette[15] != 15 {
		t.Fatalf("picker palette = %v, want indexes 0 through 15", got.Palette)
	}
	if !got.ShowPreview || got.Labels.Preview == "" || got.Glyphs.Swatch == "" || got.Glyphs.Reset != "∅ " || got.Glyphs.EdgeLower != "▂" || got.Glyphs.EdgeUpper != "🮂" || got.InactiveMarkerFG < 0 || got.InactiveMarkerFG > 255 || got.LeadingPadding != 1 || got.TrailingPadding != 1 || got.PreviewPadding != 1 || got.PreviewWidth != 9 {
		t.Fatalf("picker preview/glyphs incomplete: %+v", got)
	}
	if got.Bindings["apply"][0] != "enter" || got.Bindings["select"][0] != "space" || got.Bindings["row_bg"][0] != "down" || got.Bindings["row_fg"][0] != "up" || got.Bindings["home"][0] != "home" || got.Bindings["close"][0] != "esc" {
		t.Fatalf("picker bindings incomplete: %+v", got.Bindings)
	}
	for _, removed := range []string{"select_bg", "target_fg", "target_bg", "toggle_target", "mode_bg", "mode_fg", "mode_toggle"} {
		if _, ok := got.Bindings[removed]; ok {
			t.Errorf("removed picker binding %q remains", removed)
		}
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

func TestSessionDefaultsLoadFromSpec(t *testing.T) {
	s := SpeccedDefaults.Session
	if s.Bar != "always" || !slices.Equal(s.PrefixBindings.Help, []string{"?"}) || !slices.Equal(s.PrefixBindings.ToggleMouse, []string{"m"}) {
		t.Fatalf("session bar bindings = %q/%v/%v, want always/?/m", s.Bar, s.PrefixBindings.Help, s.PrefixBindings.ToggleMouse)
	}
	if s.ScrollbackLines != 5000 || !slices.Equal(s.PrefixBindings.Search, []string{"f"}) {
		t.Fatalf("session scrollback/search = %d/%v, want 5000/[f]", s.ScrollbackLines, s.PrefixBindings.Search)
	}
	if !slices.Equal(s.Search.Bindings.Next, []string{"enter", "up"}) || !slices.Equal(s.Search.Bindings.Previous, []string{"down"}) || !slices.Equal(s.Search.Bindings.Close, []string{"esc"}) {
		t.Fatalf("session search bindings = %+v", s.Search.Bindings)
	}
	if s.Prefix != "ctrl-b" {
		t.Fatalf("session.prefix = %q, want ctrl-b", s.Prefix)
	}
	if len(s.Bindings.Screenshot) == 0 || s.Bindings.Screenshot[0] != SpeccedDefaults.Editor.HotkeyScreenshotBinding {
		t.Fatalf("session.bindings.screenshot = %v, want editor binding %q", s.Bindings.Screenshot, SpeccedDefaults.Editor.HotkeyScreenshotBinding)
	}
	if !slices.Equal(s.Bindings.Quit, []string{"ctrl-q"}) {
		t.Fatalf("session.bindings.quit = %v, want [ctrl-q]", s.Bindings.Quit)
	}
	if !slices.Equal(s.Bindings.NewTab, []string{"ctrl-n"}) {
		t.Fatalf("session.bindings.new_tab = %v, want [ctrl-n]", s.Bindings.NewTab)
	}
	if !slices.Equal(s.PrefixBindings.Screenshot, []string{"s"}) {
		t.Fatalf("session.prefix_bindings.screenshot = %v, want [s]", s.PrefixBindings.Screenshot)
	}
	if s.Mouse != "safe" {
		t.Fatalf("session.mouse = %q, want 'safe'", s.Mouse)
	}
	if !slices.Equal(s.PrefixBindings.Quit, []string{"q"}) {
		t.Fatalf("session.prefix_bindings.quit = %v, want [q]", s.PrefixBindings.Quit)
	}
	if !slices.Equal(s.PrefixBindings.SendPrefix, []string{"ctrl-b"}) {
		t.Fatalf("session.prefix_bindings.send_prefix = %v, want [ctrl-b]", s.PrefixBindings.SendPrefix)
	}
	if !slices.Equal(s.PrefixBindings.NewTab, []string{"c"}) {
		t.Fatalf("session.prefix_bindings.new_tab = %v, want [c]", s.PrefixBindings.NewTab)
	}
	if !slices.Equal(s.PrefixBindings.NextTab, []string{"n"}) {
		t.Fatalf("session.prefix_bindings.next_tab = %v, want [n]", s.PrefixBindings.NextTab)
	}
	if !slices.Equal(s.PrefixBindings.PreviousTab, []string{"p"}) {
		t.Fatalf("session.prefix_bindings.previous_tab = %v, want [p]", s.PrefixBindings.PreviousTab)
	}
	if !slices.Equal(s.PrefixBindings.CloseTab, []string{"x"}) {
		t.Fatalf("session.prefix_bindings.close_tab = %v, want [x]", s.PrefixBindings.CloseTab)
	}
	if !slices.Equal(s.PrefixBindings.SelectTab, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}) {
		t.Fatalf("session.prefix_bindings.select_tab = %v, want 1..9", s.PrefixBindings.SelectTab)
	}
	wantSelectTab := []string{"alt-1", "alt-2", "alt-3", "alt-4", "alt-5", "alt-6", "alt-7", "alt-8", "alt-9"}
	if !slices.Equal(s.Bindings.SelectTab, wantSelectTab) {
		t.Fatalf("session.bindings.select_tab = %v, want %v", s.Bindings.SelectTab, wantSelectTab)
	}
	if err := s.validate(); err != nil {
		t.Fatalf("session.validate() failed: %v", err)
	}

	invalid := s
	invalid.Prefix = ""
	if err := invalid.validate(); err == nil {
		t.Fatal("expected error on empty session prefix")
	}
}
