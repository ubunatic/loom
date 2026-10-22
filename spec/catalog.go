// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package spec embeds Loom's application specifications.
package spec

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed widgets.yaml defaults.yaml emoji.yaml
var widgetFiles embed.FS

// WidgetsYAML returns the authoritative widget catalog document.
func WidgetsYAML() ([]byte, error) { return widgetFiles.ReadFile("widgets.yaml") }

// DefaultsYAML returns the authoritative runtime defaults document.
func DefaultsYAML() ([]byte, error) { return widgetFiles.ReadFile("defaults.yaml") }

// EmojiYAML returns the authoritative emoji width policy document.
func EmojiYAML() ([]byte, error) { return widgetFiles.ReadFile("emoji.yaml") }

// GraphDefaults describes the graph renderer defaults in defaults.yaml.
type GraphDefaults struct {
	DefaultBackgroundANSI string               `yaml:"default_background_ansi"`
	Bar                   GraphBarDefaults     `yaml:"bar"`
	Sparkline             GraphGlyphDefaults   `yaml:"sparkline"`
	StackedBar            GraphGlyphDefaults   `yaml:"stacked_bar"`
	Treemap               GraphGlyphDefaults   `yaml:"treemap"`
	Spinner               GraphSpinnerDefaults `yaml:"spinner"`
}

// GraphBarDefaults contains defaults specific to bar renderers.
type GraphBarDefaults struct {
	MaxWidth int `yaml:"max_width"`
}

// GraphGlyphDefaults stores an ordered set of glyphs as a Unicode string.
type GraphGlyphDefaults struct {
	Glyphs string `yaml:"glyphs"`
}

// GraphSpinnerDefaults contains the standard and Harnez spinner sequences.
type GraphSpinnerDefaults struct {
	Frames       string `yaml:"frames"`
	HarnezFrames string `yaml:"harnez_frames"`
}

// LoadGraphDefaults reads graph defaults from the embedded authoritative document.
func LoadGraphDefaults() (GraphDefaults, error) {
	data, err := DefaultsYAML()
	if err != nil {
		return GraphDefaults{}, fmt.Errorf("spec: read defaults.yaml: %w", err)
	}
	var defaults struct {
		Graph GraphDefaults `yaml:"graph"`
	}
	if err := yaml.Unmarshal(data, &defaults); err != nil {
		return GraphDefaults{}, fmt.Errorf("spec: parse graph defaults: %w", err)
	}
	return defaults.Graph, nil
}
