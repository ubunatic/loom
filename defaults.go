// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed spec/defaults.yaml
var defaultsYAML []byte

// LibDefaults represents specced runtime defaults loaded from spec/defaults.yaml.
type LibDefaults struct {
	KeyCaps      KeyCapDefaults       `yaml:"keycaps"`
	QuitKeys     []string             `yaml:"quit_keys"`
	EscapeQuits  bool                 `yaml:"escape_quits"`
	Pane         PaneDefaults         `yaml:"pane"`
	Scrollbar    ScrollbarDefaults    `yaml:"scrollbar"`
	Mouse        MouseDefaults        `yaml:"mouse"`
	Splash       SplashDefaults       `yaml:"splash"`
	Spinner      SpinnerDefaults      `yaml:"spinner"`
	ProgressBar  ProgressBarDefaults  `yaml:"progress_bar"`
	Media        MediaDefaults        `yaml:"media"`
	Clock        ClockDefaults        `yaml:"clock"`
	PaintCanvas  PaintCanvasDefaults  `yaml:"paint_canvas"`
	Editor       EditorDefaults       `yaml:"editor"`
	SearchBar    SearchBarDefaults    `yaml:"search_bar"`
	RichTextEdit RichTextEditDefaults `yaml:"rich_text_edit"`
}

// KeyCapDefaults defines the modifier glyphs KeyCap renders, in display order.
type KeyCapDefaults struct {
	Modifiers []KeyCapModifier `yaml:"modifiers"`
}

// KeyCapModifier maps a binding modifier name (shift, ctrl, alt) to its glyph.
type KeyCapModifier struct {
	Name  string `yaml:"name"`
	Glyph string `yaml:"glyph"`
}

// EditorDefaults defines specced defaults for the Loom editor.
type EditorDefaults struct {
	Theme                   string `yaml:"theme"`
	MouseGrab               bool   `yaml:"mousegrab"`
	AltScreen               bool   `yaml:"altscreen"`
	HotkeyFilesBinding      string `yaml:"hotkey_files_binding"`
	HotkeyFilesLabel        string `yaml:"hotkey_files_label"`
	HotkeySearchKey         string `yaml:"hotkey_search_key"`
	HotkeySearchBinding     string `yaml:"hotkey_search_binding"`
	HotkeySearchLabel       string `yaml:"hotkey_search_label"`
	HotkeyBoxBinding        string `yaml:"hotkey_box_binding"`
	HotkeyBoxLabel          string `yaml:"hotkey_box_label"`
	SearchPrompt            string `yaml:"search_prompt"`
	SearchPlaceholder       string `yaml:"search_placeholder"`
	HotkeyScreenshotBinding string `yaml:"hotkey_screenshot_binding"`
	HotkeyScreenshotLabel   string `yaml:"hotkey_screenshot_label"`
}

