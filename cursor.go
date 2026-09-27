// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"
	"math"
	"time"

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
	StarTrail CursorStarTrailSpec `yaml:"star_trail"`
}

// SpeccedCursorProximity is the immutable proximity effect configuration.
var SpeccedCursorProximity = func() CursorProximitySpec {
	var spec cursorSpecFile
	if err := yaml.Unmarshal(cursorSpecYAML, &spec); err != nil {
		panic(fmt.Sprintf("loom: parse spec/cursor.yaml: %v", err))
	}
	return spec.Proximity
}()

// CursorStarTrailSpec configures the fading trail theme.
type CursorStarTrailSpec struct {
	Enabled    bool          `yaml:"enabled"`
	Glyph      string        `yaml:"glyph"`
	Color      ThemeColor    `yaml:"color"`
	MaxPoints  int           `yaml:"max_points"`
	LifetimeMS int           `yaml:"lifetime_ms"`
	FrameMS    int           `yaml:"frame_ms"`
	Lifetime   time.Duration `yaml:"-"`
	Frame      time.Duration `yaml:"-"`
}

// SpeccedCursorStarTrail is the immutable star trail effect configuration.
var SpeccedCursorStarTrail = func() CursorStarTrailSpec {
	var spec cursorSpecFile
	if err := yaml.Unmarshal(cursorSpecYAML, &spec); err != nil {
		panic(fmt.Sprintf("loom: parse spec/cursor.yaml: %v", err))
	}
	trail := spec.StarTrail
	trail.Lifetime = time.Duration(trail.LifetimeMS) * time.Millisecond
	trail.Frame = time.Duration(trail.FrameMS) * time.Millisecond
	return trail
}()

// CursorHint describes a cell's offset from the current mouse position.
// DX and DY are cell coordinates minus cursor coordinates.
type CursorHint struct {
	DX, DY   int
	Distance float64
}

// CursorTrailPoint records a cursor position and the time it was left.
type CursorTrailPoint struct {
	X, Y int
	At   time.Time
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

// ApplyCursorBrighten brightens backgrounds, including those behind text.
// Terminal-default backgrounds have no resolvable RGB value and remain as-is.
func (c *Canvas) ApplyCursorBrighten() {
	if c == nil || !SpeccedCursorProximity.Enabled || !c.mouseKnown {
		return
	}
	for y := 0; y < c.rows; y++ {
		for x := 0; x < c.cols; x++ {
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

// ApplyCursorStarTrail draws unexpired cursor trail points, newest last.
func (c *Canvas) ApplyCursorStarTrail(points []CursorTrailPoint, now time.Time) {
	if c == nil || !SpeccedCursorStarTrail.Enabled {
		return
	}
	if len(points) > SpeccedCursorStarTrail.MaxPoints {
		points = points[len(points)-SpeccedCursorStarTrail.MaxPoints:]
	}
	r, g, b, _ := SpeccedCursorStarTrail.Color.Color().RGB()
	for _, point := range points {
		if point.X < 0 || point.X >= c.cols || point.Y < 0 || point.Y >= c.rows {
			continue
		}
		age := now.Sub(point.At)
		if age < 0 {
			age = 0
		}
		if age >= SpeccedCursorStarTrail.Lifetime {
			continue
		}
		intensity := 1 - float64(age)/float64(SpeccedCursorStarTrail.Lifetime)
		cell := Cell{
			Text: SpeccedCursorStarTrail.Glyph,
			Style: Style{FG: ColorRGB(
				uint8(math.Round(float64(r)*intensity)),
				uint8(math.Round(float64(g)*intensity)),
				uint8(math.Round(float64(b)*intensity)),
			), Dim: intensity < .55},
		}
		c.PaintDecoration(point.X, point.Y, cell)
	}
}

func brighten(value uint8, factor float64) uint8 {
	result := float64(value) + (255-float64(value))*factor
	return uint8(math.Round(result))
}
