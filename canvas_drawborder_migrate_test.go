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
// Three non-overlapping popups: short title, long truncated title, CJK title.
func TestGenerateM2PopupEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	const cols, rows = 80, 24
	canvas := loom.NewCanvas(cols, rows)

	// Header and captions occupy their own rows above a fixed three-column grid.
	canvas.Write(2, 0, "M2: Popup.Draw Migrated to DrawBox (BoxBorderStyleSharp)", loom.Style{Bold: true})
	captions := []struct {
		x, y int
		text string
	}{
		{2, 1, "Popup 1 (short)"},
		{28, 1, "Popup 2 (long)"},
		{55, 1, "Popup 3 (CJK)"},
	}
	for _, caption := range captions {
		canvas.Write(caption.x, caption.y, caption.text, loom.Reset)
	}

	popupRects := []loom.Rect{
		{X: 1, Y: 4, W: 24, H: 5},
		{X: 28, Y: 4, W: 24, H: 5},
		{X: 55, Y: 4, W: 24, H: 5},
	}
	for i, rect := range popupRects {
		for j := 0; j < i; j++ {
			other := popupRects[j]
			if rect.X < other.X+other.W && other.X < rect.X+rect.W &&
				rect.Y < other.Y+other.H && other.Y < rect.Y+rect.H {
				t.Fatalf("popup rectangles %d and %d intersect", j+1, i+1)
			}
		}
		for _, caption := range captions {
			for x := caption.x; x < caption.x+len([]rune(caption.text)); x++ {
				if x >= rect.X && x < rect.X+rect.W && caption.y >= rect.Y && caption.y < rect.Y+rect.H {
					t.Fatalf("caption %q cell (%d,%d) is inside popup rectangle %v", caption.text, x, caption.y, rect)
				}
			}
		}
	}

	popups := []*loom.Popup{
		loom.NewPopup("OK", &loom.Grid{}),
		loom.NewPopup("This is a very long title for testing truncation", &loom.Grid{}),
		loom.NewPopup("你好世界", &loom.Grid{}),
	}
	for i, popup := range popups {
		popup.Width = popupRects[i].W
		popup.Height = popupRects[i].H
		popup.Style = loom.Reset
		popup.Draw(canvas, popupRects[i])
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

	outPath := filepath.Join(outDir, "M2-popup.ansi")
	err = os.WriteFile(outPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("WriteFile %s failed: %v", outPath, err)
	}

	t.Logf("Evidence saved to %s (%d bytes)", outPath, buf.Len())
}