// RichTextEditDefaults defines selection and popover presentation defaults.
type RichTextEditDefaults struct {
	SelectionBG           int      `yaml:"selection_bg"`
	ToolbarFG             int      `yaml:"toolbar_fg"`
	ToolbarBG             int      `yaml:"toolbar_bg"`
	SeparatorGlyph        string   `yaml:"separator_glyph"`
	SeparatorFG           int      `yaml:"separator_fg"`
	SeparatorBG           int      `yaml:"separator_bg"`
	PointerUpGlyph        string   `yaml:"pointer_up_glyph"`
	PointerDownGlyph      string   `yaml:"pointer_down_glyph"`
	PointerFG             int      `yaml:"pointer_fg"`
	PopoverLabels         []string `yaml:"popover_labels"`
	BoxStyleLabels        []string `yaml:"box_style_labels"`
	BoxStyleDefault       string   `yaml:"box_style_default"`
	GhostCursorEnabled    bool     `yaml:"ghost_cursor_enabled"`
	PopoverFocusFG        int      `yaml:"popover_focus_fg"`
	PopoverFocusBG        int      `yaml:"popover_focus_bg"`
	LinkFG                int      `yaml:"link_fg"`
	LinkUnderline         bool     `yaml:"link_underline"`
	SavePopupMaxWidth     int      `yaml:"save_popup_max_width"`
	SavePopupMaxHeight    int      `yaml:"save_popup_max_height"`
	StateUntitledLabel    string   `yaml:"state_untitled_label"`
	StateSavedLabel       string   `yaml:"state_saved_label"`
	StateModifiedLabel    string   `yaml:"state_modified_label"`
	StateErrorLabel       string   `yaml:"state_error_label"`
	HotkeyHelpBinding     string   `yaml:"hotkey_help_binding"`
	HotkeyHelpLabel       string   `yaml:"hotkey_help_label"`
	HotkeySaveBinding     string   `yaml:"hotkey_save_binding"`
	HotkeySaveLabel       string   `yaml:"hotkey_save_label"`
	HotkeySaveAsBinding   string   `yaml:"hotkey_save_as_binding"`
	HotkeySaveAsLabel     string   `yaml:"hotkey_save_as_label"`
	HotkeyViewEditBinding string   `yaml:"hotkey_view_edit_binding"`
	HotkeyQuitBinding     string   `yaml:"hotkey_quit_binding"`
	HotkeyQuitLabel       string   `yaml:"hotkey_quit_label"`
}

func (d RichTextEditDefaults) validate() error {
	for _, color := range []struct {
		name  string
		value int
	}{
		{"selection_bg", d.SelectionBG}, {"toolbar_fg", d.ToolbarFG}, {"toolbar_bg", d.ToolbarBG},
		{"separator_fg", d.SeparatorFG}, {"separator_bg", d.SeparatorBG}, {"pointer_fg", d.PointerFG}, {"popover_focus_fg", d.PopoverFocusFG}, {"popover_focus_bg", d.PopoverFocusBG}, {"link_fg", d.LinkFG},
	} {
		if color.value < 0 || color.value > 255 {
			return fmt.Errorf("rich_text_edit.%s must be between 0 and 255", color.name)
		}
	}
	if d.SeparatorGlyph == "" || d.PointerUpGlyph == "" || d.PointerDownGlyph == "" {
		return fmt.Errorf("rich_text_edit glyphs must not be empty")
	}
	if len(d.PopoverLabels) != 9 {
		return fmt.Errorf("rich_text_edit.popover_labels must contain 9 labels")
	}
	if len(d.BoxStyleLabels) != 2 || d.BoxStyleLabels[0] == "" || d.BoxStyleLabels[1] == "" {
		return fmt.Errorf("rich_text_edit.box_style_labels must contain two non-empty labels")
	}
	if d.BoxStyleDefault != "plain" && d.BoxStyleDefault != "rounded" {
		return fmt.Errorf("rich_text_edit.box_style_default must be plain or rounded")
	}
	if d.StateUntitledLabel == "" || d.StateSavedLabel == "" || d.StateModifiedLabel == "" || d.StateErrorLabel == "" {
		return fmt.Errorf("rich_text_edit state labels must be non-empty")
	}
	if d.SavePopupMaxWidth < 8 || d.SavePopupMaxHeight < 7 {
		return fmt.Errorf("rich_text_edit save popup dimensions must fit a bordered picker")
	}
	for _, label := range d.PopoverLabels {
		if label == "" {
			return fmt.Errorf("rich_text_edit.popover_labels must not contain empty labels")
		}
	}
	return nil
}

// SearchBarDefaults defines specced defaults for the SearchBar widget.
type SearchBarDefaults struct {
	Prompt      string `yaml:"prompt"`
	Placeholder string `yaml:"placeholder"`
}

// PaintCanvasDefaults defines stroke tuning controls.
type PaintCanvasDefaults struct {
	TuneLabel       string `yaml:"tune_label"`
	SmoothingLabel  string `yaml:"smoothing_label"`
	SmoothingLevels []int  `yaml:"smoothing_levels"`
}

