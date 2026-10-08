// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// .tui table blocks (issue 310):
//
//	[table style=rounded border=cyan pad=1 header=bold]
//	| Name | Size |
//	|:-----|-----:|
//	| root | 50G  |
//	[/table]
//
// Rows are pipe-delimited and each cell holds inline markup. A separator row
// directly after the first row makes that row the header and sets the column
// alignment: ":--" left, ":-:" center, "--:" right. Columns fit their widest
// cell in terminal columns and never shrink or wrap. ParseMarkup draws the
// table into plain RichLines with the glyphs of spec/box.yaml.

type markupAlign int

const (
	markupAlignLeft markupAlign = iota
	markupAlignCenter
	markupAlignRight
)

type markupTableCell struct {
	text   string // cell source with \| unescaped
	offset int    // byte offset of text in its line
	spans  []RichSpan
	width  int
}

type markupTable struct {
	border      BoxBorder
	borderStyle Style
	headerStyle Style
	pad         int
	lineNo      int
	header      bool
	aligns      []markupAlign
	rows        [][]markupTableCell
	rowLines    []string
	rowNos      []int
}

// parseMarkupTableTag parses an opening table tag line into an empty table.
func parseMarkupTableTag(line string, lineNo int) (*markupTable, error) {
	table := &markupTable{pad: 1, lineNo: lineNo, headerStyle: Style{Bold: true}}
	style := BoxBorderStyleSharp
	options, err := parseMarkupBlockOptions(line, "table", lineNo)
	if err != nil {
		return nil, err
	}
	for _, opt := range options {
		fail := func(format string, args ...any) error {
			return &MarkupError{lineNo, opt.col, fmt.Sprintf(format, args...)}
		}
		switch opt.key {
		case "style":
			parsed, err := parseMarkupBorderStyle(opt.value)
			if err != nil {
				return nil, fail("unknown table style %q", opt.value)
			}
			style = parsed
		case "border":
			_, parsed, err := applyMarkupTag(Style{}, opt.value)
			if err != nil {
				return nil, fail("border: %v", err)
			}
			table.borderStyle = parsed
		case "header":
			if v := strings.ToLower(strings.TrimSpace(opt.value)); v == "" || v == "none" {
				table.headerStyle = Style{}
				continue
			}
			_, parsed, err := applyMarkupTag(Style{}, opt.value)
			if err != nil {
				return nil, fail("header: %v", err)
			}
			table.headerStyle = parsed
		case "pad":
			n, err := strconv.Atoi(opt.value)
			if err != nil || n < 0 {
				return nil, fail("pad must be a number >= 0, got %q", opt.value)
			}
			table.pad = n
		default:
			return nil, fail("unknown table option %q", opt.key)
		}
	}
	border, err := getBoxBorderGlyphs(style)
	if err != nil {
		return nil, err
	}
	table.border = border
	return table, nil
}

// addLine adds one table row or the header separator. Blank lines are
// skipped.
func (t *markupTable) addLine(line string, lineNo int) error {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}
	lead := strings.Index(line, trimmed)
	col := func(offset int) int { return utf8.RuneCountInString(line[:offset]) + 1 }
	if trimmed[0] != '|' {
		return &MarkupError{lineNo, col(lead), "table row must start with |"}
	}
	cells := splitMarkupTableRow(line, lead+1)
	if aligns, ok := parseMarkupTableSeparator(cells); ok {
		if t.header || len(t.rows) != 1 {
			return &MarkupError{lineNo, col(lead), "table separator row must follow the first row"}
		}
		t.header = true
		t.aligns = aligns
		// The first row is the header: parse it again on the header style.
		return t.parseRow(0, t.headerStyle)
	}
	t.rows = append(t.rows, cells)
	t.rowLines = append(t.rowLines, line)
	t.rowNos = append(t.rowNos, lineNo)
	return t.parseRow(len(t.rows)-1, Style{})
}

// parseRow parses the inline markup of every cell of row i on base.
func (t *markupTable) parseRow(i int, base Style) error {
	line, lineNo := t.rowLines[i], t.rowNos[i]
	for c := range t.rows[i] {
		cell := &t.rows[i][c]
		doc := &RichDocument{Lines: []RichLine{{}}}
		if err := parseMarkupLineStyled(doc, cell.text, lineNo, base); err != nil {
			me := err.(*MarkupError)
			return &MarkupError{lineNo, utf8.RuneCountInString(line[:cell.offset]) + me.Col, me.Msg}
		}
		cell.spans = doc.Lines[0].Spans
		cell.width = 0
		for _, span := range cell.spans {
			cell.width += StringWidth(span.Text)
		}
	}
	return nil
}

