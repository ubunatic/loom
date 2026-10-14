// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strconv"
	"strings"
)

// .tui box blocks (issue 309):
//
//	[box style=rounded title="Status" border="cyan" pad=1]
//	text lines, may hold inline markup and nested boxes
//	[/box]
//
// Box tags stand alone on their line. ParseMarkup draws the box into plain
// RichLines with the glyphs of spec/box.yaml, so the document holds no block
// structure and ToMarkup writes the drawn border characters.

type markupBox struct {
	border      BoxBorder
	borderStyle Style
	title       []RichSpan
	pad         int
	lineNo      int
	doc         *RichDocument
}

// parseMarkupBoxTag parses an opening box tag line into a box with an empty
// content document.
func parseMarkupBoxTag(line string, lineNo int) (*markupBox, error) {
	box := &markupBox{pad: 1, lineNo: lineNo, doc: &RichDocument{}}
	style := BoxBorderStyleSharp
	options, err := parseMarkupBlockOptions(line, "box", lineNo)
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
				return nil, fail("unknown box style %q", opt.value)
			}
			style = parsed
		case "title":
			title := &RichDocument{Lines: []RichLine{{}}}
			if err := parseMarkupLine(title, opt.value, lineNo); err != nil {
				return nil, fail("title: %v", err.(*MarkupError).Msg)
			}
			box.title = title.Lines[0].Spans
		case "border":
			_, parsed, err := applyMarkupTag(Style{}, opt.value)
			if err != nil {
				return nil, fail("border: %v", err)
			}
			box.borderStyle = parsed
		case "pad":
			n, err := strconv.Atoi(opt.value)
			if err != nil || n < 0 {
				return nil, fail("pad must be a number >= 0, got %q", opt.value)
			}
			box.pad = n
		default:
			return nil, fail("unknown box option %q", opt.key)
		}
	}
	border, err := getBoxBorderGlyphs(style)
	if err != nil {
		return nil, err
	}
	box.border = border
	return box, nil
}

// render draws the box around its content lines. The inner width fits the
// widest line plus padding and the title with one space and one border
// glyph on each side.
func (b *markupBox) render() []RichLine {
	lines := b.doc.Lines
	widths := make([]int, len(lines))
	inner := 2 * b.pad
	for i, line := range lines {
		for _, span := range line.Spans {
			widths[i] += StringWidth(span.Text)
		}
		inner = max(inner, widths[i]+2*b.pad)
	}
	titleWidth := 0
	for _, span := range b.title {
		titleWidth += StringWidth(span.Text)
	}
	if len(b.title) > 0 {
		inner = max(inner, titleWidth+4)
	}
	edge := func(s string) RichSpan { return RichSpan{Text: s, Style: b.borderStyle} }
	top := RichLine{Spans: []RichSpan{edge(b.border.TopLeft)}}
	if len(b.title) > 0 {
		top.Spans = append(top.Spans, edge(b.border.Horizontal+" "))
		top.Spans = append(top.Spans, b.title...)
		top.Spans = append(top.Spans, edge(" "+strings.Repeat(b.border.Horizontal, inner-titleWidth-3)))
	} else {
		top.Spans = append(top.Spans, edge(strings.Repeat(b.border.Horizontal, inner)))
	}
	top.Spans = append(top.Spans, edge(b.border.TopRight))
	out := []RichLine{top}
	for i, line := range lines {
		spans := []RichSpan{edge(b.border.Vertical), {Text: strings.Repeat(" ", b.pad)}}
		spans = append(spans, line.Spans...)
		spans = append(spans, RichSpan{Text: strings.Repeat(" ", inner-b.pad-widths[i])}, edge(b.border.Vertical))
		out = append(out, RichLine{Spans: mergeRichSpans(spans)})
	}
	bottom := b.border.BottomLeft + strings.Repeat(b.border.Horizontal, inner) + b.border.BottomRight
	out = append(out, RichLine{Spans: []RichSpan{edge(bottom)}})
	for i := range out {
		out[i].Spans = mergeRichSpans(out[i].Spans)
	}
	return out
}
