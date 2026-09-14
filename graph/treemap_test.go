// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom/measure"
)

func TestLayoutTreemapTilesExactlyNoOverlap(t *testing.T) {
	cases := []struct {
		values []float64
		w, h   int
	}{
		{[]float64{1, 2, 3}, 10, 6},
		{[]float64{1, 1, 1, 1, 1, 1, 1}, 9, 5},
		{[]float64{100}, 8, 4},
		{[]float64{1, 0, 3, -2, math.NaN()}, 12, 7},
		{[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3, 2},
		{[]float64{0.001, 1000}, 5, 5},
		{nil, 6, 4},
		{[]float64{0, 0}, 6, 4},
	}
	for _, c := range cases {
		rects := layoutTreemap(c.values, treemapRect{0, 0, c.w, c.h})
		if len(rects) != len(c.values) {
			t.Fatalf("layoutTreemap(%v) returned %d rects, want %d", c.values, len(rects), len(c.values))
		}
		var sum float64
		for _, v := range c.values {
			if isFinite(v) && v > 0 {
				sum += v
			}
		}
		if sum <= 0 {
			continue // nothing positive to tile; an empty grid is correct.
		}
		counts := make([][]int, c.h)
		for y := range counts {
			counts[y] = make([]int, c.w)
		}
		var totalArea int
		for _, r := range rects {
			if r.W < 0 || r.H < 0 {
				t.Fatalf("negative rect size in %v: %+v", c.values, r)
			}
			totalArea += r.W * r.H
			for y := r.Y; y < r.Y+r.H; y++ {
				for x := r.X; x < r.X+r.W; x++ {
					if x < 0 || x >= c.w || y < 0 || y >= c.h {
						t.Fatalf("rect %+v escapes grid %dx%d", r, c.w, c.h)
					}
					counts[y][x]++
				}
			}
		}
		if totalArea != c.w*c.h {
			t.Errorf("values=%v grid=%dx%d: total rect area = %d, want %d", c.values, c.w, c.h, totalArea, c.w*c.h)
		}
		for y := range counts {
			for x := range counts[y] {
				if counts[y][x] != 1 {
					t.Fatalf("values=%v grid=%dx%d: cell (%d,%d) covered %d times, want exactly 1", c.values, c.w, c.h, x, y, counts[y][x])
				}
			}
		}
	}
}

func TestLayoutTreemapZeroValueGetsNoArea(t *testing.T) {
	rects := layoutTreemap([]float64{5, 0, 5}, treemapRect{0, 0, 10, 4})
	if rects[1].W != 0 || rects[1].H != 0 {
		t.Errorf("zero-value segment got non-zero rect: %+v", rects[1])
	}
	if rects[0].W*rects[0].H+rects[2].W*rects[2].H != 40 {
		t.Errorf("remaining segments should cover the full 40 cells, got %+v %+v", rects[0], rects[2])
	}
}

func TestLayoutTreemapPrefersLandscapeSplit(t *testing.T) {
	// A 20x10 rect (2:1, already landscape) is well under the 3:1 bias
	// threshold, so it should split along height -- two 20x5 boxes,
	// stacked, each wider still -- rather than along width into two
	// 10x10 squares.
	rects := layoutTreemap([]float64{1, 1}, treemapRect{0, 0, 20, 10})
	for _, r := range rects {
		if r.W != 20 || r.H != 5 {
			t.Errorf("expected a height-split into 20x5 boxes, got %+v (all rects=%+v)", r, rects)
		}
	}
}

func TestLayoutTreemapWidthSplitOnceAlreadyLandscapeEnough(t *testing.T) {
	// A 100x10 rect (10:1) is already far past the bias threshold, so
	// layoutTreemap falls back to a width-split rather than pushing it
	// to an even more extreme aspect ratio.
	rects := layoutTreemap([]float64{1, 1}, treemapRect{0, 0, 100, 10})
	for _, r := range rects {
		if r.W != 50 || r.H != 10 {
			t.Errorf("expected a width-split into 50x10 boxes, got %+v (all rects=%+v)", r, rects)
		}
	}
}

func TestLayoutTreemapForcesWidthSplitWhenHeightIsOne(t *testing.T) {
	// With only 1 row available, a height-split can't give more than one
	// partition any area at all, so it must fall back to a width-split
	// regardless of the aspect ratio.
	rects := layoutTreemap([]float64{1, 1}, treemapRect{0, 0, 10, 1})
	var totalArea int
	for _, r := range rects {
		if r.H != 1 {
			t.Errorf("expected H=1 preserved, got %+v", r)
		}
		totalArea += r.W * r.H
	}
	if totalArea != 10 {
		t.Errorf("total area = %d, want 10 (both partitions must get real width)", totalArea)
	}
}

func TestRenderTreemapDimensions(t *testing.T) {
	segments := []TreemapSegment{{Name: "a", Value: 3}, {Name: "b", Value: 1}, {Name: "c", Value: 6}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 20, Height: 8})
	if len(rows) != 8 {
		t.Fatalf("len(rows) = %d, want 8", len(rows))
	}
	for i, row := range rows {
		if got := utf8.RuneCountInString(row); got != 20 {
			t.Errorf("row %d width = %d, want 20 (row=%q)", i, got, row)
		}
	}
}

