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

// ── M1: Evidence generation for DrawBorder and DrawBox ──────────────────────────

// TestGenerateM1BoxStyleGallery creates a visual frame showing all four box styles.
// This is the M1-boxstyle-gallery.ansi evidence frame for ticket 037.
// Only generates evidence when LOOM_EVIDENCE=1 environment variable is set.
func TestGenerateM1BoxStyleGallery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	const cols, rows = 80, 30
	canvas := loom.NewCanvas(cols, rows)

	// Header
	canvas.Write(2, 0, "M1: Canvas.DrawBorder and DrawBox Gallery", loom.Style{Bold: true})
	canvas.Write(2, 1, "All four styles with and without titles", loom.Reset)

	// Sharp style (row 3)
	y := 3
	canvas.Write(2, y, "Sharp Style:", loom.Reset)
	canvas.DrawBorder(loom.Rect{X: 15, Y: y, W: 12, H: 4}, loom.BoxBorderStyleSharp, loom.Reset)
	canvas.DrawBox(loom.Rect{X: 30, Y: y, W: 15, H: 4}, loom.BoxBorderStyleSharp, "Sharp", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 48, Y: y, W: 15, H: 4}, loom.BoxBorderStyleSharp, "Long Sharp Title", loom.Reset)

	// Rounded style (row 8)
	y = 8
	canvas.Write(2, y, "Rounded Style:", loom.Reset)
	canvas.DrawBorder(loom.Rect{X: 15, Y: y, W: 12, H: 4}, loom.BoxBorderStyleRounded, loom.Reset)
	canvas.DrawBox(loom.Rect{X: 30, Y: y, W: 15, H: 4}, loom.BoxBorderStyleRounded, "Rounded", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 48, Y: y, W: 15, H: 4}, loom.BoxBorderStyleRounded, "Long Rounded Title", loom.Reset)

	// Double style (row 13)
	y = 13
	canvas.Write(2, y, "Double Style:", loom.Reset)
	canvas.DrawBorder(loom.Rect{X: 15, Y: y, W: 12, H: 4}, loom.BoxBorderStyleDouble, loom.Reset)
	canvas.DrawBox(loom.Rect{X: 30, Y: y, W: 15, H: 4}, loom.BoxBorderStyleDouble, "Double", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 48, Y: y, W: 15, H: 4}, loom.BoxBorderStyleDouble, "Long Double Title", loom.Reset)

	// ASCII style (row 18)
	y = 18
	canvas.Write(2, y, "ASCII Style:", loom.Reset)
	canvas.DrawBorder(loom.Rect{X: 15, Y: y, W: 12, H: 4}, loom.BoxBorderStyleASCII, loom.Reset)
	canvas.DrawBox(loom.Rect{X: 30, Y: y, W: 15, H: 4}, loom.BoxBorderStyleASCII, "ASCII", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 48, Y: y, W: 15, H: 4}, loom.BoxBorderStyleASCII, "Long ASCII Title", loom.Reset)

	// Render and save
	var buf bytes.Buffer
	err := loom.RenderTo(&buf, &simpleCanvasWidget{canvas: canvas}, cols, rows)
	if err != nil {
		t.Fatalf("RenderTo failed: %v", err)
	}

	// Ensure output directory exists
	outDir := filepath.Join("docs", "progress", "037")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("mkdir %s failed: %v", outDir, err)
	}

	outPath := filepath.Join(outDir, "M1-boxstyle-gallery.ansi")
	err = os.WriteFile(outPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("WriteFile %s failed: %v", outPath, err)
	}

	t.Logf("Evidence saved to %s (%d bytes)", outPath, buf.Len())
}

// TestGenerateM1TitlesEvidence creates a visual frame showing title handling.
// Tests CJK characters, umlaut, emoji, long titles, and tiny widths.
// This is the M1-titles.ansi evidence frame for ticket 037.
func TestGenerateM1TitlesEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	const cols, rows = 80, 25
	canvas := loom.NewCanvas(cols, rows)

	// Header
	canvas.Write(2, 0, "M1: Title Handling - Display Width, Truncation, Multi-byte", loom.Style{Bold: true})

	y := 2

	// CJK title
	canvas.Write(2, y, "CJK (Chinese):", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 20, Y: y, W: 20, H: 3}, loom.BoxBorderStyleSharp, "你好世界", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 42, Y: y, W: 20, H: 3}, loom.BoxBorderStyleRounded, "日本語", loom.Reset)
	y += 4

	// Umlaut and diacritics
	canvas.Write(2, y, "Umlaut/Diacritic:", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 20, Y: y, W: 20, H: 3}, loom.BoxBorderStyleSharp, "Näïvëté", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 42, Y: y, W: 20, H: 3}, loom.BoxBorderStyleRounded, "Café", loom.Reset)
	y += 4

	// Emoji
	canvas.Write(2, y, "Emoji:", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 20, Y: y, W: 20, H: 3}, loom.BoxBorderStyleDouble, "✓ Success", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 42, Y: y, W: 20, H: 3}, loom.BoxBorderStyleDouble, "⚠ Warning", loom.Reset)
	y += 4

	// Long title that needs truncation
	canvas.Write(2, y, "Long title (truncated):", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 2, Y: y, W: 35, H: 3}, loom.BoxBorderStyleASCII,
		"This is a very long title that should be truncated with ellipsis", loom.Reset)
	y += 4

	// Tiny widths (width < 6)
	canvas.Write(2, y, "Tiny widths:", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 20, Y: y, W: 4, H: 3}, loom.BoxBorderStyleSharp, "AB", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 25, Y: y, W: 5, H: 3}, loom.BoxBorderStyleRounded, "Test", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 31, Y: y, W: 6, H: 3}, loom.BoxBorderStyleDouble, "Title", loom.Reset)

	// Render and save
	var buf bytes.Buffer
	err := loom.RenderTo(&buf, &simpleCanvasWidget{canvas: canvas}, cols, rows)
	if err != nil {
		t.Fatalf("RenderTo failed: %v", err)
	}

	// Ensure output directory exists
	outDir := filepath.Join("docs", "progress", "037")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("mkdir %s failed: %v", outDir, err)
	}

	outPath := filepath.Join(outDir, "M1-titles.ansi")
	err = os.WriteFile(outPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("WriteFile %s failed: %v", outPath, err)
	}

	t.Logf("Evidence saved to %s (%d bytes)", outPath, buf.Len())
}
