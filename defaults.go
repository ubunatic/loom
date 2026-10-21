// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"ubunatic.com/loom/measure"
	"ubunatic.com/loom/spec"
)

var defaultsYAML = func() []byte {
	data, err := spec.DefaultsYAML()
	if err != nil {
		panic(fmt.Sprintf("loom: read spec/defaults.yaml: %v", err))
	}
	return data
}()

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
	Graph        spec.GraphDefaults   `yaml:"graph"`
	ProgressBar  ProgressBarDefaults  `yaml:"progress_bar"`
	Media        MediaDefaults        `yaml:"media"`
	Clock        ClockDefaults        `yaml:"clock"`
	PaintCanvas  PaintCanvasDefaults  `yaml:"paint_canvas"`
	Editor       EditorDefaults       `yaml:"editor"`
	SearchBar    SearchBarDefaults    `yaml:"search_bar"`
	ColorPicker  ColorPickerDefaults  `yaml:"color_picker"`
	RichTextEdit RichTextEditDefaults `yaml:"rich_text_edit"`
	NumberInput  NumberInputDefaults  `yaml:"number_input"`
	Toggle       ToggleDefaults       `yaml:"toggle"`
	Paginator    PaginatorDefaults    `yaml:"paginator"`
	Tree         TreeDefaults         `yaml:"tree"`
	Split        SplitDefaults        `yaml:"split"`
	Choice       ChoiceDefaults       `yaml:"choice"`
	Table        TableDefaults        `yaml:"table"`
}

// ColorPickerDefaults defines the ColorPicker's palette, marker color, labels, glyphs and keys.
type ColorPickerDefaults struct {
	Palette       []int               `yaml:"palette"`
	InitialTarget string              `yaml:"initial_target"`
	SwatchWidth   int                 `yaml:"swatch_width"`
	MarkerFG      int                 `yaml:"marker_fg"`
	ShowPreview   bool                `yaml:"show_preview"`
	Labels        ColorPickerLabels   `yaml:"labels"`
	Glyphs        ColorPickerGlyphs   `yaml:"glyphs"`
	PreviewText   string              `yaml:"preview_text"`
	Bindings      map[string][]string `yaml:"bindings"`
}

type ColorPickerLabels struct {
	Preview string `yaml:"preview"`
}

type ColorPickerGlyphs struct {
	Swatch     string `yaml:"swatch"`
	ModeMarker string `yaml:"mode_marker"`
}

