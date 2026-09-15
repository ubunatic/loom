// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed spec/themes.yaml
var themesYAML []byte

// ThemeColors is the Go representation of one theme entry in spec/themes.yaml.
type ThemeColors struct {
	NormalFG     uint8 `yaml:"normal_fg"`
	NormalBG     uint8 `yaml:"normal_bg"`
	SelectedFG   uint8 `yaml:"selected_fg"`
	SelectedBG   uint8 `yaml:"selected_bg"`
	SelectedBold bool  `yaml:"selected_bold"`
	HeaderFG     uint8 `yaml:"header_fg"`
	HeaderBG     uint8 `yaml:"header_bg"`
	HeaderBold   bool  `yaml:"header_bold"`
	PromptFG     uint8 `yaml:"prompt_fg"`
	PromptBG     uint8 `yaml:"prompt_bg"`
	BorderFG     uint8 `yaml:"border_fg"`
	BorderBG     uint8 `yaml:"border_bg"`
	FocusBG      uint8 `yaml:"focus_bg"`
}

// themesFile is the YAML wrapper for spec/themes.yaml.
type themesFile struct {
	Themes map[string]ThemeColors `yaml:"themes"`
}

// themeColor converts a spec palette index to a Color.
// 0 is the "terminal default" sentinel and returns ColorReset().
func themeColor(index uint8) Color {
	if index == 0 {
		return ColorReset()
	}
	return ColorIndex(index)
}

// ChoiceStyle returns a ChoiceStyle derived from the theme's color roles.
func (t ThemeColors) ChoiceStyle() ChoiceStyle {
	return ChoiceStyle{
		Normal:   Style{FG: themeColor(t.NormalFG), BG: themeColor(t.NormalBG)},
		Selected: Style{FG: themeColor(t.SelectedFG), BG: themeColor(t.SelectedBG), Bold: t.SelectedBold},
		Prompt:   Style{FG: themeColor(t.PromptFG), BG: themeColor(t.PromptBG)},
		Border:   Style{FG: themeColor(t.BorderFG), BG: themeColor(t.BorderBG)},
	}
}

// TableStyle returns a TableStyle derived from the theme's color roles.
func (t ThemeColors) TableStyle() TableStyle {
	return TableStyle{
		Normal:     Style{FG: themeColor(t.NormalFG), BG: themeColor(t.NormalBG)},
		Selected:   Style{FG: themeColor(t.SelectedFG), BG: themeColor(t.SelectedBG), Bold: t.SelectedBold},
		Header:     Style{FG: themeColor(t.HeaderFG), BG: themeColor(t.HeaderBG), Bold: t.HeaderBold},
		SortHeader: Style{FG: themeColor(t.HeaderFG), BG: themeColor(t.HeaderBG), Bold: t.HeaderBold, Underline: true},
		Prompt:     Style{FG: themeColor(t.PromptFG), BG: themeColor(t.PromptBG)},
	}
}

// FocusBGColor returns the focused-cell background Color for use in Grid widgets.
func (t ThemeColors) FocusBGColor() Color {
	return themeColor(t.FocusBG)
}

// SpeccedThemes holds all themes loaded from spec/themes.yaml at init time.
// Keys are theme names ("plain", "mc", …).
var SpeccedThemes = func() map[string]ThemeColors {
	var f themesFile
	if err := yaml.Unmarshal(themesYAML, &f); err != nil {
		panic(fmt.Sprintf("loom: parse spec/themes.yaml: %v", err))
	}
	if _, ok := f.Themes["plain"]; !ok {
		panic("loom: spec/themes.yaml: required base theme \"plain\" is missing")
	}
	return f.Themes
}()

// Theme returns the named theme from SpeccedThemes.
// If name is not found, it falls back to the "plain" theme.
func Theme(name string) ThemeColors {
	if t, ok := SpeccedThemes[name]; ok {
		return t
	}
	return SpeccedThemes["plain"]
}
