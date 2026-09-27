// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"
	"math"

	"gopkg.in/yaml.v3"
)

//go:embed spec/cursor.yaml
var cursorSpecYAML []byte

// CursorProximitySpec configures the first cursor effect.
type CursorProximitySpec struct {
	Enabled    bool    `yaml:"enabled"`
	Radius     int     `yaml:"radius"`
	Brightness float64 `yaml:"brightness"`
}

type cursorSpecFile struct {
	Proximity CursorProximitySpec `yaml:"proximity"`
}

// SpeccedCursorProximity is the immutable proximity effect configuration.
var SpeccedCursorProximity = func() CursorProximitySpec {
	var spec cursorSpecFile
	if err := yaml.Unmarshal(cursorSpecYAML, &spec); err != nil {
		panic(fmt.Sprintf("loom: parse spec/cursor.yaml: %v", err))
	}
	return spec.Proximity
}()

// CursorHint describes a cell's offset from the current mouse position.
// DX and DY are cell coordinates minus cursor coordinates.
type CursorHint struct {
	DX, DY   int
	Distance float64
}

func (c *Canvas) SetCursorPosition(x, y int) {
	if c == nil || x < 0 || x >= c.cols || y < 0 || y >= c.rows {
		if c != nil {
			c.ClearCursorPosition()
		}
		return
	}
	c.mouseX, c.mouseY, c.mouseKnown = x, y, true
}

// ClearCursorPosition removes the cursor hint, such as before the first motion
// report or after a report outside the pane.
func (c *Canvas) ClearCursorPosition() {
	if c == nil {
		return
	}
	c.mouseX, c.mouseY, c.mouseKnown = 0, 0, false
}

// CursorHintAt returns a relative hint for positions within the spec radius.
func (c *Canvas) CursorHintAt(x, y int) (CursorHint, bool) {
	if c == nil || !c.mouseKnown || x < 0 || x >= c.cols || y < 0 || y >= c.rows {
		return CursorHint{}, false
	}
	dx, dy := x-c.mouseX, y-c.mouseY
	distance := math.Hypot(float64(dx), float64(dy))
	if distance > float64(SpeccedCursorProximity.Radius) {
		return CursorHint{}, false
	}
	return CursorHint{DX: dx, DY: dy, Distance: distance}, true
}

// ApplyCursorBrighten brightens unclaimed cells around the known cursor.
func (c *Canvas) ApplyCursorBrighten() {
	if c == nil || !SpeccedCursorProximity.Enabled || !c.mouseKnown {
		return
	}
	for y := 0; y < c.rows; y++ {
		for x := 0; x < c.cols; x++ {
			if c.claimed[y][x] {
				continue
			}
			hint, ok := c.CursorHintAt(x, y)
			if !ok {
				continue
			}
			bg := c.cells[y][x].Style.BG
			r, g, b, ok := bg.RGB()
			if !ok {
				continue
			}
			factor := SpeccedCursorProximity.Brightness * (1 - hint.Distance/float64(SpeccedCursorProximity.Radius+1))
			c.cells[y][x].Style.BG = ColorRGB(brighten(r, factor), brighten(g, factor), brighten(b, factor))
		}
	}
}

func brighten(value uint8, factor float64) uint8 {
	result := float64(value) + (255-float64(value))*factor
	return uint8(math.Round(result))
}