// splitMarkupTableRow splits line into trimmed cells, starting after the
// leading pipe at start. "\|" is a literal pipe; other escapes stay for the
// inline parser. A trailing pipe is optional.
func splitMarkupTableRow(line string, start int) []markupTableCell {
	var cells []markupTableCell
	var text strings.Builder
	cellStart := start
	add := func() {
		raw := text.String()
		trimmed := strings.TrimLeft(raw, " \t")
		offset := cellStart + len(raw) - len(trimmed)
		cells = append(cells, markupTableCell{text: strings.TrimRight(trimmed, " \t"), offset: offset})
		text.Reset()
	}
	for i := start; i < len(line); {
		switch {
		case line[i] == '\\' && i+1 < len(line):
			if line[i+1] == '|' {
				text.WriteByte('|')
			} else {
				text.WriteString(line[i : i+2])
			}
			i += 2
		case line[i] == '|':
			add()
			i++
			cellStart = i
		default:
			text.WriteByte(line[i])
			i++
		}
	}
	if strings.TrimSpace(text.String()) != "" {
		add()
	}
	return cells
}

// parseMarkupTableSeparator reports whether cells form a separator row such
// as "|:--|:-:|--:|" and returns the column alignments.
func parseMarkupTableSeparator(cells []markupTableCell) ([]markupAlign, bool) {
	if len(cells) == 0 {
		return nil, false
	}
	aligns := make([]markupAlign, len(cells))
	for i, cell := range cells {
		left := strings.HasPrefix(cell.text, ":")
		right := strings.HasSuffix(cell.text, ":")
		dashes := strings.TrimSuffix(strings.TrimPrefix(cell.text, ":"), ":")
		if dashes == "" || strings.Trim(dashes, "-") != "" {
			return nil, false
		}
		switch {
		case left && right:
			aligns[i] = markupAlignCenter
		case right:
			aligns[i] = markupAlignRight
		}
	}
	return aligns, true
}

// render draws the table. Each column is as wide as its widest cell plus
// padding on both sides; short rows get empty cells.
func (t *markupTable) render() []RichLine {
	cols := len(t.aligns)
	for _, row := range t.rows {
		cols = max(cols, len(row))
	}
	widths := make([]int, cols)
	for _, row := range t.rows {
		for c, cell := range row {
			widths[c] = max(widths[c], cell.width)
		}
	}
	edge := func(s string) RichSpan { return RichSpan{Text: s, Style: t.borderStyle} }
	rule := func(left, mid, right string) RichLine {
		var b strings.Builder
		b.WriteString(left)
		for c, w := range widths {
			if c > 0 {
				b.WriteString(mid)
			}
			b.WriteString(strings.Repeat(t.border.Horizontal, w+2*t.pad))
		}
		b.WriteString(right)
		return RichLine{Spans: []RichSpan{edge(b.String())}}
	}
	out := []RichLine{rule(t.border.TopLeft, t.border.TeeTop, t.border.TopRight)}
	for r, row := range t.rows {
		spans := []RichSpan{edge(t.border.Vertical)}
		for c, w := range widths {
			if c > 0 {
				spans = append(spans, edge(t.border.Vertical))
			}
			var cell markupTableCell
			if c < len(row) {
				cell = row[c]
			}
			gap := w - cell.width
			left := 0
			if c < len(t.aligns) {
				switch t.aligns[c] {
				case markupAlignCenter:
					left = gap / 2
				case markupAlignRight:
					left = gap
				}
			}
			spans = append(spans, RichSpan{Text: strings.Repeat(" ", t.pad+left)})
			spans = append(spans, cell.spans...)
			spans = append(spans, RichSpan{Text: strings.Repeat(" ", gap-left+t.pad)})
		}
		spans = append(spans, edge(t.border.Vertical))
		out = append(out, RichLine{Spans: mergeRichSpans(spans)})
		if r == 0 && t.header {
			out = append(out, rule(t.border.TeeLeft, t.border.Cross, t.border.TeeRight))
		}
	}
	out = append(out, rule(t.border.BottomLeft, t.border.TeeBottom, t.border.BottomRight))
	return out
}
