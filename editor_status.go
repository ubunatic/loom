// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// NewEditorStatusBar creates compact, spec-defined theme, mouse capture and
// alternate-screen indicators. Theme detail drops first on narrow screens.
func NewEditorStatusBar(cfg EditorConfig) *HintBar {
	icons := SpeccedDefaults.Editor.StatusIcons
	mouse, alt := icons.MouseOff, icons.AltOff
	if cfg.MouseGrab {
		mouse = icons.MouseOn
	}
	if cfg.AltScreen {
		alt = icons.AltOn
	}
	bar := NewHintBar(
		HintEntry{Key: icons.Theme, Detail: cfg.Theme, DropPriority: 1},
		HintEntry{Key: mouse},
		HintEntry{Key: alt},
	)
	bar.Style.Cap = Style{FG: ColorIndex(uint8(icons.FG)), BG: ColorIndex(uint8(icons.BG))}
	bar.Style.Label = Theme(cfg.Theme).ChoiceStyle().Normal
	return bar
}