func TestRenderTreemapDrawsBorderAndLabel(t *testing.T) {
	segments := []TreemapSegment{{Name: "A", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 10, Height: 4})
	joined := strings.Join(rows, "\n")
	for _, want := range []string{"┌", "┐", "└", "┘", "A"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q in rendered treemap:\n%s", want, joined)
		}
	}
	// Corners land exactly on the grid edges.
	if !strings.HasPrefix(rows[0], "┌") || !strings.HasSuffix(rows[0], "┐") {
		t.Errorf("row 0 = %q, want border corners at both ends", rows[0])
	}
	if !strings.HasPrefix(rows[3], "└") || !strings.HasSuffix(rows[3], "┘") {
		t.Errorf("row 3 = %q, want border corners at both ends", rows[3])
	}
}

func TestRenderTreemapSmallBoxSkipsBorder(t *testing.T) {
	// A 1x1 box can't fit a border (needs >=3 cols, >=2 rows), but its
	// single cell IS big enough for a one-digit marker.
	segments := []TreemapSegment{{Name: "solo", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 1, Height: 1})
	if len(rows) != 2 { // grid row + legend row
		t.Fatalf("len(rows) = %d, want 2 (grid + legend): %q", len(rows), rows)
	}
	if utf8.RuneCountInString(rows[0]) != 1 {
		t.Fatalf("unexpected grid row: %q", rows)
	}
	if strings.ContainsAny(rows[0], "┌┐└┘│─") {
		t.Errorf("1x1 box should not draw a border: %q", rows[0])
	}
	if rows[0] != "¹" {
		t.Errorf("1x1 box should show its marker, got %q", rows[0])
	}
	// The legend itself is also only 1 column wide here, too narrow to
	// show "¹solo", so it degrades to a bare ellipsis rather than nothing.
	if rows[1] != "…" {
		t.Errorf("1-wide legend should be a bare ellipsis, got %q", rows[1])
	}
}

func TestRenderTreemapNoBorderOption(t *testing.T) {
	segments := []TreemapSegment{{Name: "A", Value: 1}, {Name: "B", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 10, Height: 4, NoBorder: true, Glyphs: []rune("AB")})
	joined := strings.Join(rows, "\n")
	if strings.ContainsAny(joined, "┌┐└┘│─") {
		t.Errorf("NoBorder should suppress all border glyphs: %q", joined)
	}
}

func TestRenderTreemapTooLongNameGetsNumberNotFragment(t *testing.T) {
	// A box too small for its full name gets a number, never a truncated
	// fragment of the name -- "cannot be described" means "get a number,"
	// not "get an unhelpful sliver of the name."
	segments := []TreemapSegment{{Name: "a-very-long-process-name", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 10, Height: 4})
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "¹") {
		t.Errorf("expected marker ¹ in place of the untruncated name, got:\n%s", joined)
	}
	if strings.Contains(joined, "a-very-long-process-name") {
		t.Errorf("full name should not appear (too narrow to fit), got:\n%s", joined)
	}
	// Fragments like "a-ve" or "a-very-l…" must never appear either.
	for _, frag := range []string{"a-ve", "a-very", "process-name"} {
		if strings.Contains(joined, frag) {
			t.Errorf("expected no partial-name fragment %q, got:\n%s", frag, joined)
		}
	}
}