func (d ColorPickerDefaults) validate() error {
	if len(d.Palette) == 0 || len(d.Palette) > 256 || d.SwatchWidth < 1 || d.MarkerFG < 0 || d.MarkerFG > 255 {
		return fmt.Errorf("color_picker palette, swatch_width and marker_fg must be valid")
	}
	seen := make(map[int]bool, len(d.Palette))
	for _, color := range d.Palette {
		if color < 0 || color > 255 || seen[color] {
			return fmt.Errorf("color_picker palette values must be unique indexes from 0 to 255")
		}
		seen[color] = true
	}
	if d.InitialTarget != "fg" && d.InitialTarget != "bg" {
		return fmt.Errorf("color_picker.initial_target must be fg or bg")
	}
	if d.ShowPreview && (d.Labels.Preview == "" || d.PreviewText == "") {
		return fmt.Errorf("color_picker preview label and text must not be empty when preview is enabled")
	}
	if d.Glyphs.Swatch == "" || d.Glyphs.ModeMarker == "" || measure.StringWidth(d.Glyphs.Swatch) != d.SwatchWidth || measure.StringWidth(d.Glyphs.ModeMarker) != d.SwatchWidth {
		return fmt.Errorf("color_picker swatch and mode marker must match swatch_width display cells")
	}
	for _, name := range []string{"select", "mode_bg", "mode_fg", "mode_toggle", "close", "left", "right"} {
		if len(d.Bindings[name]) == 0 {
			return fmt.Errorf("color_picker.bindings.%s must not be empty", name)
		}
	}
	return nil
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
	SearchRegex              bool   `yaml:"search_regex"`
	SearchCaseInsensitive    bool   `yaml:"search_case_insensitive"`
	ScreenshotFirstNumber    int    `yaml:"screenshot_first_number"`
	ScreenshotSafeCharacters string `yaml:"screenshot_safe_characters"`
	ScreenshotReplacement    string `yaml:"screenshot_replacement"`
	ConfigDirectory          string `yaml:"config_directory"`
	ConfigFile               string `yaml:"config_file"`
	ConfigFallbackDirectory  string `yaml:"config_fallback_directory"`

	Title                   string   `yaml:"title"`
	InitialHeight           int      `yaml:"initial_height"`
	MaxCols                 int      `yaml:"max_cols"`
	Resizeable              bool     `yaml:"resizeable"`
	ShowFileBar             bool     `yaml:"show_file_bar"`
	ShowSidePanel           bool     `yaml:"show_side_panel"`
	BrowserMaxWidth         int      `yaml:"browser_max_width"`
	BrowserWidthDivisor     int      `yaml:"browser_width_divisor"`
	BrowserMinWidth         int      `yaml:"browser_min_width"`
	BrowserNarrowDivisor    int      `yaml:"browser_narrow_divisor"`
	SearchMaxWidth          int      `yaml:"search_max_width"`
	SearchMaxHeight         int      `yaml:"search_max_height"`
	FooterMinHeight         int      `yaml:"footer_min_height"`
	ShowStatusBar           bool     `yaml:"show_status_bar"`
	ShowHints               bool     `yaml:"show_hints"`
	TitlePrefix             string   `yaml:"title_prefix"`
	TitleSuffix             string   `yaml:"title_suffix"`
	DividerGlyph            string   `yaml:"divider_glyph"`
	BrowserDividerGlyph     string   `yaml:"browser_divider_glyph"`
	PositionFormat          string   `yaml:"position_format"`
	Truncation              string   `yaml:"truncation"`
	SavedPrefix             string   `yaml:"saved_prefix"`
	UnsavedTitle            string   `yaml:"unsaved_title"`
	UnsavedMessage          string   `yaml:"unsaved_message"`
	SaveLabel               string   `yaml:"save_label"`
	DiscardLabel            string   `yaml:"discard_label"`
	CancelLabel             string   `yaml:"cancel_label"`
	SearchTitle             string   `yaml:"search_title"`
	InvalidRegex            string   `yaml:"invalid_regex"`
	NoMatches               string   `yaml:"no_matches"`
	MatchFormat             string   `yaml:"match_format"`
	SearchNavigationHint    string   `yaml:"search_navigation_hint"`
	NormalLabel             string   `yaml:"normal_label"`
	RegexLabel              string   `yaml:"regex_label"`
	ModeMarker              string   `yaml:"mode_marker"`
	RegexKeyCap             string   `yaml:"regex_key_cap"`
	ScreenshotDirectory     []string `yaml:"screenshot_directory"`
	ScreenshotMarker        string   `yaml:"screenshot_marker"`
	ScreenshotNumberWidth   int      `yaml:"screenshot_number_width"`
	ScreenshotExtension     string   `yaml:"screenshot_extension"`
	ScreenshotUntitled      string   `yaml:"screenshot_untitled"`
	ScreenshotWidth         int      `yaml:"screenshot_width"`
	ScreenshotHeight        int      `yaml:"screenshot_height"`
	ScreenshotDirectoryMode uint32   `yaml:"screenshot_directory_mode"`
	ScreenshotFileMode      uint32   `yaml:"screenshot_file_mode"`
	ScreenshotSavedPrefix   string   `yaml:"screenshot_saved_prefix"`
	BoxDropPriority         int      `yaml:"box_drop_priority"`
	ScreenshotDropPriority  int      `yaml:"screenshot_drop_priority"`

	StatusIcons                  EditorStatusIcons `yaml:"status_icons"`
	Theme                        string            `yaml:"theme"`
	MouseGrab                    bool              `yaml:"mousegrab"`
	AltScreen                    bool              `yaml:"altscreen"`
	HotkeySaveKey                string            `yaml:"hotkey_save_key"`
	HotkeyFilesKey               string            `yaml:"hotkey_files_key"`
	HotkeyFilesSecondaryBinding  string            `yaml:"hotkey_files_secondary_binding"`
	HotkeySearchSecondaryBinding string            `yaml:"hotkey_search_secondary_binding"`
	HotkeyBoxKey                 string            `yaml:"hotkey_box_key"`
	HotkeyFilesBinding           string            `yaml:"hotkey_files_binding"`
	HotkeyFilesLabel             string            `yaml:"hotkey_files_label"`
	HotkeySearchKey              string            `yaml:"hotkey_search_key"`
	HotkeySearchBinding          string            `yaml:"hotkey_search_binding"`
	HotkeySearchLabel            string            `yaml:"hotkey_search_label"`
	HotkeyBoxBinding             string            `yaml:"hotkey_box_binding"`
	HotkeyBoxLabel               string            `yaml:"hotkey_box_label"`
	HotkeyMouseGrabKey           string            `yaml:"hotkey_mousegrab_key"`
	HotkeyMouseGrabBinding       string            `yaml:"hotkey_mousegrab_binding"`
	HotkeyMouseGrabLabel         string            `yaml:"hotkey_mousegrab_label"`
	MouseGrabDropPriority        int               `yaml:"mousegrab_drop_priority"`
	SearchPrompt                 string            `yaml:"search_prompt"`
	SearchPlaceholder            string            `yaml:"search_placeholder"`
	HotkeyScreenshotBinding      string            `yaml:"hotkey_screenshot_binding"`
	HotkeyScreenshotLabel        string            `yaml:"hotkey_screenshot_label"`
}

