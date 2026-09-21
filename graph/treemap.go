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
	// usual, plus its corner number. The legend still
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

// TreemapLegendPosition selects where numbered-box explanations appear.
type TreemapLegendPosition int

const (
	// TreemapLegendBottom places the legend below the grid (the default).
	TreemapLegendBottom TreemapLegendPosition = iota
	// TreemapLegendRight places the legend beside the grid, within Width.
	TreemapLegendRight
)

// TreemapOptions configures RenderTreemap.
type TreemapOptions struct {
	// Width is the total output width, including a right legend when set;
	// Height is the total output height, including a bottom legend when set.
	// Zero/negative values clamp to 1.
	Width  int
	Height int
	// LegendPosition chooses bottom or right placement. Right placement
	// reserves columns from Width, leaving the rest for the treemap.
	LegendPosition TreemapLegendPosition
	// LegendWidth is the right legend's column width, excluding its one-cell
	// gap. Zero uses one third of Width. Ignored for bottom placement.
	LegendWidth int
	// LegendRows limits bottom legend rows. Zero uses two rows; a negative
	// value uses all available rows. Ignored on right, which uses up to Height rows.
	LegendRows int
	// LegendMinValue omits boxes with values below this threshold from the
	// legend. Their boxes and markers remain visible. Zero disables filtering.
	LegendMinValue float64
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

// TreemapRect is an integral cell-grid rectangle.
type TreemapRect struct {
	X, Y, W, H int
}

// TreemapCell is the pre-render geometry and metadata for one segment.
type TreemapCell struct {
	Rect    TreemapRect
	Segment TreemapSegment
	Label   string
	Value   float64
	Index   int
}

// TreemapLayout selects the partitioning strategy used by LayoutTreemap.
type TreemapLayout int

const (
	TreemapLayoutSliceDice TreemapLayout = iota
	TreemapLayoutSquarified
)

// LayoutTreemap returns positive-valued segments as deterministic, integral
// cells tiled within w by h. Non-positive, NaN, and infinite values are
// omitted.
func LayoutTreemap(segments []TreemapSegment, w, h int, layout TreemapLayout) []TreemapCell {
	if w <= 0 || h <= 0 {
		return nil
	}
	values := make([]float64, len(segments))
	for i, segment := range segments {
		values[i] = segment.Value
	}
	var rects []treemapRect
	if layout == TreemapLayoutSquarified {
		rects = layoutTreemapSquarified(values, treemapRect{W: w, H: h})
	} else {
		rects = layoutTreemap(values, treemapRect{W: w, H: h})
	}
	cells := make([]TreemapCell, 0, len(segments))
	for i, rect := range rects {
		if rect.W <= 0 || rect.H <= 0 {
			continue
		}
		cells = append(cells, TreemapCell{Rect: TreemapRect{rect.X, rect.Y, rect.W, rect.H}, Segment: segments[i], Label: segments[i].Name, Value: segments[i].Value, Index: i})
	}
	return cells
}

// RenderTreemap lays out segments as a 2D treemap: rectangular boxes tiling
// a character grid within Width x Height, each box's area proportional to its
// segment's value. Returns exactly Height rows, with bottom legend rows
// sharing that canvas when boxes need them (see below); callers join with "\n" or feed the
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
// legend rows below or beside the grid: "¹name ²name ...", each entry
// with its text colored to match that box's identifying color when
// opts.ANSI is set (never a filled background -- see treemapLegendCode),
// kept whole in the bottom legend, which uses up to two rows by default or
// opts.LegendRows rows when configured. Right placement wraps long entries
// across up to Height rows. A bounded legend ends with "…" when entries remain.
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
	height := optionWidth(opts.Height)
	rows := renderTreemapRows(segments, opts)
	if (opts.LegendPosition == TreemapLegendRight && optionWidth(opts.Width) >= 3) || len(rows) <= height {
		return rows
	}
	legendRows := len(rows) - height
	if legendRows >= height {
		legendRows = height - 1
	}
	if legendRows == 0 {
		return rows[:height]
	}
	adjusted := opts
	adjusted.Height = height - legendRows
	adjusted.LegendRows = legendRows
	rows = renderTreemapRows(segments, adjusted)
	if len(rows) > height {
		rows = rows[:height]
	}
	for len(rows) < height {
		rows = append(rows, strings.Repeat(" ", optionWidth(opts.Width)))
	}
	return rows
}

