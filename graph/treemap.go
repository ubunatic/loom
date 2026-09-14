// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"codeberg.org/ubunatic/loom/measure"
)

// defaultTreemapGlyphs are the default per-box fill glyphs used when
// TreemapOptions.Glyphs is empty, cycling for treemaps with more segments
// than glyphs.
var defaultTreemapGlyphs = []rune("█▓▒░")

// TreemapTheme selects RenderTreemap's visual style. See issue 040.
type TreemapTheme int

const (
	// TreemapThemeClassic is the original box-drawing style: a bordered box
	// (┌─┐│└─┘) when it fits (opts.NoBorder unset, at least 3x2 cells), else
	// a solid glyph/ANSI-background fill. The zero value, so every existing
	// TreemapOptions{} literal (and every test written before theming
	// existed) renders exactly as before.
	TreemapThemeClassic TreemapTheme = iota
	// TreemapThemeBlocks renders every box full-bleed -- no border cell is
	// reserved, so labels get the box's entire area -- and softens its own
	// right and bottom edges with a half-block/quadrant glyph (▌▀▘) instead
	// of the hard one-cell-wide border line TreemapThemeClassic draws.
	// Unlike a border, these glyphs don't fill their cell solid: only the
	// half (or, at the box's own bottom-right corner, quadrant) that faces
	// this box's own interior carries its color -- the other half/quadrant
	// gets no background SGR at all, so the ambient terminal background
	// shows straight through it. See drawTreemapBlockBox's doc comment for
	// the exact per-cell rule; it is purely local to each box's own
	// rectangle (its rightmost column and bottommost row), not dependent
	// on what segment (if any) sits past that edge -- so two adjoining
	// boxes never both try to render their shared seam.
	//
	// Requires opts.ANSI and opts.BackgroundANSI (there is no color to
	// paint the glyph with otherwise); without either, TreemapThemeBlocks
	// falls back to the identical solid-glyph full-bleed fill
	// TreemapThemeClassic uses when NoBorder is set.
	TreemapThemeBlocks
)

// TreemapOptions configures RenderTreemap.
type TreemapOptions struct {
	// Width and Height are the total character grid in columns and rows.
	// Zero/negative values clamp to 1.
	Width  int
	Height int
	// Theme selects the visual style; the zero value is TreemapThemeClassic.
	Theme TreemapTheme
	// NoLabels suppresses per-box labeling entirely: no name/value text, no
	// numbered fallback marker, and no legend row.
	NoLabels bool
	// ShowValues appends each segment's formatted value after its name,
	// when there is room, using ValuePrecision decimal places and
	// ValueSuffix (e.g. "%") after the number.
	ShowValues     bool
	ValuePrecision int
	ValueSuffix    string
	// NoBorder fills every box solid instead of drawing a border; boxes
	// too small for a border (less than 3 columns or 2 rows) always fall
	// back to a solid fill regardless of this flag.
	NoBorder bool
	// Glyphs supplies one fill glyph per segment (used for a box's
	// interior when NoBorder or the box is too small to border, and for
	// its border when ANSI is off), cycling if there are more segments
	// than glyphs. Empty uses defaultTreemapGlyphs.
	Glyphs []rune
	// ANSI wraps each box's interior (and, unless NoBorder, its border) in
	// a per-segment background/foreground ANSI sequence instead of using
	// Glyphs to fill it.
	ANSI bool
	// BackgroundANSI supplies one SGR background code per segment,
	// cycling if there are more segments than codes. Empty leaves a box
	// unstyled aside from ForegroundANSI.
	BackgroundANSI []string
	// ForegroundANSI supplies one SGR foreground code per segment
	// (applied to its border and label text), cycling if there are more
	// segments than codes.
	ForegroundANSI []string
}

// treemapRect is one box's cell-grid placement, computed by layoutTreemap.
type treemapRect struct {
	X, Y, W, H int
}

