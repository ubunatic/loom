// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package media provides a loom widget backed by cati's terminal-cell renderers.
package media

import (
	"fmt"
	"image"

	"codeberg.org/ubunatic/loom"
	"ubunatic.com/cati/v1/core"
	"ubunatic.com/cati/v1/halfblock"
	"ubunatic.com/cati/v1/quadblock"
	"ubunatic.com/cati/v1/sextant"
)

// Mode selects cati's cell geometry for media rendering.
type Mode string

const (
	ModeHalfblock Mode = "halfblock"
	ModeQuadblock Mode = "quadblock"
	ModeSextant   Mode = "sextant"
)

// Widget renders a still image inside the rectangle supplied by loom.
type Widget struct {
	image image.Image
	mode  Mode
}

// NewImage creates a still-image widget using mode. A zero mode selects
// half-block rendering.
func NewImage(src image.Image, mode Mode) (*Widget, error) {
	if src == nil {
		return nil, fmt.Errorf("media: image is nil")
	}
	if mode == "" {
		mode = ModeHalfblock
	}
	switch mode {
	case ModeHalfblock, ModeQuadblock, ModeSextant:
	default:
		return nil, fmt.Errorf("media: unsupported render mode %q", mode)
	}
	return &Widget{image: src, mode: mode}, nil
}

// LoadImage decodes a PNG, JPEG, or SVG image using cati's image loader.
func LoadImage(path string, mode Mode) (*Widget, error) {
	src, err := halfblock.LoadImage(path)
	if err != nil {
		return nil, err
	}
	return NewImage(src, mode)
}

// Mode returns the widget's selected renderer mode.
func (w *Widget) Mode() Mode {
	if w == nil {
		return ""
	}
	return w.mode
}

// Draw paints a cati-rendered image, clipped to r.
func (w *Widget) Draw(c *loom.Canvas, r loom.Rect) {
	if w == nil || w.image == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	grid, err := w.render(r.W, r.H)
	if err != nil || grid == nil {
		return
	}
	for y, row := range grid.Cells {
		if y >= r.H {
			break
		}
		for x, cell := range row {
			if x >= r.W {
				break
			}
			c.Set(r.X+x, r.Y+y, canvasCell(cell))
		}
	}
}

func (w *Widget) render(cols, rows int) (*core.Grid, error) {
	switch w.mode {
	case ModeQuadblock:
		return quadblock.RenderToGrid(w.image, cols, quadblock.Options{Rows: rows})
	case ModeSextant:
		return sextant.RenderToGrid(w.image, cols, sextant.Options{Rows: rows})
	default:
		return halfblock.RenderToGrid(w.image, cols, halfblock.Options{Rows: rows})
	}
}

func canvasCell(cell core.Cell) loom.Cell {
	text := " "
	if !cell.Transparent && cell.Ch != 0 {
		text = string(cell.Ch)
	}
	style := loom.Style{}
	if cell.HasFg {
		style.FG = loom.ColorRGB(cell.Fg.R, cell.Fg.G, cell.Fg.B)
	}
	if cell.HasBg {
		style.BG = loom.ColorRGB(cell.Bg.R, cell.Bg.G, cell.Bg.B)
	}
	return loom.Cell{Text: text, Style: style}
}

// HandleKey reports that this widget does not consume keys.
func (*Widget) HandleKey(loom.KeyEvent) bool { return false }

// HandleMouse reports that this widget does not consume mouse events.
func (*Widget) HandleMouse(loom.MouseEvent) bool { return false }
