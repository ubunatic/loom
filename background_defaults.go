// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed spec/backgrounds.yaml
var backgroundSpec []byte

type backgroundDefaults struct {
	Glyphs         []string      `yaml:"glyphs"`
	DimFactor      float64       `yaml:"dim_factor"`
	Threshold      float64       `yaml:"threshold"`
	RedrawInterval time.Duration `yaml:"redraw_interval"`
	Density        float64       `yaml:"density"`
}

type backgroundSpecFile struct {
	Background backgroundDefaults `yaml:"background"`
}

func loadBackgroundDefaults() backgroundDefaults {
	var f backgroundSpecFile
	if err := yaml.Unmarshal(backgroundSpec, &f); err != nil {
		panic(fmt.Sprintf("loom: parse spec/backgrounds.yaml: %v", err))
	}
	return f.Background
}

var SpeccedBackground = loadBackgroundDefaults()
