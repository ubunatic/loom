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
	"github.com/spf13/cobra"
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

func run(path string, mode media.Mode) error {
	var widget *media.Widget
	var err error
	if isVideo(path) {
		widget, err = media.NewVideo(path, mode, 24)
	} else {
		widget, err = media.LoadImage(path, mode)
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
	return pane.Run(&demo{image: widget, path: path, mode: mode})
}

func newCommand() *cobra.Command {
	return newCommandWithRun(run)
}

func newCommandWithRun(runMedia func(string, media.Mode) error) *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "loom-media <image-or-video> [mode]",
		Short: "Display an image or video in the terminal",
		Long:  "Display an image or video in the terminal using Loom's media widget.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 2 {
				return fmt.Errorf("requires an image or video path, optionally followed by a render mode")
			}
			if len(args) == 2 && mode != string(media.ModeHalfblock) {
				return fmt.Errorf("render mode supplied both positionally and with --mode")
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 2 {
				mode = args[1]
			}
			return runMedia(args[0], media.Mode(mode))
		},
	}
	cmd.Flags().StringVar(&mode, "mode", string(media.ModeHalfblock), "render mode (halfblock, quadblock, sextant)")
	return cmd
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
	cmd := newCommand()
	cmd.SetArgs(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
