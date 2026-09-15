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

// ThemeColor distinguishes the terminal default from all 256 palette indices.
// Its zero value is the terminal default.
type ThemeColor struct {
	index   uint8
	indexed bool
}

// DefaultThemeColor returns the terminal-default theme color.
func DefaultThemeColor() ThemeColor { return ThemeColor{} }

// ThemeColorIndex returns an indexed theme color, including palette index 0.
func ThemeColorIndex(index uint8) ThemeColor {
	return ThemeColor{index: index, indexed: true}
}

// Color converts a theme color to its terminal rendering color.
func (c ThemeColor) Color() Color {
	if !c.indexed {
		return ColorReset()
	}
	return ColorIndex(c.index)
}

// IsDefault reports whether the color uses the terminal default.
func (c ThemeColor) IsDefault() bool { return !c.indexed }

// UnmarshalYAML accepts "default" or an integer palette index from 0 to 255.
func (c *ThemeColor) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode && value.Tag == "!!str" {
		if value.Value != "default" {
			return fmt.Errorf("theme color %q: want default or palette index", value.Value)
		}
		*c = DefaultThemeColor()
		return nil
	}
	var index uint8
	if err := value.Decode(&index); err != nil {
		return fmt.Errorf("theme color: want default or palette index: %w", err)
	}
	*c = ThemeColorIndex(index)
	return nil
}

// ThemeColors is the Go representation of one theme entry in spec/themes.yaml.
type ThemeColors struct {
	NormalFG           ThemeColor `yaml:"normal_fg"`
	NormalBG           ThemeColor `yaml:"normal_bg"`
	SelectedFG         ThemeColor `yaml:"selected_fg"`
	SelectedBG         ThemeColor `yaml:"selected_bg"`
	SelectedBold       bool       `yaml:"selected_bold"`
	HeaderFG           ThemeColor `yaml:"header_fg"`
	HeaderBG           ThemeColor `yaml:"header_bg"`
	HeaderBold         bool       `yaml:"header_bold"`
	PromptFG           ThemeColor `yaml:"prompt_fg"`
	PromptBG           ThemeColor `yaml:"prompt_bg"`
	PlaceholderFG      ThemeColor `yaml:"placeholder_fg"`
	PlaceholderBG      ThemeColor `yaml:"placeholder_bg"`
	PlaceholderDim     bool       `yaml:"placeholder_dim"`
	ScrollbarTrackFG   ThemeColor `yaml:"scrollbar_track_fg"`
	ScrollbarTrackBG   ThemeColor `yaml:"scrollbar_track_bg"`
	ScrollbarTrackDim  bool       `yaml:"scrollbar_track_dim"`
	ScrollbarThumbFG   ThemeColor `yaml:"scrollbar_thumb_fg"`
	ScrollbarThumbBG   ThemeColor `yaml:"scrollbar_thumb_bg"`
	ScrollbarThumbBold bool       `yaml:"scrollbar_thumb_bold"`
	StatusFG           ThemeColor `yaml:"status_fg"`
	StatusBG           ThemeColor `yaml:"status_bg"`
	StatusBold         bool       `yaml:"status_bold"`
	StatusDim          bool       `yaml:"status_dim"`
	BorderFG           ThemeColor `yaml:"border_fg"`
	BorderBG           ThemeColor `yaml:"border_bg"`
	FocusBG            ThemeColor `yaml:"focus_bg"`
}

// themesFile is the YAML wrapper for spec/themes.yaml.
type themesFile struct {
	Themes map[string]ThemeColors `yaml:"themes"`
}

// ChoiceStyle returns a ChoiceStyle derived from the theme's color roles.
func (t ThemeColors) ChoiceStyle() ChoiceStyle {
	return ChoiceStyle{
		Normal:      Style{FG: t.NormalFG.Color(), BG: t.NormalBG.Color()},
		Selected:    Style{FG: t.SelectedFG.Color(), BG: t.SelectedBG.Color(), Bold: t.SelectedBold},
		Prompt:      Style{FG: t.PromptFG.Color(), BG: t.PromptBG.Color()},
		Placeholder: Style{FG: t.PlaceholderFG.Color(), BG: t.PlaceholderBG.Color(), Dim: t.PlaceholderDim},
		Scrollbar:   t.ScrollbarStyle(),
		Border:      Style{FG: t.BorderFG.Color(), BG: t.BorderBG.Color()},
	}
}

// ScrollbarStyle returns scrollbar track and thumb styles derived from the theme.
func (t ThemeColors) ScrollbarStyle() ScrollbarStyle {
	return ScrollbarStyle{
		Track: Style{FG: t.ScrollbarTrackFG.Color(), BG: t.ScrollbarTrackBG.Color(), Dim: t.ScrollbarTrackDim},
		Thumb: Style{FG: t.ScrollbarThumbFG.Color(), BG: t.ScrollbarThumbBG.Color(), Bold: t.ScrollbarThumbBold},
	}
}

// FrameStyle returns frame background, title, and status styles derived from the theme.
func (t ThemeColors) FrameStyle() FrameStyle {
	return FrameStyle{
		Background: Style{FG: t.NormalFG.Color(), BG: t.NormalBG.Color()},
		Title:      Style{FG: t.BorderFG.Color(), BG: t.BorderBG.Color()},
		Status:     Style{FG: t.StatusFG.Color(), BG: t.StatusBG.Color(), Bold: t.StatusBold, Dim: t.StatusDim},
	}
}

// BoxStyle returns box background, border, title, and footer styles derived from the theme.
func (t ThemeColors) BoxStyle() BoxStyle {
	normal := Style{FG: t.NormalFG.Color(), BG: t.NormalBG.Color()}
	border := Style{FG: t.BorderFG.Color(), BG: t.BorderBG.Color()}
	return BoxStyle{Background: normal, Border: border, Title: border, Footer: normal}
}

// TableStyle returns a TableStyle derived from the theme's color roles.
func (t ThemeColors) TableStyle() TableStyle {
	return TableStyle{
		Normal:     Style{FG: t.NormalFG.Color(), BG: t.NormalBG.Color()},
		Selected:   Style{FG: t.SelectedFG.Color(), BG: t.SelectedBG.Color(), Bold: t.SelectedBold},
		Header:     Style{FG: t.HeaderFG.Color(), BG: t.HeaderBG.Color(), Bold: t.HeaderBold},
		SortHeader: Style{FG: t.HeaderFG.Color(), BG: t.HeaderBG.Color(), Bold: t.HeaderBold, Underline: true},
		Prompt:     Style{FG: t.PromptFG.Color(), BG: t.PromptBG.Color()},
	}
}

// FocusBGColor returns the focused-cell background Color for use in Grid widgets.
func (t ThemeColors) FocusBGColor() Color {
	return t.FocusBG.Color()
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