// ClockDefaults defines the labels for timer and stopwatch controls.
type ClockDefaults struct {
	Start string `yaml:"start"`
	Stop  string `yaml:"stop"`
	Reset string `yaml:"reset"`
}

// MediaDefaults defines user-facing media status and rendering timing defaults.
type MediaDefaults struct {
	LoadingLabel     string        `yaml:"loading_label"`
	RenderErrorLabel string        `yaml:"render_error_label"`
	LoadingThreshold time.Duration `yaml:"loading_threshold"`
}

// MouseDefaults defines specced defaults for mouse gestures.
type MouseDefaults struct {
	DoubleClickInterval time.Duration `yaml:"double_click_interval"`
	MovementTolerance   int           `yaml:"movement_tolerance"`
}

func (d MouseDefaults) validate() error {
	if d.DoubleClickInterval <= 0 {
		return fmt.Errorf("mouse.double_click_interval must be positive")
	}
	if d.MovementTolerance < 0 {
		return fmt.Errorf("mouse.movement_tolerance must not be negative")
	}
	return nil
}

// ScrollbarDefaults defines scrollbar visibility and its one-cell glyphs.
type ScrollbarDefaults struct {
	Mode           ScrollbarMode `yaml:"mode"`
	ForegroundChar string        `yaml:"foreground_char"`
	BackgroundChar string        `yaml:"background_char"`
}

func (d ScrollbarDefaults) validate() error {
	if d.Mode != ScrollbarAuto && d.Mode != ScrollbarAlways && d.Mode != ScrollbarNever {
		return fmt.Errorf("scrollbar.mode must be %q, %q, or %q", ScrollbarAuto, ScrollbarAlways, ScrollbarNever)
	}
	for _, field := range []struct{ name, glyph string }{
		{"foreground_char", d.ForegroundChar},
		{"background_char", d.BackgroundChar},
	} {
		if len(textClusters(field.glyph)) != 1 || StringWidth(field.glyph) != 1 || strings.TrimSpace(field.glyph) == "" {
			return fmt.Errorf("scrollbar.%s must be one visible terminal cell", field.name)
		}
	}
	return nil
}

// PaneDefaults defines specced defaults for Pane.
type PaneDefaults struct {
	MaxCols       int           `yaml:"max_cols"`
	EscKeyTimeout time.Duration `yaml:"esc_key_timeout"`
	GuardDuration time.Duration `yaml:"guard_duration"`
	ViewPanStep   int           `yaml:"view_pan_step"`
}

// SplashDefaults defines specced defaults for splash lifecycle and widgets.
type SplashDefaults struct {
	BracketWidth int           `yaml:"bracket_width"`
	PillGap      int           `yaml:"pill_gap"`
	FooterText   string        `yaml:"footer_text"`
	StepText     string        `yaml:"step_text"`
	TickInterval time.Duration `yaml:"tick_interval"`
	HoldDuration time.Duration `yaml:"hold_duration"`
}

// ProgressBarDefaults defines specced defaults for the ProgressBar widget.
type ProgressBarDefaults struct {
	Width        int           `yaml:"width"`
	DonePattern  string        `yaml:"done_pattern"`
	DemoInterval time.Duration `yaml:"demo_interval"`
	DemoStep     float64       `yaml:"demo_step"`
}

// SpinnerDefaults defines specced animation timing for Spinner.
type SpinnerDefaults struct {
	TickInterval time.Duration `yaml:"tick_interval"`
}

// SpeccedDefaults holds the loaded immutable defaults from spec/defaults.yaml.
var SpeccedDefaults = func() LibDefaults {
	var defs LibDefaults
	if err := yaml.Unmarshal(defaultsYAML, &defs); err != nil {
		panic(fmt.Sprintf("loom: parse spec/defaults.yaml: %v", err))
	}
	if err := defs.Scrollbar.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.Mouse.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.RichTextEdit.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	return defs
}()
