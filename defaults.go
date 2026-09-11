// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed spec/defaults.yaml
var defaultsYAML []byte

// LibDefaults represents specced runtime defaults loaded from spec/defaults.yaml.
type LibDefaults struct {
	FallbackQuitKeys []string       `yaml:"fallback_quit_keys"`
	Pane             PaneDefaults   `yaml:"pane"`
	Splash           SplashDefaults `yaml:"splash"`
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
	return defs
}()
