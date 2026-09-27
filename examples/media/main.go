// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command media displays a still image through Loom's cati-backed media widget.
package main

import (
	"fmt"
	"os"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/media"
)

type demo struct {
	image *media.Widget
}

func (d *demo) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{BG: loom.ColorRGB(17, 24, 32)})
	c.Write(r.X+1, r.Y, "Media Demo  (q quits)", loom.Style{FG: loom.ColorRGB(240, 240, 240), Bold: true})
	d.image.Draw(c, loom.Rect{X: r.X + 6, Y: r.Y + 2, W: 8, H: 4})
}

func (*demo) HandleKey(loom.KeyEvent) bool     { return false }
func (*demo) HandleMouse(loom.MouseEvent) bool { return false }

func run(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: loom-media <image.png|image.jpg|image.svg> [halfblock|quadblock|sextant]")
	}
	mode := media.ModeHalfblock
	if len(args) == 2 {
		mode = media.Mode(args[1])
	}
	widget, err := media.LoadImage(args[0], mode)
	if err != nil {
		return err
	}
	pane, err := loom.New(8)
	if err != nil {
		return err
	}
	defer pane.Close()
	return pane.Run(&demo{image: widget})
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