// RenderTreemap lays out segments as a 2D treemap: rectangular boxes tiling
// a Width x Height character grid, each box's area proportional to its
// segment's value. Returns one string per row, plus one trailing legend row
// when any box needed one (see below); callers join with "\n" or feed the
// rows directly into a larger layout.
//
// Layout is a recursive binary split: segments are sorted by value
// descending, partitioned as evenly as possible by cumulative value, and
// the current rectangle is divided between the two partitions along
// whichever axis keeps its children landscape (wider than tall, biased by
// treemapLandscapeBias) rather than aiming for a square -- repeated until
// each partition holds one segment. Boxes end up wider and shorter than a
// classic squarified treemap would produce, which is deliberate: it's
// horizontal text that needs to fit inside them. This produces a real,
// non-overlapping, gap-free 2D treemap (unlike RenderStackedBar's
// single-row proportional bar), though it does not attempt the squarified
// algorithm's optimal aspect ratios.
//
// Each box's full label (name, plus value if ShowValues) is printed inside
// it only when it fits without truncation -- a box too small to describe
// is not given a truncated fragment of its name. Instead it gets a
// superscript index number (¹, ², ³, ...), centered in the interior or, if
// even that has no room, overlaid on the top border. Every numbered box is
// listed, ranked biggest value first (independent of box layout order), on
// one legend row appended below the grid: "¹name ²name ...", each entry
// with its text colored to match that box's identifying color when
// opts.ANSI is set (never a filled background -- see treemapLegendCode),
// hard-truncated (whole entries only, never a mid-name fragment) with a
// trailing "…" the moment the next entry would exceed the grid's width --
// later entries may simply not fit and are omitted.
//
// Zero, negative, NaN, and infinite values are treated as zero and receive
// no box (zero-area). Empty or all-zero input renders a blank grid.
//
// CALLER RESPONSIBILITY -- terminal width overflow corrupts the whole
// grid, not just one line: every returned row is exactly opts.Width
// display columns (verify with measure.StringWidth on a row with its ANSI
// codes stripped, never len() or utf8.RuneCountInString on a row that may
// carry ANSI). This function has no way to know the real terminal width,
// so it trusts opts.Width completely -- if the caller sets it (or lets it
// default) wider than the terminal actually printing to, the terminal's
// own auto-wrap will wrap each overflowing row onto the next physical
// line, and this function's row-per-terminal-line assumption breaks: every
// subsequent row then lands one line lower than intended, and boxes appear
// to scramble and jump between "rows" that no longer correspond to actual
// grid rows. A single-row renderer like RenderStackedBar mostly reflows
// harmlessly if this happens; a multi-row grid like this one cannot, since
// each row's screen position is load-bearing. This is a well-documented
// trap for coding agents building terminal UIs specifically (see
// examples/treemap/main.go's resolveDimensions and its test
// TestClampDimensionsNeverExceedsTerminal for a concrete repro and fix):
// always clamp opts.Width/opts.Height to the ACTUAL current terminal size
// (e.g. via golang.org/x/term.GetSize on the real output fd) before
// calling this, never trust a caller-supplied or previously-cached size
// blindly, and re-resolve it on every redraw in a watch/live-update loop
// since the terminal can be resized between frames. Go cannot enforce this
// at compile time (it is pure runtime string content), so there is no
// substitute for actually doing the clamp -- this comment exists because
// skipping it fails silently and confusingly rather than with a clear
// error.
func RenderTreemap(segments []TreemapSegment, opts TreemapOptions) []string {
	width := optionWidth(opts.Width)
	height := optionWidth(opts.Height)

	grid := make([][]rune, height)
	for y := range grid {
		grid[y] = make([]rune, width)
		for x := range grid[y] {
			grid[y][x] = ' '
		}
	}
	style := make([][]string, height)
	for y := range style {
		style[y] = make([]string, width)
	}

	values := make([]float64, len(segments))
	for i, s := range segments {
		values[i] = s.Value
	}
	rects := layoutTreemap(values, treemapRect{0, 0, width, height})

	glyphs := opts.Glyphs
	if len(glyphs) == 0 {
		glyphs = defaultTreemapGlyphs
	}
	glyphs = oneCellGlyphs(glyphs, defaultTreemapGlyphs)

	markers := make([]string, len(segments))
	var legend []treemapLegendEntry
	if !opts.NoLabels {
		labels := make([]string, len(segments))
		var needsNumber []int
		for i, seg := range segments {
			r := rects[i]
			if r.W <= 0 || r.H <= 0 {
				continue
			}
			border := treemapHasBorder(r, opts)
			_, _, innerW, innerH := treemapInner(r, border)
			label := treemapFullLabel(seg, opts)
			labels[i] = label
			if innerH >= 1 && measure.StringWidth(label) <= innerW {
				continue // fits fully; drawTreemapBox will print it as-is.
			}
			needsNumber = append(needsNumber, i)
		}
		// Number by value descending, independent of box layout order, so
		// marker 1 is always the biggest consumer that needed a marker --
		// not just whichever unlabelable box happens to come first.
		sort.SliceStable(needsNumber, func(a, b int) bool {
			return segments[needsNumber[a]].Value > segments[needsNumber[b]].Value
		})
		for rank, i := range needsNumber {
			marker := superscriptNumber(rank + 1)
			markers[i] = marker
			legend = append(legend, treemapLegendEntry{marker: marker, label: labels[i], code: treemapLegendCode(i, opts)})
		}
	}

	for i, r := range rects {
		if r.W <= 0 || r.H <= 0 {
			continue
		}
		drawTreemapBox(grid, style, r, segments[i], glyphs[i%len(glyphs)], i, opts, markers[i])
	}

	rows := make([]string, height)
	for y := range grid {
		if !opts.ANSI {
			rows[y] = string(grid[y])
			continue
		}
		var b strings.Builder
		var open string
		for x, ch := range grid[y] {
			code := style[y][x]
			if code != open {
				if open != "" {
					b.WriteString("\x1b[0m")
				}
				if code != "" {
					b.WriteString("\x1b[" + code + "m")
				}
				open = code
			}
			b.WriteRune(ch)
		}
		if open != "" {
			b.WriteString("\x1b[0m")
		}
		rows[y] = b.String()
	}
	if len(legend) > 0 {
		rows = append(rows, buildTreemapLegend(legend, width))
	}
	return rows
}

