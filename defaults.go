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
	FallbackQuitKeys []string          `yaml:"fallback_quit_keys"`
	Pane             PaneDefaults      `yaml:"pane"`
	Scrollbar        ScrollbarDefaults `yaml:"scrollbar"`
	Splash           SplashDefaults    `yaml:"splash"`
}

// ScrollbarDefaults defines the one-cell foreground and background glyphs.
type ScrollbarDefaults struct {
	ForegroundChar string `yaml:"foreground_char"`
	BackgroundChar string `yaml:"background_char"`
}

func (d ScrollbarDefaults) validate() error {
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
	MaxCols int `yaml:"max_cols"`
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

// SpeccedDefaults holds the loaded immutable defaults from spec/defaults.yaml.
var SpeccedDefaults = func() LibDefaults {
	var defs LibDefaults
	if err := yaml.Unmarshal(defaultsYAML, &defs); err != nil {
		panic(fmt.Sprintf("loom: parse spec/defaults.yaml: %v", err))
	}
	if err := defs.Scrollbar.validate(); err != nil {
		panic(fmt.Sprintf("loom: spec/defaults.yaml: %v", err))
	}
	return defs
}()
