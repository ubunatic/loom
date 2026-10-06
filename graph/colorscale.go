// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"fmt"
	"math"
)

// CellStyle describes ANSI styling for a treemap cell.
type CellStyle struct {
	ForegroundANSI string
	BackgroundANSI string
}

// ColorScale maps a value in a range to a cell style.
type ColorScale func(value, min, max float64) CellStyle

// LinearColorScale interpolates RGB foreground colors between from and to.
// Inputs are ANSI 24-bit color strings in the form "R,G,B".
func LinearColorScale(from, to string) ColorScale {
	fr, fg, fb := parseRGB(from)
	tr, tg, tb := parseRGB(to)
	return func(value, min, max float64) CellStyle {
		t := colorScalePosition(value, min, max)
		return CellStyle{ForegroundANSI: fmt.Sprintf("38;2;%d;%d;%d", lerpByte(fr, tr, t), lerpByte(fg, tg, t), lerpByte(fb, tb, t))}
	}
}

// HeatColorScale returns a blue-to-red 24-bit foreground heat scale.
func HeatColorScale() ColorScale { return LinearColorScale("0,0,255", "255,0,0") }

func colorScalePosition(value, min, max float64) float64 {
	if math.IsNaN(value) || math.IsNaN(min) || math.IsNaN(max) || math.IsInf(value, 0) || math.IsInf(min, 0) || math.IsInf(max, 0) || min == max {
		return 0.5
	}
	t := (value - min) / (max - min)
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

func parseRGB(value string) (int, int, int) {
	var r, g, b int
	_, _ = fmt.Sscanf(value, "%d,%d,%d", &r, &g, &b)
	return r, g, b
}

func lerpByte(a, b int, t float64) int {
	return int(math.Round(float64(a) + (float64(b)-float64(a))*t))
}