// treemapLegendEntry is one numbered box awaiting a spot on the legend row.
type treemapLegendEntry struct {
	marker string
	label  string
	code   string // ANSI SGR code to style this entry with, or "" for none
}

// buildTreemapLegend renders entries as "marker label marker label ..." on
// one row padded/truncated to exactly width columns, each entry styled in
// its code (when set), with a trailing "…" the moment the next whole entry
// would not fit -- entries are never cut mid-name.
func buildTreemapLegend(entries []treemapLegendEntry, width int) string {
	if width < 1 {
		return ""
	}

	// Pass 1: how many whole entries fit, by plain (unstyled) width --
	// ANSI color codes never affect displayed width.
	plainToken := func(i int) string {
		token := entries[i].marker + entries[i].label
		if i > 0 {
			token = " " + token
		}
		return token
	}
	used, shown := 0, 0
	for shown < len(entries) {
		w := measure.StringWidth(plainToken(shown))
		if used+w > width {
			break
		}
		used += w
		shown++
	}
	truncated := shown < len(entries)
	if truncated {
		// Reserve room for the trailing ellipsis by dropping whole fitted
		// entries from the end -- never a partial, mid-name fragment --
		// until there is space for it.
		for shown > 0 && used+1 > width {
			shown--
			used -= measure.StringWidth(plainToken(shown))
		}
	}

	// Pass 2: re-render exactly those `shown` entries, now styled.
	var b strings.Builder
	for i := 0; i < shown; i++ {
		token := plainToken(i)
		if entries[i].code != "" {
			b.WriteString("\x1b[" + entries[i].code + "m" + token + "\x1b[0m")
		} else {
			b.WriteString(token)
		}
	}
	if truncated {
		b.WriteString("…")
		used++
	}
	return b.String() + strings.Repeat(" ", width-used)
}

// treemapLegendCode returns the ANSI code the legend uses to color segment
// index's entry: its identifying color as *text* color, not a filled
// background -- a background swatch there reads as another treemap box
// rather than a legend, which is exactly the confusion this avoids. When
// BackgroundANSI is configured (the usual case: it is what visually
// distinguishes segments on the grid), that color is converted to its
// foreground equivalent; otherwise it falls back to opts.ForegroundANSI
// as-is. Empty when opts.ANSI is false or neither slice is configured.
func treemapLegendCode(index int, opts TreemapOptions) string {
	if !opts.ANSI {
		return ""
	}
	if len(opts.BackgroundANSI) > 0 {
		return treemapBackgroundToForeground(opts.BackgroundANSI[index%len(opts.BackgroundANSI)])
	}
	if len(opts.ForegroundANSI) > 0 {
		return opts.ForegroundANSI[index%len(opts.ForegroundANSI)]
	}
	return ""
}

