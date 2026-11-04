// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"bytes"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

// htmlGoldenDoc is the exact document for htmlGoldenCanvas. Row 0 covers bold,
// italic, underline, a background run, a wide rune (one span, no continuation
// output) and HTML-special characters. Row 1 covers invert (FG/BG swapped with
// the resolved defaults), a 256-colour foreground and default blank cells.
const htmlGoldenDoc = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>loom screenshot</title>
<style>
body{margin:0;background:#1e1e1e;}
pre{margin:0;padding:8px;font-family:monospace;line-height:1.2;color:#d4d4d4;background:#1e1e1e;}
</style>
</head>
<body>
<pre><span style="color:#ff0000;background:#1e1e1e;font-weight:bold">A</span><span style="color:#d4d4d4;background:#1e1e1e;font-style:italic">i</span><span style="color:#d4d4d4;background:#1e1e1e;text-decoration:underline">u</span><span style="color:#d4d4d4;background:#0000ff">B </span><span style="color:#d4d4d4;background:#1e1e1e">漢&lt;&amp;&gt;&quot;</span>
<span style="color:#1e1e1e;background:#d4d4d4">x</span><span style="color:#ff0000;background:#1e1e1e">#</span><span style="color:#d4d4d4;background:#1e1e1e">         </span></pre>
</body>
</html>
`

// htmlGoldenCanvas builds the canvas described by htmlGoldenDoc.
func htmlGoldenCanvas() *loom.Canvas {
	c := loom.NewCanvas(11, 2)
	c.Set(0, 0, loom.Cell{Text: "A", Style: loom.Style{FG: loom.ColorRGB(255, 0, 0), Bold: true}})
	c.Set(1, 0, loom.Cell{Text: "i", Style: loom.Style{Italic: true}})
	c.Set(2, 0, loom.Cell{Text: "u", Style: loom.Style{Underline: true}})
	c.Set(3, 0, loom.Cell{Text: "B", Style: loom.Style{BG: loom.ColorRGB(0, 0, 255)}})
	c.Set(4, 0, loom.Cell{Text: " ", Style: loom.Style{BG: loom.ColorRGB(0, 0, 255)}})
	c.Write(5, 0, "漢", loom.Style{})
	c.Write(7, 0, "<&>\"", loom.Style{})
	c.Set(0, 1, loom.Cell{Text: "x", Style: loom.Style{Invert: true}})
	c.Set(1, 1, loom.Cell{Text: "#", Style: loom.Style{FG: loom.ColorIndex(196)}})
	return c
}

func TestWriteHTMLGolden(t *testing.T) {
	c := htmlGoldenCanvas()
	var buf bytes.Buffer
	if err := loom.WriteHTML(&buf, c, loom.HTMLOptions{}); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
	if got := buf.String(); got != htmlGoldenDoc {
		t.Errorf("WriteHTML output mismatch\n got: %q\nwant: %q", got, htmlGoldenDoc)
	}
}

func TestWriteHTMLTitle(t *testing.T) {
	var buf bytes.Buffer
	if err := loom.WriteHTML(&buf, loom.NewCanvas(1, 1), loom.HTMLOptions{Title: "a<b & \"c\""}); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
	if !strings.Contains(buf.String(), "<title>a&lt;b &amp; &quot;c&quot;</title>") {
		t.Errorf("title not escaped: %q", buf.String())
	}
}

func TestWriteHTMLRowsMatchesRenderHTML(t *testing.T) {
	view := loom.NewView([]string{"héllo <&> world", "second \"line\""})
	var direct, viaRows bytes.Buffer
	if err := loom.RenderHTML(&direct, view, 20, 3, loom.HTMLOptions{Title: "t"}); err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}
	if err := loom.WriteHTMLRows(&viaRows, loom.Render(view, 20, 3), loom.HTMLOptions{Title: "t"}); err != nil {
		t.Fatalf("WriteHTMLRows: %v", err)
	}
	if direct.String() != viaRows.String() {
		t.Errorf("WriteHTMLRows(Render) != RenderHTML\nrender: %q\nrows:   %q", direct.String(), viaRows.String())
	}
}

func TestWriteHTMLRowsResolvesColours(t *testing.T) {
	rows := []string{
		"\x1b[31mR\x1b[0m\x1b[38;5;21mB\x1b[0m\x1b[48;2;1;2;3mK\x1b[0m",
		"",         // an empty row still becomes a line of the grid
		"\x1b[2Kx", // non-SGR escapes are dropped, not printed
	}
	var buf bytes.Buffer
	if err := loom.WriteHTMLRows(&buf, rows, loom.HTMLOptions{}); err != nil {
		t.Fatalf("WriteHTMLRows: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		`<span style="color:#cd0000;background:#1e1e1e">R</span>`, // 16-colour red, index 1
		`<span style="color:#0000ff;background:#1e1e1e">B</span>`, // 256-colour index 21
		`<span style="color:#d4d4d4;background:#010203">K</span>`, // 24-bit background
		"\n\n", // blank row kept
		`<span style="color:#d4d4d4;background:#1e1e1e">x</span>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b") || strings.Contains(out, "[2K") {
		t.Errorf("non-SGR escape leaked into output: %q", out)
	}
}

func TestWriteHTMLNoExternalAssetsOrScripts(t *testing.T) {
	var buf bytes.Buffer
	if err := loom.WriteHTML(&buf, htmlGoldenCanvas(), loom.HTMLOptions{}); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
	for _, banned := range []string{"http", "src=", "<script", "<link", "url("} {
		if strings.Contains(buf.String(), banned) {
			t.Errorf("output must be self-contained, found %q", banned)
		}
	}
}

func TestRenderHTMLRejectsBadArguments(t *testing.T) {
	view := loom.NewView([]string{"x"})
	if err := loom.RenderHTML(nil, view, 5, 1, loom.HTMLOptions{}); err == nil {
		t.Error("nil writer should be an error")
	}
	if err := loom.RenderHTML(&bytes.Buffer{}, nil, 5, 1, loom.HTMLOptions{}); err == nil {
		t.Error("nil widget should be an error")
	}
	if err := loom.RenderHTML(&bytes.Buffer{}, view, 0, 1, loom.HTMLOptions{}); err == nil {
		t.Error("zero columns should be an error")
	}
	if err := loom.WriteHTML(&bytes.Buffer{}, nil, loom.HTMLOptions{}); err == nil {
		t.Error("nil canvas should be an error")
	}
	if err := loom.WriteHTMLRows(nil, []string{"x"}, loom.HTMLOptions{}); err == nil {
		t.Error("nil writer should be an error")
	}
}
