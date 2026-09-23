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
	"codeberg.org/ubunatic/loom/measure"
)

type probeCase struct {
	name      string
	hoverX    int // 0-based terminal column
	hoverY    int // 0-based terminal row
	expectedX int // expected highlighted item start column (-1 if none)
	expectedY int // expected highlighted item row (-1 if none)
}

type probeResult struct {
	probe     probeCase
	actualX   int // actual highlighted item start column (-1 if none)
	actualY   int // actual highlighted item row (-1 if none)
	boxStr    string
	offsetStr string
	dx        int
	dy        int
	status    string
	note      string
}

type screenItem struct {
	r      rune
	startX int
	width  int
	y      int
}

// findGridItems scans the initial rendered screen grid for emoji grid rows and items,
// deriving item column positions and widths directly from rendered cell runes.
func findGridItems(grid [][]ptytest.Cell) [][]screenItem {
	var gridRows [][]screenItem
	for y := 0; y < len(grid); y++ {
		var rowItems []screenItem
		for x := 0; x < len(grid[y]); {
			c := grid[y][x]
			if c.Rune != ' ' && c.Rune != 0 {
				w := max(1, measure.RuneWidth(c.Rune))
				rowItems = append(rowItems, screenItem{
					r:      c.Rune,
					startX: x,
					width:  w,
					y:      y,
				})
				x += w
			} else {
				x++
			}
		}
		// Emoji grid rows have multiple wide emoji items (w == 2) separated by whitespace
		if len(rowItems) >= 4 && rowItems[0].width == 2 {
			gridRows = append(gridRows, rowItems)
		}
	}
	return gridRows
}

// sendHover sends an SGR 1006 mouse hover report (1-based coordinates).
func sendHover(s *ptytest.Session, hx, hy int) {
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<35;%d;%dM", hx+1, hy+1)))
}

// waitForSettle polls until terminal output stabilizes (no new frames during settleQuiet).
func waitForSettle(s *ptytest.Session) [][]ptytest.Cell {
	const (
		settleQuiet = 30 * time.Millisecond
		maxWait     = 300 * time.Millisecond
		pollStep    = 5 * time.Millisecond
	)
	start := time.Now()
	lastCount := len(s.CellFrames())
	lastChange := start

	for time.Since(start) < maxWait {
		time.Sleep(pollStep)
		count := len(s.CellFrames())
		now := time.Now()
		if count != lastCount {
			lastCount = count
			lastChange = now
		} else if time.Since(lastChange) >= settleQuiet {
			break
		}
	}
	return s.Cells()
}

// diffHighlight compares the current grid against a baseline grid without hard-coded colours,
// isolating the cells carrying the newly activated highlight via frequency analysis.
func diffHighlight(baseline, current [][]ptytest.Cell) (minX, minY, maxX, maxY int, found bool) {
	type point struct{ x, y int }
	var changed []point

	for y := 0; y < len(current) && y < len(baseline); y++ {
		for x := 0; x < len(current[y]) && x < len(baseline[y]); x++ {
			_, bgBase := baseline[y][x].Style.Effective()
			_, bgCur := current[y][x].Style.Effective()
			if bgBase != bgCur {
				changed = append(changed, point{x, y})
			}
		}
	}

	if len(changed) == 0 {
		return 0, 0, 0, 0, false
	}

	// Count BG frequencies across the current grid to isolate the minority (highlight) style from ambient BG
	bgFreq := make(map[ptytest.Color]int)
	for y := 0; y < len(current); y++ {
		for x := 0; x < len(current[y]); x++ {
			_, bg := current[y][x].Style.Effective()
			bgFreq[bg]++
		}
	}

	minFreq := int(^uint(0) >> 1)
	for _, p := range changed {
		_, bg := current[p.y][p.x].Style.Effective()
		if f := bgFreq[bg]; f < minFreq {
			minFreq = f
		}
	}

	var highlighted []point
	for _, p := range changed {
		_, bg := current[p.y][p.x].Style.Effective()
		if bgFreq[bg] == minFreq {
			highlighted = append(highlighted, p)
		}
	}

	if len(highlighted) == 0 {
		highlighted = changed
	}

	minX, maxX = highlighted[0].x, highlighted[0].x
	minY, maxY = highlighted[0].y, highlighted[0].y
	for _, p := range highlighted[1:] {
		if p.x < minX {
			minX = p.x
		}
		if p.x > maxX {
			maxX = p.x
		}
		if p.y < minY {
			minY = p.y
		}
		if p.y > maxY {
			maxY = p.y
		}
	}
	return minX, minY, maxX, maxY, true
}

