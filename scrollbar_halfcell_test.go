package loom

import (
	"fmt"
	"strings"
	"testing"
)

// Issue 450: the scrollbar thumb moves in half-row units and a themed bar
// paints its track and thumb as cell backgrounds. These tests drive View and
// Viewport through their key, wheel and mouse paths and compare the rendered
// bar column cell by cell. They pin the colour profile explicitly so they do
// not depend on the caller's environment.

var (
	halfCellTrackBG = ColorIndex(27)
	halfCellThumbBG = ColorIndex(51)
	halfCellTrackFG = ColorIndex(33)
	halfCellThumbFG = ColorIndex(51)
)

// halfCellBar is a themed bar with distinct track and thumb backgrounds.
func halfCellBar() ScrollbarStyle {
	return ScrollbarStyle{
		Track: Style{FG: halfCellTrackFG, BG: halfCellTrackBG},
		Thumb: Style{FG: halfCellThumbFG, BG: halfCellThumbBG},
	}
}

// barChild is a non-claiming child with a fixed content height.
type barChild struct {
	captureProbe
	height int
}

func (c *barChild) ContentHeight() int { return c.height }

func newBarView(lines int) *View {
	v := NewView(make([]string, lines))
	v.Scrollbar = halfCellBar()
	return v
}

func drawBarView(v *View, rows int, profile ColorProfile) *Canvas {
	c := NewCanvas(6, rows)
	c.ColorProfile = profile
	v.Draw(c, Rect{W: 6, H: rows})
	return c
}

func newBarViewport(height int) *Viewport {
	vp := NewViewport(&barChild{height: height})
	vp.ScrollbarMode = ScrollbarAlways
	vp.Scrollbar = halfCellBar()
	return vp
}

func drawBarViewport(vp *Viewport, rows int, profile ColorProfile) *Canvas {
	c := NewCanvas(6, rows)
	c.ColorProfile = profile
	vp.Draw(c, Rect{W: 6, H: rows})
	return c
}

// barColumn returns the rightmost column of c, one cell per row. The bar is
// located by column index, never by byte offset into any string.
func barColumn(c *Canvas, rows int) []Cell {
	x := c.Bounds().W - 1
	cells := make([]Cell, rows)
	for y := range cells {
		cells[y] = c.Get(x, y)
	}
	return cells
}

// barPicture reduces a bar column to one comparable string: glyph, FG and BG
// per row, so two offsets that look different never compare equal.
func barPicture(cells []Cell) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		parts[i] = fmt.Sprintf("%s|%v|%v", c.Text, c.Style.FG, c.Style.BG)
	}
	return strings.Join(parts, ";")
}

func wantCell(t *testing.T, row int, got Cell, text string, fg, bg Color) {
	t.Helper()
	if got.Text != text || got.Style.FG != fg || got.Style.BG != bg {
		t.Fatalf("row %d = %q fg=%v bg=%v, want %q fg=%v bg=%v",
			row, got.Text, got.Style.FG, got.Style.BG, text, fg, bg)
	}
}

// Track cells are a space on the track background; full thumb cells are a
// space on the thumb background; a partial cell is the lower half block.
// The lower-only cell (thumb starts inside the cell) draws thumb-coloured
// lower half on the track; the upper-only cell (thumb ends inside) is the
// same block inverted: track-coloured lower half on the thumb.
func TestHalfCellSevenDistinctPicturesAcrossKeyScroll(t *testing.T) {
	v := newBarView(12) // 12 lines over 6 rows: offsets 0..6
	seen := map[string]bool{}
	for step := 0; step <= 6; step++ {
		seen[barPicture(barColumn(drawBarView(v, 6, ColorProfileTrueColor), 6))] = true
		v.ConsumeKey(KeyEvent{Key: "down"})
	}
	if v.Scroll != 6 {
		t.Fatalf("keys scrolled to %d, want 6", v.Scroll)
	}
	if len(seen) != 7 {
		t.Fatalf("12/6 pictures=%d, want 7 distinct offsets", len(seen))
	}
}

func TestHalfCellThreeDistinctPicturesAcrossWheel(t *testing.T) {
	v := newBarView(8) // 8 lines over 6 rows: offsets 0..2
	seen := map[string]bool{}
	for step := 0; step <= 2; step++ {
		seen[barPicture(barColumn(drawBarView(v, 6, ColorProfileTrueColor), 6))] = true
		v.ConsumeMouse(MouseEvent{Action: MouseScrollDown})
	}
	if len(seen) != 3 {
		t.Fatalf("8/6 pictures=%d, want 3 distinct offsets", len(seen))
	}
}