func TestRenderTreemapFullLabelPreferredOverNumber(t *testing.T) {
	// Plenty of room: the real name is shown, no number, no legend row.
	segments := []TreemapSegment{{Name: "ok", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 20, Height: 6})
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "ok") {
		t.Errorf("expected the full name \"ok\", got:\n%s", joined)
	}
	if strings.ContainsAny(joined, "¹²³⁴⁵⁶⁷⁸⁹⁰") {
		t.Errorf("a box with room for its name should not get a number, got:\n%s", joined)
	}
	if len(rows) != 6 {
		t.Errorf("len(rows) = %d, want 6 (no legend row appended)", len(rows))
	}
}

func TestRenderTreemapMarkersAreSequentialAndLegendMatches(t *testing.T) {
	// Three boxes, all too narrow for their names: markers assigned in
	// segment order, and the legend maps each back to its real name.
	var segments []TreemapSegment
	for _, name := range []string{
		"alpha-process", "bravo-process", "charlie-process",
		"delta-process", "echo-process", "foxtrot-process",
	} {
		segments = append(segments, TreemapSegment{Name: name, Value: 1})
	}
	rows := RenderTreemap(segments, TreemapOptions{Width: 60, Height: 6})
	legend := rows[len(rows)-1]
	for i, want := range []string{"¹alpha-process", "²bravo-process", "³charlie-process"} {
		if !strings.Contains(legend, want) {
			t.Errorf("legend missing entry %d (%q), got legend=%q", i+1, want, legend)
		}
	}
}

func TestRenderTreemapMarkersRankByValueNotLayoutOrder(t *testing.T) {
	// Regression: marker 1 must go to the biggest numbered consumer, not
	// whichever unlabelable box happens to appear first in segments/box
	// layout order. All four are too narrow for their names at this width.
	segments := []TreemapSegment{
		{Name: "claude", Value: 250}, // dominant, squeezes the rest into small boxes
		{Name: "tilix", Value: 4.4},
		{Name: "other", Value: 12.0},
		{Name: "firefox", Value: 13.0},
		{Name: "conmon", Value: 6.4},
	}
	rows := RenderTreemap(segments, TreemapOptions{Width: 60, Height: 3, ShowValues: true, ValuePrecision: 1, ValueSuffix: "%"})
	legend := rows[len(rows)-1]
	idx := map[string]int{}
	for rank, name := range []string{"firefox", "other", "conmon", "tilix"} {
		want := fmt.Sprintf("%s%s %.1f%%", superscriptNumber(rank+1), name, valueOf(segments, name))
		pos := strings.Index(legend, want)
		if pos < 0 {
			t.Fatalf("expected legend entry %q at rank %d, got legend=%q", want, rank+1, legend)
		}
		idx[name] = pos
	}
	if !(idx["firefox"] < idx["other"] && idx["other"] < idx["conmon"] && idx["conmon"] < idx["tilix"]) {
		t.Errorf("legend not ordered by descending value: positions=%+v legend=%q", idx, legend)
	}
}

func valueOf(segments []TreemapSegment, name string) float64 {
	for _, s := range segments {
		if s.Name == name {
			return s.Value
		}
	}
	return 0
}

func TestRenderTreemapNumberOnFrameWhenNoInteriorRoom(t *testing.T) {
	// A bordered box exactly 2 rows tall has zero interior rows (both rows
	// are border), so its marker must land on the top border itself.
	segments := []TreemapSegment{{Name: "a-long-name-that-does-not-fit", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 5, Height: 2})
	if len(rows) < 1 || !strings.Contains(rows[0], "¹") {
		t.Errorf("expected marker ¹ overlaid on the top border row, got:\n%s", strings.Join(rows, "\n"))
	}
}