// treemapBackgroundToForeground converts a single-code ANSI SGR background
// color (40-47 normal, 100-107 bright) to its foreground equivalent (30-37,
// 90-97). Codes it doesn't recognize -- already a foreground code, a
// compound "1;41"-style code, non-numeric -- are returned unchanged.
func treemapBackgroundToForeground(code string) string {
	n, err := strconv.Atoi(code)
	if err != nil {
		return code
	}
	if (n >= 40 && n <= 47) || (n >= 100 && n <= 107) {
		return strconv.Itoa(n - 10)
	}
	return code
}

// treemapSegmentCode returns segment index's combined background;foreground
// SGR code, cycling through opts.BackgroundANSI/ForegroundANSI. Empty when
// opts.ANSI is false or neither slice is configured.
func treemapSegmentCode(index int, opts TreemapOptions) string {
	if !opts.ANSI {
		return ""
	}
	var fg, bg string
	if len(opts.ForegroundANSI) > 0 {
		fg = opts.ForegroundANSI[index%len(opts.ForegroundANSI)]
	}
	if len(opts.BackgroundANSI) > 0 {
		bg = opts.BackgroundANSI[index%len(opts.BackgroundANSI)]
	}
	code := bg
	if fg != "" {
		if code != "" {
			code += ";" + fg
		} else {
			code = fg
		}
	}
	return code
}

// treemapFullLabel builds a segment's undecorated label: its name, plus a
// formatted value when opts.ShowValues is set.
func treemapFullLabel(seg TreemapSegment, opts TreemapOptions) string {
	label := seg.Name
	if opts.ShowValues {
		precision := opts.ValuePrecision
		if precision < 0 {
			precision = 0
		}
		label += fmt.Sprintf(" %.*f%s", precision, seg.Value, opts.ValueSuffix)
	}
	return label
}

// treemapHasBorder reports whether r draws a TreemapThemeClassic box-drawing
// border: never under TreemapThemeBlocks (which is always full-bleed --
// its boundary softening happens per-cell, not via a reserved border cell),
// otherwise the existing size/NoBorder rule.
func treemapHasBorder(r treemapRect, opts TreemapOptions) bool {
	return opts.Theme != TreemapThemeBlocks && !opts.NoBorder && r.W >= 3 && r.H >= 2
}

// treemapUsesBlocks reports whether r should render via drawTreemapBlockBox:
// TreemapThemeBlocks needs opts.ANSI and a background palette to have any
// color to blend -- without either, it falls back to the same solid-glyph
// full-bleed fill TreemapThemeClassic uses when NoBorder is set (the
// non-border branch of drawTreemapBox, shared code).
func treemapUsesBlocks(opts TreemapOptions) bool {
	return opts.Theme == TreemapThemeBlocks && opts.ANSI && len(opts.BackgroundANSI) > 0
}

// treemapInner returns a box's label-safe interior: the full rect when
// unbordered, or the rect inset by one cell on each side when bordered.
func treemapInner(r treemapRect, border bool) (x, y, w, h int) {
	if border {
		return r.X + 1, r.Y + 1, r.W - 2, r.H - 2
	}
	return r.X, r.Y, r.W, r.H
}

// superscriptDigits maps ASCII digits to their Unicode superscript form.
var superscriptDigits = [10]rune{'⁰', '¹', '²', '³', '⁴', '⁵', '⁶', '⁷', '⁸', '⁹'}

// superscriptNumber renders n (n >= 0) using superscript digits.
func superscriptNumber(n int) string {
	digits := strconv.Itoa(n)
	runes := make([]rune, len(digits))
	for i := 0; i < len(digits); i++ {
		runes[i] = superscriptDigits[digits[i]-'0']
	}
	return string(runes)
}

