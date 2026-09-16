// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// TestThemePlainChoiceStyleMatchesDefault asserts that the plain theme produces
// exactly the same ChoiceStyle as the hard-coded DefaultChoiceStyle().
func TestThemePlainChoiceStyleMatchesDefault(t *testing.T) {
	got := loom.Theme("plain").ChoiceStyle()
	want := loom.DefaultChoiceStyle()
	if got != want {
		t.Errorf("Theme(\"plain\").ChoiceStyle() = %+v, want %+v", got, want)
	}
}

// TestThemePlainTableStyleMatchesDefault asserts that the plain theme produces
// exactly the same TableStyle as the hard-coded DefaultTableStyle().
func TestThemePlainTableStyleMatchesDefault(t *testing.T) {
	got := loom.Theme("plain").TableStyle()
	want := loom.DefaultTableStyle()
	if got != want {
		t.Errorf("Theme(\"plain\").TableStyle() = %+v, want %+v", got, want)
	}
}

// TestThemeUnknownFallsBackToPlain verifies that an unknown theme name returns
// the plain theme without panicking.
func TestThemeUnknownFallsBackToPlain(t *testing.T) {
	got := loom.Theme("does-not-exist")
	want := loom.Theme("plain")
	if got != want {
		t.Errorf("Theme(\"does-not-exist\") = %+v, want plain %+v", got, want)
	}
}

func TestThemeColorDistinguishesDefaultFromPaletteZero(t *testing.T) {
	defaultColor := loom.DefaultThemeColor()
	indexedZero := loom.ThemeColorIndex(0)
	if !defaultColor.IsDefault() || indexedZero.IsDefault() {
		t.Fatalf("default/indexed distinction lost: default=%+v indexed=%+v", defaultColor, indexedZero)
	}
	if defaultColor.Color() != loom.ColorReset() {
		t.Errorf("default theme color = %+v, want ColorReset", defaultColor.Color())
	}
	if indexedZero.Color() != loom.ColorIndex(0) {
		t.Errorf("indexed theme color = %+v, want ColorIndex(0)", indexedZero.Color())
	}
	if got := (loom.Style{FG: indexedZero.Color()}).ANSI(); !strings.Contains(got, "\x1b[38;5;0m") {
		t.Errorf("palette index 0 ANSI = %q, want indexed foreground", got)
	}
}

// TestThemeMCHasExpectedColors spot-checks key color roles in the mc theme.
func TestThemeMCHasExpectedColors(t *testing.T) {
	mc := loom.Theme("mc")
	cs := mc.ChoiceStyle()
	if cs.Normal.BG != loom.ColorIndex(69) {
		t.Errorf("mc normal_bg: got %+v, want ColorIndex(69)", cs.Normal.BG)
	}
	if cs.Selected.BG != loom.ColorIndex(75) {
		t.Errorf("mc selected_bg: got %+v, want ColorIndex(75)", cs.Selected.BG)
	}
	if cs.Selected.FG != loom.ColorIndex(16) {
		t.Errorf("mc selected_fg: got %+v, want ColorIndex(16)", cs.Selected.FG)
	}
	if cs.Prompt.FG != loom.ColorIndex(16) {
		t.Errorf("mc prompt_fg: got %+v, want ColorIndex(16)", cs.Prompt.FG)
	}
	if cs.Placeholder != (loom.Style{FG: loom.ColorIndex(243), BG: loom.ColorIndex(75)}) {
		t.Errorf("mc placeholder: got %+v, want fixed medium grey on cyan", cs.Placeholder)
	}
	if cs.Scrollbar.Track != (loom.Style{FG: loom.ColorIndex(69), BG: loom.ColorIndex(69), Dim: true}) ||
		cs.Scrollbar.Thumb != (loom.Style{FG: loom.ColorIndex(75), BG: loom.ColorIndex(69)}) {
		t.Errorf("mc scrollbar: got %+v", cs.Scrollbar)
	}
	if got := mc.FrameStyle().Status; got != (loom.Style{FG: loom.ColorIndex(252), BG: loom.ColorIndex(69)}) {
		t.Errorf("mc frame status: got %+v, want white on blue", got)
	}
	if got := mc.BoxStyle().Border; got != (loom.Style{FG: loom.ColorIndex(252), BG: loom.ColorIndex(69)}) {
		t.Errorf("mc box border: got %+v, want white on blue", got)
	}
	ts := mc.TableStyle()
	if ts.Header.FG != loom.ColorIndex(186) {
		t.Errorf("mc header_fg: got %+v, want ColorIndex(186)", ts.Header.FG)
	}
	if mc.FocusBGColor() != loom.ColorIndex(75) {
		t.Errorf("mc focus_bg: got %+v, want ColorIndex(75)", mc.FocusBGColor())
	}
}

func TestThemeMCVariantsHaveDistinctPalettes(t *testing.T) {
	tests := []struct {
		name                 string
		normalBG, selectedBG loom.Color
	}{
		{name: "mc", normalBG: loom.ColorIndex(69), selectedBG: loom.ColorIndex(75)},
		{name: "mc-classic", normalBG: loom.ColorIndex(27), selectedBG: loom.ColorIndex(51)},
		{name: "mc-dark", normalBG: loom.ColorIndex(19), selectedBG: loom.ColorIndex(37)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			style := loom.Theme(test.name).ChoiceStyle()
			if style.Normal.BG != test.normalBG || style.Selected.BG != test.selectedBG {
				t.Errorf("%s palette: normal=%+v selected=%+v", test.name, style.Normal.BG, style.Selected.BG)
			}
		})
	}
}

// TestSpeccedThemesContainsRequiredThemes asserts that all required themes are loaded.
func TestSpeccedThemesContainsRequiredThemes(t *testing.T) {
	for _, name := range []string{"plain", "mc", "mc-classic", "mc-dark", "julia256"} {
		if _, ok := loom.SpeccedThemes[name]; !ok {
			t.Errorf("SpeccedThemes missing key %q", name)
		}
	}
}

func TestThemeJulia256(t *testing.T) {
	julia := loom.Theme("julia256")
	cs := julia.ChoiceStyle()
	if cs.Normal != (loom.Style{FG: loom.ColorIndex(250), BG: loom.ColorIndex(237)}) {
		t.Errorf("julia256 normal: got %+v, want lightgray on color237", cs.Normal)
	}
	if cs.Selected != (loom.Style{FG: loom.ColorIndex(16), BG: loom.ColorIndex(51)}) {
		t.Errorf("julia256 selected: got %+v, want black on cyan", cs.Selected)
	}
	if cs.Prompt != (loom.Style{FG: loom.ColorIndex(16), BG: loom.ColorIndex(51)}) {
		t.Errorf("julia256 prompt: got %+v, want black on cyan", cs.Prompt)
	}
	ts := julia.TableStyle()
	if ts.Header.FG != loom.ColorIndex(226) || ts.Header.BG != loom.ColorIndex(237) {
		t.Errorf("julia256 header: got %+v, want yellow on color237", ts.Header)
	}
	if got := julia.BoxStyle().Border; got != (loom.Style{FG: loom.ColorIndex(250), BG: loom.ColorIndex(237)}) {
		t.Errorf("julia256 box border: got %+v, want lightgray on color237", got)
	}
	if julia.FocusBGColor() != loom.ColorIndex(240) {
		t.Errorf("julia256 focus_bg: got %+v, want ColorIndex(240)", julia.FocusBGColor())
	}
}
