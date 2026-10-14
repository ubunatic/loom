// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package media

import (
	"path/filepath"

	"ubunatic.com/cati/v1/core"
	"ubunatic.com/cati/v1/halfblock"
	"ubunatic.com/loom"
)

// MarkupOptions returns options with BaseDir and Image set for the .tui file
// at path, so that ParseMarkupWith renders [image] tags with cati.
func MarkupOptions(path string) loom.MarkupOptions {
	return loom.MarkupOptions{BaseDir: filepath.Dir(path), Image: MarkupImage}
}

// MarkupImage renders a .tui [image] tag with cati. A relative src is
// resolved against baseDir. Without width and height the image keeps its
// native size in cells, capped by loom.MarkupImageMaxSize; otherwise it fits
// inside the given width and height, keeping its aspect ratio. SVG files need
// rsvg-convert on PATH.
func MarkupImage(img loom.MarkupImage, baseDir string) ([]loom.RichLine, error) {
	path := img.Src
	if !filepath.IsAbs(path) && baseDir != "" {
		path = filepath.Join(baseDir, path)
	}
	mode := Mode(img.Mode)
	maxW, maxH := loom.MarkupImageMaxSize()
	pxW, pxH := pxPerCell(mode)
	src, err := halfblock.LoadImageWithTarget(path, maxW*pxW, maxH*pxH)
	if err != nil {
		return nil, err
	}
	cols, rows := img.Width, img.Height
	if cols == 0 && rows == 0 {
		nativeW, _ := nativeCellSize(mode, src)
		cols = min(nativeW, maxW)
	}
	if cols == 0 {
		cols = maxW
	}
	if rows == 0 {
		rows = maxH
	}
	grid, err := renderGrid(src, mode, cols, rows)
	if err != nil {
		return nil, err
	}
	return gridLines(grid), nil
}

// gridLines converts a cati grid into rich lines, one per row and one span
// per run of equal style. Every row is grid.Width columns wide.
func gridLines(grid *core.Grid) []loom.RichLine {
	lines := make([]loom.RichLine, 0, len(grid.Cells))
	for _, row := range grid.Cells {
		var spans []loom.RichSpan
		for x := range grid.Width {
			cell := loom.Cell{Text: " "}
			if x < len(row) {
				cell = canvasCell(row[x])
			}
			if n := len(spans); n > 0 && spans[n-1].Style == cell.Style {
				spans[n-1].Text += cell.Text
				continue
			}
			spans = append(spans, loom.RichSpan{Text: cell.Text, Style: cell.Style})
		}
		lines = append(lines, loom.RichLine{Spans: spans})
	}
	return lines
}
