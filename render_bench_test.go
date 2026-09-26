// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func BenchmarkStyleANSI(b *testing.B) {
	styles := []loom.Style{
		loom.Reset,
		{Bold: true, FG: loom.ColorIndex(196)},
		{Dim: true, Underline: true, BG: loom.ColorRGB(10, 20, 30)},
		{Bold: true, Underline: true, FG: loom.ColorRGB(255, 128, 0), BG: loom.ColorIndex(234)},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := styles[i%len(styles)]
		_ = s.ANSI()
	}
}

func BenchmarkCanvasRow(b *testing.B) {
	c := loom.NewCanvas(80, 1)
	s1 := loom.Style{Bold: true, FG: loom.ColorIndex(196)}
	s2 := loom.Style{BG: loom.ColorRGB(10, 20, 30)}
	for x := 0; x < 80; x++ {
		if x%2 == 0 {
			c.Set(x, 0, loom.Cell{Text: "A", Style: s1})
		} else {
			c.Set(x, 0, loom.Cell{Text: "B", Style: s2})
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = c.Row(0)
	}
}

func BenchmarkCanvasFlush(b *testing.B) {
	c := loom.NewCanvas(80, 24)
	s1 := loom.Style{Bold: true, FG: loom.ColorIndex(196)}
	s2 := loom.Style{BG: loom.ColorRGB(10, 20, 30)}
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			if (x+y)%2 == 0 {
				c.Set(x, y, loom.Cell{Text: "X", Style: s1})
			} else {
				c.Set(x, y, loom.Cell{Text: "Y", Style: s2})
			}
		}
	}
	cfg := loom.DefaultResizeConfig()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var out strings.Builder
		c.FlushWithConfig(&out, 1, 0, cfg)
	}
}
