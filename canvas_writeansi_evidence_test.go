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

// TestGenerateM2Evidence generates a visual frame demonstrating Canvas.WriteANSI capabilities.
// This is the M2 evidence frame for ticket 034.
func TestGenerateM2Evidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}

	const cols, rows = 80, 20
	canvas := loom.NewCanvas(cols, rows)

	// Render a header
	canvas.Write(2, 0, "M2: Canvas.WriteANSI Evidence", loom.Style{Bold: true})

	// Row 2: Basic colored text
	y := 2
	canvas.Write(2, y, "Basic:", loom.Reset)
	canvas.WriteANSI(10, y, "\x1b[31mRED\x1b[0m \x1b[32mGREEN\x1b[0m \x1b[34mBLUE\x1b[0m")

	// Row 3: Bold and colored
	y = 3
	canvas.Write(2, y, "Bold+Color:", loom.Reset)
	canvas.WriteANSI(15, y, "\x1b[1;31mBold Red\x1b[0m \x1b[1;32mBold Green\x1b[0m")

	// Row 4: 256-color palette
	y = 4
	canvas.Write(2, y, "256-color:", loom.Reset)
	canvas.WriteANSI(14, y, "\x1b[38;5;196mColor196\x1b[0m \x1b[38;5;226mColor226\x1b[0m")

	// Row 5: 24-bit RGB
	y = 5
	canvas.Write(2, y, "24-bit RGB:", loom.Reset)
	canvas.WriteANSI(15, y, "\x1b[38;2;255;100;50mOrange\x1b[0m \x1b[38;2;50;100;255mBlue\x1b[0m")

	// Row 6: Background colors
	y = 6
	canvas.Write(2, y, "Backgrounds:", loom.Reset)
	canvas.WriteANSI(16, y, "\x1b[41mRed BG\x1b[0m \x1b[42mGreen BG\x1b[0m \x1b[44mBlue BG\x1b[0m")

	// Row 7: Reset sequences
	y = 7
	canvas.Write(2, y, "Reset:", loom.Reset)
	canvas.WriteANSI(10, y, "\x1b[1;31;41m BOLD RED on RED \x1b[0m \x1b[32m BACK TO DEFAULT \x1b[0m")

	// Row 8: Combined attributes
	y = 8
	canvas.Write(2, y, "Combined:", loom.Reset)
	canvas.WriteANSI(14, y, "\x1b[1;4;38;2;255;165;0;48;5;234m BOLD+UNDERLINE+ORANGE ON DARK \x1b[0m")

	// Row 9: Clipping test - text that runs off edge
	y = 9
	canvas.Write(2, y, "Clipping:", loom.Reset)
	canvas.WriteANSI(70, y, "\x1b[31mLongText")
	// Note: text should clip at column 80

	// Row 10: Wide characters in ANSI
	y = 10
	canvas.Write(2, y, "Wide chars:", loom.Reset)
	canvas.WriteANSI(16, y, "\x1b[38;2;255;215;0m✓\x1b[0m \x1b[38;2;50;200;50m✔\x1b[0m \x1b[38;2;200;50;50m✗\x1b[0m")

	// Row 11: Style reset mid-string
	y = 11
	canvas.Write(2, y, "Reset test:", loom.Reset)
	canvas.WriteANSI(15, y, "\x1b[1;31m[BOLD RED]\x1b[0m middle \x1b[32m[GREEN]\x1b[0m")

	// Row 12: Underline
	y = 12
	canvas.Write(2, y, "Underline:", loom.Reset)
	canvas.WriteANSI(14, y, "Normal \x1b[4m Underlined \x1b[0m Normal again")

	// Row 13: Dim
	y = 13
	canvas.Write(2, y, "Dim:", loom.Reset)
	canvas.WriteANSI(10, y, "\x1b[1mBold\x1b[0m \x1b[2mDim\x1b[0m \x1b[1;2mBold+Dim\x1b[0m")

	// Row 14-15: Complex multi-color sequence
	y = 14
	canvas.Write(2, y, "Multi-color:", loom.Reset)
	input := "\x1b[31mERROR\x1b[0m: \x1b[38;5;196mFailed\x1b[0m - " +
		"\x1b[1;34mDetails\x1b[0m: Check configuration"
	canvas.WriteANSI(15, y, input)

	// Save to file
	var buf bytes.Buffer
	err := loom.RenderTo(&buf, &simpleCanvasWidget{canvas: canvas}, cols, rows)
	if err != nil {
		t.Fatalf("RenderTo failed: %v", err)
	}

	outPath := filepath.Join("docs", "progress", "034", "M2-writeansi.ansi")
	err = os.WriteFile(outPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("WriteFile %s failed: %v", outPath, err)
	}

	t.Logf("Evidence saved to %s (%d bytes)", outPath, buf.Len())
}
