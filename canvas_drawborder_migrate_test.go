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

// ── M2: Evidence generation for migrated Box.Draw and Popup.Draw ──────────────

// TestGenerateM2BoxEvidence renders a frame demonstrating the migrated Box.Draw.
// This is the M2-box.ansi evidence frame for ticket 037.
// Only generates evidence when LOOM_EVIDENCE=1 environment variable is set.
func TestGenerateM2BoxEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	const cols, rows = 80, 20
	canvas := loom.NewCanvas(cols, rows)

	// Header
	canvas.Write(2, 0, "M2: Box.Draw Migrated to DrawBorder Primitive", loom.Style{Bold: true})
	canvas.Write(2, 1, "Boxes rendered through migrated Box.Draw using new primitives", loom.Reset)

	// Example 1: Simple box with title
	y := 3
	canvas.Write(2, y, "Box with title:", loom.Reset)

	box1 := &loom.Box{
		ID:    "box1",
		Title: "Box 1",
		Style: loom.BoxStyle{
			Background: loom.Reset,
			Border:     loom.Reset,
			Title:      loom.Reset,
			Footer:     loom.Reset,
		},
	}

	c := loom.NewCanvas(25, 6)
	box1.Draw(c, c.Bounds())

	// Copy rendered box to output canvas
	for fy := 0; fy < c.Rows(); fy++ {
		for fx := 0; fx < c.Cols(); fx++ {
			cell := c.Get(fx, fy)
			if cell.Text != "" {
				canvas.Set(2+fx, y+fy, cell)
			}
		}
	}

	y += 8

	// Example 2: Box with padding
	canvas.Write(2, y, "Box with padding and footer:", loom.Reset)

	box2 := &loom.Box{
		ID:      "box2",
		Title:   "Padded",
		Padding: 1,
		Footer:  "footer",
		Style: loom.BoxStyle{
			Background: loom.Reset,
			Border:     loom.Reset,
			Title:      loom.Reset,
			Footer:     loom.Reset,
		},
	}

	c2 := loom.NewCanvas(25, 6)
	box2.Draw(c2, c2.Bounds())

	// Copy to output canvas
	for fy := 0; fy < c2.Rows(); fy++ {
		for fx := 0; fx < c2.Cols(); fx++ {
			cell := c2.Get(fx, fy)
			if cell.Text != "" {
				canvas.Set(2+fx, y+fy, cell)
			}
		}
	}

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

	outPath := filepath.Join(outDir, "M2-box.ansi")
	err = os.WriteFile(outPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("WriteFile %s failed: %v", outPath, err)
	}

	t.Logf("Evidence saved to %s (%d bytes)", outPath, buf.Len())
}

// TestGenerateM2PopupEvidence renders a frame demonstrating the migrated Popup.Draw.
// This is the M2-popup.ansi evidence frame for ticket 037.
func TestGenerateM2PopupEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	const cols, rows = 80, 25
	canvas := loom.NewCanvas(cols, rows)

	// Header
	canvas.Write(2, 0, "M2: Popup.Draw Migrated to DrawBox Primitive", loom.Style{Bold: true})
	canvas.Write(2, 1, "Popups rendered through migrated Popup.Draw using new primitives", loom.Reset)

	// Background grid (simulate background content)
	for y := 3; y < 20; y++ {
		for x := 2; x < 78; x++ {
			canvas.Set(x, y, loom.Cell{Text: ".", Style: loom.Reset})
		}
	}

	// Popup 1: Simple popup with title
	popup1 := loom.NewPopup("Popup 1", &loom.Grid{})
	popup1.Width = 20
	popup1.Height = 6
	popup1.Style = loom.Reset

	popup1.Draw(canvas, loom.Rect{X: 0, Y: 0, W: cols, H: rows})

	// Popup 2: Wider popup at different location (simulated with DrawBox directly)
	canvas.Write(40, 3, "DrawBox (used by Popup):", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 40, Y: 5, W: 30, H: 8}, loom.BoxBorderStyleSharp, "Popup 2", loom.Reset)
	canvas.DrawBox(loom.Rect{X: 40, Y: 14, W: 30, H: 8}, loom.BoxBorderStyleRounded, "Long Title Popup 3", loom.Reset)

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

	outPath := filepath.Join(outDir, "M2-popup.ansi")
	err = os.WriteFile(outPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("WriteFile %s failed: %v", outPath, err)
	}

	t.Logf("Evidence saved to %s (%d bytes)", outPath, buf.Len())
}
