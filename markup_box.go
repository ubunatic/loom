// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
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

// markupBoxTag reports whether line is a box block tag: "open", "close" or "".
func markupBoxTag(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "[/box]" {
		return "close"
	}
	if rest, ok := strings.CutPrefix(trimmed, "[box"); ok && rest != "" && (rest[0] == ' ' || rest[0] == ']') {
		return "open"
	}
	return ""
}

// parseMarkupBoxTag parses an opening box tag line into a box with an empty
// content document.
func parseMarkupBoxTag(line string, lineNo int) (*markupBox, error) {
	box := &markupBox{pad: 1, lineNo: lineNo, doc: &RichDocument{}}
	style := BoxBorderStyleSharp
	start := strings.Index(line, "[box") + len("[box")
	col := func(offset int) int { return utf8.RuneCountInString(line[:offset]) + 1 }
	fail := func(offset int, format string, args ...any) error {
		return &MarkupError{lineNo, col(offset), fmt.Sprintf(format, args...)}
	}
	i := start
	for {
		for i < len(line) && line[i] == ' ' {
			i++
		}
		if i >= len(line) {
			return nil, fail(len(line), "unterminated box tag")
		}
		if line[i] == ']' {
			if strings.TrimSpace(line[i+1:]) != "" {
				return nil, fail(i+1, "box tag must stand alone on its line")
			}
			break
		}
		keyStart := i
		eq := strings.IndexByte(line[i:], '=')
		if eq < 0 {
			return nil, fail(keyStart, "box option needs key=value")
		}
		key := strings.ToLower(line[i : i+eq])
		i += eq + 1
		var value string
		if i < len(line) && line[i] == '"' {
			end := strings.IndexByte(line[i+1:], '"')
			if end < 0 {
				return nil, fail(i, "unterminated quoted value")
			}
			value = line[i+1 : i+1+end]
			i += end + 2
		} else {
			end := strings.IndexAny(line[i:], " ]")
			if end < 0 {
				end = len(line) - i
			}
			value = line[i : i+end]
			i += end
		}
		switch key {
		case "style":
			name := strings.ToLower(value)
			if name == "normal" {
				name = "sharp"
			}
			parsed, err := ParseBoxBorderStyle(name)
			if err != nil {
				return nil, fail(keyStart, "unknown box style %q", value)
			}
			style = parsed
		case "title":
			title := &RichDocument{Lines: []RichLine{{}}}
			if err := parseMarkupLine(title, value, lineNo); err != nil {
				return nil, fail(keyStart, "title: %v", err.(*MarkupError).Msg)
			}
			box.title = title.Lines[0].Spans
		case "border":
			_, parsed, err := applyMarkupTag(Style{}, value)
			if err != nil {
				return nil, fail(keyStart, "border: %v", err)
			}
			box.borderStyle = parsed
		case "pad":
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 {
				return nil, fail(keyStart, "pad must be a number >= 0, got %q", value)
			}
			box.pad = n
		default:
			return nil, fail(keyStart, "unknown box option %q", key)
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