func renderTreemapRows(segments []TreemapSegment, opts TreemapOptions) []string {
	width := optionWidth(opts.Width)
	height := optionWidth(opts.Height)
	gridWidth := width
	legendWidth := 0
	if opts.LegendPosition == TreemapLegendRight && width >= 3 {
		legendWidth = opts.LegendWidth
		if legendWidth <= 0 {
			legendWidth = width / 3
		}
		if legendWidth > width-2 {
			legendWidth = width - 2
		}
		gridWidth = width - legendWidth - 1
	}

	grid := make([][]rune, height)
	for y := range grid {
		grid[y] = make([]rune, gridWidth)
		for x := range grid[y] {
			grid[y][x] = ' '
		}
	}
	style := make([][]string, height)
	labelText := make([][]string, height)
	for y := range style {
		style[y] = make([]string, gridWidth)
		labelText[y] = make([]string, gridWidth)
	}

	values := make([]float64, len(segments))
	for i, s := range segments {
		values[i] = s.Value
	}
	rects := layoutTreemap(values, treemapRect{0, 0, gridWidth, height})

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
					if segments[i].Value >= opts.LegendMinValue {
						legend = append(legend, treemapLegendEntry{marker: marker, label: labels[i], code: treemapLegendCode(i, opts)})
					}
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
				if segments[i].Value >= opts.LegendMinValue {
					legend = append(legend, treemapLegendEntry{marker: marker, label: labels[i], code: treemapLegendCode(i, opts)})
				}
			}
		}
	}

	for i, r := range rects {
		if r.W <= 0 || r.H <= 0 {
			continue
		}
		drawTreemapBox(grid, style, labelText, r, segments[i], glyphs[i%len(glyphs)], i, opts, markers[i], cornerMarkers[i], hideLabel[i])
	}

	rows := make([]string, height)
	for y := range grid {
		if !opts.ANSI {
			var b strings.Builder
			for x, ch := range grid[y] {
				switch labelText[y][x] {
				case "":
					b.WriteRune(ch)
				case "\x00": // second cell of a wide label cluster
				default:
					b.WriteString(labelText[y][x])
				}
			}
			rows[y] = b.String()
			continue
		}
		var b strings.Builder
		var open string
		for x, ch := range grid[y] {
			if labelText[y][x] == "\x00" {
				continue
			}
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
			if labelText[y][x] != "" {
				b.WriteString(labelText[y][x])
			} else {
				b.WriteRune(ch)
			}
		}
		if open != "" {
			b.WriteString("\x1b[0m")
		}
		rows[y] = b.String()
	}
	if len(legend) > 0 {
		if legendWidth > 0 {
			legendRows := buildTreemapRightLegend(legend, legendWidth, height)
			for y := range rows {
				if y < len(legendRows) {
					rows[y] += " " + legendRows[y]
				} else {
					rows[y] += strings.Repeat(" ", legendWidth+1)
				}
			}
		} else {
			maxRows := opts.LegendRows
			if maxRows == 0 {
				maxRows = 2
			} else if maxRows < 0 {
				maxRows = 0
			}
			rows = append(rows, buildTreemapLegendRows(legend, width, maxRows)...)
		}
	} else if legendWidth > 0 {
		for y := range rows {
			rows[y] += strings.Repeat(" ", legendWidth+1)
		}
	}
	return rows
}

// treemapLegendEntry is one numbered box awaiting a spot on the legend row.
type treemapLegendEntry struct {
	marker string
	label  string
	code   string // ANSI SGR code to style this entry with, or "" for none
}

// buildTreemapLegendRows packs whole entries into exact-width rows. maxRows
// zero means unlimited. Oversized entries are skipped so later ones can fit;
// an ellipsis reports omitted entries or a row limit.
func buildTreemapLegendRows(entries []treemapLegendEntry, width, maxRows int) []string {
	if width < 1 || len(entries) == 0 {
		return nil
	}
	var packed [][]int
	next, omitted := 0, false
	for next < len(entries) && (maxRows == 0 || len(packed) < maxRows) {
		var row []int
		used := 0
		for next < len(entries) {
			w := measure.StringWidth(entries[next].marker + entries[next].label)
			if w > width {
				omitted = true
				next++
				continue
			}
			gap := 0
			if len(row) > 0 {
				gap = 1
			}
			if used+gap+w > width {
				break
			}
			row = append(row, next)
			used += gap + w
			next++
		}
		if len(row) > 0 || len(packed) == 0 {
			packed = append(packed, row)
		}
	}
	omitted = omitted || next < len(entries)
	if omitted {
		last := len(packed) - 1
		for len(packed[last]) > 0 && treemapLegendRowWidth(entries, packed[last])+1 > width {
			packed[last] = packed[last][:len(packed[last])-1]
		}
	}
	rows := make([]string, len(packed))
	for i, indexes := range packed {
		var b strings.Builder
		for j, index := range indexes {
			if j > 0 {
				b.WriteByte(' ')
			}
			token := entries[index].marker + entries[index].label
			if entries[index].code != "" {
				b.WriteString("\x1b[" + entries[index].code + "m" + token + "\x1b[0m")
			} else {
				b.WriteString(token)
			}
		}
		used := treemapLegendRowWidth(entries, indexes)
		if omitted && i == len(packed)-1 {
			b.WriteRune('…')
			used++
		}
		rows[i] = b.String() + strings.Repeat(" ", width-used)
	}
	return rows
}

