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
	// TreemapThemeNumbered renders every box full-bleed -- no border cell
	// is reserved, so labels get the box's entire area -- and softens its
	// own TOP row and RIGHT column with a seven-eighths block glyph ('▇'
	// top, '▉' right, U+2587/U+2589) instead of the hard one-cell-wide
	// border line TreemapThemeClassic draws. Unlike a border, these glyphs
	// don't fill their cell solid: only the majority (7/8) that faces this
	// box's own interior carries its color -- a thin sliver gets no
	// background SGR at all, so the ambient terminal background shows
	// straight through it. This is purely local to each box's own
	// rectangle (its topmost row and rightmost column), not dependent on
	// what segment (if any) sits past that edge -- so two adjoining boxes
	// never both try to render their shared seam.
	//
	// Every box with positive area also gets a superscript corner marker
	// (see superscriptNumber), unconditionally, stamped into its own
	// top-right corner cell -- ranked by value descending across ALL
	// boxes, not just the ones whose full label doesn't fit. A box that
	// also has room for its full label shows both: the label centered as
	// usual, plus its corner number. The legend row below the grid still
	// lists only the boxes that needed a number in place of their label
	// (the same criterion TreemapThemeClassic uses) -- a labeled box's
	// corner number isn't explained there, since the box's own visible
	// name already identifies it. The corner marker's own cell uses the
	// same foreground-only style as the rest of its row/column: it floats
	// over the ambient terminal background, not a filled swatch.
	//
	// Requires opts.ANSI and opts.BackgroundANSI (there is no color to
	// paint the glyph/marker with otherwise); without either,
	// TreemapThemeNumbered falls back to the identical solid-glyph
	// full-bleed fill TreemapThemeClassic uses when NoBorder is set, and
	// to TreemapThemeClassic's only-unlabelled-boxes marker/legend
	// behavior.
	TreemapThemeNumbered
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

	markers := make([]string, len(segments))       // interior fallback marker (Classic only)
	cornerMarkers := make([]string, len(segments)) // corner marker (Numbered only)
	hideLabel := make([]bool, len(segments))       // Numbered: label doesn't fit, corner marker alone identifies it
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
		if treemapUsesNumbered(opts) {
			// Every positive-area box gets a corner marker, ranked by value
			// across ALL of them -- not just the ones needing a number in
			// place of their label. The legend still only lists the latter
			// (a labeled box's own name already identifies it).
			var allIdx []int
			for i, r := range rects {
				if r.W > 0 && r.H > 0 {
					allIdx = append(allIdx, i)
				}
			}
			sort.SliceStable(allIdx, func(a, b int) bool {
				return segments[allIdx[a]].Value > segments[allIdx[b]].Value
			})
			needsLegend := make(map[int]bool, len(needsNumber))
			for _, i := range needsNumber {
				needsLegend[i] = true
			}
			for rank, i := range allIdx {
				marker := superscriptNumber(rank + 1)
				cornerMarkers[i] = marker
				if needsLegend[i] {
					legend = append(legend, treemapLegendEntry{marker: marker, label: labels[i], code: treemapLegendCode(i, opts)})
					// This box's full label doesn't fit. Its corner marker
					// already identifies it (see the doc comment on
					// TreemapThemeNumbered) -- drawTreemapBox must not also
					// attempt to place the label text, or it would show as
					// an ugly, hard-to-read fragment ("fir…", "v………"). "Full
					// label or a number, never a fragment" is the same rule
					// TreemapThemeClassic enforces via its own marker
					// fallback.
					hideLabel[i] = true
				}
			}
		} else {
			// Number by value descending, independent of box layout order,
			// so marker 1 is always the biggest consumer that needed a
			// marker -- not just whichever unlabelable box comes first.
			sort.SliceStable(needsNumber, func(a, b int) bool {
				return segments[needsNumber[a]].Value > segments[needsNumber[b]].Value
			})
			for rank, i := range needsNumber {
				marker := superscriptNumber(rank + 1)
				markers[i] = marker
				legend = append(legend, treemapLegendEntry{marker: marker, label: labels[i], code: treemapLegendCode(i, opts)})
			}
		}
	}

	for i, r := range rects {
		if r.W <= 0 || r.H <= 0 {
			continue
		}
		drawTreemapBox(grid, style, r, segments[i], glyphs[i%len(glyphs)], i, opts, markers[i], cornerMarkers[i], hideLabel[i])
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
// border: never under TreemapThemeNumbered (full-bleed -- its boundary
// softening happens per-cell, not via a reserved border cell), otherwise the
// existing size/NoBorder rule.
func treemapHasBorder(r treemapRect, opts TreemapOptions) bool {
	return opts.Theme != TreemapThemeNumbered && !opts.NoBorder && r.W >= 3 && r.H >= 2
}

// treemapUsesNumbered reports whether opts should render via
// drawTreemapThinBox and RenderTreemap's every-box corner-marker pre-pass:
// TreemapThemeNumbered needs opts.ANSI and a background palette to have any
// color to render its edge glyphs and corner markers with -- without
// either, it falls back to the same solid-glyph full-bleed fill
// TreemapThemeClassic uses when NoBorder is set (the non-border branch of
// drawTreemapBox, shared code), and to TreemapThemeClassic's
// only-unlabelled-boxes marker/legend behavior.
func treemapUsesNumbered(opts TreemapOptions) bool {
	return opts.Theme == TreemapThemeNumbered && opts.ANSI && len(opts.BackgroundANSI) > 0
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

// drawTreemapBox paints one segment's box into grid/style: TreemapThemeNumbered's
// boundary-softened full-bleed fill, a TreemapThemeClassic bordered box
// (when it fits and !opts.NoBorder), or a solid fill, plus a label or
// numbered marker when it fits and !opts.NoLabels. marker is "" when the
// segment's full label fits directly (computed by RenderTreemap's
// pre-pass); otherwise it is the superscript index to place instead.
// cornerMarker is only set under TreemapThemeNumbered (every positive-area
// box gets one, regardless of marker) -- see RenderTreemap's pre-pass.
func drawTreemapBox(grid [][]rune, style [][]string, r treemapRect, seg TreemapSegment, fill rune, index int, opts TreemapOptions, marker, cornerMarker string, hideLabel bool) {
	code := treemapSegmentCode(index, opts)

	// textCode is the style for a label/marker character at (x, y): on the
	// edge cells TreemapThemeNumbered softens (its own top row/right
	// column), drawTreemapThinBox deliberately writes NO background there
	// (see its doc comment) -- so stamping the combined bg;fg code over it
	// anyway would show up as a solid colored patch poking out of the
	// otherwise thin, unfilled seam around it. Use the same
	// foreground-only style that function itself uses on that edge
	// instead. Everywhere else (the overwhelming majority of the box, and
	// every cell under TreemapThemeClassic), it's the normal filled code.
	textCode := func(x, y int) string { return code }
	if treemapUsesNumbered(opts) {
		edgeCode := treemapBackgroundToForeground(opts.BackgroundANSI[index%len(opts.BackgroundANSI)])
		textCode = func(x, y int) string {
			if x == r.X+r.W-1 || y == r.Y {
				return edgeCode
			}
			return code
		}
	}

	border := treemapHasBorder(r, opts)
	setCell := func(x, y int, ch rune) {
		grid[y][x] = ch
		style[y][x] = code
	}

	switch {
	case treemapUsesNumbered(opts):
		drawTreemapThinBox(grid, style, r, index, opts, cornerMarker)
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
		drawTreemapMarker(grid, style, r, innerX, innerY, innerW, innerH, border, marker, textCode)
		return
	}

	if hideLabel {
		// TreemapThemeNumbered: this box's full label doesn't fit, and its
		// corner marker (already placed) is its only identification --
		// never show a truncated fragment instead.
		return
	}

	if innerW < 1 || innerH < 1 {
		return
	}
	label := treemapFullLabel(seg, opts)
	if measure.StringWidth(label) > innerW {
		// Doesn't fit: full label or nothing, never a truncated fragment.
		return
	}
	labelRunes := []rune(label)
	labelY := innerY + innerH/2
	for i, ch := range labelRunes {
		x := innerX + i
		grid[labelY][x] = ch
		style[labelY][x] = textCode(x, labelY)
	}
}

// drawTreemapThinBox paints one box under TreemapThemeNumbered: a purely
// local rule based only on this box's own rectangle, no neighbor lookup at
// all -- every box always renders its own topmost row and rightmost column
// this way, whether or not another box (or just the outer grid edge)
// happens to sit past it:
//
//   - top-right corner cell: reserved for cornerMarker (right-aligned
//     within the top row, ending at this cell) when it's set and fits
//     within the box's own width; otherwise '▝' (quadrant upper-right) as
//     a fallback glyph.
//   - top row (not also the right column): '▇' (seven-eighths block,
//     U+2587) -- only the bottom 7/8 (facing this box's own interior)
//     carries its color; a thin sliver at the very top is left unstyled.
//   - right column (not also the top row): '▉' (U+2589) -- only the left
//     7/8 carries this box's color; a thin sliver at the right is left
//     unstyled.
//   - every other cell: a plain filled cell (this box's solid color, same
//     as the classic NoBorder fill) -- this is the overwhelming majority
//     of the box, including its own bottom and left edges, so there is
//     plenty of room to write a label without touching a glyph cell.
//
// "Left unstyled" is the point: no background SGR code is written for
// that thin sliver at all, so it is whatever the terminal already shows
// there -- the ambient terminal background reads straight through the
// glyph's empty portion, rather than this function guessing at and
// reproducing some other color. This also means adjoining boxes never
// fight over how to render their shared seam: only the box on the
// top/right side of a split ever draws anything into it. cornerMarker's
// own cells use the same foreground-only style as the rest of its row.
func drawTreemapThinBox(grid [][]rune, style [][]string, r treemapRect, index int, opts TreemapOptions, cornerMarker string) {
	ownBG := opts.BackgroundANSI[index%len(opts.BackgroundANSI)]
	ownFG := treemapBackgroundToForeground(ownBG)

	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			top := y == r.Y
			right := x == r.X+r.W-1
			glyph, code := ' ', ownBG
			switch {
			case top && right:
				glyph, code = '▝', ownFG
			case top:
				glyph, code = '▇', ownFG
			case right:
				glyph, code = '▉', ownFG
			}
			grid[y][x] = glyph
			style[y][x] = code
		}
	}

	if cornerMarker == "" {
		return
	}
	markerRunes := []rune(cornerMarker)
	markerW := len(markerRunes) // superscript digits are always one cell wide
	if markerW > r.W {
		return // doesn't fit even the box's full width; leave the corner glyph
	}
	x := r.X + r.W - markerW
	for i, ch := range markerRunes {
		grid[r.Y][x+i] = ch
		style[r.Y][x+i] = ownFG
	}
}

// drawTreemapMarker places a numbered fallback marker for a box whose full
// label didn't fit: centered in the interior when there's room, else
// overlaid on the top border (only possible when the box has one), else
// omitted (still listed on the legend row regardless). textCode resolves
// the style for each character's exact cell (see drawTreemapBox's textCode
// for why this can't be a single flat code under TreemapThemeNumbered).
func drawTreemapMarker(grid [][]rune, style [][]string, r treemapRect, innerX, innerY, innerW, innerH int, border bool, marker string, textCode func(x, y int) string) {
	markerRunes := []rune(marker)
	markerW := len(markerRunes) // superscript digits are always one cell wide

	place := func(x, y int) {
		for i, ch := range markerRunes {
			grid[y][x+i] = ch
			style[y][x+i] = textCode(x+i, y)
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
