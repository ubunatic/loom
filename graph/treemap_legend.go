// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"strings"

	"ubunatic.com/loom/measure"
)

// TreemapLegendPosition selects where numbered-box explanations appear.
type TreemapLegendPosition int

const (
	// TreemapLegendBottom places the legend below the grid (the default).
	TreemapLegendBottom TreemapLegendPosition = iota
	// TreemapLegendRight places the legend beside the grid, within Width.
	TreemapLegendRight
)

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