func TestRenderTreemapLegendTruncatesWithEllipsis(t *testing.T) {
	var segments []TreemapSegment
	for i := 0; i < 10; i++ {
		segments = append(segments, TreemapSegment{Name: fmt.Sprintf("process-name-%d", i), Value: 1})
	}
	rows := RenderTreemap(segments, TreemapOptions{Width: 30, Height: 6})
	legend := rows[len(rows)-1]
	if got := utf8.RuneCountInString(legend); got != 30 {
		t.Fatalf("legend width = %d, want 30 (legend=%q)", got, legend)
	}
	if !strings.HasSuffix(strings.TrimRight(legend, " "), "…") {
		t.Errorf("expected legend to end with an ellipsis when it can't fit everyone, got %q", legend)
	}
}

func TestRenderTreemapNoLabelsOption(t *testing.T) {
	segments := []TreemapSegment{{Name: "visible-name", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 20, Height: 6, NoLabels: true})
	joined := strings.Join(rows, "\n")
	if strings.Contains(joined, "visible-name") {
		t.Errorf("NoLabels should suppress the label, got:\n%s", joined)
	}
}

func TestRenderTreemapShowValues(t *testing.T) {
	segments := []TreemapSegment{{Name: "cpu", Value: 42.5}}
	rows := RenderTreemap(segments, TreemapOptions{
		Width: 20, Height: 6, ShowValues: true, ValuePrecision: 1, ValueSuffix: "%",
	})
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "cpu 42.5%") {
		t.Errorf("expected value label \"cpu 42.5%%\" in:\n%s", joined)
	}
}

func TestRenderTreemapANSIResetsOnStyleChange(t *testing.T) {
	segments := []TreemapSegment{{Name: "A", Value: 1}, {Name: "B", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{
		Width: 10, Height: 4, ANSI: true, NoBorder: true,
		BackgroundANSI: []string{"41", "42"},
	})
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "\x1b[41m") || !strings.Contains(joined, "\x1b[42m") {
		t.Errorf("expected both segment background codes present:\n%q", joined)
	}
	if !strings.Contains(joined, "\x1b[0m") {
		t.Errorf("expected an ANSI reset on style change:\n%q", joined)
	}
}

func TestRenderTreemapLegendIsColoredUnderANSI(t *testing.T) {
	var segments []TreemapSegment
	for _, name := range []string{
		"alpha-process", "bravo-process", "charlie-process",
		"delta-process", "echo-process", "foxtrot-process",
	} {
		segments = append(segments, TreemapSegment{Name: name, Value: 1})
	}
	rows := RenderTreemap(segments, TreemapOptions{
		Width: 60, Height: 6, ANSI: true,
		BackgroundANSI: []string{"41", "42", "43", "44", "45", "46"},
	})
	legend := rows[len(rows)-1]
	// "41" (red background) converts to "31" (red text) -- the legend
	// colors text, never a filled background (that would read as another
	// treemap box rather than a legend entry).
	if !strings.Contains(legend, "\x1b[31m¹alpha-process\x1b[0m") {
		t.Errorf("expected legend entry 1 colored as red TEXT (not a red background), got %q", legend)
	}
	if strings.Contains(legend, "\x1b[41m") {
		t.Errorf("legend must not use a filled background color, got %q", legend)
	}
}