// TestLoomoji107HoverProbe tests loomoji mouse hover reporting against rendered VT colour cells.
// It sends SGR 1006 hover reports (\x1b[<35;X;YM) and detects highlight changes differentially
// against a baseline grid without hard-coded colours.
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
	initialGrid := waitForSettle(s)

	gridRows := findGridItems(initialGrid)
	if len(gridRows) < 3 || len(gridRows[0]) < 3 {
		t.Fatalf("could not derive grid items from rendered screen (found %d grid rows)", len(gridRows))
	}

	item0 := gridRows[0][0]
	item1 := gridRows[0][1]
	item2 := gridRows[0][2]
	lastItem0 := gridRows[0][len(gridRows[0])-1]
	row1Item0 := gridRows[1][0]
	row2Item0 := gridRows[2][0]

	// Spacing gap between item 1 and item 2
	gapX := item1.startX + item1.width

	// Probes covering edge cases per issue 107 notes:
	// - Cells next to left/top padding
	// - Both halves of wide (emoji) cells
	// - First/last cell of an item and inter-item gap
	probes := []probeCase{
		// Left padding & leftmost item (row 0)
		{name: "Left padding border", hoverX: item0.startX - 1, hoverY: item0.y, expectedX: -1, expectedY: -1},
		{name: "Item 0 left half (wide emoji)", hoverX: item0.startX, hoverY: item0.y, expectedX: item0.startX, expectedY: item0.y},
		{name: "Item 0 right half (wide emoji)", hoverX: item0.startX + 1, hoverY: item0.y, expectedX: item0.startX, expectedY: item0.y},

		// Item 1 (first, last, gap)
		{name: "Item 1 left half (first cell)", hoverX: item1.startX, hoverY: item1.y, expectedX: item1.startX, expectedY: item1.y},
		{name: "Item 1 right half (last cell)", hoverX: item1.startX + 1, hoverY: item1.y, expectedX: item1.startX, expectedY: item1.y},
		{name: "Item 1 spacing gap", hoverX: gapX, hoverY: item1.y, expectedX: item1.startX, expectedY: item1.y},

		// Item 2 (wide emoji halves)
		{name: "Item 2 left half", hoverX: item2.startX, hoverY: item2.y, expectedX: item2.startX, expectedY: item2.y},
		{name: "Item 2 right half", hoverX: item2.startX + 1, hoverY: item2.y, expectedX: item2.startX, expectedY: item2.y},

		// Last item in row 0
		{name: "Row 0 last item first half", hoverX: lastItem0.startX, hoverY: lastItem0.y, expectedX: lastItem0.startX, expectedY: lastItem0.y},
		{name: "Row 0 last item last half", hoverX: lastItem0.startX + 1, hoverY: lastItem0.y, expectedX: lastItem0.startX, expectedY: lastItem0.y},

		// Top padding & search bar (rows above first grid row)
		{name: "Top padding border", hoverX: item0.startX, hoverY: 0, expectedX: -1, expectedY: -1},
		{name: "Search bar row", hoverX: item0.startX, hoverY: 1, expectedX: -1, expectedY: -1},
		{name: "Search bar padding row", hoverX: item0.startX, hoverY: 2, expectedX: -1, expectedY: -1},

		// Row 1 items
		{name: "Row 1 item 16 left half", hoverX: row1Item0.startX, hoverY: row1Item0.y, expectedX: row1Item0.startX, expectedY: row1Item0.y},
		{name: "Row 1 item 16 right half", hoverX: row1Item0.startX + 1, hoverY: row1Item0.y, expectedX: row1Item0.startX, expectedY: row1Item0.y},

		// Row 2 items
		{name: "Row 2 item 32 left half", hoverX: row2Item0.startX, hoverY: row2Item0.y, expectedX: row2Item0.startX, expectedY: row2Item0.y},
	}

	var results []probeResult
	bugs := 0

	for _, p := range probes {
		// Move pointer off-grid between probes to ensure a clean transition
		sendHover(s, 0, 0)
		baseline := waitForSettle(s)

		sendHover(s, p.hoverX, p.hoverY)
		hoverGrid := waitForSettle(s)

		minX, minY, maxX, maxY, found := diffHighlight(baseline, hoverGrid)

		res := probeResult{
			probe:   p,
			actualX: -1,
			actualY: -1,
		}

		if !found {
			res.boxStr = "none"
			res.offsetStr = "N/A"
			if p.expectedX == -1 && p.expectedY == -1 {
				res.status = "PASS"
				res.note = "no change (as expected)"
			} else {
				res.status = "NO_CHANGE"
				res.note = "hover changed nothing"
				bugs++
			}
		} else {
			res.actualX = minX
			res.actualY = minY
			if minX == maxX && minY == maxY {
				res.boxStr = fmt.Sprintf("(%d,%d)", minX, minY)
			} else {
				res.boxStr = fmt.Sprintf("(%d,%d)..(%d,%d)", minX, minY, maxX, maxY)
			}

			if p.expectedX == -1 && p.expectedY == -1 {
				res.status = "UNEXPECTED_HL"
				res.dx = minX - p.hoverX
				res.dy = minY - p.hoverY
				res.offsetStr = fmt.Sprintf("(%+d,%+d)", res.dx, res.dy)
				res.note = fmt.Sprintf("highlighted at (%d,%d) outside grid target", minX, minY)
				bugs++
			} else {
				res.dx = minX - p.expectedX
				res.dy = minY - p.expectedY
				res.offsetStr = fmt.Sprintf("(%+d,%+d)", res.dx, res.dy)
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
	b.WriteString(fmt.Sprintf("%-30s | %-10s | %-10s | %-16s | %-12s | %-14s | %s\n",
		"Probe", "Hover(X,Y)", "Expected", "Changed Box", "Offset(dx,dy)", "Status", "Note"))
	b.WriteString(strings.Repeat("-", 130) + "\n")

	for _, r := range results {
		hoverStr := fmt.Sprintf("(%d,%d)", r.probe.hoverX, r.probe.hoverY)
		expStr := "none"
		if r.probe.expectedX != -1 {
			expStr = fmt.Sprintf("(%d,%d)", r.probe.expectedX, r.probe.expectedY)
		}

		b.WriteString(fmt.Sprintf("%-30s | %-10s | %-10s | %-16s | %-12s | %-14s | %s\n",
			r.probe.name, hoverStr, expStr, r.boxStr, r.offsetStr, r.status, r.note))
	}
	b.WriteString(strings.Repeat("-", 130) + "\n")

	tableStr := b.String()
	t.Log(tableStr)

	if bugs > 0 {
		t.Errorf("loomoji hover probe exposed %d offset/highlight bugs (expected under current bug):\n%s", bugs, tableStr)
	}
}
