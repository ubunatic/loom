// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "ubunatic.com/loom/specdefaults"

// LibDefaults represents specced runtime defaults loaded from spec/defaults.yaml.
type LibDefaults = specdefaults.LibDefaults

// ColorPickerDefaults defines the ColorPicker's palette, marker color, labels, glyphs and keys.
type ColorPickerDefaults = specdefaults.ColorPickerDefaults

// ColorPickerLabels is the runtime configuration defined in specdefaults.
type ColorPickerLabels = specdefaults.ColorPickerLabels

// ColorPickerGlyphs is the runtime configuration defined in specdefaults.
type ColorPickerGlyphs = specdefaults.ColorPickerGlyphs

// KeyCapDefaults defines the modifier glyphs KeyCap renders, in display order.
type KeyCapDefaults = specdefaults.KeyCapDefaults

// KeyCapModifier maps a binding modifier name (shift, ctrl, alt) to its glyph.
type KeyCapModifier = specdefaults.KeyCapModifier

// EditorDefaults defines specced defaults for the Loom editor.
type EditorDefaults = specdefaults.EditorDefaults

// EditorStatusIcons defines compact editor status caps and their colors.
type EditorStatusIcons = specdefaults.EditorStatusIcons

// RichTextEditDefaults defines selection and popover presentation defaults.
type RichTextEditDefaults = specdefaults.RichTextEditDefaults

// HelpEntry is one row of a help section. Keys are bindings rendered with
// KeyCap; Cap overrides the rendered keycap text; Ref names a
// RichTextEditDefaults hotkey (e.g. "HotkeySave") whose binding and label fill
// the row.
type HelpEntry = specdefaults.HelpEntry

// HelpSection is a titled group of help entries.
type HelpSection = specdefaults.HelpSection

// NumberInputDefaults defines specced defaults for NumberInput.
type NumberInputDefaults = specdefaults.NumberInputDefaults

// ToggleDefaults defines specced defaults for Toggle.
type ToggleDefaults = specdefaults.ToggleDefaults

// PaginatorDefaults defines specced defaults for Paginator.
type PaginatorDefaults = specdefaults.PaginatorDefaults

// TreeDefaults defines specced defaults for Tree.
type TreeDefaults = specdefaults.TreeDefaults

// SplitDefaults defines specced defaults for Split.
type SplitDefaults = specdefaults.SplitDefaults

// ChoiceDefaults defines specced defaults for Choice.
type ChoiceDefaults = specdefaults.ChoiceDefaults

// TableDefaults defines sort chrome for Table.
type TableDefaults = specdefaults.TableDefaults

// TabsDefaults defines specced navigation defaults for Tabs.
type TabsDefaults = specdefaults.TabsDefaults

// ViewDefaults defines specced navigation defaults for View.
type ViewDefaults = specdefaults.ViewDefaults

// MenuDefaults defines specced navigation and accelerator defaults for Menu and MenuBar.
type MenuDefaults = specdefaults.MenuDefaults

// FormDefaults defines specced navigation defaults for Form.
type FormDefaults = specdefaults.FormDefaults

// SettingsDefaults defines specced navigation and editing defaults for Settings.
type SettingsDefaults = specdefaults.SettingsDefaults

// SessionDefaults defines specced defaults for the Loom session command.
type SessionDefaults = specdefaults.SessionDefaults

// SessionBarLabels defines the text and glyphs shown in the Loom session bar.
type SessionBarLabels = specdefaults.SessionBarLabels

// SessionBindings is the runtime configuration defined in specdefaults.
type SessionBindings = specdefaults.SessionBindings

// SessionPrefixBindings is the runtime configuration defined in specdefaults.
type SessionPrefixBindings = specdefaults.SessionPrefixBindings

// SessionSearchDefaults is the runtime configuration defined in specdefaults.
type SessionSearchDefaults = specdefaults.SessionSearchDefaults

// SessionSearchBindings is the runtime configuration defined in specdefaults.
type SessionSearchBindings = specdefaults.SessionSearchBindings

// KeyConvention documents a default key convention across library widgets.
type KeyConvention = specdefaults.KeyConvention

// SearchBarDefaults defines specced defaults for the SearchBar widget.
type SearchBarDefaults = specdefaults.SearchBarDefaults

// PaintCanvasDefaults defines stroke tuning controls.
type PaintCanvasDefaults = specdefaults.PaintCanvasDefaults

// ClockDefaults defines the labels for timer and stopwatch controls.
type ClockDefaults = specdefaults.ClockDefaults

// MediaDefaults defines user-facing media status and rendering timing defaults.
type MediaDefaults = specdefaults.MediaDefaults

// MouseDefaults defines specced defaults for mouse gestures.
type MouseDefaults = specdefaults.MouseDefaults

// ScrollbarDefaults defines scrollbar visibility and its one-cell glyphs.
type ScrollbarDefaults = specdefaults.ScrollbarDefaults

// PaneDefaults defines specced defaults for Pane.
type PaneDefaults = specdefaults.PaneDefaults

// SplashDefaults defines specced defaults for splash lifecycle and widgets.
type SplashDefaults = specdefaults.SplashDefaults

// ProgressBarDefaults defines specced defaults for the ProgressBar widget.
type ProgressBarDefaults = specdefaults.ProgressBarDefaults

// SpinnerDefaults defines specced animation timing for Spinner.
type SpinnerDefaults = specdefaults.SpinnerDefaults

// SpeccedDefaults holds the loaded immutable defaults from spec/defaults.yaml.
var SpeccedDefaults = specdefaults.SpeccedDefaults
