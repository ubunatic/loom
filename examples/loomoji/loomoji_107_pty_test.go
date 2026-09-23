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

// sendAndSettle sends raw terminal input, waits for at least one new frame (within frameTimeout),
// and then waits for output to stabilize (no new frames during settleQuiet).
// Returns the settled grid and true, or current grid and false if timed out without a new frame.
func sendAndSettle(s *ptytest.Session, raw []byte) ([][]ptytest.Cell, bool) {
	const (
		frameTimeout = 2 * time.Second
		settleQuiet  = 30 * time.Millisecond
		pollStep     = 5 * time.Millisecond
	)
	startCount := len(s.CellFrames())
	s.SendRaw(raw)

	// Wait for at least one new frame
	start := time.Now()
	for len(s.CellFrames()) == startCount {
		if time.Since(start) >= frameTimeout {
			return s.Cells(), false
		}
		time.Sleep(pollStep)
	}

	// Settle quiet window
	lastCount := len(s.CellFrames())
	lastChange := time.Now()
	for {
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
	return s.Cells(), true
}

// sendHover sends an SGR 1006 mouse hover report (1-based coordinates) and waits for settlement.
func sendHover(s *ptytest.Session, hx, hy int) ([][]ptytest.Cell, bool) {
	return sendAndSettle(s, []byte(fmt.Sprintf("\x1b[<35;%d;%dM", hx+1, hy+1)))
}

// waitForQuiet polls until terminal output stabilizes (no new frames during settleQuiet).
func waitForQuiet(s *ptytest.Session) [][]ptytest.Cell {
	const (
		settleQuiet = 30 * time.Millisecond
		pollStep    = 5 * time.Millisecond
	)
	lastCount := len(s.CellFrames())
	lastChange := time.Now()
	for {
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
	initialGrid := waitForQuiet(s)

	gridRows := findGridItems(initialGrid)
	if len(gridRows) < 5 || len(gridRows[0]) < 4 {
		t.Fatalf("could not derive grid items from rendered screen (found %d grid rows)", len(gridRows))
	}

	row0Item0 := gridRows[0][0]
	row0Item1 := gridRows[0][1]
	row1Item0 := gridRows[1][0]
	row1Item1 := gridRows[1][1]

	row2Item0 := gridRows[2][0]
	row2Item1 := gridRows[2][1]
	row2Item2 := gridRows[2][2]
	row2Item3 := gridRows[2][3]
	row2Last := gridRows[2][len(gridRows[2])-1]
	row2Gap := row2Item1.startX + row2Item1.width

	row3Item0 := gridRows[3][0]
	row3Item2 := gridRows[3][2]
	row3Last := gridRows[3][len(gridRows[3])-1]

	row4Item1 := gridRows[4][1]
	row4Item4 := gridRows[4][4]

	// Probes covering edge cases per issue 107 notes:
	// - Primary edge probes moved to grid rows ≥ 5 across multiple columns
	// - Row-3/4 probes kept as edge cases
	// - Padding & border probes kept as edge cases
	probes := []probeCase{
		// ── Primary edge probes on grid rows ≥ 5 ─────────────────────────────
		// Row 2 (y=5) item 0: both emoji halves
		{name: "Row 2 item 0 left half", hoverX: row2Item0.startX, hoverY: row2Item0.y, expectedX: row2Item0.startX, expectedY: row2Item0.y},
		{name: "Row 2 item 0 right half", hoverX: row2Item0.startX + 1, hoverY: row2Item0.y, expectedX: row2Item0.startX, expectedY: row2Item0.y},

		// Row 2 (y=5) item 1: first/last cell and inter-item gap
		{name: "Row 2 item 1 first cell", hoverX: row2Item1.startX, hoverY: row2Item1.y, expectedX: row2Item1.startX, expectedY: row2Item1.y},
		{name: "Row 2 item 1 last cell", hoverX: row2Item1.startX + 1, hoverY: row2Item1.y, expectedX: row2Item1.startX, expectedY: row2Item1.y},
		{name: "Row 2 item 1 spacing gap", hoverX: row2Gap, hoverY: row2Item1.y, expectedX: row2Item1.startX, expectedY: row2Item1.y},

		// Row 2 (y=5) item 2 & 3: multiple columns across row
		{name: "Row 2 item 2 left half", hoverX: row2Item2.startX, hoverY: row2Item2.y, expectedX: row2Item2.startX, expectedY: row2Item2.y},
		{name: "Row 2 item 2 right half", hoverX: row2Item2.startX + 1, hoverY: row2Item2.y, expectedX: row2Item2.startX, expectedY: row2Item2.y},
		{name: "Row 2 item 3 left half", hoverX: row2Item3.startX, hoverY: row2Item3.y, expectedX: row2Item3.startX, expectedY: row2Item3.y},

		// Row 2 (y=5) last item of row
		{name: "Row 2 last item first half", hoverX: row2Last.startX, hoverY: row2Last.y, expectedX: row2Last.startX, expectedY: row2Last.y},
		{name: "Row 2 last item last half", hoverX: row2Last.startX + 1, hoverY: row2Last.y, expectedX: row2Last.startX, expectedY: row2Last.y},

		// Row 3 (y=6) items across columns
		{name: "Row 3 item 0 left half", hoverX: row3Item0.startX, hoverY: row3Item0.y, expectedX: row3Item0.startX, expectedY: row3Item0.y},
		{name: "Row 3 item 2 left half", hoverX: row3Item2.startX, hoverY: row3Item2.y, expectedX: row3Item2.startX, expectedY: row3Item2.y},
		{name: "Row 3 last item last half", hoverX: row3Last.startX + 1, hoverY: row3Last.y, expectedX: row3Last.startX, expectedY: row3Last.y},

		// Row 4 (y=7) items across columns
		{name: "Row 4 item 1 left half", hoverX: row4Item1.startX, hoverY: row4Item1.y, expectedX: row4Item1.startX, expectedY: row4Item1.y},
		{name: "Row 4 item 4 left half", hoverX: row4Item4.startX, hoverY: row4Item4.y, expectedX: row4Item4.startX, expectedY: row4Item4.y},

		// ── Row-3/4 edge cases ───────────────────────────────────────────────
		{name: "Row 0 (y=3) item 0 left half", hoverX: row0Item0.startX, hoverY: row0Item0.y, expectedX: row0Item0.startX, expectedY: row0Item0.y},
		{name: "Row 0 (y=3) item 1 left half", hoverX: row0Item1.startX, hoverY: row0Item1.y, expectedX: row0Item1.startX, expectedY: row0Item1.y},
		{name: "Row 1 (y=4) item 0 left half", hoverX: row1Item0.startX, hoverY: row1Item0.y, expectedX: row1Item0.startX, expectedY: row1Item0.y},
		{name: "Row 1 (y=4) item 1 left half", hoverX: row1Item1.startX, hoverY: row1Item1.y, expectedX: row1Item1.startX, expectedY: row1Item1.y},

		// ── Padding & border edge cases ──────────────────────────────────────
		{name: "Left padding border (row 0)", hoverX: row0Item0.startX - 1, hoverY: row0Item0.y, expectedX: -1, expectedY: -1},
		{name: "Left padding border (row 2)", hoverX: row2Item0.startX - 1, hoverY: row2Item0.y, expectedX: -1, expectedY: -1},
		{name: "Top padding border", hoverX: row0Item0.startX, hoverY: 0, expectedX: -1, expectedY: -1},
		{name: "Search bar row", hoverX: row0Item0.startX, hoverY: 1, expectedX: -1, expectedY: -1},
		{name: "Search bar padding row", hoverX: row0Item0.startX, hoverY: 2, expectedX: -1, expectedY: -1},
	}

	var results []probeResult
	bugs := 0
	observedChanges := 0

	for _, p := range probes {
		// Move pointer off-grid between probes to ensure a clean transition
		baseline, okReset := sendHover(s, 0, 0)
		hoverGrid, okHover := sendHover(s, p.hoverX, p.hoverY)

		res := probeResult{
			probe:   p,
			actualX: -1,
			actualY: -1,
		}

		if !okReset || !okHover {
			res.boxStr = "none"
			res.offsetStr = "N/A"
			if p.expectedX == -1 && p.expectedY == -1 {
				res.status = "PASS"
				res.note = "no change (as expected)"
			} else {
				res.status = "NO_CHANGE"
				res.note = "no visible target"
			}
			results = append(results, res)
			continue
		}

		minX, minY, maxX, maxY, found := diffHighlight(baseline, hoverGrid)

		if !found {
			res.boxStr = "none"
			res.offsetStr = "N/A"
			if p.expectedX == -1 && p.expectedY == -1 {
				res.status = "PASS"
				res.note = "no change (as expected)"
			} else {
				res.status = "NO_CHANGE"
				res.note = "no visible target"
			}
		} else {
			if p.expectedX != -1 || p.expectedY != -1 {
				observedChanges++
			}
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

	if observedChanges == 0 {
		t.Fatal("hover not observable: no on-grid hover produced a screen change")
	}

	// Calculate dominant (dx, dy) across measured probes
	type offsetKey struct{ dx, dy int }
	offsetCounts := make(map[offsetKey]int)
	var dominant offsetKey
	maxCount := 0
	measuredCount := 0

	for _, r := range results {
		if r.probe.expectedX != -1 && r.actualX != -1 {
			k := offsetKey{r.dx, r.dy}
			offsetCounts[k]++
			if offsetCounts[k] > maxCount {
				maxCount = offsetCounts[k]
				dominant = k
			}
			measuredCount++
		}
	}

	var summaryLine string
	if measuredCount > 0 {
		summaryLine = fmt.Sprintf("Dominant offset: (dx=%+d, dy=%+d) across %d/%d measured probes", dominant.dx, dominant.dy, maxCount, measuredCount)
	} else {
		summaryLine = "No probes measured a highlight change"
	}

	// Format results table
	var b strings.Builder
	b.WriteString("\n=== LOOMOJI HOVER PROBE OFFSET TABLE ===\n")
	b.WriteString(fmt.Sprintf("%-32s | %-11s | %-11s | %-16s | %-13s | %-14s | %s\n",
		"Probe", "Hover(X,Y)", "Expected", "Changed Box", "Offset(dx,dy)", "Status", "Note"))
	b.WriteString(strings.Repeat("-", 135) + "\n")

	for _, r := range results {
		hoverStr := fmt.Sprintf("(%d,%d)", r.probe.hoverX, r.probe.hoverY)
		expStr := "none"
		if r.probe.expectedX != -1 {
			expStr = fmt.Sprintf("(%d,%d)", r.probe.expectedX, r.probe.expectedY)
		}

		b.WriteString(fmt.Sprintf("%-32s | %-11s | %-11s | %-16s | %-13s | %-14s | %s\n",
			r.probe.name, hoverStr, expStr, r.boxStr, r.offsetStr, r.status, r.note))
	}
	b.WriteString(strings.Repeat("-", 135) + "\n")
	b.WriteString(summaryLine + "\n")

	tableStr := b.String()
	t.Log(tableStr)

	if bugs > 0 {
		t.Errorf("loomoji hover probe exposed %d offset bugs (expected under current bug):\n%s", bugs, tableStr)
	}
}
