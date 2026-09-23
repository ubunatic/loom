// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

type probeCase struct {
	name      string
	hoverX    int // 0-based terminal column
	hoverY    int // 0-based terminal row
	expectedX int // expected highlighted item start column (-1 if none)
	expectedY int // expected highlighted item row (-1 if none)
}

type probeResult struct {
	probe   probeCase
	actualX int // actual highlighted item start column (-1 if none)
	actualY int // actual highlighted item row (-1 if none)
	dx      int
	dy      int
	status  string
	note    string
}

// TestLoomoji107HoverProbe tests loomoji mouse hover reporting against rendered VT colour cells.
// It sends SGR 1006 hover reports (\x1b[<35;X;YM) and asserts that the hovered cell
// matches the highlighted cell in the terminal grid.
//
// This test exposes the offset bug in loomoji where mouse coordinates are
// miscalculated relative to the grid viewport and padding.
func TestLoomoji107HoverProbe(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loomoji")
	if out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/loomoji").CombinedOutput(); err != nil {
		t.Fatalf("build loomoji: %v\n%s", err, out)
	}

	s := ptytest.Start(t, 80, 24, bin)
	s.WaitFor("search…", 5*time.Second)

	// Accent color used by loomoji for item highlight (RGB 32, 151, 185)
	hoverBG := ptytest.ColorRGB(32, 151, 185)

	// Probes covering edge cases per issue 107 notes:
	// - Cells next to left/top padding
	// - Both halves of wide (emoji) cells
	// - First/last cell of an item and inter-item gap
	probes := []probeCase{
		// Left padding & leftmost item (row 0: termY=3)
		{name: "Left padding border", hoverX: 0, hoverY: 3, expectedX: -1, expectedY: -1},
		{name: "Item 0 left half (wide emoji)", hoverX: 1, hoverY: 3, expectedX: 1, expectedY: 3},
		{name: "Item 0 right half (wide emoji)", hoverX: 2, hoverY: 3, expectedX: 1, expectedY: 3},

		// Item 1 (first, last, gap)
		{name: "Item 1 left half (first cell)", hoverX: 4, hoverY: 3, expectedX: 4, expectedY: 3},
		{name: "Item 1 right half (last cell)", hoverX: 5, hoverY: 3, expectedX: 4, expectedY: 3},
		{name: "Item 1 spacing gap", hoverX: 6, hoverY: 3, expectedX: 4, expectedY: 3},

		// Item 2 (wide emoji halves)
		{name: "Item 2 left half", hoverX: 7, hoverY: 3, expectedX: 7, expectedY: 3},
		{name: "Item 2 right half", hoverX: 8, hoverY: 3, expectedX: 7, expectedY: 3},

		// Last item in row 0 (col 15 at X=46..47)
		{name: "Row 0 last item first half", hoverX: 46, hoverY: 3, expectedX: 46, expectedY: 3},
		{name: "Row 0 last item last half", hoverX: 47, hoverY: 3, expectedX: 46, expectedY: 3},

		// Top padding & search bar (termY=0..2)
		{name: "Top padding border", hoverX: 1, hoverY: 0, expectedX: -1, expectedY: -1},
		{name: "Search bar row", hoverX: 1, hoverY: 1, expectedX: -1, expectedY: -1},
		{name: "Search bar padding row", hoverX: 1, hoverY: 2, expectedX: -1, expectedY: -1},

		// Row 1 items (termY=4)
		{name: "Row 1 item 16 left half", hoverX: 1, hoverY: 4, expectedX: 1, expectedY: 4},
		{name: "Row 1 item 16 right half", hoverX: 2, hoverY: 4, expectedX: 1, expectedY: 4},
		{name: "Row 2 item 32 left half", hoverX: 1, hoverY: 5, expectedX: 1, expectedY: 5},
	}

	findHighlighted := func(grid [][]ptytest.Cell) (int, int, bool) {
		// Scan grid area (y: 3..21) for cells with effective BG == hoverBG
		for y := 3; y < len(grid) && y <= 21; y++ {
			for x := 0; x < len(grid[y]); x++ {
				_, bg := grid[y][x].Style.Effective()
				if bg == hoverBG {
					return x, y, true
				}
			}
		}
		return -1, -1, false
	}

	sendHoverAndWait := func(hx, hy int) [][]ptytest.Cell {
		prevCount := len(s.CellFrames())
		// SGR 1006 mouse reports are 1-based: \x1b[<35;X;YM
		s.SendRaw([]byte(fmt.Sprintf("\x1b[<35;%d;%dM", hx+1, hy+1)))
		deadline := time.Now().Add(250 * time.Millisecond)
		for time.Now().Before(deadline) {
			frames := s.CellFrames()
			if len(frames) > prevCount {
				return frames[len(frames)-1]
			}
			time.Sleep(5 * time.Millisecond)
		}
		return s.Cells()
	}

	var results []probeResult
	bugs := 0

	for _, p := range probes {
		grid := sendHoverAndWait(p.hoverX, p.hoverY)
		actX, actY, found := findHighlighted(grid)

		res := probeResult{
			probe:   p,
			actualX: actX,
			actualY: actY,
		}

		if !found {
			if p.expectedX == -1 && p.expectedY == -1 {
				res.status = "PASS"
				res.note = "no highlight (as expected)"
			} else {
				res.status = "NO_HIGHLIGHT"
				res.note = "expected highlight, none found"
				bugs++
			}
		} else {
			if p.expectedX == -1 && p.expectedY == -1 {
				res.status = "UNEXPECTED_HL"
				res.dx = actX - p.hoverX
				res.dy = actY - p.hoverY
				res.note = fmt.Sprintf("highlighted at (%d,%d) outside grid target", actX, actY)
				bugs++
			} else {
				res.dx = actX - p.expectedX
				res.dy = actY - p.expectedY
				if res.dx == 0 && res.dy == 0 {
					res.status = "PASS"
					res.note = "exact match"
				} else {
					res.status = "OFFSET_BUG"
					res.note = fmt.Sprintf("offset (dx=%+d, dy=%+d) from expected", res.dx, res.dy)
					bugs++
				}
			}
		}
		results = append(results, res)
	}

	// Format results table
	var b strings.Builder
	b.WriteString("\n=== LOOMOJI HOVER PROBE OFFSET TABLE ===\n")
	b.WriteString(fmt.Sprintf("%-30s | %-10s | %-12s | %-12s | %-12s | %-14s | %s\n",
		"Probe", "Hover(X,Y)", "Expected", "Actual", "Offset(dx,dy)", "Status", "Note"))
	b.WriteString(strings.Repeat("-", 120) + "\n")

	for _, r := range results {
		hoverStr := fmt.Sprintf("(%d,%d)", r.probe.hoverX, r.probe.hoverY)
		expStr := "none"
		if r.probe.expectedX != -1 {
			expStr = fmt.Sprintf("(%d,%d)", r.probe.expectedX, r.probe.expectedY)
		}
		actStr := "none"
		if r.actualX != -1 {
			actStr = fmt.Sprintf("(%d,%d)", r.actualX, r.actualY)
		}
		offsetStr := "N/A"
		if r.actualX != -1 && r.probe.expectedX != -1 {
			offsetStr = fmt.Sprintf("(%+d,%+d)", r.dx, r.dy)
		} else if r.actualX != -1 {
			offsetStr = fmt.Sprintf("(%+d,%+d)", r.dx, r.dy)
		}

		b.WriteString(fmt.Sprintf("%-30s | %-10s | %-12s | %-12s | %-12s | %-14s | %s\n",
			r.probe.name, hoverStr, expStr, actStr, offsetStr, r.status, r.note))
	}
	b.WriteString(strings.Repeat("-", 120) + "\n")

	tableStr := b.String()
	t.Log(tableStr)

	if bugs > 0 {
		t.Errorf("loomoji hover probe exposed %d offset/highlight bugs (expected under current bug):\n%s", bugs, tableStr)
	}
}