// drawTreemapBox paints one segment's box into grid/style: TreemapThemeBlocks
// boundary-blended full-bleed fill, a TreemapThemeClassic bordered box (when
// it fits and !opts.NoBorder), or a solid fill, plus a label or numbered
// marker when it fits and !opts.NoLabels. marker is "" when the segment's
// full label fits directly (computed by RenderTreemap's pre-pass); otherwise
// it is the superscript index to place instead.
func drawTreemapBox(grid [][]rune, style [][]string, r treemapRect, seg TreemapSegment, fill rune, index int, opts TreemapOptions, marker string) {
	code := treemapSegmentCode(index, opts)

	border := treemapHasBorder(r, opts)
	setCell := func(x, y int, ch rune) {
		grid[y][x] = ch
		style[y][x] = code
	}

	switch {
	case treemapUsesBlocks(opts):
		drawTreemapBlockBox(grid, style, r, index, opts)
	case !border:
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				if opts.ANSI {
					setCell(x, y, ' ')
				} else {
					setCell(x, y, fill)
				}
			}
		}
	default:
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				top, bottom := y == r.Y, y == r.Y+r.H-1
				left, right := x == r.X, x == r.X+r.W-1
				switch {
				case top && left:
					setCell(x, y, '┌')
				case top && right:
					setCell(x, y, '┐')
				case bottom && left:
					setCell(x, y, '└')
				case bottom && right:
					setCell(x, y, '┘')
				case top, bottom:
					setCell(x, y, '─')
				case left, right:
					setCell(x, y, '│')
				default:
					if opts.ANSI {
						setCell(x, y, ' ')
					} else {
						setCell(x, y, fill)
					}
				}
			}
		}
	}

	if opts.NoLabels {
		return
	}
	innerX, innerY, innerW, innerH := treemapInner(r, border)

	if marker != "" {
		drawTreemapMarker(grid, style, r, innerX, innerY, innerW, innerH, border, marker, code)
		return
	}

	if innerW < 1 || innerH < 1 {
		return
	}
	label := treemapFullLabel(seg, opts)
	labelRunes := truncateLabel(label, innerW)
	if len(labelRunes) == 0 {
		return
	}
	labelY := innerY + innerH/2
	for i, ch := range labelRunes {
		grid[labelY][innerX+i] = ch
		style[labelY][innerX+i] = code
	}
}

// drawTreemapBlockBox paints one box under TreemapThemeBlocks: a purely
// local rule based only on this box's own rectangle, no neighbor lookup at
// all -- every box always renders its own rightmost column and bottommost
// row this way, whether or not another box (or just the outer grid edge)
// happens to sit past it:
//
//   - bottom-right corner cell (both the box's rightmost column AND its
//     bottommost row): '▘' -- only the upper-left quadrant carries this
//     box's color; the other three quadrants are left unstyled.
//   - rightmost column (not also the bottom row): '▌' -- only the left
//     half carries this box's color; the right half is left unstyled.
//   - bottommost row (not also the rightmost column): '▀' -- only the top
//     half carries this box's color; the bottom half is left unstyled.
//   - every other cell: a plain filled cell (this box's solid color, same
//     as the classic NoBorder fill) -- this is the overwhelming majority
//     of the box, including its own top and left edges, so there is
//     plenty of room to write a label without touching a glyph cell.
//
// "Left unstyled" is the point: no background SGR code is written for
// that portion at all, so it is whatever the terminal already shows there
// -- the ambient terminal background reads straight through the glyph's
// empty half/quadrant, rather than this function guessing at and
// reproducing some other color. This also means adjoining boxes never
// fight over how to render their shared seam: only the box on the
// right/bottom side of a split ever draws anything into it.
func drawTreemapBlockBox(grid [][]rune, style [][]string, r treemapRect, index int, opts TreemapOptions) {
	ownBG := opts.BackgroundANSI[index%len(opts.BackgroundANSI)]
	ownFG := treemapBackgroundToForeground(ownBG)

	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			right := x == r.X+r.W-1
			bottom := y == r.Y+r.H-1
			glyph, code := ' ', ownBG
			switch {
			case right && bottom:
				glyph, code = '▘', ownFG
			case right:
				glyph, code = '▌', ownFG
			case bottom:
				glyph, code = '▀', ownFG
			}
			grid[y][x] = glyph
			style[y][x] = code
		}
	}
}

// drawTreemapMarker places a numbered fallback marker for a box whose full
// label didn't fit: centered in the interior when there's room, else
// overlaid on the top border (only possible when the box has one), else
// omitted (still listed on the legend row regardless).
func drawTreemapMarker(grid [][]rune, style [][]string, r treemapRect, innerX, innerY, innerW, innerH int, border bool, marker, code string) {
	markerRunes := []rune(marker)
	markerW := len(markerRunes) // superscript digits are always one cell wide

	place := func(x, y int) {
		for i, ch := range markerRunes {
			grid[y][x+i] = ch
			style[y][x+i] = code
		}
	}

	if innerW >= markerW && innerH >= 1 {
		x := innerX + (innerW-markerW)/2
		y := innerY + innerH/2
		place(x, y)
		return
	}
	if border && r.W >= markerW {
		x := r.X + (r.W-markerW)/2
		if x < r.X {
			x = r.X
		}
		place(x, r.Y)
	}
}