// EditorStatusIcons defines compact editor status caps and their colors.
type EditorStatusIcons struct {
	Theme    string `yaml:"theme"`
	MouseOn  string `yaml:"mouse_on"`
	MouseOff string `yaml:"mouse_off"`
	AltOn    string `yaml:"alt_on"`
	AltOff   string `yaml:"alt_off"`
	FG       int    `yaml:"fg"`
	BG       int    `yaml:"bg"`
}

// RichTextEditDefaults defines selection and popover presentation defaults.
type RichTextEditDefaults struct {
	Bindings          map[string][]string `yaml:"bindings"`
	UndoLimit         int                 `yaml:"undo_limit"`
	MultiClickWindow  time.Duration       `yaml:"multi_click_window"`
	ShowCursor        bool                `yaml:"show_cursor"`
	ShowPopover       bool                `yaml:"show_popover"`
	HelpTitle         string              `yaml:"help_title"`
	HelpMaxWidth      int                 `yaml:"help_max_width"`
	HelpMaxHeight     int                 `yaml:"help_max_height"`
	HelpPageStep      int                 `yaml:"help_page_step"`
	HelpFooter        string              `yaml:"help_footer"`
	SaveTitle         string              `yaml:"save_title"`
	FileMode          uint32              `yaml:"file_mode"`
	FinalNewline      bool                `yaml:"final_newline"`
	CreateMissingFile bool                `yaml:"create_missing_file"`
	ViewLabel         string              `yaml:"view_label"`
	EditLabel         string              `yaml:"edit_label"`
	UntitledName      string              `yaml:"untitled_name"`
	BoxHint           string              `yaml:"box_hint"`
	StatusSeparator   string              `yaml:"status_separator"`
	StatusTruncation  string              `yaml:"status_truncation"`
	PopoverMinWidth   int                 `yaml:"popover_min_width"`

	ActionClosesMenu          bool   `yaml:"action_closes_menu"`
	MenuFocusKey              string `yaml:"menu_focus_key"`
	ArrowsCancelUnfocusedMenu bool   `yaml:"arrows_cancel_unfocused_menu"`
	UnselectAfterAction       bool   `yaml:"unselect_after_action"`
	UnselectOnCancel          bool   `yaml:"unselect_on_cancel"`
	PopupOpenKey              string `yaml:"popup_open_key"`
	AutoPopupOnSelection      bool   `yaml:"auto_popup_on_selection"`
	PopupOpenDelayMS          int    `yaml:"popup_open_delay_ms"`
	CycleNavigation           bool   `yaml:"cycle_navigation"`
	PositionPreference        string `yaml:"position_preference"`
	DismissOnScroll           bool   `yaml:"dismiss_on_scroll"`
	MouseHoverFocus           bool   `yaml:"mouse_hover_focus"`

	SelectionBG             int           `yaml:"selection_bg"`
	ToolbarFG               int           `yaml:"toolbar_fg"`
	ToolbarBG               int           `yaml:"toolbar_bg"`
	SeparatorGlyph          string        `yaml:"separator_glyph"`
	SeparatorFG             int           `yaml:"separator_fg"`
	SeparatorBG             int           `yaml:"separator_bg"`
	PointerUpGlyph          string        `yaml:"pointer_up_glyph"`
	PointerDownGlyph        string        `yaml:"pointer_down_glyph"`
	PointerFG               int           `yaml:"pointer_fg"`
	BoxDrawLabel            string        `yaml:"box_draw_label"`
	PopoverLabels           []string      `yaml:"popover_labels"`
	PopoverHamburgerLabel   string        `yaml:"popover_hamburger_label"`
	PopoverHamburgerItems   []string      `yaml:"popover_hamburger_items"`
	PopoverHamburgerActions []string      `yaml:"popover_hamburger_actions"`
	BoxStyleLabels          []string      `yaml:"box_style_labels"`
	BoxStyleDefault         string        `yaml:"box_style_default"`
	GhostCursorEnabled      bool          `yaml:"ghost_cursor_enabled"`
	PopoverFocusFG          int           `yaml:"popover_focus_fg"`
	PopoverFocusBG          int           `yaml:"popover_focus_bg"`
	LinkFG                  int           `yaml:"link_fg"`
	LinkUnderline           bool          `yaml:"link_underline"`
	SavePopupMaxWidth       int           `yaml:"save_popup_max_width"`
	SavePopupMaxHeight      int           `yaml:"save_popup_max_height"`
	StateUntitledLabel      string        `yaml:"state_untitled_label"`
	StateSavedLabel         string        `yaml:"state_saved_label"`
	StateModifiedLabel      string        `yaml:"state_modified_label"`
	StateErrorLabel         string        `yaml:"state_error_label"`
	FileMenuTitle           string        `yaml:"file_menu_title"`
	FileMenuMnemonic        string        `yaml:"file_menu_mnemonic"`
	HotkeyHelpBinding       string        `yaml:"hotkey_help_binding"`
	HotkeyHelpLabel         string        `yaml:"hotkey_help_label"`
	HotkeySaveBinding       string        `yaml:"hotkey_save_binding"`
	HotkeySaveLabel         string        `yaml:"hotkey_save_label"`
	HotkeySaveAsLabel       string        `yaml:"hotkey_save_as_label"`
	HotkeyOpenBinding       string        `yaml:"hotkey_open_binding"`
	HotkeyOpenLabel         string        `yaml:"hotkey_open_label"`
	HotkeyCloseBinding      string        `yaml:"hotkey_close_binding"`
	HotkeyCloseLabel        string        `yaml:"hotkey_close_label"`
	HotkeyViewEditBinding   string        `yaml:"hotkey_view_edit_binding"`
	HotkeyQuitBinding       string        `yaml:"hotkey_quit_binding"`
	HotkeyQuitLabel         string        `yaml:"hotkey_quit_label"`
	HelpSections            []HelpSection `yaml:"help_sections"`
}