// A 60/6 bar has a one-cell thumb. At offset 3 its two half units straddle
// rows 0 and 1: the lower half of row 0 and the upper half of row 1.
func TestHalfCellStraddleAtOneCellThumb(t *testing.T) {
	trackBG, thumbBG := halfCellTrackBG, halfCellThumbBG
	trackFG := halfCellTrackFG
	want := []Cell{
		{Text: "▄", Style: Style{FG: thumbBG, BG: trackBG}}, // thumb starts inside row 0
		{Text: "▄", Style: Style{FG: trackBG, BG: thumbBG}}, // thumb ends inside row 1
		{Text: " ", Style: Style{FG: trackFG, BG: trackBG}},
		{Text: " ", Style: Style{FG: trackFG, BG: trackBG}},
		{Text: " ", Style: Style{FG: trackFG, BG: trackBG}},
		{Text: " ", Style: Style{FG: trackFG, BG: trackBG}},
	}
	t.Run("view", func(t *testing.T) {
		v := newBarView(60)
		v.Scroll = 3
		got := barColumn(drawBarView(v, 6, ColorProfileTrueColor), 6)
		for y, w := range want {
			wantCell(t, y, got[y], w.Text, w.Style.FG, w.Style.BG)
		}
	})
	t.Run("viewport", func(t *testing.T) {
		vp := newBarViewport(60)
		vp.ScrollY = 3
		got := barColumn(drawBarViewport(vp, 6, ColorProfileTrueColor), 6)
		for y, w := range want {
			wantCell(t, y, got[y], w.Text, w.Style.FG, w.Style.BG)
		}
	})
}

func TestHalfCellEndsReachTheTrackEnds(t *testing.T) {
	trackFG, trackBG := halfCellTrackFG, halfCellTrackBG
	thumbFG, thumbBG := halfCellThumbFG, halfCellThumbBG
	v := newBarView(60)
	v.Scroll = 0
	top := barColumn(drawBarView(v, 6, ColorProfileTrueColor), 6)
	wantCell(t, 0, top[0], " ", thumbFG, thumbBG)
	wantCell(t, 5, top[5], " ", trackFG, trackBG)

	v.Scroll = 54
	bottom := barColumn(drawBarView(v, 6, ColorProfileTrueColor), 6)
	wantCell(t, 5, bottom[5], " ", thumbFG, thumbBG)
	wantCell(t, 0, bottom[0], " ", trackFG, trackBG)
}

// Dragging keeps the grabbed half-row phase. The thumb is grabbed in its lower
// half (offset 3, thumb units 1-2); one row of pointer travel moves it two
// units, so it stays a lower-half thumb on row 1 at offset 16.
func TestViewportHalfCellDragKeepsGrabbedPhase(t *testing.T) {
	vp := newBarViewport(60)
	vp.ScrollY = 3
	c := drawBarViewport(vp, 6, ColorProfileTrueColor)
	if got := barColumn(c, 6)[0]; got.Text != "▄" {
		t.Fatalf("setup: row 0 = %q, want the lower half block", got.Text)
	}
	vp.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 5, Y: 0})
	vp.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 5, Y: 1})
	if vp.ScrollY != 16 {
		t.Fatalf("drag one row: ScrollY=%d, want 16", vp.ScrollY)
	}
	cells := barColumn(drawBarViewport(vp, 6, ColorProfileTrueColor), 6)
	wantCell(t, 1, cells[1], "▄", halfCellThumbBG, halfCellTrackBG)
}

func TestViewHalfCellDragKeepsGrabbedPhase(t *testing.T) {
	v := newBarView(60)
	v.Scroll = 3
	drawBarView(v, 6, ColorProfileTrueColor)
	v.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 5, Y: 0})
	v.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 5, Y: 1})
	if v.Scroll != 16 {
		t.Fatalf("drag one row: Scroll=%d, want 16", v.Scroll)
	}
	cells := barColumn(drawBarView(v, 6, ColorProfileTrueColor), 6)
	wantCell(t, 1, cells[1], "▄", halfCellThumbBG, halfCellTrackBG)
}

