// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
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

// TestThemeMCHasExpectedColors spot-checks key color roles in the mc theme.
func TestThemeMCHasExpectedColors(t *testing.T) {
	mc := loom.Theme("mc")
	cs := mc.ChoiceStyle()
	if cs.Normal.BG != loom.ColorIndex(27) {
		t.Errorf("mc normal_bg: got %+v, want ColorIndex(27)", cs.Normal.BG)
	}
	if cs.Selected.BG != loom.ColorIndex(51) {
		t.Errorf("mc selected_bg: got %+v, want ColorIndex(51)", cs.Selected.BG)
	}
	ts := mc.TableStyle()
	if ts.Header.FG != loom.ColorIndex(226) {
		t.Errorf("mc header_fg: got %+v, want ColorIndex(226)", ts.Header.FG)
	}
	if mc.FocusBGColor() != loom.ColorIndex(33) {
		t.Errorf("mc focus_bg: got %+v, want ColorIndex(33)", mc.FocusBGColor())
	}
}

// TestSpeccedThemesContainsBothThemes asserts that both required themes are loaded.
func TestSpeccedThemesContainsBothThemes(t *testing.T) {
	for _, name := range []string{"plain", "mc"} {
		if _, ok := loom.SpeccedThemes[name]; !ok {
			t.Errorf("SpeccedThemes missing key %q", name)
		}
	}
}