// HelpEntry is one row of a help section. Keys are bindings rendered with
// KeyCap; Cap overrides the rendered keycap text; Ref names a
// RichTextEditDefaults hotkey (e.g. "HotkeySave") whose binding and label fill
// the row.
type HelpEntry struct {
	Keys  []string `yaml:"keys"`
	Cap   string   `yaml:"cap"`
	Ref   string   `yaml:"ref"`
	Label string   `yaml:"label"`
}

// HelpSection is a titled group of help entries.
type HelpSection struct {
	Title   string      `yaml:"title"`
	Entries []HelpEntry `yaml:"entries"`
}

// helpRef resolves a HelpEntry.Ref such as "HotkeySave" to its binding and label.
func (d RichTextEditDefaults) helpRef(ref string) (binding, label string, ok bool) {
	v := reflect.ValueOf(d)
	b, l := v.FieldByName(ref+"Binding"), v.FieldByName(ref+"Label")
	if !b.IsValid() || !l.IsValid() || b.Kind() != reflect.String || l.Kind() != reflect.String {
		return "", "", false
	}
	return b.String(), l.String(), true
}

func (d RichTextEditDefaults) validate() error {
	if d.UndoLimit < 1 || d.MultiClickWindow <= 0 || d.HelpPageStep < 1 || d.HelpMaxWidth < 1 || d.HelpMaxHeight < 1 || d.PopoverMinWidth < 1 {
		return fmt.Errorf("rich_text_edit limits and timings must be positive")
	}
	for name, keys := range d.Bindings {
		if len(keys) == 0 {
			return fmt.Errorf("rich_text_edit.bindings.%s must not be empty", name)
		}
	}

	if d.PopupOpenKey == "" {
		return fmt.Errorf("rich_text_edit popup_open_key must not be empty")
	}
	if d.PopupOpenDelayMS < 0 {
		return fmt.Errorf("rich_text_edit.popup_open_delay_ms must be nonnegative")
	}
	switch d.PositionPreference {
	case "above", "below", "smart":
	default:
		return fmt.Errorf("rich_text_edit.position_preference must be above, below or smart")
	}
	if len(d.HelpSections) == 0 {
		return fmt.Errorf("rich_text_edit.help_sections must not be empty")
	}
	for _, section := range d.HelpSections {
		if section.Title == "" || len(section.Entries) == 0 {
			return fmt.Errorf("rich_text_edit.help_sections need a title and entries")
		}
		for _, entry := range section.Entries {
			if _, _, ok := d.helpRef(entry.Ref); entry.Ref != "" && !ok {
				return fmt.Errorf("rich_text_edit.help_sections: unknown ref %q", entry.Ref)
			}
			if entry.Ref == "" && (entry.Label == "" || (len(entry.Keys) == 0 && entry.Cap == "")) {
				return fmt.Errorf("rich_text_edit.help_sections %q: entry needs label and keys or cap", section.Title)
			}
		}
	}
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
	if len(d.PopoverLabels) != 7 {
		return fmt.Errorf("rich_text_edit.popover_labels must contain 7 labels")
	}
	if d.PopoverHamburgerLabel == "" || len(d.PopoverHamburgerItems) == 0 || len(d.PopoverHamburgerItems) != len(d.PopoverHamburgerActions) {
		return fmt.Errorf("rich_text_edit hamburger label and items must not be empty")
	}
	for i, action := range d.PopoverHamburgerActions {
		if action != "select_all" || d.PopoverHamburgerItems[i] == "" {
			return fmt.Errorf("rich_text_edit hamburger item %d has invalid label or action", i)
		}
	}
	if d.BoxDrawLabel == "" {
		return fmt.Errorf("rich_text_edit.box_draw_label must not be empty")
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

// NumberInputDefaults defines specced defaults for NumberInput.
type NumberInputDefaults struct {
	Step         float64 `yaml:"step"`
	LeftGlyph    string  `yaml:"left_glyph"`
	RightGlyph   string  `yaml:"right_glyph"`
	InvalidError string  `yaml:"invalid_error"`
	BelowMinFmt  string  `yaml:"below_min_fmt"`
	AboveMaxFmt  string  `yaml:"above_max_fmt"`
}

func (d NumberInputDefaults) validate() error {
	if d.Step <= 0 {
		return fmt.Errorf("number_input.step must be positive")
	}
	if d.LeftGlyph == "" || d.RightGlyph == "" {
		return fmt.Errorf("number_input glyphs must not be empty")
	}
	if d.InvalidError == "" || d.BelowMinFmt == "" || d.AboveMaxFmt == "" {
		return fmt.Errorf("number_input error formats must not be empty")
	}
	return nil
}

// ToggleDefaults defines specced defaults for Toggle.
type ToggleDefaults struct {
	OnMark  string `yaml:"on_mark"`
	OffMark string `yaml:"off_mark"`
}

func (d ToggleDefaults) validate() error {
	if d.OnMark == "" || d.OffMark == "" {
		return fmt.Errorf("toggle marks must not be empty")
	}
	return nil
}

// PaginatorDefaults defines specced defaults for Paginator.
type PaginatorDefaults struct {
	ActiveDot     string `yaml:"active_dot"`
	InactiveDot   string `yaml:"inactive_dot"`
	NumericFormat string `yaml:"numeric_format"`
}

func (d PaginatorDefaults) validate() error {
	if d.ActiveDot == "" || d.InactiveDot == "" || d.NumericFormat == "" {
		return fmt.Errorf("paginator defaults must not be empty")
	}
	return nil
}

// TreeDefaults defines specced defaults for Tree.
type TreeDefaults struct {
	CollapsedMarker string `yaml:"collapsed_marker"`
	ExpandedMarker  string `yaml:"expanded_marker"`
	LeafMarker      string `yaml:"leaf_marker"`
}

func (d TreeDefaults) validate() error {
	if d.CollapsedMarker == "" || d.ExpandedMarker == "" {
		return fmt.Errorf("tree markers must not be empty")
	}
	return nil
}

// SplitDefaults defines specced defaults for Split.
type SplitDefaults struct {
	VerticalDivider   string  `yaml:"vertical_divider"`
	HorizontalDivider string  `yaml:"horizontal_divider"`
	DefaultRatio      float64 `yaml:"default_ratio"`
	DefaultGap        int     `yaml:"default_gap"`
}

func (d SplitDefaults) validate() error {
	if d.VerticalDivider == "" || d.HorizontalDivider == "" {
		return fmt.Errorf("split dividers must not be empty")
	}
	if d.DefaultRatio < 0 || d.DefaultRatio > 1 {
		return fmt.Errorf("split.default_ratio must be between 0 and 1")
	}
	if d.DefaultGap < 0 {
		return fmt.Errorf("split.default_gap must not be negative")
	}
	return nil
}

// ChoiceDefaults defines specced defaults for Choice.
type ChoiceDefaults struct {
	DefaultPlaceholder string `yaml:"default_placeholder"`
	SelectionMarker    string `yaml:"selection_marker"`
	UnselectedMarker   string `yaml:"unselected_marker"`
	CheckedMarker      string `yaml:"checked_marker"`
	UncheckedMarker    string `yaml:"unchecked_marker"`
}

// TableDefaults defines sort chrome for Table.
type TableDefaults struct {
	SortAscGlyph    string `yaml:"sort_asc_glyph"`
	SortDescGlyph   string `yaml:"sort_desc_glyph"`
	ControlsHintFmt string `yaml:"controls_hint_fmt"`
}

func (d ChoiceDefaults) validate() error {
	if d.SelectionMarker == "" || d.CheckedMarker == "" || d.UncheckedMarker == "" {
		return fmt.Errorf("choice markers must not be empty")
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
	KittyKeyboard bool          `yaml:"kitty_keyboard"`
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
	if err := defs.ColorPicker.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.RichTextEdit.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.NumberInput.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.Toggle.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.Paginator.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.Tree.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.Split.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	if err := defs.Choice.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	return defs
}()