// A press on either half of the thumb keeps the offset. Only a press on a row
// the thumb does not touch moves it.
func TestHalfCellPressOnPartialThumbRowKeepsOffset(t *testing.T) {
	t.Run("view", func(t *testing.T) {
		v := newBarView(60)
		v.Scroll = 3
		drawBarView(v, 6, ColorProfileTrueColor)
		for _, row := range []int{0, 1} { // lower-only row 0, upper-only row 1
			v.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 5, Y: row})
			if v.Scroll != 3 {
				t.Fatalf("press on thumb row %d moved Scroll to %d", row, v.Scroll)
			}
		}
	})
	t.Run("viewport", func(t *testing.T) {
		vp := newBarViewport(60)
		vp.ScrollY = 3
		drawBarViewport(vp, 6, ColorProfileTrueColor)
		for _, row := range []int{0, 1} {
			vp.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 5, Y: row})
			if vp.ScrollY != 3 {
				t.Fatalf("press on thumb row %d moved ScrollY to %d", row, vp.ScrollY)
			}
		}
	})
}

// A track click on the Viewport centers the thumb in half-row units: row 4 of
// a 40/10 bar (thumb 5 units) puts the thumb on units 7-11 at offset 14.
func TestViewportHalfCellTrackClickCentersThumb(t *testing.T) {
	vp := newBarViewport(40)
	drawBarViewport(vp, 10, ColorProfileTrueColor)
	vp.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 5, Y: 4})
	if vp.ScrollY != 14 {
		t.Fatalf("track click: ScrollY=%d, want 14", vp.ScrollY)
	}
	cells := barColumn(drawBarViewport(vp, 10, ColorProfileTrueColor), 10)
	wantCell(t, 2, cells[2], " ", halfCellTrackFG, halfCellTrackBG)
	wantCell(t, 3, cells[3], "▄", halfCellThumbBG, halfCellTrackBG)
	wantCell(t, 4, cells[4], " ", halfCellThumbFG, halfCellThumbBG)
	wantCell(t, 5, cells[5], " ", halfCellThumbFG, halfCellThumbBG)
	wantCell(t, 6, cells[6], " ", halfCellTrackFG, halfCellTrackBG)
}

// The theme's own track and thumb colours reach the bar. A distinct thumb must
// not be painted over with the surface colour.
func TestViewportThemeKeepsDistinctTrackAndThumbBackgrounds(t *testing.T) {
	vp := newBarViewport(60)
	vp.ApplyTheme(ThemeColors{
		NormalBG:         ThemeColorIndex(16),
		ScrollbarTrackBG: ThemeColorIndex(27),
		ScrollbarThumbBG: ThemeColorIndex(51),
	})
	if vp.Scrollbar.Track.BG != halfCellTrackBG || vp.Scrollbar.Thumb.BG != halfCellThumbBG {
		t.Fatalf("bar BG track=%v thumb=%v, want the theme colours", vp.Scrollbar.Track.BG, vp.Scrollbar.Thumb.BG)
	}
	cells := barColumn(drawBarViewport(vp, 6, ColorProfileTrueColor), 6)
	wantCell(t, 0, cells[0], " ", vp.Scrollbar.Thumb.FG, halfCellThumbBG)
	wantCell(t, 5, cells[5], " ", vp.Scrollbar.Track.FG, halfCellTrackBG)
}

// Without colour, and with a plain theme that has no track or thumb
// background, the bar stays whole-cell glyphs. Rounding is per cell: the thumb
// covers each cell that its half units touch.
func TestHalfCellNoColorFallbackStaysWholeCellGlyphs(t *testing.T) {
	thumb := SpeccedDefaults.Scrollbar.ForegroundChar
	track := SpeccedDefaults.Scrollbar.BackgroundChar
	for _, tc := range []struct {
		scroll int
		first  int // first thumb row
		last   int // last thumb row (inclusive)
	}{
		{0, 0, 2},
		{3, 2, 4},
		{6, 3, 5},
	} {
		v := newBarView(12)
		v.Scroll = tc.scroll
		cells := barColumn(drawBarView(v, 6, ColorProfileNone), 6)
		for y, cell := range cells {
			want := track
			if y >= tc.first && y <= tc.last {
				want = thumb
			}
			if cell.Text != want {
				t.Fatalf("scroll %d row %d = %q, want %q", tc.scroll, y, cell.Text, want)
			}
		}
	}
}

func TestViewportPlainThemeKeepsGlyphBar(t *testing.T) {
	vp := newBarViewport(12)
	vp.ApplyTheme(ThemeColors{})
	cells := barColumn(drawBarViewport(vp, 6, ColorProfileTrueColor), 6)
	if cells[0].Text != SpeccedDefaults.Scrollbar.ForegroundChar || cells[5].Text != SpeccedDefaults.Scrollbar.BackgroundChar {
		t.Fatalf("plain theme bar = %q..%q, want the glyph thumb and track", cells[0].Text, cells[5].Text)
	}
}
