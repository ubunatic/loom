// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"strings"
	"testing"

	"ubunatic.com/loom/measure"
	"ubunatic.com/loom/spec"
)

func TestRenderDefaultsComeFromSpec(t *testing.T) {
	want, err := spec.LoadGraphDefaults()
	if err != nil {
		t.Fatal(err)
	}
	if graphDefaults != want {
		t.Fatalf("loaded graph defaults = %+v, spec has %+v", graphDefaults, want)
	}
	if MaxWidth != want.Bar.MaxWidth {
		t.Fatalf("MaxWidth = %d, spec has %d", MaxWidth, want.Bar.MaxWidth)
	}
	if DefaultBackgroundANSI != want.DefaultBackgroundANSI {
		t.Fatalf("default background ANSI = %q, spec has %q", DefaultBackgroundANSI, want.DefaultBackgroundANSI)
	}
	if string(percentSparkChars) != want.Sparkline.Glyphs || string(defaultStackedBarGlyphs) != want.StackedBar.Glyphs || string(defaultTreemapGlyphs) != want.Treemap.Glyphs {
		t.Fatal("graph glyph defaults do not match spec")
	}
	if string(DefaultSpinnerFrames) != want.Spinner.Frames || string(HarnezSpinnerFrames) != want.Spinner.HarnezFrames {
		t.Fatal("spinner frames do not match spec")
	}
	spinnerFrames := []rune(want.Spinner.Frames)
	if len(spinnerFrames) == 0 || SpinnerString(0) != string(spinnerFrames[0]) {
		t.Errorf("first spinner frame = %q, want first spec frame", SpinnerString(0))
	}

	bar := RenderBar(100, BarOptions{NoWrapper: true})
	if got := measure.StringWidth(bar); got != want.Bar.MaxWidth {
		t.Errorf("default bar width = %d, want spec width %d", got, want.Bar.MaxWidth)
	}
	ansiBar := RenderBar(100, BarOptions{Width: 1, NoWrapper: true, ANSI: true})
	if !strings.HasPrefix(ansiBar, "\x1b["+want.DefaultBackgroundANSI+"m") {
		t.Errorf("default ANSI background = %q, want spec code %q", ansiBar, want.DefaultBackgroundANSI)
	}
	sparkline := []rune(RenderPercentSparkline([]float64{0, 100}, SparklineOptions{Width: 1}))
	sparkGlyphs := []rune(want.Sparkline.Glyphs)
	if len(sparkGlyphs) == 0 || len(sparkline) != 1 || sparkline[0] != sparkGlyphs[len(sparkGlyphs)-1] {
		t.Errorf("default percent sparkline = %q, want highest spec glyph", string(sparkline))
	}

	stacked := []rune(RenderStackedBar([]float64{1, 1}, StackedBarOptions{Width: 2, NoWrapper: true}))
	stackedGlyphs := []rune(want.StackedBar.Glyphs)
	if len(stackedGlyphs) < 2 || len(stacked) != 2 || stacked[0] != stackedGlyphs[0] || stacked[1] != stackedGlyphs[1] {
		t.Errorf("default stacked bar = %q, want first spec glyphs %q", string(stacked), string(stackedGlyphs[:min(2, len(stackedGlyphs))]))
	}

	treemap := RenderTreemap([]TreemapSegment{{Name: "a", Value: 1}}, TreemapOptions{
		Width: 1, Height: 1, NoBorder: true, NoLabels: true,
	})
	treemapGlyphs := []rune(want.Treemap.Glyphs)
	if len(treemapGlyphs) == 0 || len(treemap) != 1 || treemap[0] != string(treemapGlyphs[0]) {
		t.Errorf("default treemap = %q, want first spec glyph %q", treemap, string(treemapGlyphs[:min(1, len(treemapGlyphs))]))
	}
}