func TestTreemapBackgroundToForeground(t *testing.T) {
	cases := map[string]string{
		"40": "30", "41": "31", "47": "37", // normal background range
		"100": "90", "107": "97", // bright background range
		"31":   "31",   // already foreground: unchanged
		"1;41": "1;41", // compound code: unchanged (not a plain integer)
		"none": "none", // non-numeric: unchanged
		"":     "",     // empty: unchanged
	}
	for in, want := range cases {
		if got := treemapBackgroundToForeground(in); got != want {
			t.Errorf("treemapBackgroundToForeground(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderTreemapLegendPlainWithoutANSI(t *testing.T) {
	segments := []TreemapSegment{{Name: "a-very-long-process-name", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 10, Height: 4})
	legend := rows[len(rows)-1]
	if strings.Contains(legend, "\x1b[") {
		t.Errorf("expected no ANSI codes in legend without opts.ANSI, got %q", legend)
	}
}

func TestRenderTreemapLegendNeverCutsMidName(t *testing.T) {
	var segments []TreemapSegment
	for i := 0; i < 10; i++ {
		segments = append(segments, TreemapSegment{Name: fmt.Sprintf("process-name-%d", i), Value: 1})
	}
	rows := RenderTreemap(segments, TreemapOptions{Width: 30, Height: 6})
	legend := strings.TrimRight(rows[len(rows)-1], " ")
	legend = strings.TrimSuffix(legend, "…")
	for _, tok := range strings.Split(legend, " ") {
		if tok == "" {
			continue
		}
		// Every token must be a full, untruncated "markerprocess-name-N".
		trimmedMarker := strings.TrimLeftFunc(tok, func(r rune) bool {
			return strings.ContainsRune("⁰¹²³⁴⁵⁶⁷⁸⁹", r)
		})
		if !strings.HasPrefix(trimmedMarker, "process-name-") {
			t.Errorf("legend token %q looks like a cut-off fragment, got legend=%q", tok, rows[len(rows)-1])
		}
	}
}

func TestRenderTreemapNeverPanics(t *testing.T) {
	segmentSets := [][]TreemapSegment{
		nil,
		{},
		{{Name: "", Value: 0}},
		{{Name: "x", Value: math.NaN()}},
		{{Name: "x", Value: math.Inf(1)}},
		{{Name: "x", Value: math.Inf(-1)}},
		{{Name: "x", Value: -5}},
	}
	for i := 0; i < 20; i++ {
		segmentSets = append(segmentSets, []TreemapSegment{{Name: "s", Value: float64(i + 1)}})
	}
	var many []TreemapSegment
	for i := 0; i < 50; i++ {
		many = append(many, TreemapSegment{Name: "p", Value: float64(i + 1)})
	}
	segmentSets = append(segmentSets, many)

	for _, segments := range segmentSets {
		for _, dim := range []int{-1, 0, 1, 2, 3, 40} {
			_ = RenderTreemap(segments, TreemapOptions{Width: dim, Height: dim})
			_ = RenderTreemap(segments, TreemapOptions{Width: dim, Height: dim, ANSI: true, ShowValues: true})
			_ = RenderTreemap(segments, TreemapOptions{Width: dim, Height: dim, NoBorder: true, NoLabels: true})
		}
	}
}

func TestRenderTreemapRealProcSnapshot(t *testing.T) {
	root := buildProcTree(t, realProcSnapshot)
	segments := AggregateTreemap(root, MaxTreemapNodes)
	rows := RenderTreemap(segments, TreemapOptions{
		Width: 60, Height: 15, ShowValues: true, ValuePrecision: 1,
	})
	if len(rows) != 15 && len(rows) != 16 { // grid rows, plus an optional legend row
		t.Fatalf("len(rows) = %d, want 15 or 16", len(rows))
	}
	for i, row := range rows {
		if got := utf8.RuneCountInString(row); got != 60 {
			t.Errorf("row %d width = %d, want 60", i, got)
		}
	}
	t.Logf("treemap:\n%s", strings.Join(rows, "\n"))
}

func TestTreemapThemeClassicIsZeroValue(t *testing.T) {
	var theme TreemapTheme
	if theme != TreemapThemeClassic {
		t.Errorf("zero value of TreemapTheme = %v, want TreemapThemeClassic", theme)
	}
}

func blocksOptions(width, height int) TreemapOptions {
	return TreemapOptions{
		Width: width, Height: height, Theme: TreemapThemeBlocks, ANSI: true,
		BackgroundANSI: []string{"41", "42", "43", "44"},
	}
}

func TestRenderTreemapBlocksDimensions(t *testing.T) {
	segments := []TreemapSegment{{Name: "A", Value: 3}, {Name: "B", Value: 1}}
	rows := RenderTreemap(segments, blocksOptions(20, 6))
	if len(rows) != 6 {
		t.Fatalf("len(rows) = %d, want 6 (no legend needed)", len(rows))
	}
	for i, row := range rows {
		if w := measure.StringWidth(row); w != 20 {
			t.Errorf("row %d width = %d, want 20", i, w)
		}
	}
}

func TestRenderTreemapBlocksNeverDrawsClassicBorderGlyphs(t *testing.T) {
	// TreemapThemeBlocks is always full-bleed -- no reserved border cell,
	// so none of the box-drawing glyphs TreemapThemeClassic uses should
	// ever appear.
	segments := []TreemapSegment{{Name: "A", Value: 3}, {Name: "B", Value: 1}, {Name: "C", Value: 1}}
	rows := RenderTreemap(segments, blocksOptions(20, 8))
	joined := strings.Join(rows, "\n")
	if strings.ContainsAny(joined, "┌┐└┘│─") {
		t.Errorf("TreemapThemeBlocks must never draw a classic border glyph, got:\n%s", joined)
	}
}

func TestRenderTreemapBlocksBoundaryCellsUseHalfBlockGlyphs(t *testing.T) {
	// Two adjacent, differently-sized segments guarantee at least one box
	// whose own right/bottom edge is exercised.
	segments := []TreemapSegment{{Name: "A", Value: 3}, {Name: "B", Value: 1}}
	rows := RenderTreemap(segments, blocksOptions(20, 6))
	joined := strings.Join(rows, "\n")
	if !strings.ContainsAny(joined, "▌▀▘") {
		t.Errorf("expected at least one edge/corner glyph (▌▀▘), got:\n%s", joined)
	}
}

func TestRenderTreemapBlocksEdgeAndCornerGlyphsAreLocalToOwnBox(t *testing.T) {
	// Two equal segments in a 10x2 grid: landscape bias (W=10 >= H=2*3)
	// picks a width split, giving box A rect{X:0, W:5, H:2} -- its own
	// rightmost column is col 4, its own bottommost row is row 1.
	segments := []TreemapSegment{{Name: "A", Value: 1}, {Name: "B", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{
		Width: 10, Height: 2, Theme: TreemapThemeBlocks, ANSI: true, NoLabels: true,
		BackgroundANSI: []string{"41", "42"},
	})
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	// row 0, col 4: right-edge-only cell -> '▌', foreground-only "31" (41
	// converted to foreground), regardless of B (the box past this edge).
	if !strings.Contains(rows[0], "\x1b[31m▌\x1b[0m") {
		t.Errorf("expected a foreground-only-styled '▌' on box A's right edge, row0=%q", rows[0])
	}
	// row 1 (bottom row): interior bottom-only cells -> '▀', the corner
	// cell (col 4, also the right edge) -> '▘'. Both use the same
	// foreground-only "31" code, so they render as one unbroken style run
	// (no code re-emitted mid-run) ending "...▀▘".
	if !strings.Contains(rows[1], "\x1b[31m▀▀▀▀▘\x1b[0m") {
		t.Errorf("expected box A's bottom row to read '▀▀▀▀▘' under one foreground-only \"31\" run, row1=%q", rows[1])
	}
}

func TestRenderTreemapBlocksGlyphCellsLeaveBackgroundUnset(t *testing.T) {
	// The whole point: a glyph cell's unfilled half/quadrant has no
	// background SGR at all, so the ambient terminal background shows
	// through it -- never a color this code makes up or borrows from a
	// neighbor.
	segments := []TreemapSegment{{Name: "A", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{
		Width: 6, Height: 3, Theme: TreemapThemeBlocks, ANSI: true, NoLabels: true,
		BackgroundANSI: []string{"41"},
	})
	joined := strings.Join(rows, "\n")
	if strings.Contains(joined, "\x1b[41;31m") || strings.Contains(joined, "\x1b[31;41m") {
		t.Errorf("glyph cells must never combine this box's color with a background code, got:\n%q", joined)
	}
	if !strings.Contains(joined, "\x1b[41m") {
		t.Errorf("interior cells should still use the plain background fill \"41\", got:\n%q", joined)
	}
}

func TestRenderTreemapBlocksWithoutBackgroundPaletteFallsBackToSolidFill(t *testing.T) {
	// ANSI is on, but no BackgroundANSI palette: there's no color to blend,
	// so this must render identically to the shared classic NoBorder+ANSI
	// fill path (blank cells styled by treemapSegmentCode), not attempt to
	// blend with an empty palette (which would panic on modulo-by-zero).
	segments := []TreemapSegment{{Name: "A", Value: 3}, {Name: "B", Value: 1}}
	blocks := RenderTreemap(segments, TreemapOptions{
		Width: 20, Height: 6, Theme: TreemapThemeBlocks, ANSI: true, ForegroundANSI: []string{"97"},
	})
	classic := RenderTreemap(segments, TreemapOptions{
		Width: 20, Height: 6, NoBorder: true, ANSI: true, ForegroundANSI: []string{"97"},
	})
	if strings.Join(blocks, "\n") != strings.Join(classic, "\n") {
		t.Errorf("TreemapThemeBlocks without a background palette should match classic NoBorder+ANSI fill exactly:\nblocks:  %q\nclassic: %q", blocks, classic)
	}
}

func TestRenderTreemapBlocksWithoutANSIFallsBackToSolidGlyphFill(t *testing.T) {
	// Without ANSI at all, there's no color to blend either: must match the
	// shared classic NoBorder glyph-fill path exactly.
	segments := []TreemapSegment{{Name: "A", Value: 3}, {Name: "B", Value: 1}}
	glyphs := []rune("AB")
	blocks := RenderTreemap(segments, TreemapOptions{
		Width: 20, Height: 6, Theme: TreemapThemeBlocks, Glyphs: glyphs,
	})
	classic := RenderTreemap(segments, TreemapOptions{
		Width: 20, Height: 6, NoBorder: true, Glyphs: glyphs,
	})
	if strings.Join(blocks, "\n") != strings.Join(classic, "\n") {
		t.Errorf("TreemapThemeBlocks without ANSI should match classic NoBorder glyph fill exactly:\nblocks:  %q\nclassic: %q", blocks, classic)
	}
}

func TestRenderTreemapBlocksLabelsUseFullBoxAreaNoInset(t *testing.T) {
	// No border cell is reserved under TreemapThemeBlocks, so a label can
	// start right at a box's edge column -- verify a single, full-width
	// box's label starts at column 0, not column 1 (which a classic
	// bordered box would require). Plain (non-ANSI) so the row is content
	// only, no escape codes to skip over.
	segments := []TreemapSegment{{Name: "hi", Value: 1}}
	rows := RenderTreemap(segments, TreemapOptions{Width: 10, Height: 3, Theme: TreemapThemeBlocks})
	middle := []rune(rows[1])
	if middle[0] != 'h' {
		t.Errorf("expected label to start at column 0 (full-bleed, no border inset), row = %q", rows[1])
	}
}

func TestRenderTreemapBlocksMarkerLegendStillWorks(t *testing.T) {
	// Six same-value segments subdivide the grid small enough that none of
	// their full names fit, forcing the marker/legend fallback.
	var segments []TreemapSegment
	for _, name := range []string{
		"alpha-process", "bravo-process", "charlie-process",
		"delta-process", "echo-process", "foxtrot-process",
	} {
		segments = append(segments, TreemapSegment{Name: name, Value: 1})
	}
	rows := RenderTreemap(segments, blocksOptions(60, 6))
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "¹") {
		t.Errorf("expected marker fallback to still work under TreemapThemeBlocks, got:\n%s", joined)
	}
	legend := rows[len(rows)-1]
	if !strings.Contains(legend, "¹alpha-process") {
		t.Errorf("expected legend entry for the numbered box, got %q", legend)
	}
}

func TestRenderTreemapBlocksNeverPanics(t *testing.T) {
	segmentSets := [][]TreemapSegment{
		nil,
		{},
		{{Name: "", Value: 0}},
		{{Name: "x", Value: math.NaN()}},
		{{Name: "x", Value: math.Inf(1)}},
		{{Name: "x", Value: -5}},
	}
	for i := 0; i < 10; i++ {
		segmentSets = append(segmentSets, []TreemapSegment{{Name: "s", Value: float64(i + 1)}})
	}
	var many []TreemapSegment
	for i := 0; i < 30; i++ {
		many = append(many, TreemapSegment{Name: "p", Value: float64(i + 1)})
	}
	segmentSets = append(segmentSets, many)

	for _, segments := range segmentSets {
		for _, dim := range []int{-1, 0, 1, 2, 3, 40} {
			_ = RenderTreemap(segments, TreemapOptions{Width: dim, Height: dim, Theme: TreemapThemeBlocks})
			_ = RenderTreemap(segments, blocksOptions(dim, dim))
			_ = RenderTreemap(segments, TreemapOptions{Width: dim, Height: dim, Theme: TreemapThemeBlocks, ANSI: true})
		}
	}
}