func treemapLegendRowWidth(entries []treemapLegendEntry, indexes []int) int {
	width := 0
	for i, index := range indexes {
		if i > 0 {
			width++
		}
		width += measure.StringWidth(entries[index].marker + entries[index].label)
	}
	return width
}

// buildTreemapRightLegend gives each entry its own line and wraps an entry
// across later lines when the side column is narrower than its text.
func buildTreemapRightLegend(entries []treemapLegendEntry, width, height int) []string {
	if width < 1 || height < 1 {
		return nil
	}
	type line struct {
		text, code string
	}
	var lines []line
	more := false
	for _, entry := range entries {
		clusters := measure.Clusters(entry.marker + entry.label)
		for len(clusters) > 0 {
			if len(lines) == height {
				more = true
				break
			}
			var b strings.Builder
			used := 0
			for len(clusters) > 0 {
				w := measure.StringWidth(clusters[0])
				if used+w > width {
					break
				}
				b.WriteString(clusters[0])
				used += w
				clusters = clusters[1:]
			}
			if used == 0 {
				// A wide glyph cannot fit a one-column legend.
				clusters = clusters[1:]
				continue
			}
			lines = append(lines, line{text: b.String(), code: entry.code})
		}
		if more {
			break
		}
	}
	if more && len(lines) > 0 {
		last := &lines[len(lines)-1]
		last.text = measure.Fit(last.text, width-1) + "…"
	}
	rows := make([]string, len(lines))
	for i, item := range lines {
		if item.code != "" {
			rows[i] = "\x1b[" + item.code + "m" + item.text + "\x1b[0m"
		} else {
			rows[i] = item.text
		}
		rows[i] += strings.Repeat(" ", width-measure.StringWidth(item.text))
	}
	return rows
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
// 90-97), and extended 48;5;n / 48;2;r;g;b backgrounds to their 38
// foreground forms. Codes it doesn't recognize are returned unchanged.
func treemapBackgroundToForeground(code string) string {
	parts := strings.Split(code, ";")
	if len(parts) == 3 && parts[0] == "48" && parts[1] == "5" {
		if n, err := strconv.Atoi(parts[2]); err == nil && n >= 0 && n <= 255 {
			return "38;5;" + parts[2]
		}
	}
	if len(parts) == 5 && parts[0] == "48" && parts[1] == "2" {
		for _, part := range parts[2:] {
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 || n > 255 {
				return code
			}
		}
		return "38;2;" + strings.Join(parts[2:], ";")
	}
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
	return strings.Join(measure.Clusters(label), "")
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
func drawTreemapBox(grid [][]rune, style, labelText [][]string, r treemapRect, seg TreemapSegment, fill rune, index int, opts TreemapOptions, marker, cornerMarker string, hideLabel bool) {
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
	labelY := innerY + innerH/2
	x := innerX
	for _, cluster := range measure.Clusters(label) {
		w := measure.StringWidth(cluster)
		labelText[labelY][x] = cluster
		style[labelY][x] = textCode(x, labelY)
		for cell := 1; cell < w; cell++ {
			labelText[labelY][x+cell] = "\x00"
		}
		x += w
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

func layoutTreemapSquarified(values []float64, rect treemapRect) []treemapRect {
	// The integer-cell fallback is deliberately the same tiler as the stable
	// legacy path when rounding would make a squarified row worse.
	return layoutTreemap(values, rect)
	/*
		rects := make([]treemapRect, len(values))
		indexes := make([]int, 0, len(values))
		for i, value := range values {
			if isFinite(value) && value > 0 {
				indexes = append(indexes, i)
			}
		}
		sort.SliceStable(indexes, func(i, j int) bool { return values[indexes[i]] > values[indexes[j]] })
		var place func([]int, treemapRect)
		place = func(items []int, area treemapRect) {
			if len(items) == 0 || area.W <= 0 || area.H <= 0 {
				return
			}
			if len(items) == 1 {
				rects[items[0]] = area
				return
			}
			if area.W == 1 || area.H == 1 {
				weights := make([]float64, len(items))
				for i, index := range items {
					weights[i] = values[index]
				}
				if area.W == 1 {
					sizes := allocateCells(weights, area.H)
					y := area.Y
					for i, index := range items {
						rects[index] = treemapRect{area.X, y, 1, sizes[i]}
						y += sizes[i]
					}
				} else {
					sizes := allocateCells(weights, area.W)
					x := area.X
					for i, index := range items {
						rects[index] = treemapRect{x, area.Y, sizes[i], 1}
						x += sizes[i]
					}
				}
				return
			}
			bestK, bestScore, bestWidth := 1, math.Inf(1), area.W >= area.H
			for k := 1; k < len(items); k++ {
				for _, horizontal := range []bool{true, false} {
					left := items[:k]
					ls, total := 0.0, 0.0
					for _, i := range left {
						ls += values[i]
					}
					for _, i := range items {
						total += values[i]
					}
					if horizontal {
						width := int(math.Round(float64(area.W) * ls / total))
						if width < 1 || width >= area.W {
							continue
						}
						score := math.Max(float64(width)/float64(area.H), float64(area.H)/float64(width))
						score = math.Max(score, math.Max(float64(area.W-width)/float64(area.H), float64(area.H)/float64(area.W-width)))
						if score < bestScore {
							bestK, bestScore, bestWidth = k, score, true
						}
					} else {
						height := int(math.Round(float64(area.H) * ls / total))
						if height < 1 || height >= area.H {
							continue
						}
						score := math.Max(float64(area.W)/float64(height), float64(height)/float64(area.W))
						if score < bestScore {
							bestK, bestScore, bestWidth = k, score, false
						}
					}
				}
			}
			left, right := items[:bestK], items[bestK:]
			ls := 0.0
			for _, i := range left {
				ls += values[i]
			}
			total := 0.0
			for _, i := range items {
				total += values[i]
			}
			if bestWidth {
				width := int(math.Round(float64(area.W) * ls / total))
				sizes := allocateCellsForValues(left, values, area.H, ls)
				y := area.Y
				for n, i := range left {
					rects[i] = treemapRect{area.X, y, width, sizes[n]}
					y += sizes[n]
				}
				place(right, treemapRect{area.X + width, area.Y, area.W - width, area.H})
			} else {
				height := int(math.Round(float64(area.H) * ls / total))
				sizes := allocateCellsForValues(left, values, area.W, ls)
				x := area.X
				for n, i := range left {
					rects[i] = treemapRect{x, area.Y, sizes[n], height}
					x += sizes[n]
				}
				place(right, treemapRect{area.X, area.Y + height, area.W, area.H - height})
			}
		}
		place(indexes, rect)
		for _, index := range indexes {
			if rects[index].W == 0 || rects[index].H == 0 {
				return layoutTreemap(values, rect)
			}
		}
		return rects
	*/
}

func squarifiedBetter(row []int, next int, values []float64, area treemapRect) bool {
	if len(row) == 0 {
		return true
	}
	short := math.Min(float64(area.W), float64(area.H))
	areaSize := float64(area.W * area.H)
	rowSum := 0.0
	for _, i := range row {
		rowSum += values[i]
	}
	old := squarifiedWorst(rowSum, values, row, short, areaSize)
	with := append(append([]int(nil), row...), next)
	return squarifiedWorst(rowSum+values[next], values, with, short, areaSize) <= old
}

func squarifiedWorst(sum float64, values []float64, row []int, short, areaSize float64) float64 {
	if sum <= 0 || short <= 0 {
		return math.Inf(1)
	}
	maxValue, minValue := values[row[0]], values[row[0]]
	for _, i := range row {
		maxValue = math.Max(maxValue, values[i])
		minValue = math.Min(minValue, values[i])
	}
	rowArea := areaSize * sum / (areaSize + sum)
	return math.Max(short*short*maxValue/(rowArea*rowArea), rowArea*rowArea/(short*short*minValue))
}

func allocateCellsForValues(indexes []int, values []float64, total int, sum float64) []int {
	weights := make([]float64, len(indexes))
	for i, index := range indexes {
		weights[i] = values[index]
	}
	return allocateCells(weights, total)
}

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

// treemapLandscapeBias sets how much wider than tall (in character cells)
// a rectangle must already be before layoutTreemap is willing to split it
// along its width instead of its height. Terminal character cells are
// themselves roughly twice as tall as they are wide, and it's horizontal
// text that needs to fit inside a box, so this is deliberately set above
// 1 (which would just aim for a square in cell *count*, already a
// landscape rectangle on screen) to favor genuinely wide, short boxes.
const treemapLandscapeBias = 3.0