// truncateLabel trims label to at most maxWidth display columns, appending
// an ellipsis when it was cut. Returns nil for maxWidth < 1.
func truncateLabel(label string, maxWidth int) []rune {
	if maxWidth < 1 {
		return nil
	}
	runes := []rune(label)
	if measure.StringWidth(label) <= maxWidth {
		return runes
	}
	if maxWidth == 1 {
		return []rune{'…'}
	}
	kept := runes[:0]
	width := 1 // reserve for the ellipsis
	for _, r := range runes {
		w := measure.RuneWidth(r)
		if width+w > maxWidth {
			break
		}
		kept = append(kept, r)
		width += w
	}
	return append(append([]rune(nil), kept...), '…')
}

// layoutTreemap recursively partitions values into non-overlapping,
// integer-cell rectangles tiling rect with no gaps. See RenderTreemap for
// the algorithm. The returned slice has one rect per value, in the same
// order as values; segments with a non-positive value get a zero-size
// rect.
func layoutTreemap(values []float64, rect treemapRect) []treemapRect {
	rects := make([]treemapRect, len(values))
	idx := make([]int, 0, len(values))
	for i, v := range values {
		if isFinite(v) && v > 0 {
			idx = append(idx, i)
		}
	}
	var recurse func(idx []int, r treemapRect)
	recurse = func(idx []int, r treemapRect) {
		if len(idx) == 0 {
			return
		}
		if len(idx) == 1 {
			rects[idx[0]] = r
			return
		}
		sorted := append([]int(nil), idx...)
		sort.SliceStable(sorted, func(a, b int) bool { return values[sorted[a]] > values[sorted[b]] })

		total := 0.0
		for _, i := range sorted {
			total += values[i]
		}
		splitAt, bestDiff, cum := 1, math.Inf(1), 0.0
		for k := 1; k < len(sorted); k++ {
			cum += values[sorted[k-1]]
			if diff := math.Abs(cum - total/2); diff < bestDiff {
				bestDiff, splitAt = diff, k
			}
		}
		left, right := sorted[:splitAt], sorted[splitAt:]
		var leftSum, rightSum float64
		for _, i := range left {
			leftSum += values[i]
		}
		for _, i := range right {
			rightSum += values[i]
		}

		// Splitting along width (left/right) keeps the full height but
		// shrinks each child's width, making them relatively TALLER --
		// more portrait. Splitting along height (top/bottom) keeps the
		// full width and shrinks height, making children relatively
		// WIDER -- more landscape, which fits a horizontal text label
		// far more easily. So default to a height-split, and only cut
		// width once the rect is already comfortably landscape (see
		// treemapLandscapeBias) or height has no room left to give.
		splitWidth := r.H <= 1 || (r.W > 1 && float64(r.W) >= float64(r.H)*treemapLandscapeBias)
		if splitWidth {
			sizes := allocateCells([]float64{leftSum, rightSum}, r.W)
			recurse(left, treemapRect{r.X, r.Y, sizes[0], r.H})
			recurse(right, treemapRect{r.X + sizes[0], r.Y, sizes[1], r.H})
		} else {
			sizes := allocateCells([]float64{leftSum, rightSum}, r.H)
			recurse(left, treemapRect{r.X, r.Y, r.W, sizes[0]})
			recurse(right, treemapRect{r.X, r.Y + sizes[0], r.W, sizes[1]})
		}
	}
	recurse(idx, rect)
	return rects
}

// treemapLandscapeBias sets how much wider than tall (in character cells)
// a rectangle must already be before layoutTreemap is willing to split it
// along its width instead of its height. Terminal character cells are
// themselves roughly twice as tall as they are wide, and it's horizontal
// text that needs to fit inside a box, so this is deliberately set above
// 1 (which would just aim for a square in cell *count*, already a
// landscape rectangle on screen) to favor genuinely wide, short boxes.
const treemapLandscapeBias = 3.0
