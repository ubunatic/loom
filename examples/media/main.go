// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command media displays a still image through Loom's cati-backed media widget.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/media"
)

type demo struct {
	image *media.Widget
	path  string
	mode  media.Mode
}

func (d *demo) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{BG: loom.ColorRGB(17, 24, 32)})
	c.Write(r.X+1, r.Y, "Media Demo  (q quits)", loom.Style{FG: loom.ColorRGB(240, 240, 240), Bold: true})
	if r.H < 2 {
		return
	}
	d.image.Draw(c, loom.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 2})
	c.Write(r.X, r.Y+r.H-1, fmt.Sprintf("Cols: %d  Rows: %d  Mode: %s  Media: %s", c.Cols(), c.Rows(), d.mode, d.path), loom.Style{FG: loom.ColorRGB(240, 240, 240), BG: loom.ColorRGB(38, 48, 60)})
}

func (*demo) HandleKey(loom.KeyEvent) bool     { return false }
func (*demo) HandleMouse(loom.MouseEvent) bool { return false }

func (d *demo) TickInterval() time.Duration { return d.image.TickInterval() }
func (d *demo) Tick(now time.Time)          { d.image.Tick(now) }

func run(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: loom-media <image-or-video> [halfblock|quadblock|sextant]")
	}
	mode := media.ModeHalfblock
	if len(args) == 2 {
		mode = media.Mode(args[1])
	}
	var widget *media.Widget
	var err error
	if isVideo(args[0]) {
		widget, err = media.NewVideo(args[0], mode, 24)
	} else {
		widget, err = media.LoadImage(args[0], mode)
	}
	if err != nil {
		return err
	}
	defer widget.Close()
	pane, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	pane.MaxCols = 0
	defer pane.Close()
	return pane.Run(&demo{image: widget, path: args[0], mode: mode})
}

func isVideo(path string) bool {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(path), ".")) {
	case "mp4", "m4v", "mov", "mkv", "webm", "avi", "mpeg", "mpg":
		return true
	default:
		return false
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
