// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"io"
	"strings"
)

// HTMLOptions controls HTML screenshot output. The zero value is valid.
type HTMLOptions struct {
	Title string // <title>; empty gives "loom screenshot"
}

// Default colours for cells whose FG or BG is the terminal default. A canvas
// has no colours of its own, so these fixed values stand in for "terminal
// default": light text on a dark background.
const (
	htmlDefaultFG = "#d4d4d4"
	htmlDefaultBG = "#1e1e1e"
)

// htmlRunStyle is the resolved appearance of one run of cells. It is comparable,
// so adjacent cells with equal values merge into one span.
type htmlRunStyle struct {
	fg, bg                  string
	bold, italic, underline bool
}

// WriteHTML writes canvas c as one self-contained HTML document: a single <pre>
// monospace grid with inline styles only, no external assets and no scripts.
// Colours are written as resolved hex values; the canvas colour profile does not
// change the output.
func WriteHTML(w io.Writer, c *Canvas, opts HTMLOptions) error {
	if w == nil {
		return fmt.Errorf("loom: html writer required")
	}
	if c == nil {
		return fmt.Errorf("loom: html canvas required")
	}
	return writeHTMLDocument(w, opts, c.cells)
}

// RenderHTML draws root off-screen at cols x rows and writes it with WriteHTML.
// It opens no terminal, so it is safe in headless contexts, tests, and screenshots.
func RenderHTML(w io.Writer, root Widget, cols, rows int, opts HTMLOptions) error {
	if w == nil {
		return fmt.Errorf("loom: html writer required")
	}
	if root == nil {
		return fmt.Errorf("loom: html widget required")
	}
	if cols < 1 || rows < 1 {
		return fmt.Errorf("loom: html size must be positive, got %dx%d", cols, rows)
	}
	c := NewCanvas(cols, rows)
	root.Draw(c, c.Bounds())
	return WriteHTML(w, c, opts)
}

// WriteHTMLRows writes ANSI rows (as from Render, or any SGR-styled lines of any
// count) as one HTML document. SGR sequences set the style; other escape
// sequences are dropped. Rows keep their own lengths.
func WriteHTMLRows(w io.Writer, rows []string, opts HTMLOptions) error {
	if w == nil {
		return fmt.Errorf("loom: html writer required")
	}
	grid := make([][]Cell, len(rows))
	for y, row := range rows {
		grid[y] = ParseANSI(row)
	}
	return writeHTMLDocument(w, opts, grid)
}

// writeHTMLDocument writes the full document around the grid of cells.
func writeHTMLDocument(w io.Writer, opts HTMLOptions, grid [][]Cell) error {
	title := opts.Title
	if title == "" {
		title = "loom screenshot"
	}
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<title>")
	b.WriteString(htmlEscape(title))
	b.WriteString("</title>\n<style>\n")
	fmt.Fprintf(&b, "body{margin:0;background:%s;}\n", htmlDefaultBG)
	fmt.Fprintf(&b, "pre{margin:0;padding:8px;font-family:monospace;line-height:1.2;color:%s;background:%s;}\n", htmlDefaultFG, htmlDefaultBG)
	b.WriteString("</style>\n</head>\n<body>\n<pre>")
	writeHTMLGrid(&b, grid)
	b.WriteString("</pre>\n</body>\n</html>\n")
	_, err := io.WriteString(w, b.String())
	return err
}

// writeHTMLGrid writes each row as a line of spans, merging equal-style runs.
func writeHTMLGrid(b *strings.Builder, grid [][]Cell) {
	for y, row := range grid {
		if y > 0 {
			b.WriteByte('\n')
		}
		var run strings.Builder
		var cur htmlRunStyle
		open := false
		flush := func() {
			if open {
				writeHTMLSpan(b, cur, run.String())
				run.Reset()
			}
		}
		for _, cell := range row {
			if cell.Continuation {
				continue
			}
			style := htmlStyleOf(cell.Style)
			if open && style != cur {
				flush()
			}
			cur, open = style, true
			text := safeCellText(cell.Text)
			if text == "" {
				text = " "
			}
			run.WriteString(text)
		}
		flush()
	}
}

// writeHTMLSpan writes one run as a span with its resolved colours and attributes.
func writeHTMLSpan(b *strings.Builder, s htmlRunStyle, text string) {
	fmt.Fprintf(b, "<span style=\"color:%s;background:%s", s.fg, s.bg)
	if s.bold {
		b.WriteString(";font-weight:bold")
	}
	if s.italic {
		b.WriteString(";font-style:italic")
	}
	if s.underline {
		b.WriteString(";text-decoration:underline")
	}
	b.WriteString("\">")
	b.WriteString(htmlEscape(text))
	b.WriteString("</span>")
}

// htmlStyleOf resolves a cell style to hex colours. Invert swaps the resolved
// FG and BG, as a terminal shows it.
func htmlStyleOf(s Style) htmlRunStyle {
	fg, bg := htmlColorHex(s.FG, htmlDefaultFG), htmlColorHex(s.BG, htmlDefaultBG)
	if s.Invert {
		fg, bg = bg, fg
	}
	return htmlRunStyle{fg: fg, bg: bg, bold: s.Bold, italic: s.Italic, underline: s.Underline}
}

// htmlColorHex returns c as "#rrggbb", or fallback for the terminal default.
func htmlColorHex(c Color, fallback string) string {
	r, g, b, ok := c.RGB()
	if !ok {
		return fallback
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

var htmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;")

// htmlEscape escapes the characters that are significant in HTML text and
// attribute values.
func htmlEscape(s string) string {
	return htmlEscaper.Replace(s)
}
