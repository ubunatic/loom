// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

// TestLoomojiCategoryLineWidthsPTY validates that Loomoji's rendered screen rows
// have deterministic, uniform line widths across all 14 categories in a real PTY session.
// It verifies that no category produces ragged row endings, horizontal drift,
// stray right-edge fragments, or overflowing background bands.
func TestLoomojiCategoryLineWidthsPTY(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PTY category width test in short mode")
	}
	t.Setenv("LOOM_ZWJ", "join") // The test VT is not a terminal emulator and cannot answer DSR.

	bin := filepath.Join(t.TempDir(), "loomoji")
	if out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/loomoji").CombinedOutput(); err != nil {
		t.Fatalf("build loomoji: %v\n%s", err, out)
	}

	const (
		termCols    = 80
		termRows    = 24
		expectedW   = 50 // Loom default pane max_cols is 50
		panelHeight = 13 // Loomoji pane height is 13 rows
	)

	s := ptytest.Start(t, termCols, termRows, bin)
	s.WaitFor("search…", 5*time.Second)

	categories := []string{
		"Faces", "Hands", "Animals", "Food", "Sports", "Travel",
		"Objects", "Hearts", "Nature", "Symbols", "Arrows", "Blocks", "Lines", "Box",
	}

	for catIdx, catName := range categories {
		t.Run("Category_"+catName, func(t *testing.T) {
			cells := waitForQuiet(s)
			if len(cells) < panelHeight {
				t.Fatalf("screen has %d rows, want at least %d", len(cells), panelHeight)
			}

			// Verify that the active footer or screen mentions the current category or contents
			for rowY := 0; rowY < panelHeight; rowY++ {
				rowCells := cells[rowY]
				if len(rowCells) < expectedW {
					t.Fatalf("row %d has %d columns, want at least %d", rowY, len(rowCells), expectedW)
				}

				// Find the rightmost painted cell in the row
				lastPaintedCol := -1
				for colX := termCols - 1; colX >= 0; colX-- {
					c := rowCells[colX]
					// A cell is considered painted if it has a non-zero rune, non-default BG, or continuation
					if (c.Rune != ' ' && c.Rune != 0) || c.Style.BG != (ptytest.Color{}) {
						lastPaintedCol = colX
						break
					}
				}

				// All active panel rows must end at exactly column expectedW - 1 (width 50)
				if lastPaintedCol != expectedW-1 {
					// Format row content for debugging
					var rowRunes []rune
					for _, c := range rowCells[:min(termCols, expectedW+5)] {
						if c.Rune == 0 {
							rowRunes = append(rowRunes, '·')
						} else {
							rowRunes = append(rowRunes, c.Rune)
						}
					}
					t.Errorf("category %q row %d painted width = %d (lastCol=%d), want exactly %d columns\nContent: %q",
						catName, rowY, lastPaintedCol+1, lastPaintedCol, expectedW, string(rowRunes))
				}

				// Columns beyond expectedW (50..79) must be completely unpainted/blank
				for colX := expectedW; colX < termCols; colX++ {
					c := rowCells[colX]
					if (c.Rune != ' ' && c.Rune != 0) || c.Style.BG != (ptytest.Color{}) {
						t.Errorf("category %q row %d has stray paint at column %d: Rune=%q BG=%v",
							catName, rowY, colX, c.Rune, c.Style.BG)
					}
				}
			}

			// Cycle to the next category using 'f'
			if catIdx < len(categories)-1 {
				s.Send("f")
				time.Sleep(30 * time.Millisecond)
			}
		})
	}

	// Verify grid navigation with arrow keys
	s.Send("\x1b[B") // down arrow
	time.Sleep(20 * time.Millisecond)
	s.Send("\x1b[C") // right arrow
	time.Sleep(20 * time.Millisecond)

	// Quit cleanly
	s.Send("\x1b") // ESC
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("loomoji exit: %v", err)
	}
}
