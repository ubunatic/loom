// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// simpleCanvasWidget wraps a canvas for rendering.
type simpleCanvasWidget struct {
	canvas *loom.Canvas
}

func (w *simpleCanvasWidget) Draw(c *loom.Canvas, r loom.Rect) {
	for y := 0; y < w.canvas.Rows(); y++ {
		for x := 0; x < w.canvas.Cols(); x++ {
			cell := w.canvas.Get(x, y)
			c.Set(x, y, cell)
		}
	}
}

func (w *simpleCanvasWidget) HandleKey(e loom.KeyEvent) bool   { return false }
func (w *simpleCanvasWidget) HandleMouse(e loom.MouseEvent) bool { return false }

// TestGenerateM1Evidence generates a visual frame demonstrating ParseANSI capabilities.
// This is the M1 evidence frame for ticket 034.
// Only generates evidence when LOOM_EVIDENCE=1 environment variable is set.
func TestGenerateM1Evidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	const cols, rows = 80, 20
	canvas := loom.NewCanvas(cols, rows)

	// Render a header
	canvas.Write(2, 0, "M1: ParseANSI Evidence", loom.Style{Bold: true})

	// Row 2: 16-color foreground
	y := 2
	canvas.Write(2, y, "16-color FG:", loom.Reset)
	x := 15
	colors16 := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	for _, idx := range colors16 {
		cells := loom.ParseANSI("\x1b[3" + string(rune('0'+idx%10)) + "m█")
		for _, cell := range cells {
			if !cell.Continuation {
				canvas.Set(x, y, cell)
				x += 2
			}
		}
	}

	// Row 3: 256-color samples
	y = 3
	canvas.Write(2, y, "256-color FG:", loom.Reset)
	x = 15
	colors256 := []uint8{16, 52, 88, 124, 160, 196, 226}
	for _, idx := range colors256 {
		str := "\x1b[38;5;" + string(rune('0'+idx/100)) + string(rune('0'+(idx/10)%10)) + string(rune('0'+idx%10)) + "m█"
		cells := loom.ParseANSI(str)
		for _, cell := range cells {
			if !cell.Continuation {
				canvas.Set(x, y, cell)
				x += 2
			}
		}
	}

	// Row 4: 24-bit RGB samples
	y = 4
	canvas.Write(2, y, "24-bit RGB FG:", loom.Reset)
	x = 15
	// Red, Green, Blue gradients
	canvas.Write(x, y, "R", loom.Style{FG: loom.ColorRGB(255, 0, 0)})
	x += 2
	canvas.Write(x, y, "G", loom.Style{FG: loom.ColorRGB(0, 255, 0)})
	x += 2
	canvas.Write(x, y, "B", loom.Style{FG: loom.ColorRGB(0, 0, 255)})
	x += 2
	canvas.Write(x, y, "M", loom.Style{FG: loom.ColorRGB(255, 0, 255)})
	x += 2
	canvas.Write(x, y, "C", loom.Style{FG: loom.ColorRGB(0, 255, 255)})
	x += 2
	canvas.Write(x, y, "Y", loom.Style{FG: loom.ColorRGB(255, 255, 0)})

	// Row 5: Attributes
	y = 5
	canvas.Write(2, y, "Attributes:", loom.Reset)
	x = 15

	cells := loom.ParseANSI("\x1b[1mBold")
	for _, cell := range cells {
		if !cell.Continuation {
			canvas.Set(x, y, cell)
			x++
		}
	}
	x += 2

	cells = loom.ParseANSI("\x1b[2mDim")
	for _, cell := range cells {
		if !cell.Continuation {
			canvas.Set(x, y, cell)
			x++
		}
	}
	x += 2

	cells = loom.ParseANSI("\x1b[4mUnderline")
	for _, cell := range cells {
		if !cell.Continuation {
			canvas.Set(x, y, cell)
			x++
		}
	}

	// Row 6: Background colors
	y = 6
	canvas.Write(2, y, "BG colors:", loom.Reset)
	x = 15
	cells = loom.ParseANSI("\x1b[41m \x1b[0m")
	canvas.Set(x, y, cells[0])
	x += 2
	cells = loom.ParseANSI("\x1b[42m \x1b[0m")
	canvas.Set(x, y, cells[0])
	x += 2
	cells = loom.ParseANSI("\x1b[44m \x1b[0m")
	canvas.Set(x, y, cells[0])
	x += 2

	// Row 7: Combined (bold + color + background)
	y = 7
	canvas.Write(2, y, "Combined:", loom.Reset)
	x = 15
	cells = loom.ParseANSI("\x1b[1;38;2;255;165;0;48;5;234m Styled ")
	for _, cell := range cells {
		if !cell.Continuation {
			canvas.Set(x, y, cell)
			x++
		}
	}

	// Row 8: Multi-color sequence (color changes)
	y = 8
	canvas.Write(2, y, "Multi-color:", loom.Reset)
	x = 15
	input := "\x1b[31mRED\x1b[0m " +
		"\x1b[32mGREEN\x1b[0m " +
		"\x1b[34mBLUE\x1b[0m"
	cells = loom.ParseANSI(input)
	for _, cell := range cells {
		if !cell.Continuation {
			canvas.Set(x, y, cell)
			x++
		}
	}

	// Row 9: Wide character (emoji)
	y = 9
	canvas.Write(2, y, "Wide char:", loom.Reset)
	x = 15
	cells = loom.ParseANSI("\x1b[38;2;255;215;0m👍")
	for _, cell := range cells {
		if !cell.Continuation {
			canvas.Set(x, y, cell)
			x++
		}
	}

	// Row 10: Reset sequence
	y = 10
	canvas.Write(2, y, "Reset test:", loom.Reset)
	x = 15
	cells = loom.ParseANSI("\x1b[1;31m[BOLD RED]\x1b[0m [RESET]")
	for _, cell := range cells {
		if !cell.Continuation {
			canvas.Set(x, y, cell)
			x++
		}
	}

	// Save to file
	var buf bytes.Buffer
	err := loom.RenderTo(&buf, &simpleCanvasWidget{canvas: canvas}, cols, rows)
	if err != nil {
		t.Fatalf("RenderTo failed: %v", err)
	}

	outPath := filepath.Join("docs", "progress", "034", "M1-parse-samples.ansi")
	err = os.WriteFile(outPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("WriteFile %s failed: %v", outPath, err)
	}

	t.Logf("Evidence saved to %s (%d bytes)", outPath, buf.Len())
}
